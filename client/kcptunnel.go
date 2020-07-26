package client

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Qingluan/Kcpee/utils"
	"github.com/xtaci/smux"
)

type KcpTunnel struct {
	utils.KcpBase
	redirectHost string
	alive        int
	tunMode      string
}

func NewKcpTunnel(config *utils.Config, kconfig *utils.KcpConfig) (kclient *KcpTunnel) {
	kclient = new(KcpTunnel)
	kclient.Numconn = 20
	kclient.SetConfig(config)
	kclient.SetKcpConfig(kconfig)
	return
}

func (tun *KcpTunnel) Connect(listenAddr, host string) {
	if strings.HasSuffix(host, "/cmd") {
		host = strings.SplitN(host, "/cmd", 2)[0]
		tun.SetTunMode("bash")
		if err := tun.ConnectCmd(host); err != nil {
			log.Fatal("cmdCmd:", err)
		}
		return
	} else {
		log.Println("<reverse proxy><", host, ">")
	}
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal("tun listen error:", err)
	}
	tun.Init(nil)
	rr := uint16(0)
	for {
		p1, err := ln.Accept()
		if err != nil {
			log.Fatalf("%+v", err)
		}

		session := tun.WithSession(nil, rr)
		if p2, err := session.OpenStream(); err == nil {
			// d := NewCmdStruct("redirect://cc://" + host).ToData().Bytes()
			// fmt.Println("Err Test:", d)
			for {
				_, err = p2.Write(NewCmdStruct("redirect://cc://" + host).ToData().Bytes())
				if err == nil {
					break
				}
				utils.ColorL("remote error , try again", time.Now(), "\r")
			}

			if tun.tunMode == "map" {
				if err := utils.Socks5HandShake(&p1); err != nil {
					utils.ColorL("socks handshake:", err)
					return
				}
				_, err = p1.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x43})
				if err != nil {
					utils.ColorL("Err socks5:", err)
				}
				if raw, _, isUdp, err := utils.GetLocalRequest(&p1); err == nil {
					if isUdp {
						// utils.ColorL()
					}
					p2.Write(raw)
				}
			}
			go utils.Pipe(p2, p1)
		}
		rr++
	}
}

// func (tun *KcpTunnel) mapConnect(listenAddr, host string){

// 	defer p1.Close()
// 	if err := utils.Socks5HandShake(&p1); err != nil {
// 		utils.ColorL("socks handshake:", err)
// 		return
// 	}

// 	raw, host, err := utils.GetLocalRequest(&p1)
// }

func (tun *KcpTunnel) TryPadding(host string) {
	i := 0
	st := time.Now()
	go func() {
		for {
			if time.Now().Sub(st) > time.Second*5 {
				log.Fatal("test End")
			}
			time.Sleep(1 * time.Second)
		}
	}()
	for {
		tun.ConnectCmd(host)
		i++
		if i >= 30 {
			break
		}
	}
	utils.ColorL("i try to boom ", 30, " times .... , now try again")
}

func (tun *KcpTunnel) ConnectCmdInit(host string) (p1 Stdio, p2 net.Conn, err error) {
	rr := uint16(rand.Uint32())
	p1 = NewStdio(host)
	session := tun.WithSession(nil, rr)
	if p2, err = session.OpenStream(); err == nil {
		p2.Write(NewCmdStruct("redirect://cc://" + host).ToData().Bytes())
		if tun.tunMode == "bash" {
			if err != nil {
				utils.ColorL("Err stdio:", err)
			}
			raw := NewCmdStruct("connect://cmd").ToData().Bytes()
			if _, err := p2.Write(raw); err != nil {
				return p1, p2, err
			}
		}
	} else {
		log.Fatal("may server down?")
	}
	return
}

func (tun *KcpTunnel) ConnectCmd(host string) error {
	p1, p2, err := tun.ConnectCmdInit(host)
	if err != nil {
		return err
	}
	_, err = p1.Write([]byte("Sio  ->  " + host + "\n"))
	err = utils.Pipe(p2, &p1)
	if err != nil && err.Error() != "EOF" {
		log.Fatal("Pipe: ", err)
	}
	if p1.ci == 0 {
		tun.TryPadding(host)
		log.Fatal("~~ bye~~")
	}
	return nil
}

func (kcpBase *KcpTunnel) HeartBeatC(stream net.Conn) {
	kcpBase.IsHeartBeat = true
	defer func() {
		stream.Close()
		utils.ColorL("~ Dead because hearbeat is stop")
		os.Exit(0)
	}()
	for {
		now := time.Now()
		buf := make([]byte, 123)
		if !kcpBase.IsHeartBeat {
			utils.ColorL("stop heart ", "Client")
			break
		}
		if n, err := stream.Read(buf); err == nil {
			// time.Sleep(1 * time.Second)
			if string(buf[:n]) == "[BOARING]" {
				if kcpBase.alive <= 1 {
					time.Sleep(5 * time.Second)
				}
				if _, err := stream.Write([]byte("[ME TOO]")); err != nil {
					break
				}
				fmt.Print(utils.FGCOLORS[2](kcpBase.alive), utils.FGCOLORS[1](now.Format(time.UnixDate))+"\r")

			} else {
				if kcpBase.alive <= 1 {
					time.Sleep(1 * time.Second)
				} else {
					time.Sleep(1 * time.Second)
				}

				kcpBase.SendMsg(stream.RemoteAddr().String(), string(buf[:n]))
				if _, err := stream.Write([]byte("OK")); err != nil {
					break
				}
				utils.ColorL(now.Format(time.UnixDate), string(buf[:n]))
			}
		} else {
			kcpBase.IsHeartBeat = false
			break
		}
	}
	// kcpBase.IsHeartBeat = false
}

func (tun *KcpTunnel) Tunnel(host string) {
	tun.SetRedirectHost(host)
	// tun.SetTunnelNum(10)
	var rr uint16 = 1
	if session := tun.WithSession(nil, 0); session != nil {
		if p2, err := session.OpenStream(); err == nil {
			p2.Write(NewCmdStruct("redirect://TUNNEL_INIT").ToData().Bytes())
			buf := make([]byte, 256)
			if n, err := p2.Read(buf); err == nil {
				fmt.Println(utils.FGCOLORS[1]("global address [ ", string(buf[:n]), " ]"))
			} else {
				return
			}
			go tun.HeartBeatC(p2)
			// p2.Close()
		}
	}

	time.Sleep(2 * time.Second)
	// firstBuf := make([]byte, 1024)
	// maxLimit := make(chan int, 50)
	for {
		if rr == 0 {
			rr++
			continue
		}

		if _, content := tun.RecvMsg(); content == "" {
			if tun.alive > 7 {
				if tun.GetActivateConn() > 1 {
					time.Sleep(100 * time.Microsecond)
				} else {
					time.Sleep(1 * time.Second)
				}
				rr++
				continue
			}

		} else {
			utils.ColorL("conetnt:", content, tun.alive)
		}
		// for {
		for no := 0; no < 2; no++ {
			go tun.TunnelOne(host)
			// rr++
			// rr %= uint16(tun.Numconn)
			// go func() {
			// 	maxLimit <- 1
			// 	session := tun.WithSession(nil, rr)
			// 	firstBuf := make([]byte, 1024)
			// 	if p2, err := session.OpenStream(); err == nil {
			// 		utils.ColorL("more tunnel", p2.ID(), "\r")
			// 		p2.Write(NewCmdStruct("TUNNEL_CONNECT").ToData().Bytes())
			// 		if tun.tunMode == "map" {
			// 			tun.handle(p2, []byte{}, host)
			// 		} else {
			// 			if n, err := p2.Read(firstBuf); err != nil {
			// 				utils.ColorL("error in first!", err)
			// 				<-maxLimit
			// 				return
			// 			} else {
			// 				tun.handle(p2, firstBuf[:n], host)
			// 			}
			// 		}

			// 	}
			// 	<-maxLimit
			// }()
		}

	}
}

func (tun *KcpTunnel) TunnelOne(host string) {
	rr := uint16(tun.alive)
	tun.alive++
	defer func() {
		tun.alive--
	}()
	// maxLimit <- 1
	session := tun.WithSession(nil, rr)
	firstBuf := make([]byte, 1024)
	if p2, err := session.OpenStream(); err == nil {
		utils.ColorL("more tunnel", p2.ID(), "\r")
		p2.Write(NewCmdStruct("TUNNEL_CONNECT").ToData().Bytes())
		if tun.tunMode == "map" {
			tun.handle(p2, []byte{}, host)
		} else {
			if n, err := p2.Read(firstBuf); err != nil {
				utils.ColorL("error in first!", err)
				// <-maxLimit
				return
			} else {
				tun.handle(p2, firstBuf[:n], host)
			}
		}

	}

	// <-maxLimit
}

func (tun *KcpTunnel) SetRedirectHost(host string) {
	tun.redirectHost = host
}

func (tun *KcpTunnel) handle(p1 *smux.Stream, firstData []byte, hostBackup string) {
	tun.alive++
	if tun.tunMode != "map" {
		if data := string(firstData); strings.HasPrefix(data, "connect://") {
			hostOrCmd := strings.Split(data, "connect://")[1]
			if strings.Contains(hostOrCmd, ":") {
				defer p1.Close()
				tun.SetRedirectHost(strings.TrimSpace(hostOrCmd))
				p1.Write([]byte(fmt.Sprintf("connect to => %s", hostOrCmd)))
			} else if hostOrCmd == "cmd" {
				// redirect as bash
				NewCmdRunner(p1)
			}
		} else {
			tun.handleStream(p1, firstData)
		}
	} else {
		tun.handleStream(p1, firstData)
	}

	tun.alive--

}

func GetUDPCon(host string) (con *net.UDPConn, err error) {
	ts := strings.SplitN(host, ":", 2)
	port, _ := strconv.Atoi(ts[1])
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: net.ParseIP(ts[0]), Port: port}
	con, err = net.DialUDP("udp", srcAddr, dstAddr)
	return
}

func (tun *KcpTunnel) handleStream(p1 *smux.Stream, firstData []byte) (err error) {
	if tun.tunMode == "map" {
		if host, _, isudp, err := utils.GetServerRequest(p1); err == nil {
			utils.ColorL("First:", host)
			if strings.HasPrefix(host, "ben:") {
				host = "localhost:" + strings.SplitN(host, ":", 2)[1]
			} else if strings.HasPrefix(host, "99.99.99.99:") {
				host = "localhost:" + strings.SplitN(host, ":", 2)[1]
			} else if strings.HasPrefix(host, "connect://cmd") {
				NewCmdRunner(p1)
				return nil
			}
			if err := tun.Activate(func() error {
				if isudp {

					p2, err := GetUDPCon(host)
					if err != nil {
						if strings.Contains(err.Error(), "use of closed udp network connection") {
						} else if strings.Contains(err.Error(), "EOF") {
						} else {
							utils.ColorL("err", "udp", err)
						}
						p1.Close()
						return err
					}
					utils.Pipe(p1, p2)
					return nil
				} else {
					p2, err := net.Dial("tcp", host)
					if err != nil {
						if strings.Contains(err.Error(), "use of closed network connection") {
						} else if strings.Contains(err.Error(), "EOF") {
						} else {
							utils.ColorL("err", "tcp", err)
						}
						p1.Close()
						return err
					}
					utils.Pipe(p1, p2)
					return nil
				}

			}); err != nil {
				return nil
			}

			// if _, err := p2.Write(firstData); err == nil {
			// }
		}
	} else {
		if err := tun.Activate(func() error {
			utils.ColorL("Connect", tun.redirectHost)
			p2, err := net.Dial("tcp", tun.redirectHost)
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
				} else if strings.Contains(err.Error(), "EOF") {
				} else {
					utils.ColorL("err", "tcp", err)
				}

			}
			utils.ColorL("tunnel://", tun.redirectHost)
			if _, err := p2.Write(firstData); err == nil {
				utils.Pipe(p1, p2)
			}
			return nil
		}); err != nil {
			return err
		}

	}

	return
}

func (tun *KcpTunnel) SetTunMode(mode string) {
	tun.tunMode = mode
}
