package server

import (
	"time"

	"github.com/gjbae1212/grpc-vpn/auth"

	"google.golang.org/grpc"
)

// Option is to use a dependency injection for handler.
type Option interface {
	apply(cfg *config)
}

type config struct {
	vpnSubNet              string
	vpnJwtSalt             string
	vpnJwtExpiration       time.Duration
	grpcPort               string
	grpcTlsCertification   string
	grpcTlsPem             string
	grpcUnaryInterceptors  []grpc.UnaryServerInterceptor
	grpcStreamInterceptors []grpc.StreamServerInterceptor
	grpcOptions            []grpc.ServerOption
	grpcAuthMethods        []auth.ServerAuthMethod
}

// OptionFunc is a function for Option interface.
type OptionFunc func(c *config)

func (o OptionFunc) apply(c *config) { o(c) }

// WithVpnSubNet returns OptionFunc for inserting VPN SUBNET.
func WithVpnSubNet(vpnSubNet string) OptionFunc {
	return func(c *config) {
		c.vpnSubNet = vpnSubNet
	}
}

// WithVpnJwtSalt returns OptionFunc for inserting VPN JWT SALT.
func WithVpnJwtSalt(vpnJwtSalt string) OptionFunc {
	return func(c *config) {
		c.vpnJwtSalt = vpnJwtSalt
	}
}

// WithVpnJwtExpiration returns OptionFunc for inserting VPN expiration time.
func WithVpnJwtExpiration(exp time.Duration) OptionFunc {
	return func(c *config) {
		c.vpnJwtExpiration = exp
	}
}

// WithGrpcPort returns OptionFunc for inserting GRPC PORT.
func WithGrpcPort(port string) OptionFunc {
	return func(c *config) {
		c.grpcPort = port
	}
}

// WithGrpcTlsCertification returns OptionFunc for inserting GRPC TLS Certification.
func WithGrpcTlsCertification(cert string) OptionFunc {
	return func(c *config) {
		c.grpcTlsCertification = cert
	}
}

// WithGrpcTlsPem returns OptionFunc for inserting GRPC TLS PRIVATE PEM.
func WithGrpcTlsPem(pem string) OptionFunc {
	return func(c *config) {
		c.grpcTlsPem = pem
	}
}

// WithAuthMethods returns OptionFunc for inserting GRPC authentication method.
func WithAuthMethods(methods []auth.ServerAuthMethod) OptionFunc {
	return func(c *config) {
		c.grpcAuthMethods = methods
	}
}

// WithGrpcUnaryInterceptors returns OptionFunc for inserting GRPC Unary Interceptors( such as auth(Google OpenId, AWS IAM) )
func WithGrpcUnaryInterceptors(interceptors []grpc.UnaryServerInterceptor) OptionFunc {
	return func(c *config) {
		c.grpcUnaryInterceptors = interceptors
	}
}

// WithGrpcStreamInterceptors returns OptionFunc for inserting GRPC Stream Interceptors( such as checking header associated JWT ).
func WithGrpcStreamInterceptors(interceptors []grpc.StreamServerInterceptor) OptionFunc {
	return func(c *config) {
		c.grpcStreamInterceptors = interceptors
	}
}

// WithGrpcOptions returns OptionFunc for inserting GRPC OPTIONS.
func WithGrpcOptions(opts []grpc.ServerOption) OptionFunc {
	return func(c *config) {
		c.grpcOptions = opts
	}
}
