package auth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewServerManagerForGoogleOpenID(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		clientId     string
		clientSecret string
		hd           string
		allowEmails  []string
		isErr        bool
	}{
		"fail":    {isErr: true},
		"success": {clientId: "client-id", clientSecret: "client-secret", hd: "hd-id", allowEmails: []string{"gjbae1212@gmail.com", "example@gmail.com"}},
	}

	for _, t := range tests {
		s, err := NewServerManagerForGoogleOpenID(t.clientId, t.clientSecret, t.hd, t.allowEmails)
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			assert.Equal(t.clientId, s.(*GoogleOpenIDConfig).ClientId)
			assert.Equal(t.clientSecret, s.(*GoogleOpenIDConfig).ClientSecret)
			assert.Equal(t.hd, s.(*GoogleOpenIDConfig).HD)
			assert.Equal(t.allowEmails, s.(*GoogleOpenIDConfig).AllowEmails)
		}
	}
}

func TestNewClientManagerForGoogleOpenID(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		clientId     string
		clientSecret string
		isErr        bool
	}{
		"fail":    {isErr: true},
		"success": {clientId: "client-id", clientSecret: "client-secret"},
	}

	for _, t := range tests {
		s, err := NewClientManagerForGoogleOpenID(t.clientId, t.clientSecret)
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			assert.Equal(t.clientId, s.(*GoogleOpenIDConfig).ClientId)
			assert.Equal(t.clientSecret, s.(*GoogleOpenIDConfig).ClientSecret)
		}
	}
}

func TestGoogleOpenIDConfig_ServerAuth(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		cfg *GoogleOpenIDConfig
		ok  bool
	}{
		"fail":    {cfg: &GoogleOpenIDConfig{}},
		"success": {cfg: &GoogleOpenIDConfig{ClientId: "id", ClientSecret: "secret"}, ok: true},
	}

	for _, t := range tests {
		_, ok := t.cfg.ServerAuth()
		assert.Equal(t.ok, ok)
	}
}

func TestGoogleOpenIDConfig_ClientAuth(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		cfg *GoogleOpenIDConfig
		ok  bool
	}{
		"fail":    {cfg: &GoogleOpenIDConfig{}},
		"success": {cfg: &GoogleOpenIDConfig{ClientId: "id", ClientSecret: "secret"}, ok: true},
	}

	for _, t := range tests {
		_, ok := t.cfg.ClientAuth()
		assert.Equal(t.ok, ok)
	}
}
