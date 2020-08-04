package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Qingluan/Kcpee/utils"
	"github.com/fatih/color"

	"github.com/xtaci/smux"
	// "github.com/cs8425/smux"
)

const (
	SINGLE_MODE = 0
	AUTO_MODE   = 1
	FLOW_MODE   = 20
	SOCKS5TYPE  = 0
	HTTPTYPE    = 1
)

var (
	waitChannel = make(chan int, 10)
	TO_STOP     = false
)

// KcpClient for client and cmd
type KcpClient struct {
	utils.KcpBase
	useAutoMap      bool
	routeMode       int
	listenAddr      string
	clientProxyType int
	RRR             uint16
}

type waitTest struct {
	Url string
	i   uint16
}

// NewKcpClient init KcpClient
func NewKcpClient(config *utils.Config, kconfig *utils.KcpConfig) (kclient *KcpClient) {
	kclient = new(KcpClient)
	kclient.Numconn = 512
	kclient.SetConfig(config)
	kclient.SetMode(AUTO_MODE)
	kclient.SetKcpConfig(kconfig)
	return
}

// Set client Proxy type as http/socks5
func (conn *KcpClient) SetProxyType(tp int) {
	conn.clientProxyType = tp
}

// Listen for socks5
func (conn *KcpClient) Listen(listenAddr string, ifStartUdpListener ...bool) (err error) {
	if ifStartUdpListener != nil && ifStartUdpListener[0] {
		go conn.ListenUDP(listenAddr)
	}
	ln, err := net.Listen("tcp", listenAddr)
	utils.ColorL("Local Listen:", listenAddr)
	conn.listenAddr = listenAddr
	if err != nil {
		log.Fatal("listen error:", err)
	}
	// rr := uint16(0)
	if len(utils.AutoMap) > 0 {
		conn.useAutoMap = true
	}
	// acceptnum := 0
	if conn.Role == "client" {
		go utils.SpeedShow()
	}
	conn.ShowConfig()

	for {
		if TO_STOP {
			break
		}
		// if conn.Role == "tester" && conn.GetAliveNum() > conn.Numconn {
		// 	time.Sleep(10 * time.Millisecond)
		// 	continue
		// }
		p1, err := ln.Accept()

		if err != nil {
			if !strings.Contains(err.Error(), "too many open files") {
				log.Println("%+v", err)
			}

			continue
		}
		go conn.handleSocks5TcpAndUDP(p1)

	}
	return
}

func (conn *KcpClient) ListenUDP(listenAddr string) (err error) {
	ts := strings.SplitN(listenAddr, ":", 2)
	port, _ := strconv.Atoi(ts[1])
	port = port << 2
	conn.listenAddr = listenAddr
	utils.ColorL("Local Listen udp: ", port)
	// rr := uint16(0)
	if len(utils.AutoMap) > 0 {
		conn.useAutoMap = true
	}
	acceptnum := 0
	if conn.Role == "client" {
		go utils.SpeedShow()
	}
	for {
		if TO_STOP {
			break
		}
		if conn.Role == "tester" && conn.GetAliveNum() > conn.Numconn {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port}
		p1, err := net.ListenUDP("udp", addr)
		// p1.ReadFromUDP()
		if err != nil {
			log.Fatal("listen error:", err)
		}

		acceptnum++
		if acceptnum%20 == 0 {
			// utils.ColorL("accept tcp:", acceptnum)
		}

		if err != nil {
			if !strings.Contains(err.Error(), "too many open files") {
				log.Println(err)
			}
			continue
		}
		conn.handleSocks5TcpAndUDP(p1)

	}
	return
}

func (kclient *KcpClient) SetMode(mode int) {
	kclient.routeMode = mode
}

func (kclient *KcpClient) handleSocks5TcpAndUDP(p1 net.Conn) {
	defer p1.Close()
	if err := utils.Socks5HandShake(&p1); err != nil {
		utils.ColorL("socks handshake:", err)
		return
	}

	raw, host, _, err := utils.GetLocalRequest(&p1)
	if err != nil {
		fmt.Println(err)
	}
	// if isUdp {

	// 	utils.ColorL("socks5 UDP-->", host)
	// } else {

	// 	utils.ColorL("socks5 -->", host)
	// }
	if err != nil {
		log.Println("error getting request:", err)
		return
	}
	kclient.handleBody(p1, host, raw)
}

func (kclient *KcpClient) handleBodyDo(p1 net.Conn, host string, do func(p1, p2 net.Conn, raw []byte)) {
	defer func() {
		p1.Close()
		kclient.RRR++
		kclient.RRR %= uint16(kclient.Numconn)
	}()
	rr := kclient.RRR
	if strings.HasPrefix(host, "ss://") {
		if kclient.listenAddr != utils.TestProxyAddr {
			utils.ColorL("set new :", host)
		}
		kclient.handleSS(host, p1)
		return
	}
	raw := utils.HostToRaw(host, -1)

	// do auto expiration && rekclientection
	config := kclient.GetConfig()
	// utils.ColorL(config.Method, "mode:", kclient.routeMode)
	if kclient.routeMode == AUTO_MODE {
		if kclient.useAutoMap && kclient.listenAddr != utils.TestProxyAddr {
			// utils.ColorL("try use", host, fmt.Sprintf("(%s)", utils.GetMainDomain(host)))
			if v, ok := utils.AutoMap[utils.GetMainDomain(host)]; ok {
				config = utils.BOOK.Get(v[0].Server)
			} else {
				config = utils.BOOK.RandGet()
				// to test in background
				if kclient.listenAddr != utils.TestProxyAddr {
					go kclient.testURL(host)
				}
			}
		}
	} else if kclient.routeMode == FLOW_MODE {
		config = utils.BOOK.FlowGet()
	} else if kclient.routeMode == SINGLE_MODE {
		config = kclient.GetConfig()
	} else {
		config = kclient.GetConfig()
	}
	if config == nil {
		log.Fatal("nil : no config")
		// config = kclient.GetConfig()
	}

	if config.Method != "tls" {
		session := kclient.WithSession(config, rr)
		p2, err := session.OpenStream()
		if err != nil {
			return
		}
		defer p2.Close()
		if do == nil {
			log.Fatal("No handler to deal with two net.Conn")
		}

		do(p2, p1, raw)
		// utils.ColorL("Raw:", raw)
		// kclient.handleClient(session, p1, false, raw)
	} else {
		tconfig, err := config.ToTlsConfig()
		if err != nil {
			log.Fatal("create tls kclient error:", err)
		}
		ptls, err := tconfig.WithConn()
		defer ptls.Close()

		if err != nil {
			log.Fatal("create tls kclient error:", err)
		}
		if do == nil {
			log.Fatal("No handler to deal with two net.Conn")
		}

		do(ptls, p1, raw)
		// kclient.handleClientCon(ptls, p1, false, raw)
	}
}

func (kclient *KcpClient) handleBody(p1 net.Conn, host string, raw []byte) {
	defer func() {
		kclient.RRR++
		kclient.RRR %= uint16(kclient.Numconn)
	}()
	rr := kclient.RRR
	if strings.HasPrefix(host, "ss://") {
		if kclient.listenAddr != utils.TestProxyAddr {
			utils.ColorL("set new :", host)
		}
		kclient.handleSS(host, p1)
		return
	}
	if raw == nil {
		raw = utils.HostToRaw(host, -1)
	}
	// do auto expiration && rekclientection
	config := kclient.GetConfig()

	if kclient.routeMode == AUTO_MODE {
		utils.ColorL(config.Method, "mode:", "Auto", host)
		if kclient.useAutoMap && kclient.listenAddr != utils.TestProxyAddr {
			// utils.ColorL("try use", host, fmt.Sprintf("(%s)", utils.GetMainDomain(host)))
			if v, ok := utils.AutoMap[utils.GetMainDomain(host)]; ok {
				config = utils.BOOK.Get(v[0].Server)
			} else {
				config = utils.BOOK.RandGet()
				// to test in background
				if kclient.listenAddr != utils.TestProxyAddr {
					go kclient.testURL(host)
				}
			}
		}
	} else if kclient.routeMode == FLOW_MODE {

		utils.ColorL(config.Method, "mode:", "Flow", host)
		config = utils.BOOK.FlowGet()
	} else if kclient.routeMode == SINGLE_MODE {

		utils.ColorL(config.Method, "mode:", "Single", host)
		config = kclient.GetConfig()
	} else {

		utils.ColorL(config.Method, "mode:", "Default", host)
		config = kclient.GetConfig()
	}
	if config == nil {
		log.Fatal("nil : no config")
		// config = kclient.GetConfig()
	}
	if config.Method != "tls" {

		// utils.ColorL("Stream", "....")
		session := kclient.WithSession(config, rr)
		// utils.ColorL("Raw:", raw)

		kclient.handleClient(session, p1, false, raw)
	} else {
		tconfig, err := config.ToTlsConfig()

		if err != nil {
			log.Fatal("create tls kclient error:", err)
		}
		ptls, err := tconfig.WithConn()
		if err != nil {
			log.Fatal("create tls kclient error:", err)
		}
		kclient.handleClientCon(ptls, p1, false, raw)
	}
}

func (conn *KcpClient) testURL(url string) {
	waitChannel <- 1
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	resultT := make(chan utils.ConfigSpeed, 10)
	fff := make(map[string][]utils.ConfigSpeed)
	go func() {
		for {
			c := <-resultT
			if c.Server == "[EOF]" {
				break
			}
			if _, ok := fff[c.Url]; ok {
				fff[c.Url] = append(fff[c.Url], c)
			} else {
				fff[c.Url] = []utils.ConfigSpeed{c}
			}
		}
	}()
	testSingle := func(url, host string) {
		resultT <- utils.TestURLUsedTime(url, host)
	}
	for _, c := range utils.BOOK.Books() {
		utils.SetConfigS(&c, utils.TestProxyAddr)
		testSingle(url, c.Server.(string))
	}
	resultT <- utils.ConfigSpeed{Server: "[EOF]"}
	for _, arr := range fff {
		if len(arr) == 0 {
			continue
		}
		sort.Slice(arr, func(i, j int) bool {
			return arr[i].Used < arr[j].Used
		})
		// utils.ColorL(u, "Min", arr[0], "Max", arr[len(arr)-1])
	}
	utils.AutoMap[utils.GetMainDomain(url)] = fff[utils.GetMainDomain(url)]
	<-waitChannel
}

func (conn *KcpClient) handleSS(ssuri string, con net.Conn) {
	defer func() {
		// time.Sleep(1 * time.Second)
		con.Close()
	}()
	if ssuri == "ss://ls" {
		output := make(map[string]string)
		for _, b := range utils.BOOK.Books() {
			output[b.Server.(string)] = b.LocalAddress
		}
		if d, err := json.Marshal(output); err == nil {
			// utils.ColorL(d)
			//fmt.Println(string(d))
			if _, err := con.Write(d); err != nil {
				// utils.ColorL("Err", err)
			}
		}
	} else if ssuri == "ss://route" {
		if d, err := json.Marshal(utils.AutoMap); err == nil {
			if _, err := con.Write(d); err != nil {
				utils.ColorL("ss://route| Err", err)
			}
		}
	} else if ssuri == "ss://stop" {
		TO_STOP = true
		if GlobalStatus {
			ProxySet("")
		}
		if IsStartHttpProxy {
			WaitClose <- 1
		}
		utils.ColorM("*", "*", "*", " byte~ ", "*", "*", "*")
		os.Exit(0)
	} else if ssuri == "ss://flow" {
		conn.SetMode(FLOW_MODE)
		con.Write([]byte(fmt.Sprintf("To Flow Mode Use  [%d]", len(utils.BOOK.Books()))))
		return
	} else if strings.HasPrefix(ssuri, "ss://use") {
		if usedRouteIP := strings.Replace(ssuri, "ss://use", "", 1); len(usedRouteIP) > 0 {
			if config := utils.BOOK.Get(usedRouteIP); config != nil {
				utils.ColorL("Single Mode Use:", config.Server.(string), "pwd:", config.Password, " Port:", config.ServerPort)
				conn.SetMode(SINGLE_MODE)
				conn.SetConfig(config)
				con.Write([]byte("Single Mode Use:" + config.Server.(string)))
				return
			}
		}

	} else if strings.HasPrefix(ssuri, "ss://auto") {
		conn.SetMode(AUTO_MODE)
		utils.ColorL("Turn Auto Route Mode ")
		con.Write([]byte("Turn Auto Route Mode "))
		return
	} else if strings.HasPrefix(ssuri, "ss://show/") {
		parts := strings.SplitN(ssuri, "show/", 2)
		ip := strings.TrimSpace(parts[1])
		c := utils.BOOK.Get(ip)
		con.Write([]byte(c.ToUri()))
	} else {
		conn.UpdateConfig(ssuri)
		con.Write([]byte("Change config ->" + ssuri))
		return
	}
}

func (conn *KcpClient) cmd(data []byte) {
	log.Println("{before init session}")
	session := conn.WithSession(nil, 2)
	log.Println("{before stream opened out}")
	if p2, err := session.OpenStream(); err == nil {

		log.Println("stream opened out:", fmt.Sprint(p2.RemoteAddr(), "(", p2.ID(), ")"))

		if _, err := p2.Write(data); err != nil {
			log.Fatal("wriet:", err)
		} else {
			// log.Println(data, "-> [", n, "]")
			g := color.New(color.FgGreen, color.Bold)
			buf := make([]byte, 1024)
			if n, err := p2.Read(buf); err != nil {
				// g.Println("cmd: ", string(buf[:n]))
				// utils.ColorL(err)
			} else {
				g.Println("cmd: ", string(buf[:n]))
			}
		}
		// wg.Wait()
	} else {
		log.Fatal("cmd:", err)
	}

}

// CmdString run cmd like redirect://ss://xxxxxxxx | redirect://ls | redirect://stop
func (conn *KcpClient) CmdString(cmd string) {
	conn.cmd(NewCmdStruct(cmd).ToData().Bytes())
}
func (conn *KcpClient) handleConns(p1, p2 net.Conn, quiet bool, hostData []byte) {
	defer p1.Close()
	defer p2.Close()
	conn.Pipe(p1, p2)

}

// handleClientCon aggregates connection p1 on mux with 'writeLock'
func (conn *KcpClient) handleClientCon(p2, p1 net.Conn, quiet bool, hostData []byte) {
	defer p1.Close()
	defer p2.Close()
	if _, err := p2.Write(hostData); err != nil {
		log.Fatal("no host/addr")
		return
	}
	conn.Pipe(p1, p2)
}

// handleClient aggregates connection p1 on mux with 'writeLock'
func (conn *KcpClient) handleClient(session *smux.Session, p1 net.Conn, quiet bool, hostData []byte) {
	defer p1.Close()
	p2, err := session.OpenStream()
	if err != nil {
		utils.ColorL("StreamErr", err)
		if session, err = conn.ReConnection(); err != nil {
			utils.ColorL("StreamErrAgain", err)
			return
		}
		return
	}
	// utils.ColorL("Stream", "ready")
	defer p2.Close()
	if _, err := p2.Write(hostData); err != nil {
		log.Fatal("no host/addr")
		return
	}
	socksReplyBuf := make([]byte, 128)
	n, err := p2.Read(socksReplyBuf)
	if err != nil {
		if err.Error() == "io: read/write on closed pipe" {
			// utils.ColorL("Close ")
			return
		}
		log.Println("no socks5 reply host/addr, err:", err)
		p1.Close()
		return
	} else {
		// utils.ColorL("Stream", socksReplyBuf[:n])
	}
	p1.Write(socksReplyBuf[:n])
	conn.Pipe(p1, p2)
}
