package auth

import (
	auth_aws_iam "github.com/gjbae1212/grpc-vpn/auth/aws_iam"
	auth_google_openid "github.com/gjbae1212/grpc-vpn/auth/google_openid"
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

// ServerConfig is a config for authorization in server.
type ServerConfig struct {
	GoogleOpenId *auth_google_openid.Config
	AwsIAM       *auth_aws_iam.Config
}

type AuthMethod func(conn protocol.VPNClient) (jwt string, err error)

// AuthForGoogleOpenID returns interceptor and bool value(whether exist or not).
func (c *ServerConfig) AuthForGoogleOpenID() (interceptor grpc.UnaryServerInterceptor, ok bool) {
	if c.GoogleOpenId == nil || c.GoogleOpenId.ClientId == "" ||
		c.GoogleOpenId.ClientSecret == "" || c.GoogleOpenId.RedirectURL == "" {
		return nil, false
	}
	return auth_google_openid.UnaryServerInterceptor(c.GoogleOpenId.ClientId,
		c.GoogleOpenId.ClientSecret, c.GoogleOpenId.RedirectURL), true
}

// AuthForGoogleOpenID returns interceptor and bool value(whether exist or not).
func (c *ServerConfig) AuthForAwsIAM() (interceptor grpc.UnaryServerInterceptor, ok bool) {
	if c.AwsIAM == nil || c.AwsIAM.AccessKey == "" || c.AwsIAM.SecretAccessKey == "" {
		return nil, false
	}
	return auth_aws_iam.UnaryServerInterceptor(c.AwsIAM.AccessKey, c.AwsIAM.SecretAccessKey), true
}

// JWTAuthHeaderForGRPC returns JWT Auth Header for GRPC.
func JWTAuthHeaderForGRPC(jwtToken string) metadata.MD {
	header := Bearer + " " + jwtToken
	return metadata.New(map[string]string{AuthorizationHeader: header})
}
