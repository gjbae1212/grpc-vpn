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

// Config is a config for authorization in server.
type Config struct {
	GoogleOpenId *GoogleOpenIDConfig
	AwsIAM       *AwsIamConfig
}

type ClientAuthMethod func(conn protocol.VPNClient) (jwt string, err error)

// ServerAuthForGoogleOpenID returns interceptor and bool value(whether exist or not).
func (c *Config) ServerAuthForGoogleOpenID() (grpc.UnaryServerInterceptor, bool) {
	if c.GoogleOpenId == nil || c.GoogleOpenId.ClientId == "" ||
		c.GoogleOpenId.ClientSecret == "" {
		return nil, false
	}
	return c.GoogleOpenId.unaryServerInterceptor(), true
}

// ServerAuthForGoogleOpenID returns interceptor and bool value(whether exist or not).
func (c *Config) ServerAuthForAwsIAM() (grpc.UnaryServerInterceptor, bool) {
	if c.AwsIAM == nil || c.AwsIAM.ServerAccountId == "" {
		return nil, false
	}
	return c.AwsIAM.unaryServerInterceptor(), true
}

// ServerAuthForTest returns interceptor and bool value(whether exist or not).
func (c *Config) ServerAuthForTest() (grpc.UnaryServerInterceptor, bool) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		auth, ok := req.(*protocol.AuthRequest)
		if !ok {
			return handler(ctx, req)
		}
		if auth.AuthType != protocol.AuthType_AT_GOOGLE_TEST {
			return handler(ctx, req)
		}
		// inject user(vpn-test)
		newCtx := context.WithValue(ctx, UserCtxName, "vpn-test")
		return handler(newCtx, req)
	}, true
}

// ClientAuthForGoogleOpenID is returns Auth Method for Google Open ID.
func (c *Config) ClientAuthForGoogleOpenID() (ClientAuthMethod, bool) {
	if c.GoogleOpenId == nil || c.GoogleOpenId.ClientId == "" ||
		c.GoogleOpenId.ClientSecret == "" {
		return nil, false
	}
	return c.GoogleOpenId.clientAuthMethod(), true
}

// ClientAuthForAwsIAM is returns Auth Method for AWS IAM.
func (c *Config) ClientAuthForAwsIAM() (ClientAuthMethod, bool) {
	if c.AwsIAM == nil || c.AwsIAM.ClientAccessKey == "" || c.AwsIAM.ClientSecretAccessKey == "" {
		return nil, false
	}
	return c.AwsIAM.clientAuthMethod(), true
}

// ClientAuthForTest is returns Auth Method for Test.
func (c *Config) ClientAuthForTest() (ClientAuthMethod, bool) {
	return func(conn protocol.VPNClient) (jwt string, err error) {
		// Timeout 30 seconds
		ctx := context.Background()
		timeout := 30 * time.Second
		newctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// call default auth
		result, err := conn.Auth(newctx, &protocol.AuthRequest{
			AuthType: protocol.AuthType_AT_GOOGLE_TEST,
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

// JWTAuthHeaderForGRPC returns JWT Auth Header for GRPC.
func JWTAuthHeaderForGRPC(jwtToken string) metadata.MD {
	header := Bearer + " " + jwtToken
	return metadata.New(map[string]string{AuthorizationHeader: header})
}
