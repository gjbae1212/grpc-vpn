package client

import (
	"github.com/gjbae1212/grpc-vpn/auth"
	"github.com/sirupsen/logrus"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetDefaultLogger(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]struct {
		input *logrus.Logger
	}{
		"success": {input: nil},
	}

	for _, t := range tests {
		SetDefaultLogger(t.input)
	}
	_ = assert
}

func TestNewVpnClient(t *testing.T) {
	assert := assert.New(t)

	testAuth, _ := auth.NewClientManagerForTest()
	testAuthMethod, _ := testAuth.ClientAuth()

	googleAuth, _ := auth.NewClientManagerForGoogleOpenID("a", "a")
	googleAuthMethod, _ := googleAuth.ClientAuth()

	tests := map[string]struct {
		opts  []Option
		check *config
		err   error
	}{
		"default": {
			opts: []Option{
				WithServerAddr("1.1.1.1"),
				WithServerPort("80"),
			},
			check: &config{
				serverAddr:   "1.1.1.1",
				serverPort:   "80",
				grpcInsecure: false,
				authMethod:   testAuthMethod,
			},
		},
		"additional": {
			opts: []Option{
				WithServerAddr("1.1.1.1"),
				WithServerPort("80"),
				WithAuthMethod(googleAuthMethod),
			},
			check: &config{
				serverAddr:   "1.1.1.1",
				serverPort:   "80",
				grpcInsecure: false,
				authMethod:   googleAuthMethod,
			},
		},
	}

	for _, t := range tests {
		v, err := NewVpnClient(t.opts...)
		assert.Equal(t.err, err)
		assert.Equal(t.check.serverAddr, v.(*vpnClient).cfg.serverAddr)
		assert.Equal(t.check.serverPort, v.(*vpnClient).cfg.serverPort)
		assert.Equal(t.check.grpcInsecure, v.(*vpnClient).cfg.grpcInsecure)
	}
}
