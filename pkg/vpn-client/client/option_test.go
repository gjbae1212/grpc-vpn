package client

import (
	"github.com/gjbae1212/grpc-vpn/auth"
	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/tj/assert"
	"testing"
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

func TestWithTlsCertification(t *testing.T) {
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
		f := WithTlsCertification(t.input)
		f(c)
		assert.Equal(t.output, c.tlsCertification)
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
