package server

import (
	"context"

	"github.com/pkg/errors"

	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/gjbae1212/grpc-vpn/internal"
)

// Handler is an object for exchanging packet between client and server.
type handler struct {
	srv *vpnServer
}

// NewHandler returns new handler object.
func NewHandler(srv *vpnServer) (*handler, error) {
	if srv == nil {
		return nil, errors.Wrapf(internal.ErrorInvalidParams, "Method: %s", "NewHandler")
	}
	return &handler{srv: srv}, nil
}

func (h *handler) Auth(ctx context.Context, req *protocol.AuthRequest) (*protocol.AuthResponse, error) {
	return nil, nil
}

func (h *handler) Exchange(stream protocol.VPN_ExchangeServer) error {
	return nil
	//stream.Recv()
	//stream.Send()
}
