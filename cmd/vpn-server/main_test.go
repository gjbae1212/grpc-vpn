// +build integration

package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gjbae1212/grpc-vpn/auth"
	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	health_pb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

func TestIntegration(t *testing.T) {
	assert := assert.New(t)

	// Test Health Check
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", "8080"), grpc.WithInsecure())
	assert.NoError(err)
	healthClient := health_pb.NewHealthClient(conn)
	result1, err := healthClient.Check(context.Background(), &health_pb.HealthCheckRequest{Service: ""})
	assert.Equal(result1.Status.String(), "SERVING")

	// Test Auth
	vpnClient := protocol.NewVPNClient(conn)
	result2, err := vpnClient.Auth(context.Background(), &protocol.AuthRequest{})
	assert.NoError(err)
	spew.Dump(result2)

	// Test Stream
	ctx := metadata.NewOutgoingContext(context.Background(), auth.JWTAuthHeaderForGRPC(result2.Jwt))
	result3, err := vpnClient.Exchange(ctx)
	assert.NoError(err)
	_ = result3
	time.Sleep(10 * time.Second)
}
