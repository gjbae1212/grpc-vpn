package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchedStringFromSlice(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]struct {
		check string
		slice []string
		ok    bool
	}{
		"not matched": {check: "data", slice: []string{"hi", "hello"}, ok: false},
		"matched":     {check: "data", slice: []string{"hi", "hello", "data"}, ok: true},
	}

	for _, t := range tests {
		assert.Equal(t.ok, IsMatchedStringFromSlice(t.check, t.slice))
	}
}
