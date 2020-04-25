package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVPNStreamContext_Context(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	tests := map[string]struct {
		ctx context.Context
	}{
		"success": {ctx: ctx},
	}

	for _, t := range tests {
		vc := AuthorizedContext{Ctx: ctx}
		assert.Equal(t.ctx, vc.Context())
	}
}
