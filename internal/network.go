package internal

import "net"

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
