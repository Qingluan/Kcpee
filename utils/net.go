package utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

const (
	idType  = 0 // address type index
	idIP0   = 1 // ip address start index
	idDmLen = 1 // domain address length index
	idDm0   = 2 // domain address start index

	typeIPv4     = 1 // type is ipv4 address
	typeDm       = 3 // type is domain address
	typeIPv6     = 4 // type is ipv6 address
	typeRedirect = 9

	lenIPv4              = net.IPv4len + 2 // ipv4 + 2port
	lenIPv6              = net.IPv6len + 2 // ipv6 + 2port
	lenDmBase            = 2               // 1addrLen + 2port, plus addrLen
	AddrMask        byte = 0xf
	socksVer5            = 5
	socksCmdConnect      = 1
	socksCmdUdp          = 3
	// lenHmacSha1 = 10
)

var (
	errAddrType      = errors.New("socks addr type not supported")
	errVer           = errors.New("socks version not supported")
	errMethod        = errors.New("socks only support 1 method now")
	errAuthExtraData = errors.New("socks authentication get extra data")
	errReqExtraData  = errors.New("socks request get extra data")
	errCmd           = errors.New("socks command not supported")

	debug DebugLog
	// smuxConfig = smux.DefaultConfig()

)

type CanDeadLine interface {
	SetReadDeadline(t time.Time) error
}

func SetReadTimeout(c *net.Conn) {
	if readTimeout != 0 {
		(*c).SetReadDeadline(time.Now().Add(readTimeout))
	}
}

// func SetRTimeout(c io.Reader) {
// 	if readTimeout != 0 {
// 		c.SetReadDeadline(time.Now().Add(readTimeout))
// 	}
// }
func HostToRaw(host string, port int) (raw []byte) {
	raw = []byte{}
	if port == -1 {
		tmp := strings.SplitN(host, ":", 2)
		host = tmp[0]
		port, _ = strconv.Atoi(tmp[1])
	}
	raw = append(raw, 0x03, byte(len(host)))
	raw = append(raw, []byte(host)...)
	bb := make([]byte, 2)
	binary.BigEndian.PutUint16(bb, uint16(port))
	raw = append(raw, bb...)
	return
}

func SetStreamReadTimeout(c *smux.Stream) {
	if readTimeout != 0 {
		(*c).SetReadDeadline(time.Now().Add(readTimeout))
	}
}

func SetKcpReadTimeout(c *kcp.UDPSession) {
	if readTimeout != 0 {
		(*c).SetReadDeadline(time.Now().Add(readTimeout))
	}
}

// DialWithRawAddr is intended for use by users implementing a local socks proxy.
// rawaddr shoud contain part of the data in socks request, starting from the
// ATYP field. (Refer to rfc1928 for more information.)
func DialWithRawAddr(rawaddr []byte, server string) (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", server)
	if err != nil {
		return
	}

	if _, err = conn.Write(rawaddr); err != nil {
		conn.Close()
		return nil, err
	}
	return
}
func GetLocalRequestUDP(conn *net.Conn) (rawaddr []byte, host string, err error) {
	const (
		idVer   = 0
		idCmd   = 1
		idType  = 3 // address type index
		idIP0   = 4 // ip address start index
		idDmLen = 4 // domain address length index
		idDm0   = 5 // domain address start index

		typeIPv4   = 1 // type is ipv4 address
		typeDm     = 3 // type is domain address
		typeIPv6   = 4 // type is ipv6 address
		typeChange = 5 // type is ss change config

		lenIPv4   = 3 + 1 + net.IPv4len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv4 + 2port
		lenIPv6   = 3 + 1 + net.IPv6len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv6 + 2port
		lenDmBase = 3 + 1 + 1 + 2           // 3 + 1addrType + 1addrLen + 2port, plus addrLen
	)

	// refer to getRequest in server.go for why set buffer size to 263
	buf := make([]byte, 263)
	var n int
	SetReadTimeout(conn)
	// read till we get possible domain length field
	if n, err = io.ReadAtLeast(*conn, buf, idDmLen+1); err != nil {
		return
	}
	// ColorL("->", buf[:10])
	// check version and cmd
	if buf[idVer] != socksVer5 {

		err = errors.New("Sock5 error: " + string(buf[idVer]))
		return
	}
	if buf[idCmd] != socksCmdUdp {
		err = errCmd
		return
	}

	reqLen := -1
	switch buf[idType] {
	case typeIPv4:
		reqLen = lenIPv4
	case typeIPv6:
		reqLen = lenIPv6
	case typeDm:
		reqLen = int(buf[idDmLen]) + lenDmBase
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	case typeChange:
		reqLen = int(buf[idDmLen]) + lenDmBase - 2
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
		// ColorL("hh", host)
	default:
		err = errAddrType
		return
	}
	// ColorL("hq", buf[:10])

	if n == reqLen {
		// common case, do nothing
	} else if n < reqLen { // rare case
		if _, err = io.ReadFull(*conn, buf[n:reqLen]); err != nil {
			return
		}
	} else {
		fmt.Println(n, reqLen, buf)
		err = errReqExtraData
		return
	}

	// rawaddr = buf[idType:reqLen]
	rawaddr = buf[:reqLen]

	// ColorL("hm", buf[:reqLen])

	// debug.Println("addr:", rawaddr)
	if debug {
		switch buf[idType] {
		case typeIPv4:
			host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
		case typeIPv6:
			host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
		case typeDm:
			host = string(buf[idDm0 : idDm0+buf[idDmLen]])
		case typeChange:
			host = string(buf[idDm0 : idDm0+buf[idDmLen]])
			// ColorL("hm", host)

			return
		}
		port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])
		host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	}
	return
}

func GetLocalRequest(conn *net.Conn) (rawaddr []byte, host string, isUdp bool, err error) {
	const (
		idVer   = 0
		idCmd   = 1
		idType  = 3 // address type index
		idIP0   = 4 // ip address start index
		idDmLen = 4 // domain address length index
		idDm0   = 5 // domain address start index

		typeIPv4   = 1 // type is ipv4 address
		typeDm     = 3 // type is domain address
		typeIPv6   = 4 // type is ipv6 address
		typeChange = 5 // type is ss change config

		lenIPv4   = 3 + 1 + net.IPv4len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv4 + 2port
		lenIPv6   = 3 + 1 + net.IPv6len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv6 + 2port
		lenDmBase = 3 + 1 + 1 + 2           // 3 + 1addrType + 1addrLen + 2port, plus addrLen
	)

	// refer to getRequest in server.go for why set buffer size to 263
	buf := make([]byte, 263)
	var n int
	SetReadTimeout(conn)
	// read till we get possible domain length field
	if n, err = io.ReadAtLeast(*conn, buf, idDmLen+1); err != nil {
		return
	}
	// ColorL("->", buf[:10])
	// check version and cmd
	if buf[idVer] != socksVer5 {

		err = errors.New("Sock5 error: " + string(buf[idVer]))
		return
	}
	if buf[idCmd] != socksCmdConnect && buf[idCmd] != socksCmdUdp {
		err = errCmd
		return
	}
	if buf[idCmd] == socksCmdUdp {
		isUdp = true
	}

	reqLen := -1
	switch buf[idType] {
	case typeIPv4:
		reqLen = lenIPv4
	case typeIPv6:
		reqLen = lenIPv6
	case typeDm:
		reqLen = int(buf[idDmLen]) + lenDmBase
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	case typeChange:
		reqLen = int(buf[idDmLen]) + lenDmBase - 2
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
		// ColorL("hh", host)
	default:
		err = errAddrType
		return
	}
	// ColorL("hq", buf[:10])

	if n == reqLen {
		// common case, do nothing
	} else if n < reqLen { // rare case
		if _, err = io.ReadFull(*conn, buf[n:reqLen]); err != nil {
			return
		}
	} else {
		fmt.Println(n, reqLen, buf)
		err = errReqExtraData
		return
	}

	rawaddr = buf[:reqLen]

	// ColorL("hm", buf[:reqLen])

	// debug.Println("addr:", rawaddr)
	if debug {
		switch buf[idType] {
		case typeIPv4:
			host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
		case typeIPv6:
			host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
		case typeDm:
			host = string(buf[idDm0 : idDm0+buf[idDmLen]])
		case typeChange:
			host = string(buf[idDm0 : idDm0+buf[idDmLen]])
			// ColorL("hm", host)

			return
		}
		port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])
		host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	}
	return
}

func GetServerRequest(conn net.Conn) (host string, raw []byte, isUdp bool, err error) {

	// utils.SetStreamReadTimeout(*conn)
	SetReadTimeout(&conn)
	const (
		idVer   = 0
		idCmd   = 1
		idType  = 3 // address type index
		idIP0   = 4 // ip address start index
		idDmLen = 4 // domain address length index
		idDm0   = 5 // domain address start index

		typeIPv4   = 1 // type is ipv4 address
		typeDm     = 3 // type is domain address
		typeIPv6   = 4 // type is ipv6 address
		typeChange = 5 // type is ss change config

		lenIPv4   = 3 + 1 + net.IPv4len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv4 + 2port
		lenIPv6   = 3 + 1 + net.IPv6len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv6 + 2port
		lenDmBase = 3 + 1 + 1 + 2           // 3 + 1addrType + 1addrLen + 2port, plus addrLen
	)
	// buf size should at least have the same size with the largest possible
	// request size (when addrType is 3, domain name has at most 256 bytes)
	// 1(ver) + 1(cmd) + 1(0) + 1(addrType) + 1(lenByte) + 255(max length address) + 2(port) + 10(hmac-sha1)
	// g := color.New(color.FgBlue)
	buf := make([]byte, 269)
	// read till we get possible domain length field

	if _, err = io.ReadFull(conn, buf[:idType+1]); err != nil {

		return
	}

	if buf[idCmd] == socksCmdUdp {
		isUdp = true
	}

	// g.Printf("read %v \n", buf[:20])
	var reqStart, reqEnd int
	addrType := buf[idType]
	switch addrType & AddrMask {
	case typeIPv4:
		reqStart, reqEnd = idIP0, lenIPv4
		raw = buf[:reqEnd]
	case typeIPv6:
		reqStart, reqEnd = idIP0, idIP0+lenIPv6
		raw = buf[:reqEnd]
	case typeDm:
		if _, err = io.ReadFull(conn, buf[idType+1:idDmLen+1]); err != nil {
			return
		}
		reqStart, reqEnd = idDm0, int(buf[idDmLen])+lenDmBase
		raw = buf[:reqEnd]
		// ColorL("Raw:", raw)
	case typeRedirect:

		if _, err = io.ReadFull(conn, buf[idType+1:idDmLen+1]); err != nil {
			return
		}
		// g.Printf("read %v \n", buf[:20])
		reqStart, reqEnd = idDm0, idDm0+int(buf[idDmLen])
		raw = buf[:reqEnd]
	default:
		fmt.Println("Err buf:", buf)
		err = fmt.Errorf("addr type %d not supported:%s", addrType&AddrMask, buf)
		return
	}

	if _, err = io.ReadFull(conn, buf[reqStart:reqEnd]); err != nil {
		// g.Printf("read %v \n", buf[:100])
		return
	}
	// data = buf[:reqEnd]
	// Return string for typeIP is not most efficient, but browsers (Chrome,
	// Safari, Firefox) all seems using typeDm exclusively. So this is not a
	// big problem.
	switch addrType & AddrMask {
	case typeIPv4:
		host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
	case typeIPv6:
		host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idDm0 : idDm0+int(buf[idDmLen])])
	case typeRedirect:
		host = string(buf[idDm0 : idDm0+int(buf[idDmLen])])

		return
	}

	// parse port
	port := binary.BigEndian.Uint16(buf[reqEnd-2 : reqEnd])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	// raw = buf[:reqEnd]
	return
}

func Socks5HandShake(conn *net.Conn) (err error) {
	const (
		idVer     = 0
		idNmethod = 1
	)
	// version identification and method selection message in theory can have
	// at most 256 methods, plus version and nmethod field in total 258 bytes
	// the current rfc defines only 3 authentication methods (plus 2 reserved),
	// so it won't be such long in practice
	SetReadTimeout(conn)
	buf := make([]byte, 258)
	var n int
	if n, err = io.ReadAtLeast(*conn, buf, idNmethod+1); err != nil {
		return
	}
	if buf[idVer] != socksVer5 {
		log.Println(buf)
		return errVer
	}
	nmethod := int(buf[idNmethod])
	msgLen := nmethod + 2
	if n == msgLen { // handshake done, common case
		// do nothing, jump directly to send confirmation
	} else if n < msgLen { // has more methods to read, rare case
		if _, err = io.ReadFull(*conn, buf[n:msgLen]); err != nil {
			return
		}
	} else { // error, should not get extra data
		log.Println(buf)
		return errAuthExtraData
	}
	// send confirmation: version 5, no authentication required
	if _, err = (*conn).Write([]byte{socksVer5, 0}); err != nil {
		return err
	}
	return
}

func ParseUDPSocks5(buf []byte) (host string, rawaddr []byte, err error) {
	const (
		idVer   = 0
		idCmd   = 1
		idType  = 3 // address type index
		idIP0   = 4 // ip address start index
		idDmLen = 4 // domain address length index
		idDm0   = 5 // domain address start index

		typeIPv4   = 1 // type is ipv4 address
		typeDm     = 3 // type is domain address
		typeIPv6   = 4 // type is ipv6 address
		typeChange = 5 // type is ss change config

		lenIPv4   = 3 + 1 + net.IPv4len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv4 + 2port
		lenIPv6   = 3 + 1 + net.IPv6len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv6 + 2port
		lenDmBase = 3 + 1 + 1 + 2           // 3 + 1addrType + 1addrLen + 2port, plus addrLen
	)

	// ColorL("->", buf[:10])
	// check version and cmd
	if buf[idVer] != 5 {

		err = errors.New("Sock5 error: " + string(buf[idVer]))
		return
	}
	if buf[idCmd] != 3 {
		err = errCmd
		return
	}
	reqLen := -1
	switch buf[idType] {
	case typeIPv4:
		reqLen = lenIPv4
	case typeIPv6:
		reqLen = lenIPv6
	case typeDm:
		reqLen = int(buf[idDmLen]) + lenDmBase
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	case typeChange:
		reqLen = int(buf[idDmLen]) + lenDmBase - 2
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
		// ColorL("hh", host)
	default:
		err = errAddrType
		return
	}

	rawaddr = buf[:reqLen]

	// ColorL("hm", buf[:reqLen])

	// debug.Println("addr:", rawaddr)
	if debug {
		switch buf[idType] {
		case typeIPv4:
			host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
		case typeIPv6:
			host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
		case typeDm:
			host = string(buf[idDm0 : idDm0+buf[idDmLen]])
		case typeChange:
			host = string(buf[idDm0 : idDm0+buf[idDmLen]])
			// ColorL("hm", host)

			return
		}
		port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])
		host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	}
	return

}
