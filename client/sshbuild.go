package client

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/Qingluan/Kcpee/utils"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func Auth(name, host, port, passwd string, callbcak func(sess *ssh.Session)) {

	sshConfig := &ssh.ClientConfig{
		User: name,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
		Timeout:         15 * time.Second,
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	host += ":" + port
	client, err := ssh.Dial("tcp", host, sshConfig)

	if err != nil {
		fmt.Println("connect:", err)
		return
	}
	defer client.Close()

	// start session
	sess, err := client.NewSession()
	if err != nil {
		log.Fatal("session:", err)
	}
	defer sess.Close()
	callbcak(sess)
}

func credentials() string {
	// reader := bufio.NewReader(os.Stdin)

	// fmt.Print("Enter Username: ")
	// username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		fmt.Println("\nPassword typed: " + string(bytePassword))
	}
	password := string(bytePassword)

	return strings.TrimSpace(password)
}

func Build(name, host, port, pwd string, config utils.Config) {
	Auth(name, host, port, pwd, func(sess *ssh.Session) {
		// setup standard out and error
		// uses writer interface
		var out bytes.Buffer
		sess.Stdout = &out
		// sess.Stdout = os.Stdout
		// sess.Stderr = os.Stderr
		// run single command
		// fmt.Println(host, "Connected", pwd)
		// cmdStr := fmt.Sprintf(`apt install -y wget;yum install -y wget; rm Kcpee-linux; wget -c -q 'https://github.com/Qingluan/kcpee/releases/download/v1.2/Kcpee-linux' && chmod +x Kcpee-linux ; ulimit -n 4096 ;  ./Kcpee-linux -S -R  -d -k "%s" -p %d  && rm ./Kcpee-linux`, config.Password, config.ServerPort)
		cmdStr := fmt.Sprintf(`wget -c -q 'https://github.com/Qingluan/kcpee/releases/download/v1.3/Kcpee-linux' && chmod +x Kcpee-linux; ps aux | grep Kcpee-linux | egrep -v '(grep|egrep|wget)' | awk '{print $2}' |xargs kill -9 ;ulimit -n 4096 ;  ./Kcpee-linux -S -R  -d -k "%s" -p %d -P ss; rm ./Kcpee-linux`, config.Password, config.ServerPort)
		// fmt.Println(cmdStr)
		err := sess.Run(cmdStr)
		// fmt.Println(host, "finished")
		if err != nil {
			log.Fatal("run:", err, utils.FGCOLORS[3](cmdStr))
			// }
		} else {
			if strings.Contains(out.String(), "./Kcpee-linux [PID]") {
				utils.Stat(host+"start kcpee", true)

			} else {
				utils.Stat(host+"stop kcpee", false)
			}
		}

	})
}

func Stop(name, host, port, pwd string) {
	Auth(name, host, port, pwd, func(sess *ssh.Session) {
		sess.Stdout = os.Stdout
		sess.Stderr = os.Stderr
		// run single command
		err := sess.Run("ps aux | grep Kcpee | egrep -v '(grep|egrep)' | awk '{print $2}' |xargs kill -9 ")
		if err != nil {
			log.Fatal(err)
		}
	})
}
