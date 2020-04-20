package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input string
		isErr bool
	}{
		"success-1": {input: "", isErr: false},
		"success-2": {input: "../tmp.txt", isErr: false},
	}

	for _, t := range tests {
		l, err := NewLogger(t.input)
		assert.Equal(t.isErr, err != nil)
		l.Info("TestNewLogger test")
	}
}

func TestLogger_PanicWithError(t *testing.T) {
	assert := assert.New(t)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	logger, _ := NewLogger("")
	tests := map[string]struct {
		input error
	}{
		"test": {input: fmt.Errorf("print TestLogger_PanicWithMessage")},
	}

	for _, t := range tests {
		logger.PanicWithError(t.input)
	}
	_ = assert
}
