package client

import (
	"testing"

	"github.com/gjbae1212/grpc-vpn/auth"
	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/stretchr/testify/assert"
)

func TestWithServerAddr(t *testing.T) {
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
		f := WithServerAddr(t.input)
		f(c)
		assert.Equal(t.output, c.serverAddr)
	}
}

func TestWithServerPort(t *testing.T) {
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
		f := WithServerPort(t.input)
		f(c)
		assert.Equal(t.output, c.serverPort)
	}
}

func TestWithAuthMethod(t *testing.T) {
	assert := assert.New(t)

	temp := func(conn protocol.VPNClient) (jwt string, err error) {
		return "allan", nil
	}

	tests := map[string]struct {
		input  auth.ClientAuthMethod
		output auth.ClientAuthMethod
	}{
		"success": {
			input:  temp,
			output: temp,
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithAuthMethod(t.input)
		f(c)
		a, _ := t.output(nil)
		b, _ := c.authMethod(nil)
		assert.Equal(a, b)

	}
}

func TestWithSelfSignedCertification(t *testing.T) {
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
		f := WithSelfSignedCertification(t.input)
		f(c)
		assert.Equal(t.output, c.selfSignedCertification)
	}
}

func TestWithGRPCInsecure(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  bool
		output bool
	}{
		"success": {
			input:  true,
			output: true,
		},
	}

	for _, t := range tests {
		c := &config{}
		f := WithGRPCInsecure(t.input)
		f(c)
		assert.Equal(t.output, c.grpcInsecure)
	}
}
