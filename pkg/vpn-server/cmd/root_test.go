package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetConfig(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		path  string
		isErr bool
	}{
		"success": {path: "../sample.yaml"},
	}

	for _, t := range tests {
		err := setConfig(t.path)
		assert.Equal(t.isErr, err != nil)
	}
}
