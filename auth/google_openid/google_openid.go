package auth_google_openid

import (
	"context"

	"google.golang.org/grpc"
)

const (
	userCtxName = "user"
)

// https://developers.google.com/identity/protocols/oauth2/native-app
type Config struct {
	ClientId     string // google client id
	ClientSecret string // google secret
	RedirectURL  string // redirect url ([required] localhost)
}

// UnaryServerInterceptor returns new unary server interceptor that checks an authorization with google openID.
func UnaryServerInterceptor(clientId, clientSecret, redirectURL string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// if authentication already completed.
		if ctx.Value(userCtxName) != nil {
			return handler(ctx, req)
		}

		// TODO: google openID authentication
		return nil, nil
	}
}
