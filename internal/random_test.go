package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]struct {
		input int
	}{
		"0": {
			input: 9,
		},
		"4": {
			input: 4,
		},
	}

	for _, t := range tests {
		result := GenerateRandomString(t.input)
		assert.Len(result, t.input)
	}
}
