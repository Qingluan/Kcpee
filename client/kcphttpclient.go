package client

import (
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DEFAULT_TIMEOUT = 5 * time.Second
)

var (
	flagAuthUser     = flag.String("httpuser", "", "Server authentication username")
	flagAuthPass     = flag.String("httppass", "", "Server authentication password")
	IsStartHttpProxy = false
)

func (kclient *KcpClient) ListenHttpProxy(listenAddr string) (err error) {
	c := zap.NewProductionConfig()
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := c.Build()
	if err != nil {
		log.Fatalln("Error: failed to initiate logger")
	}
	defer logger.Sync()
	stdLogger := zap.NewStdLog(logger)

	proxy := &Proxy{
		ForwardingHTTPProxy: NewForwardingHTTPProxy(stdLogger),
		Logger:              logger,
		AuthUser:            *flagAuthUser,
		AuthPass:            *flagAuthPass,
		DestDialTimeout:     DEFAULT_TIMEOUT,
		DestReadTimeout:     DEFAULT_TIMEOUT,
		DestWriteTimeout:    DEFAULT_TIMEOUT,
		ClientReadTimeout:   DEFAULT_TIMEOUT,
		ClientWriteTimeout:  DEFAULT_TIMEOUT,
		// Avoid:              ,
		HandleBody: func(p1 net.Conn, host string, afterConnected func(p1, p2 net.Conn)) {
			kclient.handleBodyDo(p1, host, func(p2 net.Conn, p1 net.Conn, raw []byte) {
				if _, err := p2.Write(raw); err != nil {
					log.Fatal("no host/addr")
					return
				} else {
					buf := make([]byte, 10)
					_, err = p2.Read(buf)
					if err != nil {

						p1.Close()
					}
				}
				afterConnected(p1, p2)
				kclient.Pipe(p1, p2)
				// log.Println(raw)
			})
		},
	}
	if listenAddr == "" {
		listenAddr = ":10091"
	}
	srv := &http.Server{
		Addr:    listenAddr,
		Handler: proxy,
	}
	IsStartHttpProxy = true
	srv.ListenAndServe()
	return
}
