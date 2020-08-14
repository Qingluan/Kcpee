package kcpserver

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"gitee.com/dark.H/go-remote-repl/remote"
	"github.com/Qingluan/Kcpee/utils"
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

func (server *KcpServer) SetRedirectIRC(man *remote.Man, ip string) {
	authName, err := man.Talk(true, "username")
	if err != nil {
		log.Println("talk error:", err)
		return
	}
	authPasswd, err := man.Talk(true, "pwd")
	if err != nil {
		log.Println("talk error:", err)
		return
	}
	defer os.RemoveAll("Kcpconfig")
	if configFile, err := utils.Credient(authName, authPasswd); err != nil {
		log.Println("redirect error:", err)
		return
	} else {

		if utils.IsDir(configFile) {
			book := utils.NewBook()
			book.Scan(configFile)
			cs := book.GetServers()
			if err != nil {
				log.Println("talk error:", err)
				return
			}
			choose, err := man.Talk(false, cs...)
			if err != nil {
				log.Println("talk error:", err)
				return
			}
			usetime, err := man.Talk(true, "86400", "3600")
			if err != nil {
				log.Println("talk time error:", err)
				return
			}
			if usetime != "" {
				t, err := strconv.Atoi(usetime)
				if err != nil {
					t = 86400
				}
				if choose != "" {
					keys := strings.SplitN(choose, ":", 2)
					conf := book.Get(keys[0])
					route := new(utils.Route)
					route.SetMode("proxy")
					route.SetExireTime(t)
					route.SetConfig(conf)
					server.RedirectBooks[ip] = route
					utils.ColorL("redirect ->", ip)
				}
			}

		}
	}
}
