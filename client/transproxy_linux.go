// // +build linux,!windows,!darwin,!js

package client

// import (
// 	"log"
// 	"net"
// 	"strings"
// 	"time"

// 	"github.com/Qingluan/HookNet/ptrace"
// 	"github.com/Qingluan/Kcpee/utils"
// )

// // KcpClient for client and cmd
// type KcpClientLinux struct {
// 	KcpClient
// }

// // NewKcpClient init KcpClient
// func NewKcpClientLinux(config *utils.Config, kconfig *utils.KcpConfig) (kclient *KcpClient) {
// 	kclient = new(KcpClient)
// 	kclient.Numconn = 512
// 	kclient.SetConfig(config)
// 	kclient.SetMode(AUTO_MODE)
// 	kclient.SetKcpConfig(kconfig)
// 	return
// }

// func LocalUnixServer(addr string) {
// 	unixServer := ptrace.NewCacheUnixSocket("/tmp/addr.sock")
// 	go unixServer.StartService()
// 	return
// }

// func (client *KcpClientLinux) Listen(listenAddr string) {
// 	ln, err := net.Listen("tcp", listenAddr)
// 	client.listenAddr = listenAddr
// 	if err != nil {
// 		log.Fatal("listen error:", err)
// 	}
// 	// rr := uint16(0)
// 	if len(utils.AutoMap) > 0 {
// 		client.useAutoMap = true
// 	}
// 	acceptnum := 0
// 	if client.Role == "client" {
// 		go utils.SpeedShow()
// 	}
// 	for {
// 		if TO_STOP {
// 			break
// 		}
// 		if client.Role == "tester" && client.GetAliveNum() > client.Numclient {
// 			time.Sleep(10 * time.Millisecond)
// 			continue
// 		}
// 		p1, err := ln.Accept()
// 		utils.ColorL("socks5 <-- ", p1.RemoteAddr())
// 		acceptnum++
// 		if acceptnum%20 == 0 {
// 			// utils.ColorL("accept tcp:", acceptnum)
// 		}

// 		if err != nil {
// 			if !strings.Contains(err.Error(), "too many open files") {
// 				log.Println("%+v", err)
// 			}

// 			continue
// 		}
// 		go client.handleSocks5Tcp(p1)

// 	}
// 	return
// }
