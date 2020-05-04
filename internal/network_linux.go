package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// SetTunStatus is to up or down network device for TUN.
func SetTunStatus(tun string, up bool) error {
	status := "down"
	if up {
		status = "up"
	}
	sub := fmt.Sprintf("link set dev %s %s mtu %d qlen %d", tun, status, TunMtuSize, TunTxLen)
	args := strings.Split(sub, " ")
	return CommandExec("ip", args)
}

// SetTunIP sets the local IP address of a network interface.
func SetTunIP(tun string, localAddr net.IP, addr *net.IPNet) error {
	sub := fmt.Sprintf("%s %s netmask %s", tun, localAddr.String(), net.IP(addr.Mask).String())
	args := strings.Split(sub, " ")
	return CommandExec("ifconfig", args)
}

// SetDefaultGateway sets the systems gateway to the IP / device specified.
func SetDefaultGateway(gw, tun string) error {
	sub := fmt.Sprintf("add default gw %s dev %s", gw, tun)
	args := strings.Split(sub, " ")
	return CommandExec("route", args)
}

// SetPacketForward sets ip packet forward.
func SetPacketForward(ok bool) error {
	sub := fmt.Sprintf("net.ipv4.ip_forward=1")
	if !ok {
		sub = fmt.Sprintf("net.ipv4.ip_forward=0")
	}
	return CommandExec("sysctl", []string{sub})
}

// SetPostRoutingMasquerade sets outbound packets masquerade.
func SetPostRoutingMasquerade(ok bool) error {
	var sub string
	if ok {
		sub = "-t nat -A POSTROUTING -j MASQUERADE"
	} else {
		sub = "-t nat -D POSTROUTING 1"
	}
	args := strings.Split(sub, " ")
	return CommandExec("iptables", args)
}

// SetGoogleDNS sets google dns
func SetGoogleDNS() error {
	// TODO: Don't support
	return nil
}

// SetDeleteDNS sets google dns
func SetDeleteDNS() error {
	// TODO: Don't support
	return nil
}

// AddRoute routes all traffic for addr via interface tunName.
func AddRoute(addr, viaAddr net.IP, tun string) error {
	sub := fmt.Sprintf("add %s gw %s dev %s", addr.String(), viaAddr.String(), tun)
	args := strings.Split(sub, " ")
	return CommandExec("route", args)
}

// DelRoute deletes the route in the system routing table to a specific destination.
func DelRoute(addr, viaAddr net.IP, tun string) error {
	sub := fmt.Sprintf("del %s gw %s dev %s", addr.String(), viaAddr.String(), tun)
	args := strings.Split(sub, " ")
	return CommandExec("route", args)
}

// GetNetGateway return net gateway (default route) and nic.
func GetNetGateway() (gw, dev string, err error) {
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return "", "", err
	}

	defer file.Close()
	rd := bufio.NewReader(file)

	s2byte := func(s string) byte {
		b, _ := strconv.ParseUint(s, 16, 8)
		return byte(b)
	}

	for {
		line, isPrefix, err := rd.ReadLine()

		if err != nil {
			return "", "", fmt.Errorf("[err] GetNetGateway %w", err)
		}
		if isPrefix {
			return "", "", fmt.Errorf("[err] GetNetGateway Parse error: Line too long")
		}
		buf := bytes.NewBuffer(line)
		scanner := bufio.NewScanner(buf)
		scanner.Split(bufio.ScanWords)
		tokens := make([]string, 0, 8)

		for scanner.Scan() {
			tokens = append(tokens, scanner.Text())
		}

		iface := tokens[0]
		dest := tokens[1]
		gw := tokens[2]
		mask := tokens[7]

		if bytes.Equal([]byte(dest), []byte("00000000")) &&
			bytes.Equal([]byte(mask), []byte("00000000")) {
			a := s2byte(gw[6:8])
			b := s2byte(gw[4:6])
			c := s2byte(gw[2:4])
			d := s2byte(gw[0:2])

			ip := net.IPv4(a, b, c, d)

			return ip.String(), iface, nil
		}

	}
}
