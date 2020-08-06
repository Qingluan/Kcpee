package shadowaead

import (
	"net"
	"sync"
)

type packetConn struct {
	net.PacketConn
	Cipher
	sync.Mutex
	buf []byte // write lock
}

// NewPacketConn wraps a net.PacketConn with cipher
func NewPacketConn(c net.PacketConn, ciph Cipher) net.PacketConn {
	const maxPacketSize = 64 * 1024
	return &packetConn{PacketConn: c, Cipher: ciph, buf: make([]byte, maxPacketSize)}
}
