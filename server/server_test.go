package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/gjbae1212/grpc-vpn/internal"

	health_pb "google.golang.org/grpc/health/grpc_health_v1"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/stretchr/testify/assert"
)

// mock vpn
type mockVPN struct{}

func (m *mockVPN) Run() error {
	time.Sleep(10 * time.Second)
	return nil
}
func (m *mockVPN) Close() error {
	return nil
}
func (m *mockVPN) Exchange(stream protocol.VPN_ExchangeServer) error { return nil }
func (m *mockVPN) GetJwtSalt() string                                { return "mock" }
func (m *mockVPN) Auth(ctx context.Context, req *protocol.AuthRequest) (*protocol.AuthResponse, error) {
	return nil, nil
}

func TestNewVpnServer(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		opts  []Option
		check *config
		err   error
	}{
		"default": {
			check: &config{
				vpnSubNet: "10.10.10.1/24",
				grpcPort:  "8080",
				grpcOptions: []grpc.ServerOption{
					grpc.MaxRecvMsgSize(maxGRPCMsgSize),
					grpc.MaxSendMsgSize(maxGRPCMsgSize),
					grpc.KeepaliveParams(keepalive.ServerParameters{
						Time:    5 * time.Minute,  // keepalive 5 min
						Timeout: 20 * time.Second, // keepalive timeout
					})},
			},
		},
		"additional": {
			opts: []Option{
				WithVpnSubNet("10.0.0.1/24"),
				WithGrpcPort("2020"),
				WithGrpcStreamInterceptors(
					[]grpc.StreamServerInterceptor{
						grpc_ctxtags.StreamServerInterceptor(),
						grpc_recovery.StreamServerInterceptor()},
				),
				WithGrpcUnaryInterceptors(
					[]grpc.UnaryServerInterceptor{
						grpc_ctxtags.UnaryServerInterceptor(),
						grpc_recovery.UnaryServerInterceptor()},
				),
			},
			check: &config{
				vpnSubNet: "10.0.0.1/24",
				grpcPort:  "2020",
				grpcStreamInterceptors: []grpc.StreamServerInterceptor{
					grpc_ctxtags.StreamServerInterceptor(),
					grpc_recovery.StreamServerInterceptor(),
				},
				grpcUnaryInterceptors: []grpc.UnaryServerInterceptor{
					grpc_ctxtags.UnaryServerInterceptor(),
					grpc_recovery.UnaryServerInterceptor(),
				},
				grpcOptions: []grpc.ServerOption{
					grpc.MaxRecvMsgSize(maxGRPCMsgSize),
					grpc.MaxSendMsgSize(maxGRPCMsgSize),
					grpc.KeepaliveParams(keepalive.ServerParameters{
						Time:    5 * time.Minute,  // keepalive 5 min
						Timeout: 20 * time.Second, // keepalive timeout
					})},
			},
		},
	}

	for _, t := range tests {
		s, err := NewVpnServer(t.opts...)
		assert.Equal(t.err, err)
		vpn := s.(*vpnServer)
		assert.Equal(t.check.vpnSubNet, vpn.config.vpnSubNet)
		assert.Equal(t.check.grpcPort, vpn.config.grpcPort)
		assert.Len(vpn.config.grpcUnaryInterceptors, len(t.check.grpcUnaryInterceptors)+1)
		assert.Len(vpn.config.grpcStreamInterceptors, len(t.check.grpcStreamInterceptors)+1)
		assert.Len(vpn.config.grpcOptions, len(t.check.grpcOptions))
	}
}

func TestVpnServer_Run(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input []Option
		isErr bool
	}{
		"success": {},
	}

	for _, t := range tests {
		vpn, err := NewVpnServer(t.input...)
		assert.Equal(t.isErr, err != nil)
		vpn.(*vpnServer).vpn = &mockVPN{}
		go vpn.Run()
		time.Sleep(2 * time.Second)

		// call health check
		conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", vpn.(*vpnServer).config.grpcPort), grpc.WithInsecure())
		assert.NoError(err)
		client := health_pb.NewHealthClient(conn)
		result, err := client.Check(context.Background(), &health_pb.HealthCheckRequest{Service: ""})
		assert.NoError(err)
		assert.Equal(result.Status.String(), "SERVING")
	}

}

func TestSetDefaultLogger(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]struct {
		input *internal.Logger
	}{
		"success": {input: nil},
	}

	for _, t := range tests {
		SetDefaultLogger(t.input)
	}
	_ = assert
}
