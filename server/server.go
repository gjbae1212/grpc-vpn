package server

// reference
// https://grpc.io/docs/quickstart/go/
// https://grpc.io/docs/tutorials/basic/go/
// https://github.com/grpc/grpc-go
// https://github.com/grpc-ecosystem/go-grpc-middleware
// https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gjbae1212/grpc-vpn/auth"

	"google.golang.org/grpc/peer"

	"github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/gjbae1212/grpc-vpn/internal"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpchealth "google.golang.org/grpc/health"
	health_pb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

const (
	maxGRPCMsgSize = 1 << 30 // MAX 1GB
)

const (
	ipCtxName  = "ip"
	jwtCtxName = "jwt"
)

var (
	defaultOptions = []Option{
		WithVpnSubNet("10.10.10.1/24"),
		WithVpnJwtSalt(internal.GenerateRandomString(16)),
		WithGrpcPort("8080"),
	}

	defaultLogger *logrus.Logger
)

// VpnServer is an interface for utilizing vpn operations.
type VpnServer interface {
	Run() error
}

type vpnServer struct {
	config *config      // config
	grpc   *grpc.Server // grpc server
	vpn    VPN          // vpn
}

// SetDefaultLogger is to set logger for vpn server.
func SetDefaultLogger(logger *logrus.Logger) {
	defaultLogger = logger
}

// NewVpnServer returns vpn server.
func NewVpnServer(opts ...Option) (VpnServer, error) {
	cfg := &config{}
	// make default options
	tmpOpts := make([]Option, len(defaultOptions))
	copy(tmpOpts, defaultOptions)

	// merge custom options
	tmpOpts = append(tmpOpts, opts...)

	for _, opt := range tmpOpts {
		opt.apply(cfg)
	}

	// apply default grpc interceptors
	cfg.grpcUnaryInterceptors = append([]grpc.UnaryServerInterceptor{defaultUnaryServerInterceptors()},
		cfg.grpcUnaryInterceptors...)
	cfg.grpcStreamInterceptors = append([]grpc.StreamServerInterceptor{defaultStreamServerInterceptors()},
		cfg.grpcStreamInterceptors...)

	// apply default grpc options
	cfg.grpcOptions = append([]grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxGRPCMsgSize),
		grpc.MaxSendMsgSize(maxGRPCMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    5 * time.Minute,  // keepalive 5 min
			Timeout: 20 * time.Second, // keepalive timeout
		})},
		cfg.grpcOptions...)

	// apply tls
	if cfg.grpcTlsPem != "" && cfg.grpcTlsCertification != "" {
		cert, err := tls.X509KeyPair([]byte(cfg.grpcTlsCertification), []byte(cfg.grpcTlsPem))
		if err != nil {
			return nil, errors.Wrapf(err, "Method: NewVpnServer")
		}
		cred := credentials.NewServerTLSFromCert(&cert)
		cfg.grpcOptions = append([]grpc.ServerOption{grpc.Creds(cred)}, cfg.grpcOptions...)
	}

	// merge all of grpc option
	allOpts := []grpc.ServerOption{
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(cfg.grpcStreamInterceptors...)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(cfg.grpcUnaryInterceptors...)),
	}
	allOpts = append(allOpts, cfg.grpcOptions...)

	// create vpn server
	server := &vpnServer{
		config: cfg,
		grpc:   grpc.NewServer(allOpts...),
	}

	// make vpn
	vpn, err := newVPN(cfg.vpnSubNet, cfg.vpnJwtSalt)
	if err != nil {
		return nil, errors.Wrapf(err, "Method: NewVpnServer")
	}
	server.vpn = vpn

	// register api
	protocol.RegisterVPNServer(server.grpc, vpn)

	// register health check handler
	health_pb.RegisterHealthServer(server.grpc, grpchealth.NewServer())

	return server, nil
}

// Run executes VPN Server.
func (s *vpnServer) Run() error {
	if s.config.grpcPort == "80" {
		return fmt.Errorf("VPN-SERVER dosen't use 80 Port. Retry Other Port(Ex 443, 8080 ...)")
	}

	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", s.config.grpcPort))
	if err != nil {
		return errors.Wrapf(err, "Method: Run")
	}
	defer listen.Close()

	// run GRPC Server
	go s.grpc.Serve(listen)

	// run VPN Server and block
	if err := s.vpn.Run(); err != nil {
		return errors.Wrapf(err, "Method: Run")
	}

	return nil
}

func defaultStreamServerInterceptors() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		defer func() {
			if err := recover(); err != nil {
				defaultLogger.Error(color.RedString("[err][stream][recover] %s", err.(error).Error()))
			}
		}()

		// parse ip
		var ip net.IP
		peer, ok := peer.FromContext(stream.Context())
		if ok {
			ip = net.ParseIP(strings.Split(peer.Addr.String(), ":")[0])
			if ip == nil {
				ip = net.ParseIP("127.0.0.1")
			}
		}

		// check jwt
		jwt, err := checkJwt(srv, stream)
		if err != nil {
			defaultLogger.Error(color.RedString("[err] %s %s", err.Error(), ip.String()))
			return err
		}

		// insert ip and jwt.
		ctx := stream.Context()
		ctx = context.WithValue(ctx, ipCtxName, ip)
		ctx = context.WithValue(ctx, jwtCtxName, jwt)
		wrap := &AuthorizedContext{ServerStream: stream, Ctx: ctx}

		if err := handler(srv, wrap); err != nil {
			defaultLogger.Error(color.RedString("[err] %s", err.Error()))
			return err
		}
		return nil
	}
}

func defaultUnaryServerInterceptors() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		defer func() {
			if err := recover(); err != nil {
				defaultLogger.Error(color.RedString("[err][unary][recover] %s", err.(error).Error()))
			}
		}()

		// if it is an api for health-check.
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		// parse ip
		var ip net.IP
		peer, ok := peer.FromContext(ctx)
		if ok {
			ip = net.ParseIP(strings.Split(peer.Addr.String(), ":")[0])
			if ip == nil {
				ip = net.ParseIP("127.0.0.1")
			}
		}

		newCtx := context.WithValue(ctx, ipCtxName, ip)
		result, err := handler(newCtx, req)
		if err != nil {
			defaultLogger.Error(color.RedString("[err] %s %s", err.Error(), ip.String()))
			return nil, err
		}
		return result, nil
	}
}

// checkJwt is to check jwt.
// rfc2617 (e.g. Authorization: basic token, Authorization: bearer token)
func checkJwt(srv interface{}, stream grpc.ServerStream) (*jwt.Token, error) {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return nil, errors.Wrapf(internal.ErrorUnauthorized, "Method: auth")
	}

	if len(md[auth.AuthorizationHeader]) == 0 {
		return nil, errors.Wrapf(internal.ErrorUnauthorized, "Method: auth")
	}

	seps := strings.SplitN(md[auth.AuthorizationHeader][0], " ", 2)
	if len(seps) != 2 {
		return nil, errors.Wrapf(internal.ErrorUnauthorized, "Method: auth")
	}

	if seps[0] != auth.Basic && seps[0] != auth.Bearer {
		return nil, errors.Wrapf(internal.ErrorUnauthorized, "Method: auth")
	}

	jwt, err := internal.DecodeJWT(seps[1], []byte(srv.(VPN).GetJwtSalt()))
	if err != nil {
		return nil, errors.Wrapf(internal.ErrorInvalidJWT, "Method: auth")
	}

	return jwt, nil
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
