package utils

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

type KcpBase struct {
	kconfig          *KcpConfig
	config           *Config
	smuxConfig       *smux.Config
	Numconn          int
	aliveConn        int
	activateConn     int
	aliveReverseConn int
	aliveRefreshRate time.Duration
	chScavenger      chan *smux.Session
	IsHeartBeat      bool
	Role             string
	Messages         chan string
	muxes            []struct {
		session *smux.Session
		ttl     time.Time
	}
	testmuxes chan struct {
		session *smux.Session
		ttl     time.Time
	}
}

func (kcpBase *KcpBase) AddReverseCon() {
	kcpBase.aliveReverseConn++
}

func (kcpBase *KcpBase) DelReverseCon() {
	kcpBase.aliveReverseConn--
}

func (kcpBase *KcpBase) GetAliveReverseCon() int {
	return kcpBase.aliveReverseConn
}

func (kcpBase *KcpBase) SetRefreshRate(interval int) {
	kcpBase.aliveRefreshRate = time.Duration(interval) * time.Second
}

func (kcpBase *KcpBase) GetRefreshRate() time.Duration {
	if kcpBase.aliveRefreshRate == 0 {
		kcpBase.aliveRefreshRate = 2 * time.Second
	}
	return kcpBase.aliveRefreshRate
}

func (kcpBase *KcpBase) UpdateKcpConfig(kcpconn *kcp.UDPSession) {
	kcpconn.SetStreamMode(true)
	kcpconn.SetWriteDelay(false)
	kcpconn.SetNoDelay(kcpBase.kconfig.NoDelay, kcpBase.kconfig.Interval, kcpBase.kconfig.Resend, kcpBase.kconfig.NoCongestion)
	kcpconn.SetWindowSize(kcpBase.kconfig.SndWnd, kcpBase.kconfig.RcvWnd)
	kcpconn.SetMtu(kcpBase.kconfig.MTU)
	kcpconn.SetACKNoDelay(kcpBase.kconfig.AckNodelay)
}

func (kcpBase *KcpBase) GetSmuxConfig() *smux.Config {
	if kcpBase.smuxConfig == nil {
		kcpBase.smuxConfig = kcpBase.kconfig.GenerateConfig()
	}
	return kcpBase.smuxConfig
}

func (kcpBase *KcpBase) SetTunnelNum(n int) int {
	kcpBase.Numconn = n
	return kcpBase.Numconn
}

func (kcpBase *KcpBase) createConn(config *Config) (session *smux.Session, err error) {
	if config.Method == "tls" {
		tlsConfig, err2 := config.ToTlsConfig()
		if err != nil {
			log.Fatal("Create With Tls Error:", err2)
		}

		// serverString := fmt.Sprintf("%s:%d", config.GetServerArray()[0], config.ServerPort)
		conn, err2 := tlsConfig.WithConn()
		// ColorL("-> create conn in tls:", serverString)
		if err2 != nil {
			log.Println("Create With Tls Error:", err2)
			return nil, err2
		}
		if kcpBase.smuxConfig == nil {
			kcpBase.smuxConfig = kcpBase.kconfig.GenerateConfig()
		}

		if session, err = smux.Client(conn, kcpBase.smuxConfig); err == nil {

			// ColorL("-> create conn in tls session:", session)
			return session, nil

		} else {
			log.Fatal("tls conn -> smux session error:", err)
		}
		return
	}
	block := config.GeneratePassword()
	serverString := fmt.Sprintf("%s:%d", config.GetServerArray()[0], config.ServerPort)
	var kcpconn *kcp.UDPSession
	if kcpconn, err = kcp.DialWithOptions(serverString, block, kcpBase.kconfig.DataShard, kcpBase.kconfig.ParityShard); err == nil {
		kcpBase.UpdateKcpConfig(kcpconn)
		if kcpBase.smuxConfig == nil {
			kcpBase.smuxConfig = kcpBase.kconfig.GenerateConfig()
		}
		if session, err = smux.Client(kcpconn, kcpBase.smuxConfig); err == nil {
			return session, nil
		}

	}
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (kcpBase *KcpBase) Activate(do_some func() error) error {
	kcpBase.activateConn++
	defer func() {
		kcpBase.activateConn--
	}()
	return do_some()
}

func (kcpBase *KcpBase) GetActivateConn() int {
	return kcpBase.activateConn
}

func (kcpBase *KcpBase) WaitConn(config *Config) *smux.Session {
	if config == nil {
		config = kcpBase.config
	}
	for {
		if session, err := kcpBase.createConn(config); err == nil {
			return session
		} else {
			if err.Error() == "listen udp4 :0: socket: too many open files" {

			} else {
				log.Println("re-connecting:", err)
			}

			time.Sleep(time.Second)
		}
	}
}

func (kcpBase *KcpBase) Init(config *Config) {
	// ColorD(kcpBase.kconfig)
	// utils.ColorL(fmt.Sprint("Ds/Ps", kcp), fmt.Sprint("Conn/AutoExpire:", conn.Numconn, conn.AutoExpire))
	// ColorL("Start tunnel:", kcpBase.Numconn, "Config with:", kcpBase.config.Method, kcpBase.config.Password, kcpBase.config.Server.(string), kcpBase.config.ServerPort)
	// ColorL("num:", kcpBase.Numconn)
	kcpBase.Messages = make(chan string, 2)
	numconn := uint16(kcpBase.Numconn)
	if kcpBase.muxes == nil {
		kcpBase.muxes = make([]struct {
			session *smux.Session
			ttl     time.Time
		}, numconn)
		kcpBase.testmuxes = make(chan struct {
			session *smux.Session
			ttl     time.Time
		}, 5)
	}

	if config == nil {
		config = kcpBase.config
	}
	if config.Method != "tls" {
		for k := range kcpBase.muxes {
			kcpBase.muxes[k].session = kcpBase.WaitConn(config)
			kcpBase.muxes[k].ttl = time.Now().Add(time.Duration(kcpBase.kconfig.AutoExpire) * time.Second)

		}
	}

	if kcpBase.chScavenger == nil {
		kcpBase.chScavenger = make(chan *smux.Session, 256)
	}
	go scavenger(kcpBase.chScavenger, kcpBase.kconfig.ScavengeTTL)
}

func (kcpBase *KcpBase) SendMsg(pre string, content string) {
	if strings.HasPrefix(pre, ":") {
		ColorL("xxx", "pre can not include ':' !!!")
		return
	}
	kcpBase.Messages <- fmt.Sprintf("%s:%s", pre, content)
}

func (kcpBase *KcpBase) RecvMsg(filter ...string) (pre string, content string) {
	select {
	case msg := <-kcpBase.Messages:
		parts := strings.SplitN(msg, ":", 2)
		pre = parts[0]
		if len(filter) > 0 {
			if filter[0] == pre {
				content = parts[1]
			} else {
				kcpBase.Messages <- msg
			}
		}
		content = parts[1]
	default:
		time.Sleep(kcpBase.GetRefreshRate())
		content = ""
	}

	return
}

func (kcpBase *KcpBase) HeartBeatC(stream net.Conn) {
	kcpBase.IsHeartBeat = true
	defer func() {
		stream.Close()
		ColorL("~ Dead because hearbeat is stop")
		os.Exit(0)
	}()
	for {
		now := time.Now()
		buf := make([]byte, 123)
		if !kcpBase.IsHeartBeat {
			ColorL("stop heart ", "Client")
			break
		}
		if n, err := stream.Read(buf); err == nil {
			// time.Sleep(1 * time.Second)
			if string(buf[:n]) == "[BOARING]" {
				if kcpBase.activateConn <= 1 {
					time.Sleep(3 * time.Second)
				}
				if _, err := stream.Write([]byte("[ME TOO]")); err != nil {
					break
				}
				fmt.Print(FGCOLORS[2](kcpBase.activateConn), FGCOLORS[1](now.Format(time.UnixDate))+"\r")

			} else {
				if kcpBase.activateConn <= 1 {
					time.Sleep(kcpBase.GetRefreshRate())
				} else {
					time.Sleep(1 * time.Second)
				}

				kcpBase.SendMsg(stream.RemoteAddr().String(), string(buf[:n]))
				if _, err := stream.Write([]byte("OK")); err != nil {
					break
				}
				ColorL(now.Format(time.UnixDate), string(buf[:n]))
			}
		} else {
			ColorL("Heartbeat Err", err)
			kcpBase.IsHeartBeat = false
			break
		}
	}
	// kcpBase.IsHeartBeat = false
}

func (kcpBase *KcpBase) HeartBeatS(stream net.Conn, fromHost string, clearAfter func(h string)) {
	nowT := 0
	kcpBase.IsHeartBeat = true
	used := false
	defer stream.Close()
	for {
		now := time.Now()
		buf := make([]byte, 1024)
		if !kcpBase.IsHeartBeat {
			ColorL("stop heart", "server")
			break
		}
		if pre, msg := kcpBase.RecvMsg(); msg == "" {
			if _, err := stream.Write([]byte("[BOARING]")); err == nil {
				if n, err := stream.Read(buf); err == nil {
					if kcpBase.activateConn > 1 {
						used = true
					}
					if string(buf[:n]) == "[ME TOO]" {
						fmt.Print(FGCOLORS[1](now.Format(time.UnixDate))+" T: ", kcpBase.GetAliveReverseCon(), "...\r")
					} else {
						ColorL("close no me too", "error when write boaring!!")
						break
					}
				} else {
					ColorL("close by remote", "so i closed too")
					break
				}
			} else {
				ColorL("close some error", err)
				break
			}
		} else if msg == "[[EOF]]" && pre == "HEART" {
			ColorL("*", "*", "stop heart beet")
			break
		} else {
			if msg == "more" {
				if _, err := stream.Write([]byte("9")); err == nil {

					n, err := stream.Read(buf)
					if err == nil {
						ColorL("reply msg:", string(buf[:n]))
						if string(buf[:n]) == "OK" {
							nowT += 9
						}
					} else {
						ColorL("close by remote", "so i closed too", err)
						break
					}
					// kcpBase.SendMsg(fromHost, "OK")
				} else {
					ColorL("close some error", err)
					break
				}
			}
		}

	}
	if used {
		clearAfter(fromHost)
	}
	// kcpBase.IsHeartBeat = false

}

func (kcpBase *KcpBase) UpdateConfig(uri string) {
	parseURI(uri, kcpBase.config)
}

func (kcpBase *KcpBase) WithTestSession(config *Config, howTo func(sess *smux.Session)) {
	ss := new(struct {
		session *smux.Session
		ttl     time.Time
	})

	ss.session = kcpBase.WaitConn(config)
	ss.ttl = time.Now().Add(time.Duration(kcpBase.kconfig.AutoExpire) * time.Second)
	kcpBase.testmuxes <- *ss
	howTo(ss.session)
	<-kcpBase.testmuxes

}

func (kcpBase *KcpBase) WithSession(config *Config, id ...uint16) (session *smux.Session) {
	var idx uint16
	kcpBase.aliveConn++
	if len(id) > 0 {
		idx = id[0] % uint16(kcpBase.Numconn)
	}
	if config == nil {
		config = kcpBase.config
	}

	if kcpBase.muxes == nil {
		kcpBase.Init(config)
	}
	// Closed / Timeout / Session host incorrect will reconnect
	if kcpBase.muxes[idx].session == nil {
		kcpBase.chScavenger <- kcpBase.muxes[idx].session
		kcpBase.muxes[idx].session = kcpBase.WaitConn(config)
		kcpBase.muxes[idx].ttl = time.Now().Add(time.Duration(kcpBase.kconfig.AutoExpire) * time.Second)
	} else {
		if kcpBase.muxes[idx].session.IsClosed() || (kcpBase.kconfig.AutoExpire > 0 && time.Now().After(kcpBase.muxes[idx].ttl)) {
			kcpBase.chScavenger <- kcpBase.muxes[idx].session
			kcpBase.muxes[idx].session = kcpBase.WaitConn(config)
			kcpBase.muxes[idx].ttl = time.Now().Add(time.Duration(kcpBase.kconfig.AutoExpire) * time.Second)
		} else if kcpBase.muxes[idx].session.RemoteAddr().String() != config.Server.(string) {
			sss := kcpBase.WaitConn(config)
			kcpBase.chScavenger <- sss
			return sss
		}
	}

	session = kcpBase.muxes[idx].session
	return
}

func (kcpBase *KcpBase) GetSession(idx uint16) (session *smux.Session) {
	if !kcpBase.muxes[idx].session.IsClosed() {
		session = kcpBase.muxes[idx].session
	}
	return
}

func (kcpBase *KcpBase) SetConfig(config *Config) {
	kcpBase.config = config
}

func (kcpBase *KcpBase) SetKcpConfig(config *KcpConfig) {
	kcpBase.kconfig = config
}

func (kcpBase *KcpBase) GetConfig() *Config {
	return kcpBase.config
}

func (kcpBase *KcpBase) GetKcpConfig() *KcpConfig {
	return kcpBase.kconfig
}

func (kcpBase *KcpBase) GetAliveNum() int {
	return kcpBase.aliveConn
}

// func (kcpBase *KcpBase) PipeTest(p1, p2 net.Conn) (p1data, p2data []byte) {
// 	p1.

// }

func (kcpBase *KcpBase) Pipe(p1, p2 net.Conn) {
	// start tunnel & wait for tunnel termination
	// p1.SetWriteDeadline(5 * time.Second)
	// p2.SetWriteDeadline(5 * time.Second)
	streamCopy := func(dst io.Writer, src io.ReadCloser, fr, to net.Addr) {
		// startAt := time.Now()
		Copy(dst, src)

		// if err != nil {
		// 	r := color.New(color.FgRed)
		// 	r.Println("error : ", err)
		// }
		// speedF := float64(n) / time.Since(startAt).Seconds()

		// if kcpBase.Role == "client" {
		// go SendSpeedMsg(p2.RemoteAddr().String(), n, speedF)
		// } else if kcpBase.Role == "tester" {

		// } else {
		// endAt := time.Now().Sub(startAt)
		// chn := float32(n) / 1024.0
		// if n == 0 && err != nil {
		// 	log.Print(err.Error() + "\r")
		// }
		// ColorL(fmt.Sprint("alive: ", kcpBase.aliveConn), fmt.Sprint("passed: ", FGCOLORS[1](chn))+"Kb", fmt.Sprint(FGCOLORS[0](p1.RemoteAddr()), "->", FGCOLORS[0](p2.RemoteAddr())), fmt.Sprint("Used:", endAt), "\r")
		// }

		p1.Close()
		p2.Close()
		// }()
	}
	go streamCopy(p1, p2, p2.RemoteAddr(), p1.RemoteAddr())
	streamCopy(p2, p1, p1.RemoteAddr(), p2.RemoteAddr())
	kcpBase.aliveConn--
}

type scavengeSession struct {
	session *smux.Session
	ts      time.Time
}

func scavenger(ch chan *smux.Session, ttl int) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var sessionList []scavengeSession
	for {
		select {
		case sess := <-ch:
			sessionList = append(sessionList, scavengeSession{sess, time.Now()})
			// log.Println("session marked as expired")
			// log.Println("session marked as expired", sess.RemoteAddr())
		case <-ticker.C:
			var newList []scavengeSession
			for k := range sessionList {
				s := sessionList[k]
				if s.session == nil {
					continue
				}
				if s.session.NumStreams() == 0 || s.session.IsClosed() {
					// log.Println("session normally closed", s.session.RemoteAddr())
					// log.Println("session normally closed")
					s.session.Close()
				} else if ttl >= 0 && time.Since(s.ts) >= time.Duration(ttl)*time.Second {
					// log.Println("session reached scavenge ttl", s.session.RemoteAddr())
					// log.Println("session reached scavenge ttl")
					s.session.Close()
				} else {
					newList = append(newList, sessionList[k])
				}
			}
			sessionList = newList
		}
	}
}
