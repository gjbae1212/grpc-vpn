package internal

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestIncreaseIP(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  net.IP
		output net.IP
	}{
		"ex1": {
			input:  net.IPv4(uint8(10), uint8(10), uint8(10), uint8(10)),
			output: net.IPv4(uint8(10), uint8(10), uint8(10), uint8(11)),
		},
		"ex2": {
			input:  net.IPv4(uint8(10), uint8(10), uint8(10), uint8(255)),
			output: net.IPv4(uint8(10), uint8(10), uint8(11), uint8(0)),
		},
		"ex3": {
			input:  net.IPv4(uint8(10), uint8(10), uint8(10), uint8(0)),
			output: net.IPv4(uint8(10), uint8(10), uint8(10), uint8(1)),
		},
	}

	for _, t := range tests {
		IncreaseIP(t.input)
		assert.Equal(t.output, t.input)
	}
}
