package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandExec(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		command string
		args    []string
		isErr   bool
	}{
		"ls": {command: "ls", args: []string{"-al"}},
	}

	for _, t := range tests {
		err := CommandExec(t.command, t.args)
		assert.Equal(t.isErr, err != nil)
	}
}

func TestSetCommandLogger(t *testing.T) {
	assert := assert.New(t)

	tempLogger, _ := NewLogger("")
	tests := map[string]struct {
		input  *Logger
		output *Logger
	}{
		"success": {input: tempLogger, output: tempLogger},
	}

	for _, t := range tests {
		SetCommandLogger(t.input)
		assert.Equal(t.output, t.output)
	}
}
