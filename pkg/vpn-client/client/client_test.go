package client

import (
	"testing"

	"github.com/gjbae1212/grpc-vpn/internal"
	"github.com/tj/assert"
)

func TestSetDefaultLogger(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]struct {
		input *internal.Logger
	}{
		"success": {input: nil},
	}

	for _, t := range tests {
		SetDefaultLogger(t.input)
	}
	_ = assert
}
