package auth

import (
	"context"
	"time"

	"github.com/gjbae1212/grpc-vpn/internal"

	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	UserCtxName = "user"
)

const (
	// rfc2617 (e.g. Authorization: basic token, Authorization: bearer token)
	AuthorizationHeader = "authorization"
	Basic               = "basic"
	Bearer              = "bearer"
)

type ClientAuthMethod func(conn protocol.VPNClient) (jwt string, err error)
type ServerAuthMethod grpc.UnaryServerInterceptor

type ServerManager interface {
	ServerAuth() (ServerAuthMethod, bool)
}

type ClientManager interface {
	ClientAuth() (ClientAuthMethod, bool)
}

type defaultConfig struct{}

// AuthMethod returns ServerAuth and bool value(whether exist or not).
func (c *defaultConfig) ServerAuth() (ServerAuthMethod, bool) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		auth, ok := req.(*protocol.AuthRequest)
		if !ok {
			return handler(ctx, req)
		}
		if auth.AuthType != protocol.AuthType_AT_TEST {
			return handler(ctx, req)
		}
		// inject user(vpn-test)
		newCtx := context.WithValue(ctx, UserCtxName, "vpn-test")
		return handler(newCtx, req)
	}, true
}

// AuthMethod is returns ClientAuth for Test.
func (c *defaultConfig) ClientAuth() (ClientAuthMethod, bool) {
	return func(conn protocol.VPNClient) (jwt string, err error) {
		// Timeout 30 seconds
		ctx := context.Background()
		timeout := 30 * time.Second
		newctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// call default auth
		result, err := conn.Auth(newctx, &protocol.AuthRequest{
			AuthType: protocol.AuthType_AT_TEST,
		})
		if err != nil {
			return "", err
		}

		// extract JWT
		switch result.ErrorCode {
		case protocol.ErrorCode_EC_SUCCESS:
			return result.Jwt, nil
		default:
			return "", internal.ErrorUnauthorized
		}
	}, true
}

// NewServerManagerForTest returns ServerManager for test.
func NewServerManagerForTest() (ServerManager, error) {
	return &defaultConfig{}, nil
}

// NewClientManagerForTest returns ClientManager for test
func NewClientManagerForTest() (ClientManager, error) {
	return &defaultConfig{}, nil
}

// JWTAuthHeaderForGRPC returns JWT Auth Header for GRPC.
func JWTAuthHeaderForGRPC(jwtToken string) metadata.MD {
	header := Bearer + " " + jwtToken
	return metadata.New(map[string]string{AuthorizationHeader: header})
}
