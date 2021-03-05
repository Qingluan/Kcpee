package shadowaead

import (
	"net"
)

type ShadowsocksStream struct {
	conn net.Conn
	w    *writer
	r    *reader
}

func (s *ShadowsocksStream) Read(p []byte) (n int, err error) {
	return s.r.Read(p)
}
