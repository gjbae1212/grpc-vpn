package internal

import (
	"fmt"
	"net"
)

const (
	// TunMtuSize is mtu size in TUN.
	TunMtuSize = 1500

	// TunPacketBufferSize is buffer size in TUN.
	TunPacketBufferSize = 4 * 1024

	// TunTxLen is a sending queue size in TUN.
	TunTxLen = 300
)

// IncreaseIP is to increase 1.
func IncreaseIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// GetIPByAddr returns ip.
func GetIPByAddr(addr string) (net.IP, error) {
	addrs, err := net.LookupIP(addr)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("[ERR] not found ip")
	}
	return addrs[seededRand.Int()%len(addrs)], nil
}
