package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"

	"google.golang.org/grpc/keepalive"

	"github.com/fatih/color"
	"github.com/gjbae1212/grpc-vpn/auth"
	"github.com/songgao/water/waterutil"
	"google.golang.org/grpc/metadata"

	"github.com/briandowns/spinner"
	"github.com/gjbae1212/grpc-vpn/internal"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	health_pb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/cenkalti/backoff"
	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/songgao/water"
)

const (
	queueSize = 1000
)

var (
	defaultOptions = []Option{
		WithTlsInsecure(true),
	}
)

var (
	defaultLogger *logrus.Logger
)

// VpnClient is an interface for connecting to VPN server.
type VpnClient interface {
	Run() error
	Close() error
	JWT() string
	MyVpnIp() string
}

type vpnClient struct {
	cfg      *config
	dialOpts []grpc.DialOption
	auth     auth.ClientAuthMethod
	jwt      string

	tun         *water.Interface
	tunName     string
	vpnMyIP     net.IP // vpn my ip
	vpnSubnet   *net.IPNet
	vpnGateway  net.IP       // vpn gateway
	networkLock sync.RWMutex // network lock

	originServerIP   net.IP // origin vpn server ip
	originGateway    net.IP // origin gateway
	originDeviceName string // origin device name

	conn     protocol.VPNClient          // vpn connection
	connPipe protocol.VPN_ExchangeClient // vpn connection read, write pipe
	connLock sync.RWMutex                // conn lock

	in  chan *protocol.IPPacket // in queue
	out chan *protocol.IPPacket // out queue

	retryLock         sync.RWMutex // retry lock
	lastConnectedTime time.Time    // last connected time

	networkRollback *Rollback                   // network rollback
	backoff         *backoff.ExponentialBackOff // backoff
	exit            chan bool                   // exit channel
}

func (vc *vpnClient) Run() error {
	defer vc.Close()

	s := spinner.New(spinner.CharSets[7], 100*time.Millisecond) // Build our new spinner
	s.Start()

	// extract current gateway
	gw, gwDevice, err := internal.GetNetGateway()
	if err != nil {
		return errors.Wrapf(err, "Method: Run")
	}

	gateway := net.ParseIP(gw)
	defaultLogger.Info(color.GreenString("Default gateway is %s on %s", gateway, gwDevice))

	vc.originGateway = gateway
	vc.originDeviceName = gwDevice
	vc.networkRollback.AddRoute(vc.originServerIP, vc.originGateway, vc.originDeviceName)

	// connect GRPC
	if err := vc.setGRPCConnection(); err != nil {
		return errors.Wrapf(err, "Method: Run")
	}

	// authorization
	jwt, err := vc.auth(vc.conn)
	if err != nil {
		return errors.Wrapf(err, "Method: Run")
	}
	vc.jwt = jwt

	// connect VPN
	vc.vpnConnect(jwt)

	// Read TUN
	go vc.readTun()
	// Write TUN
	go vc.writeTun()
	// read packet from grpc connection
	go vc.readToGRPC()
	// write packet to grpc connection
	go vc.writeToGRPC()

	s.Stop()

	// block
	vc.trapSignal()

	// close
	vc.Close()
	return nil
}

func (vc *vpnClient) Close() error {
	vc.networkRollback.Close()
	if runtime.GOOS == "darwin" {
		internal.SetDeleteDNS()
	}
	defaultLogger.Error(color.RedString("[EXIT] BYE"))
	return nil
}

// JWT returns jwt string
func (vc *vpnClient) JWT() string {
	return vc.jwt
}

// MyVpnIP returns my vpn ip.
func (vc *vpnClient) MyVpnIp() string {
	vc.networkLock.RLock()
	defer vc.networkLock.RUnlock()
	if vc.vpnMyIP != nil {
		return vc.vpnMyIP.String()
	}
	return ""
}

func (vc *vpnClient) vpnConnect(jwt string) error {
	ctx := metadata.NewOutgoingContext(context.Background(), auth.JWTAuthHeaderForGRPC(jwt))
	sock, err := vc.conn.Exchange(ctx)
	if err != nil {
		return errors.Wrapf(err, "Method: connect")
	}
	vc.connPipe = sock

	// must be VPN assigned packet.
	packet, err := vc.connPipe.Recv()
	if err != nil {
		return errors.Wrapf(err, "Method: connect")
	}

	if packet.ErrorCode != protocol.ErrorCode_EC_SUCCESS {
		return errors.Wrapf(internal.ErrorReceiveUnknownPacket, "Method: connect")
	}
	if packet.PacketType != protocol.IPPacketType_IPPT_VPN_ASSIGN {
		return errors.Wrapf(internal.ErrorReceiveUnknownPacket, "Method: connect")
	}
	if packet.Packet2 == nil {
		return errors.Wrapf(internal.ErrorReceiveUnknownPacket, "Method: connect")
	}

	// assign VPN IP
	vpnIP := net.IP(packet.Packet2.VpnAssignedIp)
	vpnGateway := net.IP(packet.Packet2.VpnGateway)
	vpnSubnet := &net.IPNet{
		IP:   net.IP(packet.Packet2.VpnSubnetIp),
		Mask: net.IPMask(packet.Packet2.VpnSubnetMask),
	}
	if err := vc.setVPN(vpnIP, vpnGateway, vpnSubnet); err != nil {
		return errors.Wrapf(internal.ErrorReceiveUnknownPacket, "Method: connect")
	}
	vc.lastConnectedTime = time.Now()
	return nil
}

func (vc *vpnClient) retryVpnConnect() error {
	vc.retryLock.Lock()
	defer vc.retryLock.Unlock()
	// pass retry
	if vc.lastConnectedTime.Add(5*time.Second).Unix() > time.Now().Unix() {
		return nil
	}

	vc.networkRollback.Close()
	for i := 0; i < 10; i++ {
		defaultLogger.Warn(color.YellowString("[RETRY] vpn connect %d", i+1))
		time.Sleep(vc.backoff.NextBackOff())
		// connect GRPC
		if err := vc.setGRPCConnection(); err != nil {
			defaultLogger.Error(color.RedString(err.Error()))
			continue
		}

		if err := vc.vpnConnect(vc.jwt); err != nil {
			defaultLogger.Error(color.RedString(err.Error()))
			continue
		}

		vc.backoff.Reset()
		defaultLogger.Info(color.GreenString("[SUCCESS] vpn reconnect"))
		return nil
	}

	return fmt.Errorf("[FAIL] FAIL RETRY")
}

func (vc *vpnClient) setVPN(vpnIP, vpnGateway net.IP, vpnSubnet *net.IPNet) error {
	vc.networkLock.Lock()
	defer vc.networkLock.Unlock()

	vc.vpnMyIP = vpnIP
	vc.vpnGateway = vpnGateway
	vc.vpnSubnet = vpnSubnet

	// make tun device
	tun, err := water.New(water.Config{DeviceType: water.TUN})
	if err != nil {
		return errors.Wrapf(err, "Method: setVPN")
	}
	vc.tun = tun
	vc.tunName = tun.Name()
	defaultLogger.Info(color.GreenString("[create] tun device %s", tun.Name()))

	if runtime.GOOS == "darwin" {
		// write reset gateway, if vpn client is closed.
		vc.networkRollback.ResetGatewayOSX(vc.tun, vc.originGateway.String())
		// write dns 8.8.8.8 Wi-Fi device
		internal.SetGoogleDNS()
	}

	// set ip to tun.
	if err := internal.SetTunIP(vc.tunName, vc.vpnMyIP, vc.vpnSubnet); err != nil {
		return errors.Wrapf(err, "Method: setVPN")
	}

	// route all traffic to the VPN server(real ip) through the current gateway device
	if err := internal.AddRoute(vc.originServerIP, vc.originGateway, vc.originDeviceName); err != nil {
		return errors.Wrapf(err, "Method: setVPN")
	}

	// redirect default traffic via our VPN
	if err := internal.SetDefaultGateway(vc.vpnGateway.String(), vc.tun.Name()); err != nil {
		return errors.Wrapf(err, "Method: setVPN")
	}

	// tun up
	if err := internal.SetTunStatus(vc.tun.Name(), true); err != nil {
		return fmt.Errorf("[err] Run %w", err)
	}

	return nil

}

func (vc *vpnClient) setGRPCConnection() error {
	vc.connLock.Lock()
	defer vc.connLock.Unlock()

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", vc.originServerIP.String(), vc.cfg.serverPort), vc.dialOpts...)
	if err != nil {
		return errors.Wrapf(err, "Method: Run")
	}

	// health check
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	healthClient := health_pb.NewHealthClient(conn)
	result, err := healthClient.Check(ctx, &health_pb.HealthCheckRequest{Service: ""})
	if err != nil {
		return errors.Wrapf(err, "Method: Run")
	}

	if result.Status != health_pb.HealthCheckResponse_SERVING {
		return errors.Wrapf(internal.ErrorReceiveUnknownPacket, "Method: Run")
	}

	vc.conn = protocol.NewVPNClient(conn)
	return nil
}

func (vc *vpnClient) getGRPCConnection() protocol.VPNClient {
	vc.connLock.RLock()
	defer vc.connLock.RUnlock()
	return vc.conn
}

func (vc *vpnClient) getGRPCConnectionPipe() protocol.VPN_ExchangeClient {
	vc.connLock.RLock()
	defer vc.connLock.RUnlock()
	return vc.connPipe
}

func (vc *vpnClient) getTun() *water.Interface {
	vc.networkLock.RLock()
	defer vc.networkLock.RUnlock()
	return vc.tun
}

func (vc *vpnClient) readTun() {
	for {
		tun := vc.getTun()

		packet := make([]byte, internal.TunPacketBufferSize)
		n, err := tun.Read(packet)
		if err != nil {
			defaultLogger.Error(color.RedString("[ERR] READ TUN DEVICE %d %s %s",
				unsafe.Pointer(tun), tun.Name(), err.Error()))
			//break ReadTun
			time.Sleep(1 * time.Second)
			continue
		}

		raw := packet[:n]
		dest := waterutil.IPv4Destination(raw)
		// bypass multicast
		if dest.IsMulticast() {
			continue
		}

		// out queue
		vc.out <- &protocol.IPPacket{
			ErrorCode:  protocol.ErrorCode_EC_SUCCESS,
			PacketType: protocol.IPPacketType_IPPT_RAW,
			Packet1:    &protocol.IPPacket_Raw{Raw: raw},
		}
	}

	defaultLogger.Errorf(color.RedString("[STOP] readTun"))
}

func (vc *vpnClient) writeTun() {
	for {
		select {
		case packet := <-vc.in:
			tun := vc.getTun()

			size, err := tun.Write(packet.Packet1.Raw)
			if err != nil {
				defaultLogger.Error(color.RedString("[ERR] WRITE TUN DEVICE %d %s %s",
					unsafe.Pointer(tun), tun.Name(), err.Error()))
				time.Sleep(1 * time.Second)
				continue
			}

			if size != len(packet.Packet1.Raw) {
				defaultLogger.Warn(color.YellowString("[WARNING] TunWriteRoutine %s mismatched %d != %d",
					tun.Name(), size, len(packet.Packet1.Raw)))
			}
		}
	}

	defaultLogger.Errorf(color.RedString("[STOP] writeTun"))
}

func (vc *vpnClient) writeToGRPC() {
WriteGRPC:
	for {
		select {
		case packet := <-vc.out:
			pipe := vc.getGRPCConnectionPipe()
			if err := pipe.Send(packet); err != nil {
				defaultLogger.Error(color.RedString("[ERR] writeToGRPC %s", err.Error()))
				time.Sleep(1 * time.Second)
				// retry connection
				if suberr := vc.retryVpnConnect(); suberr != nil {
					defaultLogger.Error(color.RedString("[ERR] writeToGRPC %s", suberr.Error()))
					// Good Bye
					vc.exit <- true
					break WriteGRPC
				}
			}
		}
	}
}

func (vc *vpnClient) readToGRPC() {
ReadGRPC:
	for {
		pipe := vc.getGRPCConnectionPipe()
		packet, err := pipe.Recv()
		if err != nil {
			defaultLogger.Error(color.RedString("[ERR] readToGRPC %s", err.Error()))
			time.Sleep(1 * time.Second)
			// retry connection
			if suberr := vc.retryVpnConnect(); suberr != nil {
				defaultLogger.Error(color.RedString("[ERR] readToGRPC %s", suberr.Error()))
				// Good Bye
				vc.exit <- true
				break ReadGRPC
			} else {
				continue
			}
		}

		// exit when jwt is expired.
		if packet.ErrorCode == protocol.ErrorCode_EC_EXPIRED_JWT {
			defaultLogger.Error(color.RedString("[ERR] readToGRPC JWT Expired"))
			vc.exit <- true
			break ReadGRPC
		}

		if packet.ErrorCode != protocol.ErrorCode_EC_SUCCESS {
			defaultLogger.Error(color.RedString("[ERR] readToGRPC %s", internal.ErrorReceiveUnknownPacket.Error()))
			continue
		}

		if packet.PacketType != protocol.IPPacketType_IPPT_RAW {
			defaultLogger.Error(color.RedString("[ERR] readToGRPC %s", internal.ErrorReceiveUnknownPacket.Error()))
			continue
		}

		// mismatched VPN IP.
		dest := waterutil.IPv4Destination(packet.Packet1.Raw)
		if !vc.vpnMyIP.Equal(dest) {
			defaultLogger.Error(color.RedString("[ERR] readToGRPC %s", internal.ErrorMismatchVpnIP.Error()))
			continue
		}

		vc.in <- packet
	}
}

// trap signal
func (vc *vpnClient) trapSignal() {
	sig := make(chan os.Signal, 2)
	done := make(chan bool, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		done <- true
	}()

	select {
	case <-done:
		defaultLogger.Info(color.YellowString("[trap-signal] signal stopping..."))
	case <-vc.exit:
		defaultLogger.Info(color.YellowString("[trap-signal] event stopping..."))
	}
}

// NewVpnClient returns new vpn client.
func NewVpnClient(opts ...Option) (VpnClient, error) {
	cfg := &config{}

	tmpOpts := make([]Option, len(defaultOptions))
	copy(tmpOpts, defaultOptions)

	// merge custom options
	tmpOpts = append(tmpOpts, opts...)

	for _, opt := range tmpOpts {
		opt.apply(cfg)
	}

	originServerIP, err := internal.GetIPByAddr(cfg.serverAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "Method: NewVpnClient")
	}

	// make dial options
	dialOpts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    1 * time.Minute,
			Timeout: 10 * time.Second,
		}),
	}

	if cfg.tlsCertification != "" {
		roots := x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM([]byte(cfg.tlsCertification)); !ok {
			return nil, errors.Wrapf(internal.ErrorInvalidParams, "TLS Certification Invalid Method: NewVpnClient")
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			RootCAs: roots, InsecureSkipVerify: cfg.tlsInsecure})))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	return &vpnClient{
		cfg:             cfg,
		dialOpts:        dialOpts,
		auth:            cfg.authMethod,
		originServerIP:  originServerIP,
		networkRollback: &Rollback{},
		in:              make(chan *protocol.IPPacket, queueSize),
		out:             make(chan *protocol.IPPacket, queueSize),
		backoff:         backoff.NewExponentialBackOff(),
		exit:            make(chan bool, 1),
	}, nil
}

// SetDefaultLogger is to set logger for vpn client.
func SetDefaultLogger(logger *logrus.Logger) {
	defaultLogger = logger
}

func init() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		DisableColors: false,
		FullTimestamp: true,
	})
	logger.SetOutput(os.Stdout)
	defaultLogger = logger
}
