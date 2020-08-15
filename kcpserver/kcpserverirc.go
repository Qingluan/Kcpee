package kcpserver

import (
	"log"
	"net"
	"strings"

	"gitee.com/dark.H/go-remote-repl/remote"
)

func (server *KcpServer) ConfigWithIrc(conn net.Conn) {
	defer conn.Close()
	ip := strings.SplitN(conn.RemoteAddr().String(), ":", 2)[0]
	man := remote.ManWraper(conn)
	choose, err := man.Talk(false, "redirect", "ssconfig")
	if choose == "" || err != nil {
		if err != nil {
			log.Println("ConfigRC error:", err)
		}
		return
	}
	switch choose {
	case "redirect":
		server.SetRedirectIRC(man, ip)
	case "ssconfig":
		server.SetSSConfigIRC(man, ip)
	}

}

func (server *KcpServer) SetSSConfigIRC(man *remote.Man, ip string) {
	// method, err := man.Talk(true, "ss method")
	// if err != nil {
	// 	log.Println("Talk err:", err)
	// }
	// passwd, err := man.Talk(true, "ss password")
	// if err != nil {
	// 	log.Println("Talk err:", err)
	// }
	// server.k
}
