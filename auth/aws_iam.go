package auth

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/gjbae1212/grpc-vpn/internal"

	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"

	"google.golang.org/grpc"
)

type AwsIamConfig struct {
	ClientAccessKey       string
	ClientSecretAccessKey string

	ServerAllowUsers []string // allow users
	ServerAccountId  string   // server allow users
}

// ServerAuth returns ServerAuthMethod and bool value(whether exist or not).
func (c *AwsIamConfig) ServerAuth() (ServerAuthMethod, bool) {
	if c.ServerAccountId == "" {
		return nil, false
	}
	return ServerAuthMethod(c.unaryServerInterceptor()), true
}

// ClientAuth is returns ClientAuthMethod for AWS IAM.
func (c *AwsIamConfig) ClientAuth() (ClientAuthMethod, bool) {
	if c.ClientAccessKey == "" || c.ClientSecretAccessKey == "" {
		return nil, false
	}
	return c.clientAuthMethod(), true
}

// unaryServerInterceptor returns new unary server interceptor that checks an authorization with aws iam.
func (c *AwsIamConfig) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		auth, ok := req.(*protocol.AuthRequest)
		if !ok {
			return handler(ctx, req)
		}
		if auth.AuthType != protocol.AuthType_AT_AWS_IAM {
			return handler(ctx, req)
		}

		accountId := c.ServerAccountId
		allowUsers := c.ServerAllowUsers

		sess, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(auth.AwsIam.AccessKey, auth.AwsIam.SecretAccessKey, ""),
		})
		if err != nil {
			return nil, internal.ErrorUnauthorized
		}

		identity, err := sts.New(sess).GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			return nil, internal.ErrorUnauthorized
		}

		// must be equal to account
		if accountId != *identity.Account {
			return nil, internal.ErrorUnauthorized
		}

		args := strings.Split(*identity.Arn, "/")
		user := args[len(args)-1]
		if len(allowUsers) != 0 {
			if !internal.IsMatchedStringFromSlice(user, allowUsers) {
				return nil, internal.ErrorUnauthorized
			}
		}

		// inject user
		newCtx := context.WithValue(ctx, UserCtxName, user)
		return handler(newCtx, req)
	}
}

func (c *AwsIamConfig) clientAuthMethod() ClientAuthMethod {
	return func(conn protocol.VPNClient) (jwt string, err error) {
		if conn == nil {
			return "", errors.Wrapf(internal.ErrorInvalidParams, "AWS IAM  ClientAuthMethod")
		}

		// extract information
		accessKey := c.ClientAccessKey
		secretAccessKey := c.ClientSecretAccessKey
		if accessKey == "" || secretAccessKey == "" {
			return "", errors.Wrapf(internal.ErrorInvalidParams, "AWS IAM  ClientAuthMethod")
		}

		// call authentication request to VPN server.
		authCtx, authCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer authCancel()
		response, err := conn.Auth(authCtx, &protocol.AuthRequest{
			AuthType: protocol.AuthType_AT_AWS_IAM,
			AwsIam: &protocol.AuthRequest_AwsIam{
				AccessKey:       accessKey,
				SecretAccessKey: secretAccessKey,
			},
		})
		if err != nil {
			return "", errors.Wrapf(internal.ErrorInvalidParams, "AWS IAM  ClientAuthMethod")
		}
		if response.ErrorCode != protocol.ErrorCode_EC_SUCCESS || response.Jwt == "" {
			return "", errors.Wrapf(internal.ErrorInvalidParams, "AWS IAM  ClientAuthMethod")
		}

		return response.Jwt, nil
	}
}

// NewServerManagerForAwsIAM returns ServerManager implementing awsIam.
func NewServerManagerForAwsIAM(accountId string, allowUsers []string) (ServerManager, error) {
	if accountId == "" {
		return nil, internal.ErrorInvalidParams
	}

	if allowUsers == nil {
		allowUsers = []string{}
	}

	return &AwsIamConfig{
		ServerAccountId:  accountId,
		ServerAllowUsers: allowUsers,
	}, nil
}

// NewClientManagerForAwsIAM returns ClientManager implementing awsIam.
func NewClientManagerForAwsIAM(accessKey, accessSecret string) (ClientManager, error) {
	if accessKey == "" || accessSecret == "" {
		return nil, internal.ErrorInvalidParams
	}

	return &AwsIamConfig{
		ClientAccessKey:       accessKey,
		ClientSecretAccessKey: accessSecret,
	}, nil
}
