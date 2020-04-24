package auth

import (
	"context"
	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"

	"google.golang.org/grpc"
)

type AwsIamConfig struct {
	AccessKey       string // aws access key
	SecretAccessKey string // aws secret key
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

		// TODO: AWS IAM
		return handler(ctx, req)
	}
}

func (c *AwsIamConfig) clientAuthMethod() ClientAuthMethod {
	return nil
}
