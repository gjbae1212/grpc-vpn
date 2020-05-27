package client

import (
	"github.com/gjbae1212/grpc-vpn/auth"
)

// Option is to use a dependency injection for handler.
type Option interface {
	apply(cfg *config)
}

type config struct {
	serverAddr              string
	serverPort              string
	grpcInsecure            bool
	selfSignedCertification string
	authMethod              auth.ClientAuthMethod
}

// OptionFunc is a function for Option interface.
type OptionFunc func(c *config)

func (o OptionFunc) apply(c *config) { o(c) }

// WithServerAddr returns OptionFunc for inserting server addr.
func WithServerAddr(addr string) OptionFunc {
	return func(c *config) {
		c.serverAddr = addr
	}
}

// WithServerPort returns OptionFunc for inserting server port.
func WithServerPort(port string) OptionFunc {
	return func(c *config) {
		c.serverPort = port
	}
}

// WithAuthMethod returns OptionFunc for inserting auth method.
func WithAuthMethod(f auth.ClientAuthMethod) OptionFunc {
	return func(c *config) {
		c.authMethod = f
	}
}

// WithGRPCInsecure returns OptionFunc for inserting grpc insecure.
func WithGRPCInsecure(b bool) OptionFunc {
	return func(c *config) {
		c.grpcInsecure = b
	}
}

// WithSelfSignedCertification returns OptionFunc for inserting grpc custom certification
func WithSelfSignedCertification(cert string) OptionFunc {
	return func(c *config) {
		c.selfSignedCertification = cert
	}
}
