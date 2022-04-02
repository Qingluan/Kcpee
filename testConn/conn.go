package testConn

import (
	"fmt"
	"net"
	"time"
)

type TestNetConn struct {
	PREDATA  []byte
	RBUF     []byte
	st       time.Time
	UsedTime time.Duration
	ReadTime int
	ReadedNo int
}

func NewTestConn(data []byte, readTime int) *TestNetConn {
	return &TestNetConn{
		PREDATA:  data,
		st:       time.Now(),
		ReadTime: readTime,
	}
}

// func (t *TestNetConn) SetWBuf(buf []byte)
func (t *TestNetConn) Read(buf []byte) (n int, err error) {
	n = len(t.PREDATA)
	copy(buf[:n], t.PREDATA)
	t.st = time.Now()
	t.ReadedNo += 1
	if t.ReadedNo >= t.ReadTime {
		err = fmt.Errorf("readed to end ! exit this conn")
	}
	return
}

func (t *TestNetConn) Write(buf []byte) (n int, err error) {
	// fmt.Println()
	t.UsedTime = time.Now().Sub(t.st)
	return
}

func (t *TestNetConn) Close() (err error) {
	return
}

func (test *TestNetConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (t *TestNetConn) SetWriteDeadline(t2 time.Time) error {
	return nil
}

func (t *TestNetConn) SetDeadline(t2 time.Time) error {
	return nil
}

func (t *TestNetConn) LocalAddr() net.Addr {
	return nil
}

func (t *TestNetConn) RemoteAddr() net.Addr {
	return nil
}
