package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	health_pb "google.golang.org/grpc/health/grpc_health_v1"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/stretchr/testify/assert"
)

func TestNewVpnServer(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		opts  []Option
		check *config
		err   error
	}{
		"default": {
			check: &config{
				logPath:   "",
				vpnSubNet: "10.10.10.10/24",
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
				WithLogPath("../tmp.txt"),
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
				logPath:   "../tmp.txt",
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
		assert.Equal(t.check.logPath, vpn.cfg.logPath)
		assert.Equal(t.check.vpnSubNet, vpn.cfg.vpnSubNet)
		assert.Equal(t.check.grpcPort, vpn.cfg.grpcPort)
		assert.Len(vpn.cfg.grpcUnaryInterceptors, len(t.check.grpcUnaryInterceptors)+1)
		assert.Len(vpn.cfg.grpcStreamInterceptors, len(t.check.grpcStreamInterceptors)+1)
		assert.Len(vpn.cfg.grpcOptions, len(t.check.grpcOptions))
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
		go vpn.Run()
		time.Sleep(3 * time.Second)

		// call health check
		conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", vpn.(*vpnServer).cfg.grpcPort), grpc.WithInsecure())
		assert.NoError(err)
		client := health_pb.NewHealthClient(conn)
		result, err := client.Check(context.Background(), &health_pb.HealthCheckRequest{Service: ""})
		assert.NoError(err)
		assert.Equal(result.Status.String(), "SERVING")
	}

}
