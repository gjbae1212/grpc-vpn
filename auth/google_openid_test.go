// +build integration

package auth

import (
	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"testing"

	"github.com/tj/assert"
)

func TestUnaryServerInterceptor(t *testing.T) {
	assert := assert.New(t)

	_ = UnaryServerInterceptor("", "")
	_ = assert
}

func TestConfig_ClientAuthMethod(t *testing.T) {
	assert := assert.New(t)

	// Google OAuth
	conf := &Config{
		ClientId:     "",
		ClientSecret: "",
	}

	method := conf.ClientAuthMethod()
	method(protocol.NewVPNClient(nil))
	_ = assert
}
