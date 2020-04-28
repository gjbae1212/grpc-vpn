package client

import (
	"testing"

	"github.com/sirupsen/logrus"

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
