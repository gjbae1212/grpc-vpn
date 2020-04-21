package internal

import (
	"github.com/dgrijalva/jwt-go"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEncodeAndDecodeJWT(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		claims *jwt.StandardClaims
		salt   []byte
		isErr  bool
		valid  bool
	}{
		"fail": {isErr: true},
		"expired": {
			claims: &jwt.StandardClaims{
				Audience:  "hello",
				ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
				Issuer:    "vpn-server",
				Subject:   "vpn jwt token",
			},
			salt:  []byte("allan"),
			valid: false,
		},
		"success": {
			claims: &jwt.StandardClaims{
				Audience:  "hello",
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				Issuer:    "vpn-server",
				Subject:   "vpn jwt token",
			},
			salt:  []byte("allan"),
			valid: true,
		},
	}

	for _, t := range tests {
		b64, err := EncodeJWT(t.claims, t.salt)
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			_, err := DecodeJWT(b64, t.salt)
			assert.Equal(t.valid, err == nil)
		}
	}
}
