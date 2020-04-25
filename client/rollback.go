package client

import (
	"net"

	"github.com/fatih/color"
	"github.com/gjbae1212/grpc-vpn/internal"
	"github.com/songgao/water"
)

type Rollback struct {
	Routes        []route
	reset         bool
	originGateway string
	tun           *water.Interface
}

type route struct {
	dest net.IP
	via  net.IP
	dev  string
}

// AddRoute adds a route to the deletion set when it is reset.
func (r *Rollback) AddRoute(destination net.IP, via net.IP, dev string) {
	r.Routes = append(r.Routes, route{
		dest: destination,
		via:  via,
		dev:  dev,
	})
}

// ResetGatewayOSX tells the rollback object what gateway should be set on exit.
func (r *Rollback) ResetGatewayOSX(tun *water.Interface, gw string) {
	r.reset = true
	r.originGateway = gw
	r.tun = tun
}

// Close is to rollback applied network settings.
func (r *Rollback) Close() {
	for _, route := range r.Routes {
		e := internal.DelRoute(route.dest, route.via, route.dev)
		if e == nil {
			defaultLogger.Info(color.GreenString("Deleted route to %s via %s on %s\n", route.dest.String(), route.via.String(), route.dev))
		} else {
			defaultLogger.Info(color.RedString("Error: Route delete %s (%s on %s) - %s\n", route.dest.String(), route.via.String(), route.dev, e.Error()))
		}
	}

	if r.reset {
		r.tun.Close()
		internal.CommandExec("route", []string{"add", "default", r.originGateway})
	}
}
