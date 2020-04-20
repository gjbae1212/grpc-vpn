package server

// reference
// https://grpc.io/docs/quickstart/go/
// https://grpc.io/docs/tutorials/basic/go/
// https://github.com/grpc/grpc-go
// https://github.com/grpc-ecosystem/go-grpc-middleware

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/gjbae1212/grpc-vpn/internal"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpchealth "google.golang.org/grpc/health"
	health_pb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

const (
	maxGRPCMsgSize = 1 << 30 // MAX 1GB
)

var (
	defaultOptions = []Option{
		WithLogPath(""),
		WithVpnSubNet("10.10.10.10/24"),
		WithVpnJwtSalt(internal.GenerateRandomString(16)),
		WithGrpcPort("8080"),
	}
)

type VpnServer interface {
	Run() error
}

type vpnServer struct {
	cfg    *config
	gs     *grpc.Server
	logger *internal.Logger
	// TODO: VPN 설정
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

	logger, err := internal.NewLogger(cfg.logPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Method: NewVpnServer")
	}

	// create vpn server
	server := &vpnServer{
		cfg:    cfg,
		gs:     grpc.NewServer(allOpts...),
		logger: logger,
	}

	// make handler
	h, err := NewHandler(server)
	if err != nil {
		return nil, errors.Wrapf(err, "Method: NewVpnServer")
	}

	// register api
	protocol.RegisterVPNServer(server.gs, h)

	// register health check handler
	health_pb.RegisterHealthServer(server.gs, grpchealth.NewServer())

	return server, nil
}

// Run executes VPN Server.
func (s *vpnServer) Run() error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", s.cfg.grpcPort))
	if err != nil {
		return errors.Wrapf(err, "Method: Run")
	}
	defer listen.Close()
	return s.gs.Serve(listen)
}

// TODO: JWT VALID CHECK
// TODO: add context
func defaultStreamServerInterceptors() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, stream)
	}
}

// TODO: add context
func defaultUnaryServerInterceptors() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}
