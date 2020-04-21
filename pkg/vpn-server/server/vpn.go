package server

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"net"
	"strings"
	"sync"

	"github.com/gjbae1212/grpc-vpn/internal"
	"github.com/pkg/errors"

	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/songgao/water"
	"go.uber.org/atomic"
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
}

// Auth is to authorize user, and it's GRPC METHOD.
func (v *vpn) Auth(ctx context.Context, req *protocol.AuthRequest) (*protocol.AuthResponse, error) {
	return nil, nil
}

// Exchange is to exchange packets, and it's GRPC METHOD.
func (v *vpn) Exchange(stream protocol.VPN_ExchangeServer) error {
	// TODO: extract context Data(있는지 확인)
	return nil
	// TODO block := chan bool
	//stream.Recv()
	//stream.Send()

	// TODO: recv gorutine (client exit monitoring)
	// TODO: exit block and deregister api call

	// <- block
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

	return nil
}

// GetJwtSalt returns JWT Salt.
func (v *vpn) GetJwtSalt() string {
	return v.jwtSalt
}

// addClient is to add client to map.
func (v *vpn) addClient(c *client) {
	if c == nil {
		return
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
		break
	}
}

// deleteClient is to delete client to map.
func (v *vpn) deleteClient(c *client) {
	if c == nil {
		return
	}
	v.clientsLock.Lock()
	defer v.clientsLock.Unlock()

	// delete
	if _, ok := v.clients[c.vpnIP.String()]; ok {
		delete(v.clients, c.vpnIP.String())
	}
}

// getClient is to give client object.
func (v *vpn) getClient(key net.IP) *client {
	v.clientsLock.RLock()
	defer v.clientsLock.RUnlock()
	return v.clients[key.String()]
}

// TODO: method client read
// TODO: method client write

// TODO:
func (v *vpn) loopReadFromTun()    {}
func (v *vpn) loopClientToServer() {}
func (v *vpn) loopServerToClient() {}

//

type client struct {
	user     string                      // user
	originIP net.IP                      // user origin ip
	vpnIP    net.IP                      // user vpn ip
	jwt      *jwt.Token                  // user jwt token
	stream   protocol.VPN_ExchangeServer // stream
	loop     *atomic.Bool                // whether break loop or not
	exit     chan bool                   // exit
}

func (c *client) process() {
	for c.loop.Load() {
		// TODO: recv
	}
}

func (c *client) read() (*protocol.IPPacket, error) {
	return nil, nil
}

func (c *client) write(packet *protocol.IPPacket) error {
	return nil
}

func (c *client) hasVpnIP() bool {
	return c.vpnIP != nil
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
