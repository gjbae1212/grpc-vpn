package server

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/gjbae1212/grpc-vpn/auth"

	"google.golang.org/grpc"

	"github.com/stretchr/testify/assert"
)

func TestWithVpnSubNet(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input string
	}{
		"success": {
			input: "allan",
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithVpnSubNet(t.input)
		f(c)
		assert.Equal(t.input, c.vpnSubNet)
	}
}

func TestWithJwtSalt(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input string
	}{
		"success": {
			input: "allan",
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithVpnJwtSalt(t.input)
		f(c)
		assert.Equal(t.input, c.vpnJwtSalt)
	}
}

func TestWithVpnJwtExpiration(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input time.Duration
	}{
		"success": {
			input: time.Hour,
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithVpnJwtExpiration(t.input)
		f(c)
		assert.Equal(t.input, c.vpnJwtExpiration)
	}
}

func TestWithGrpcPort(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input string
	}{
		"success": {
			input: "allan",
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcPort(t.input)
		f(c)
		assert.Equal(t.input, c.grpcPort)
	}
}

func TestWithGrpcTlsCertification(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input string
	}{
		"success": {
			input: "allan",
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcTlsCertification(t.input)
		f(c)
		assert.Equal(t.input, c.grpcTlsCertification)
	}
}

func TestWithGrpcTlsPem(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input string
	}{
		"success": {
			input: "allan",
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcTlsPem(t.input)
		f(c)
		assert.Equal(t.input, c.grpcTlsPem)
	}
}

func TestWithGrpcUnaryInterceptors(t *testing.T) {
	assert := assert.New(t)

	f := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		return nil, nil
	}
	input := []grpc.UnaryServerInterceptor{f}

	tests := map[string]struct {
		input []grpc.UnaryServerInterceptor
	}{
		"success": {
			input: input,
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcUnaryInterceptors(t.input)
		f(c)
		assert.True(reflect.DeepEqual(t.input, c.grpcUnaryInterceptors))
	}
}

func TestWithGrpcStreamInterceptors(t *testing.T) {
	assert := assert.New(t)

	f := func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return nil
	}
	input := []grpc.StreamServerInterceptor{f}

	tests := map[string]struct {
		input []grpc.StreamServerInterceptor
	}{
		"success": {
			input: input,
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcStreamInterceptors(t.input)
		f(c)
		assert.True(reflect.DeepEqual(t.input, c.grpcStreamInterceptors))
	}
}

func TestWithGrpcOptions(t *testing.T) {
	assert := assert.New(t)

	input := []grpc.ServerOption{grpc.WriteBufferSize(10)}

	tests := map[string]struct {
		input []grpc.ServerOption
	}{
		"success": {
			input: input,
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcOptions(t.input)
		f(c)
		assert.True(reflect.DeepEqual(t.input, c.grpcOptions))
	}
}

func TestWithAuthMethods(t *testing.T) {
	assert := assert.New(t)

	f := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		return nil, nil
	}
	input := []auth.ServerAuthMethod{f}

	tests := map[string]struct {
		input []auth.ServerAuthMethod
	}{
		"success": {
			input: input,
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithAuthMethods(t.input)
		f(c)
		assert.True(reflect.DeepEqual(t.input, c.grpcAuthMethods))
	}
}
