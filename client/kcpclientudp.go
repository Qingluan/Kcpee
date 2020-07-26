package client

import (
	"bytes"
	"errors"
	"net"
	"sync"
)

// func (conn *KcpClientU) handleSocks5Udp(src net.Conn) {

// 	utils.ColorL("socks5 UDP <-- ", src.RemoteAddr())
// 	defer src.Close()
// 	if err := utils.Socks5HandShake(&src); err != nil {
// 		utils.ColorL("socks handshake:", err)
// 		return
// 	}
// 	raw, host, err := utils.GetLocalRequestUDP(&src)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	conn.handleBody(src, host, raw)
// }

const (
	STAGE_INIT     = 0
	STAGE_CONNECT  = 1
	STAGE_DONE     = 2
	STAGE_UNEXISTS = -1
)

type UDpSocks5Conn struct {
	StackSrc map[string]int

	Dst   *net.UDPAddr
	Stage int
	data  []byte
}

var (
	UDPSTACK = &UDpSocks5Conn{
		StackSrc: make(map[string]int),
	}
	Locker = sync.RWMutex{}
)

func GetStage(src *net.UDPAddr) int {

	Stage, ok := UDPSTACK.StackSrc[src.String()]
	if !ok {
		return STAGE_UNEXISTS
	}
	return Stage
}

func UpdateStage(src *net.UDPAddr, stage int) {
	Locker.Lock()
	defer Locker.Unlock()
	if stage == STAGE_DONE {
		delete(UDPSTACK.StackSrc, src.String())
	} else {
		UDPSTACK.StackSrc[src.String()] = stage
	}
}

func HandleBySrc(src *net.UDPAddr, data []byte) (out []byte, err error) {
	Stage := GetStage(src)
	switch Stage {
	case STAGE_UNEXISTS:
		if bytes.HasPrefix(data, []byte{0x05, 0x01, 0x00}) {
			out = []byte{0x5, 0x0}
			defer UpdateStage(src, STAGE_INIT)
		}
		return nil, errors.New("INIT ERROR")
	case STAGE_INIT:

	case STAGE_CONNECT:
	default:
	}
	return

}
