package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	"github.com/gjbae1212/grpc-vpn/auth"
	"github.com/songgao/water/waterutil"

	"github.com/gjbae1212/grpc-vpn/internal"
	"github.com/pkg/errors"

	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/songgao/water"
)

const (
	queueSizeForClientToServer = 500
	queueSizeForServerToClient = 500
)

type VPN interface {
	// Start VPN
	Run() error

	GetJwtSalt() string

	// GRPC METHODS
	Exchange(stream protocol.VPN_ExchangeServer) error
	Auth(ctx context.Context, req *protocol.AuthRequest) (*protocol.AuthResponse, error)
}

type vpn struct {
	tun *water.Interface // tun device

	localIP      net.IP     // vpn server ip
	localNetmask *net.IPNet // vpn server netmask

	clients     map[string]*client // clients(map[vpn-ip]*client)
	clientsLock sync.RWMutex       // clients lock

	clientToServer chan *protocol.IPPacket // packets which flow from client to server.
	serverToClient chan *protocol.IPPacket // packets which flow from server to client.

	jwtSalt string // JWT Salt

	exit     chan bool // exit channel
	stopping bool
}

// Auth is to authorize user, and it's GRPC METHOD.
func (v *vpn) Auth(ctx context.Context, req *protocol.AuthRequest) (*protocol.AuthResponse, error) {
	if v.stopping {
		return nil, errors.Wrapf(internal.ErrorStoppingServer, "Method: Auth")
	}

	var user string
	if ctx.Value(auth.UserCtxName) == nil {
		user = "unknown"
	} else {
		user = ctx.Value(auth.UserCtxName).(string)
	}

	// make jwt token
	claims := &jwt.StandardClaims{
		Audience:  user,
		Subject:   "grpc-vpn-auth",
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "grpc-vpn",
	}
	encode, err := internal.EncodeJWT(claims, []byte(v.jwtSalt))
	if err != nil {
		return nil, errors.Wrapf(err, "Method: Auth")
	}

	return &protocol.AuthResponse{
		ErrorCode: protocol.ErrorCode_EC_SUCCESS,
		Jwt:       encode,
	}, nil
}

// Exchange is to exchange packets, and it's GRPC METHOD.
func (v *vpn) Exchange(stream protocol.VPN_ExchangeServer) error {
	if v.stopping {
		return errors.Wrapf(internal.ErrorStoppingServer, "Method: Exchange")
	}

	cli, err := newClient(stream, v.clientToServer)
	if err != nil {
		return errors.Wrapf(err, "Method: Exchange")
	}

	// add client
	if err := v.addClient(cli); err != nil {
		return errors.Wrapf(err, "Method: Exchange")
	}

	// assign vpn ip to client.
	packet := &protocol.IPPacket{
		ErrorCode:  protocol.ErrorCode_EC_SUCCESS,
		PacketType: protocol.IPPacketType_IPPT_VPN_ASSIGN,
		Packet2: &protocol.IPPacket_Vpn{
			VpnAssignedIp: cli.vpnIP,
			VpnGateway:    v.localIP,
			VpnSubnetIp:   v.localNetmask.IP,
			VpnSubnetMask: v.localNetmask.Mask,
		},
	}
	if err := stream.Send(packet); err != nil {
		return errors.Wrapf(err, "Method: Exchange")
	}

	defaultLogger.Info(color.GreenString("[login] %s (%s, %s)",
		cli.user, cli.originIP.String(), cli.vpnIP.String()))

	// receive packets
	go cli.processReading()

	// write packets and block
	cli.processWriting()

	// delete client
	_ = v.deleteClient(cli)

	defaultLogger.Info(color.GreenString("[logout] %s (%s, %s)",
		cli.user, cli.originIP.String(), cli.vpnIP.String()))

	return nil
}

// Run is to run Tun device for VPN.
func (v *vpn) Run() error {
	// make tun device.
	tun, err := water.New(water.Config{DeviceType: water.TUN})
	if err != nil {
		return errors.Wrapf(err, "Method: %s", "newVPN")
	}
	v.tun = tun

	// set ip to tun device
	if err := internal.SetTunIP(v.tun.Name(), v.localIP, v.localNetmask); err != nil {
		return errors.Wrapf(err, "Method: %s", "newVPN")
	}

	// read packets from TUN
	go v.loopReadFromTun()

	// process packets
	go v.loopClientToServer()
	go v.loopServerToClient()

	// trap signal(block)
	v.trapSignal()
	return nil
}

// GetJwtSalt returns JWT Salt.
func (v *vpn) GetJwtSalt() string {
	return v.jwtSalt
}

// addClient is to add client to map.
func (v *vpn) addClient(c *client) error {
	if c == nil {
		return errors.Wrapf(internal.ErrorInvalidParams, "Method: addClient")
	}
	v.clientsLock.Lock()
	defer v.clientsLock.Unlock()

	// issue vpn ip.
	for ip := v.localIP.Mask(v.localNetmask.Mask); v.localNetmask.Contains(ip); internal.IncreaseIP(ip) {
		// continue when such as below conditions.
		if ip.String() == v.localIP.String() || strings.HasSuffix(ip.String(), ".0") ||
			strings.HasSuffix(ip.String(), ".255") || strings.HasSuffix(ip.String(), ":") {
			continue
		}

		// continue if other user is used.
		if _, ok := v.clients[ip.String()]; ok {
			continue
		}

		// assign vpn ip
		c.vpnIP = ip

		// register
		v.clients[ip.String()] = c

		return nil
	}

	return errors.Wrapf(internal.ErrorExceedClientPool, "Method: addClient")
}

// deleteClient is to delete client to map.
func (v *vpn) deleteClient(c *client) error {
	if c == nil {
		return errors.Wrapf(internal.ErrorInvalidParams, "Method: deleteClient")
	}
	v.clientsLock.Lock()
	defer v.clientsLock.Unlock()

	// delete
	if _, ok := v.clients[c.vpnIP.String()]; ok {
		delete(v.clients, c.vpnIP.String())
	}

	return nil
}

// getClient is to give client object.
func (v *vpn) getClient(key net.IP) *client {
	v.clientsLock.RLock()
	defer v.clientsLock.RUnlock()
	return v.clients[key.String()]
}

// loopReadFromTun reads packet from tun device.
func (v *vpn) loopReadFromTun() {
	for {
		raw := make([]byte, internal.TunPacketBufferSize)
		n, err := v.tun.Read(raw)
		if err != nil {
			defaultLogger.Error(color.RedString("[ERR] Read Tun Device %s", err.Error()))
			break
		}

		// send packet to queue.
		v.serverToClient <- &protocol.IPPacket{
			ErrorCode:  protocol.ErrorCode_EC_SUCCESS,
			PacketType: protocol.IPPacketType_IPPT_RAW,
			Packet1: &protocol.IPPacket_Raw{
				Raw: raw[:n],
			},
		}
	}
	v.exit <- true
}

// loopClientToServer processes packets which should flow out of server.
func (v *vpn) loopClientToServer() {
	// support only packet1
	for {
		select {
		case packet := <-v.clientToServer:
			if packet.Packet1 == nil {
				continue
			}

			// extract destination
			dest := waterutil.IPv4Destination(packet.Packet1.Raw)

			// ignore multicast
			if dest.IsMulticast() {
				continue
			}

			// if destination is a vpn client.
			innerVpnClient := v.getClient(dest)
			if innerVpnClient != nil {
				innerVpnClient.in <- packet
				continue
			}

			// send packets to tun device.
			size, err := v.tun.Write(packet.Packet1.Raw)
			if err != nil {
				defaultLogger.Error(color.RedString("[ERR] Client To Server %s", err.Error()))
				break
			}

			if size == len(packet.Packet1.Raw) {
				defaultLogger.Error(color.RedString("[ERR] Mismatched Sending Packet %s != %s",
					size, len(packet.Packet1.Raw)))
			}
		}
	}
	v.exit <- true
}

// loopServerToClient processes packets which should flow into clients.
func (v *vpn) loopServerToClient() {
	// support only packet1
	for {
		select {
		case packet := <-v.serverToClient:
			if packet.Packet1 == nil {
				continue
			}

			// extract destination
			dest := waterutil.IPv4Destination(packet.Packet1.Raw)

			// ignore multicast
			if dest.IsMulticast() {
				continue
			}

			// if destination is a vpn client.
			innerVpnClient := v.getClient(dest)
			if innerVpnClient != nil {
				innerVpnClient.in <- packet
			} else {
				defaultLogger.Error(color.RedString("[ERR] Server To Client Unknown Destination %s", dest.String()))
			}
		}
	}
	v.exit <- true
}

// newVPN return new vpn object.
func newVPN(subnet, jwtSalt string) (VPN, error) {
	if subnet == "" || jwtSalt == "" {
		return nil, errors.Wrapf(internal.ErrorInvalidParams, "Method: %s", "newVPN")
	}

	v := &vpn{
		clients:        map[string]*client{},
		clientToServer: make(chan *protocol.IPPacket, queueSizeForClientToServer),
		serverToClient: make(chan *protocol.IPPacket, queueSizeForServerToClient),
		jwtSalt:        jwtSalt,
		exit:           make(chan bool, 1),
	}

	// parse ip and netmask from subnet.
	ip, netmask, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, errors.Wrapf(err, "Method: %s", "newVPN")
	}

	// if suffix of ip is .0, ip increase +1.
	if strings.HasSuffix(ip.String(), ".0") {
		internal.IncreaseIP(ip)
	}

	v.localIP = ip
	v.localNetmask = netmask

	return v, nil
}

// trap signal
func (v *vpn) trapSignal() {
	sig := make(chan os.Signal, 2)
	done := make(chan bool, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		done <- true
	}()

	select {
	case <-done:
		v.stopping = true
		defaultLogger.Info(color.YellowString("[trap-signal] signal stopping..."))
	case <-v.exit:
		defaultLogger.Info(color.YellowString("[trap-signal] event stopping..."))
		v.stopping = true
	}

	// close clients.
	v.clientsLock.Lock()
	defer v.clientsLock.Unlock()
	for _, c := range v.clients {
		go func(c *client) {
			c.exit <- true
		}(c)
	}
	time.Sleep(5 * time.Second)
}
