package internal

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

// SetTunStatus is to up or down network device for TUN.
func SetTunStatus(tun string, up bool) error {
	status := "down"
	if up {
		status = "up"
	}
	sub := fmt.Sprintf("%s %s mtu %d", tun, status, TunMtuSize)
	args := strings.Split(sub, " ")
	return CommandExec("ifconfig", args)
}

// SetTunIP sets the local IP address of a network interface.
func SetTunIP(tun string, localAddr net.IP, addr *net.IPNet) error {
	sub := fmt.Sprintf("set %s MANUAL %s 0x%s", tun, localAddr.String(), addr.Mask)
	args := strings.Split(sub, " ")
	return CommandExec("ipconfig", args)
}

// SetDefaultGateway sets the systems gateway to the IP / device specified.
func SetDefaultGateway(gw, tun string) error {
	sub := fmt.Sprintf("-n change default -interface %s", tun)
	args := strings.Split(sub, " ")
	return CommandExec("route", args)
}

// SetGoogleDNS sets google dns
func SetGoogleDNS() error {
	return CommandExec("networksetup", []string{"-setdnsservers", "Wi-Fi", "8.8.8.8"})
}

// SetDeleteDNS sets google dns
func SetDeleteDNS() error {
	return CommandExec("networksetup", []string{"-setdnsservers", "Wi-Fi", "empty"})
}

// AddRoute routes all traffic for addr via interface iName.
func AddRoute(addr, viaAddr net.IP, tun string) error {
	sub := fmt.Sprintf("-n add %s %s -ifscope %s", addr.String(), viaAddr.String(), tun)
	args := strings.Split(sub, " ")
	return CommandExec("route", args)
}

// DelRoute deletes the route in the system routing table to a specific destination.
func DelRoute(addr, viaAddr net.IP, tun string) error {
	sub := fmt.Sprintf("-n delete %s %s -ifscope %s", addr.String(), viaAddr.String(), tun)
	args := strings.Split(sub, " ")
	return CommandExec("route", args)
}

// GetNetGateway returns net gateway (default route) and nic.
func GetNetGateway() (gw, dev string, err error) {
	cmd := exec.Command("route", "-n", "get", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	rx := regexp.MustCompile(`(?m)^\W*([^\:]+):\W(.*)$`)
	matches := rx.FindAllSubmatch(output, -1)
	defaultRouteInfo := map[string]string{}
	for _, match := range matches {
		defaultRouteInfo[string(match[1])] = string(match[2])
	}

	_, gatewayExists := defaultRouteInfo["gateway"]
	_, interfaceExists := defaultRouteInfo["interface"]
	if !gatewayExists || !interfaceExists {
		return "", "", fmt.Errorf("[err] GetNetGateway could not read gateway or interface")
	}

	return defaultRouteInfo["gateway"], defaultRouteInfo["interface"], nil
}
