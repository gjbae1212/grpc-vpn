package server

import (
	"context"

	"google.golang.org/grpc"
)

// AuthorizedContext is a wrapper for stream context in GRPC.
type AuthorizedContext struct {
	grpc.ServerStream
	Ctx context.Context
}

// Context returns an internal context
func (rs *AuthorizedContext) Context() context.Context {
	return rs.Ctx
}
