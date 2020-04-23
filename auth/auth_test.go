package auth

import (
	auth_aws_iam "github.com/gjbae1212/grpc-vpn/auth/aws_iam"
	auth_google_openid "github.com/gjbae1212/grpc-vpn/auth/google_openid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServerConfig_AuthForAwsIAM(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		cfg *ServerConfig
		ok  bool
	}{
		"false": {
			cfg: &ServerConfig{},
			ok:  false,
		},
		"true": {
			cfg: &ServerConfig{
				GoogleOpenId: &auth_google_openid.Config{
					ClientId:     "allan",
					ClientSecret: "allan",
					RedirectURL:  "allan",
				},
			},
			ok: true,
		},
	}

	for _, t := range tests {
		_, ok := t.cfg.AuthForGoogleOpenID()
		assert.Equal(t.ok, ok)
	}
}

func TestServerConfig_AuthForGoogleOpenID(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		cfg *ServerConfig
		ok  bool
	}{
		"false": {
			cfg: &ServerConfig{},
			ok:  false,
		},
		"true": {
			cfg: &ServerConfig{
				AwsIAM: &auth_aws_iam.Config{
					AccessKey:       "allan",
					SecretAccessKey: "allan",
				},
			},
			ok: true,
		},
	}

	for _, t := range tests {
		_, ok := t.cfg.AuthForAwsIAM()
		assert.Equal(t.ok, ok)
	}
}

func TestJWTAuthHeaderForGRPC(t *testing.T) {
	assert := assert.New(t)
	opt := JWTAuthHeaderForGRPC("allan")
	_ = assert
	_ = opt
}
