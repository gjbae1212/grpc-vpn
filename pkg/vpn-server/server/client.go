package server

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/gjbae1212/grpc-vpn/internal"
	"github.com/pkg/errors"
	"github.com/songgao/water/waterutil"
	"go.uber.org/atomic"
	"io"
	"net"
)

const (
	queueSizeForClientIn = 500
)

type client struct {
	user     string                      // user
	originIP net.IP                      // user origin ip
	vpnIP    net.IP                      // user vpn ip
	jwt      *jwt.Token                  // user jwt token
	stream   protocol.VPN_ExchangeServer // stream
	loop     *atomic.Bool                // whether break loop or not
	exit     chan bool                   // exit

	out chan *protocol.IPPacket // out queue
	in  chan *protocol.IPPacket // in queue
}

// read packet
func (c *client) processReading() {
	for c.loop.Load() {
		packet, err := c.stream.Recv()
		if err == io.EOF {
			defaultLogger.Error(color.RedString("[ERR] %s (%s, %s) EOF",
				c.user, c.originIP.String(), c.vpnIP.String()))
			break
		}
		if err != nil {
			defaultLogger.Error(color.RedString("[ERR] %s (%s, %s) %s",
				c.user, c.originIP.String(), c.vpnIP.String(), err.Error()))
			break
		}

		// check error code
		switch packet.ErrorCode {
		case protocol.ErrorCode_EC_SUCCESS:
		default:
			// send unknown packet
			c.stream.Send(&protocol.IPPacket{ErrorCode: protocol.ErrorCode_EC_UNKNOWN})
			defaultLogger.Error(color.RedString("[ERR] %s (%s, %s) %s",
				c.user, c.originIP.String(), c.vpnIP.String(), internal.ErrorReceiveUnknownPacket.Error()))
			break
		}

		// check packet type
		switch packet.PacketType {
		case protocol.IPPacketType_IPPT_RAW:
		default:
			defaultLogger.Error(color.RedString("[ERR] %s (%s, %s) %s",
				c.user, c.originIP.String(), c.vpnIP.String(), internal.ErrorReceiveUnknownPacket.Error()))
			break
		}

		// check packet
		raw := packet.Packet1
		if raw == nil {
			defaultLogger.Error(color.RedString("[ERR] %s (%s, %s) %s",
				c.user, c.originIP.String(), c.vpnIP.String(), internal.ErrorReceiveUnknownPacket.Error()))
			break
		}

		// check source ip(equals vpn ip)
		srcIP := waterutil.IPv4Source(raw.Raw)
		if !srcIP.Equal(c.vpnIP) {
			defaultLogger.Error(color.RedString("[ERR] %s (%s, %s) %s(%s)",
				c.user, c.originIP.String(), c.vpnIP.String(), internal.ErrorReceiveUnknownPacket.Error(), srcIP))
			break
		}

		// out to server
		c.out <- packet
	}

	// flag off
	c.loop.Store(false)
	// writing exit
	c.exit <- true
}

// write packet
func (c *client) processWriting() {
	for c.loop.Load() {
		select {
		case packet := <-c.in:
			if err := c.stream.Send(packet); err != nil {
				defaultLogger.Error(color.RedString("[ERR] %s (%s, %s) %s",
					c.user, c.originIP.String(), c.vpnIP.String(), err.Error()))
				break
			}
		case <-c.exit:
			defaultLogger.Error(color.RedString("[ERR] %s (%s, %s) exit signal",
				c.user, c.originIP.String(), c.vpnIP.String()))
			break
		}
	}
	// flag off
	c.loop.Store(false)
}

// hasVpnIP is to check whether to be assigned vpn ip in client or not.
func (c *client) hasVpnIP() bool {
	return c.vpnIP != nil
}

// newClient is to create new client.
func newClient(stream protocol.VPN_ExchangeServer, clientToServer chan *protocol.IPPacket) (*client, error) {
	if stream == nil {
		return nil, errors.Wrapf(internal.ErrorInvalidParams, "Method: newClient")
	}

	// extract params
	ctx := stream.Context()

	ip := ctx.Value(ipCtxName)
	if ip == nil {
		return nil, errors.Wrapf(internal.ErrorInvalidContext, "Method: newClient")
	}
	j := ctx.Value(jwtCtxName)
	if j == nil {
		return nil, errors.Wrapf(internal.ErrorInvalidContext, "Method: newClient")
	}

	c := &client{
		user:     j.(*jwt.Token).Claims.(*jwt.StandardClaims).Audience,
		originIP: ip.(net.IP),
		jwt:      j.(*jwt.Token),
		stream:   stream,
		loop:     atomic.NewBool(true),
		exit:     make(chan bool, 1),
		out:      clientToServer,                                      // vpn server  queue
		in:       make(chan *protocol.IPPacket, queueSizeForClientIn), // only exclusive client queue
	}

	return c, nil
}
