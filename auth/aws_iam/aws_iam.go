package auth_aws_iam

import (
	"context"

	"google.golang.org/grpc"
)

const (
	userCtxName = "user"
)

type Config struct {
	AccessKey       string // aws access key
	SecretAccessKey string // aws secret key
}

// UnaryServerInterceptor returns new unary server interceptor that checks an authorization with aws iam.
func UnaryServerInterceptor(awsAccess, awsSecretAccess string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// if authentication already completed.
		if ctx.Value(userCtxName) != nil {
			return handler(ctx, req)
		}

		// TODO: AWS IAM
		return nil, nil
	}
}
