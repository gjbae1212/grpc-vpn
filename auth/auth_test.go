package auth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_ServerAuthForGoogleOpenID(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		cfg *Config
		ok  bool
	}{
		"false": {
			cfg: &Config{},
			ok:  false,
		},
		"true": {
			cfg: &Config{
				GoogleOpenId: &GoogleOpenIDConfig{
					ClientId:     "allan",
					ClientSecret: "allan",
				},
			},
			ok: true,
		},
	}

	for _, t := range tests {
		_, ok := t.cfg.ServerAuthForGoogleOpenID()
		assert.Equal(t.ok, ok)
	}
}

func TestConfig_ServerAuthForAwsIAM(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		cfg *Config
		ok  bool
	}{
		"false": {
			cfg: &Config{},
			ok:  false,
		},
		"true": {
			cfg: &Config{
				AwsIAM: &AwsIamConfig{
					AccessKey:       "allan",
					SecretAccessKey: "allan",
				},
			},
			ok: true,
		},
	}

	for _, t := range tests {
		_, ok := t.cfg.ServerAuthForAwsIAM()
		assert.Equal(t.ok, ok)
	}
}

func TestConfig_ServerAuthForTest(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		cfg *Config
		ok  bool
	}{
		"true": {
			cfg: &Config{},
			ok:  true,
		},
	}

	for _, t := range tests {
		_, ok := t.cfg.ServerAuthForTest()
		assert.Equal(t.ok, ok)
	}
}

func TestConfig_ClientAuthForAwsIAM(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		cfg *Config
		ok  bool
	}{
		"false": {
			cfg: &Config{},
			ok:  false,
		},
		"true": {
			cfg: &Config{
				AwsIAM: &AwsIamConfig{
					AccessKey:       "allan",
					SecretAccessKey: "allan",
				},
			},
			ok: true,
		},
	}

	for _, t := range tests {
		_, ok := t.cfg.ClientAuthForAwsIAM()
		assert.Equal(t.ok, ok)
	}

}

func TestConfig_ClientAuthForGoogleOpenID(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		cfg *Config
		ok  bool
	}{
		"false": {
			cfg: &Config{},
			ok:  false,
		},
		"true": {
			cfg: &Config{
				GoogleOpenId: &GoogleOpenIDConfig{
					ClientId:     "allan",
					ClientSecret: "allan",
				},
			},
			ok: true,
		},
	}

	for _, t := range tests {
		_, ok := t.cfg.ClientAuthForGoogleOpenID()
		assert.Equal(t.ok, ok)
	}
}

func TestConfig_ClientAuthForTest(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		cfg *Config
		ok  bool
	}{
		"true": {
			cfg: &Config{},
			ok:  true,
		},
	}

	for _, t := range tests {
		_, ok := t.cfg.ClientAuthForTest()
		assert.Equal(t.ok, ok)
	}
}

func TestJWTAuthHeaderForGRPC(t *testing.T) {
	assert := assert.New(t)
	opt := JWTAuthHeaderForGRPC("allan")
	_ = assert
	_ = opt
}
