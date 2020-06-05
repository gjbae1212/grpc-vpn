package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerManagerForTest(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		isErr bool
	}{
		"success": {isErr: false},
	}

	for _, t := range tests {
		_, err := NewServerManagerForTest()
		assert.Equal(t.isErr, err != nil)
	}
}

func TestNewClientManagerForTest(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		isErr bool
	}{
		"success": {isErr: false},
	}

	for _, t := range tests {
		_, err := NewClientManagerForTest()
		assert.Equal(t.isErr, err != nil)
	}
}

func TestDefaultConfig_ServerAuth(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		ok bool
	}{
		"success": {ok: true},
	}

	for _, t := range tests {
		s, _ := NewServerManagerForTest()
		_, ok := s.ServerAuth()
		assert.Equal(t.ok, ok)
	}
}

func TestDefaultConfig_ClientAuth(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		ok bool
	}{
		"success": {ok: true},
	}

	for _, t := range tests {
		s, _ := NewClientManagerForTest()
		_, ok := s.ClientAuth()
		assert.Equal(t.ok, ok)
	}
}

func TestJWTAuthHeaderForGRPC(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  string
		output string
	}{
		"success": {
			input:  "test-token",
			output: Bearer + " " + "test-token",
		},
	}

	for _, t := range tests {
		m := JWTAuthHeaderForGRPC(t.input)
		assert.Equal(t.output, m.Get(AuthorizationHeader)[0])
	}
}
