package server

import (
	"context"
	"reflect"
	"testing"

	"google.golang.org/grpc"

	"github.com/stretchr/testify/assert"
)

func TestWithVpnSubNet(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  string
		output string
	}{
		"success": {
			input:  "allan",
			output: "allan",
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithVpnSubNet(t.input)
		f(c)
		assert.Equal(t.output, c.vpnSubNet)
	}
}

func TestWithJwtSalt(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  string
		output string
	}{
		"success": {
			input:  "allan",
			output: "allan",
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithVpnJwtSalt(t.input)
		f(c)
		assert.Equal(t.output, c.vpnJwtSalt)
	}
}

func TestWithGrpcPort(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  string
		output string
	}{
		"success": {
			input:  "allan",
			output: "allan",
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcPort(t.input)
		f(c)
		assert.Equal(t.output, c.grpcPort)
	}
}

func TestWithGrpcTlsCertification(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  string
		output string
	}{
		"success": {
			input:  "allan",
			output: "allan",
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcTlsCertification(t.input)
		f(c)
		assert.Equal(t.output, c.grpcTlsCertification)
	}
}

func TestWithGrpcTlsPem(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  string
		output string
	}{
		"success": {
			input:  "allan",
			output: "allan",
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcTlsPem(t.input)
		f(c)
		assert.Equal(t.output, c.grpcTlsPem)
	}
}

func TestWithGrpcUnaryInterceptors(t *testing.T) {
	assert := assert.New(t)

	f := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		return nil, nil
	}
	input := []grpc.UnaryServerInterceptor{f}

	tests := map[string]struct {
		input  []grpc.UnaryServerInterceptor
		output []grpc.UnaryServerInterceptor
	}{
		"success": {
			input:  input,
			output: input,
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcUnaryInterceptors(t.input)
		f(c)
		assert.True(reflect.DeepEqual(t.input, t.output))
	}
}

func TestWithGrpcStreamInterceptors(t *testing.T) {
	assert := assert.New(t)

	f := func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return nil
	}
	input := []grpc.StreamServerInterceptor{f}

	tests := map[string]struct {
		input  []grpc.StreamServerInterceptor
		output []grpc.StreamServerInterceptor
	}{
		"success": {
			input:  input,
			output: input,
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcStreamInterceptors(t.input)
		f(c)
		assert.True(reflect.DeepEqual(t.input, t.output))
	}
}

func TestWithGrpcOptions(t *testing.T) {
	assert := assert.New(t)

	input := []grpc.ServerOption{grpc.WriteBufferSize(10)}

	tests := map[string]struct {
		input  []grpc.ServerOption
		output []grpc.ServerOption
	}{
		"success": {
			input:  input,
			output: input,
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGrpcOptions(t.input)
		f(c)
		assert.True(reflect.DeepEqual(t.input, t.output))
	}
}
