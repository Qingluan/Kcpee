package kcpserver

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net"
// 	"os"
// 	"strings"
// 	"syscall"
// 	"time"

// 	"github.com/Qingluan/Kcpee/utils"

// )

// // KcpServer used for server
// type TlsServer struct {
// 	KcpServer

// }

// // Listen for kcp
// func (serve *TlsServer) Listen() {
// 	config := serve.GetConfig()
// 	kconfig := serve.GetKcpConfig()
// 	block := config.GeneratePassword()
// 	severString := fmt.Sprintf("%s:%d", config.GetServerArray()[0], config.ServerPort)
// 	if listener, err := kcp.ListenWithOptions(severString, block, kconfig.DataShard, kconfig.ParityShard); err == nil {
// 		listener.SetReadBuffer(4194304)
// 		listener.SetWriteBuffer(4194304)

// 		g := color.New(color.FgGreen)
// 		g.Printf("accept ready \r")
// 		for {
// 			conn, err := listener.AcceptKCP()
// 			serve.UpdateKcpConfig(conn)
// 			if err != nil {
// 				if !strings.Contains(err.Error(), "too many open files") {
// 					log.Println(err)
// 				}
// 				continue
// 			}
// 			go serve.handleMux(conn)
// 		}
// 	} else {
// 		log.Fatal(err)
// 	}
// }
