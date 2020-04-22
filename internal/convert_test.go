package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterfaceToString(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  interface{}
		output string
	}{
		"empty":   {input: nil},
		"struct":  {input: map[string]string{}},
		"string":  {input: "string", output: "string"},
		"int":     {input: 10, output: "10"},
		"int64":   {input: int64(10), output: "10"},
		"float32": {input: float32(10), output: "10"},
		"float64": {input: float64(10), output: "10"},
		"bool":    {input: true, output: "true"},
	}

	for _, t := range tests {
		v := InterfaceToString(t.input)
		assert.Equal(t.output, v)
	}
}
