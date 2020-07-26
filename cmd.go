package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	// "github.com/dr/KcpEnumaElish/client"
	"github.com/Qingluan/Kcpee/client"
	"github.com/Qingluan/Kcpee/kcpserver"
	"github.com/Qingluan/Kcpee/utils"
	"github.com/fatih/color"
	// "./client"
)

var (
	configFile   string
	routeMapFile string
	authName     string
	authPasswd   string

	bindString       string
	server           string
	pwd              string
	bookcmd          string
	bookuri          string
	tunnelTo         string
	urlsFile         string
	dirRoot          string
	testURL          string
	conNum           int
	ttl              int
	isChangeConfig   bool
	isRedirect       bool
	isTunnel         bool
	isServer         bool
	isConnect        bool
	isBuild          bool
	isGenerate       bool
	isHttpProxy      bool
	godaemon         bool
	isTest           bool
	isSync           bool
	isCredient       bool
	isCmdLs          bool
	isCmdShowRoute   bool
	isCmdStop        bool
	isStatus         bool
	isStartTest      bool
	isToUri          bool
	isCmdAutoRoute   bool
	isCmdFlowRoute   bool
	isAuthEdit       bool
	configToUrl      bool
	isCmdSingleRoute string
	thisnodeproxyto  string
	doSomeString     string
	SaveToFile       string
	isGBK            bool
	ifStartUDPClient bool

	// Config area
	refreshRate int
)

func DoMain() {
	var cmdConfig utils.Config
	var kcpConfig utils.KcpConfig
	gprint := utils.BGCOLORS[0]
	flag.StringVar(&configFile, "c", "", "specify config uri/file/dir | uri: ss://{base64} "+gprint("if specify a dir , will auto scan multi configs"))
	flag.StringVar(&server, "s", "", "server ip")
	flag.StringVar(&cmdConfig.LocalAddress, "b", "0.0.0.0", "local address, listen only to this address if specified")
	flag.StringVar(&cmdConfig.Password, "k", "Hello World!", "password")
	flag.IntVar(&cmdConfig.ServerPort, "p", 10443, "server port")
	flag.IntVar(&cmdConfig.Timeout, "t", 300, "timeout in seconds")
	flag.IntVar(&cmdConfig.LocalPort, "l", 1091, "local socks5 proxy port")
	flag.StringVar(&cmdConfig.Method, "m", "", "encryption method, default: aes-256-cfb")

	flag.StringVar(&urlsFile, "history.url", "", "url file ")
	flag.StringVar(&routeMapFile, "route", "", "set route map json file path")
	flag.StringVar(&authName, "name", "", "-Auth -name xxx to set name")
	flag.StringVar(&authPasswd, "pwd", "", "-Auth -pwd xxx to set pwd")
	flag.BoolVar(&isChangeConfig, "update", false, "change running config file")
	flag.BoolVar(&isGBK, "gbk", false, "change stdio charset to gbk!")
	flag.IntVar(&ttl, "ttl", 600, "set ttl ")
	// flag.StringVar(&server, "s", "127.0.0.1:18081", "set server addr")
	// flag.StringVar(&pwd, "p", "hello world", "set password")

	flag.IntVar(&conNum, "channelNum", 128, "set connect to kcp conn")
	flag.BoolVar(&isServer, "S", false, "set true when use server mode")
	flag.BoolVar(&ifStartUDPClient, "U", false, "if start udp listen client")
	flag.BoolVar(&isTunnel, "T", false, "set true when use server mode")
	flag.BoolVar(&configToUrl, "L", false, "trans config to urls -c /somedir/ -L ")
	flag.BoolVar(&isHttpProxy, "H", true, "add http proxy local listener in :10091 ")
	// flag.BoolVar(&isConnect, "C", false, "set true when use connect mode | this mode connect[client] -> tunnel[server] ")
	flag.StringVar(&tunnelTo, "C", "", "reverse Map {host} , {host} can see in -T mode")
	flag.BoolVar(&isRedirect, "R", false, "set true when redirect mode")
	flag.BoolVar(&godaemon, "d", false, "run as a daemon !")
	flag.BoolVar(&isTest, "Test", false, "set true if test urls for configs")
	flag.BoolVar(&isStartTest, "B", false, "if start background tester for all routes")
	flag.BoolVar(&isGenerate, "Test.gen", false, "if genreate config set true")
	flag.StringVar(&testURL, "Test.url", "", "test this url for all route")
	flag.BoolVar(&isBuild, "Test.build", false, "set true if want to build from config.json's server host vps")
	flag.BoolVar(&isStatus, "Stat", false, "start status bar in terminal?")
	flag.BoolVar(&isSync, "Sync", false, "will generate a config.en file by -c configs and route.map.file")
	flag.BoolVar(&isCredient, "Auth", false, "will sync personal configs and route by decrypted git file.")
	flag.BoolVar(&isCmdLs, "book.ls", false, "[use in local]:show local route info ")
	flag.BoolVar(&isCmdAutoRoute, "book.auto", false, "[use in local]:show local route info example: -book.auto ")
	flag.StringVar(&isCmdSingleRoute, "book.single", "", "[use in local]:use config's single route ! example: -book.single 'ip' ")
	flag.BoolVar(&isCmdFlowRoute, "book.flow", false, "[use in local]: example: -book.flow set book flow mode ")
	flag.BoolVar(&isCmdShowRoute, "book.show", false, "[use in local]:show local route json  ")
	flag.BoolVar(&isCmdStop, "book.stop", false, "[use in local]: stop local progress by this  ")
	flag.StringVar(&bookuri, "book.uri", "", "[use in local]: show ip's detail ssuri")
	flag.StringVar(&bookcmd, "book.cmd", "", "[use in server]:run book cmd: redirect://ls | redirect://stop | redirect://ss://uri | redirect://scan@/path")
	flag.BoolVar(&isToUri, "Url", false, "true to print config file's uri stirng")
	flag.StringVar(&kcpConfig.Mode, "kcpmode", "fast3", "set kcp mode normal,fast, fast1, fast2, fast3")
	flag.StringVar(&thisnodeproxyto, "red", "", "node redirect to another node like -red ss://xxxx= ")
	flag.StringVar(&doSomeString, "Do", "", "cmd string run hear include test, gernerate and do some")
	flag.StringVar(&SaveToFile, "output", "", "output string dst or some output ")

	flag.IntVar(&refreshRate, "config.rate", 3, "set recv msg refresh rate, default: 3s")
	flag.Parse()

	if isTest && utils.IsDir(configFile) {
		tester := utils.NewSpeedTest()
		tester.GetHistoryUrls()
		utils.ColorL("to history")
		if testURL == "" {
			testRouteMapJSONString := tester.TestSpeed([]string{}, configFile)
			ddd, _ := json.Marshal(testRouteMapJSONString)
			if err := ioutil.WriteFile(filepath.Join(configFile, "route.map.json"), ddd, 0644); err != nil {
				log.Fatal(err)
			}
		} else {
			rou := make(map[string][]utils.ConfigSpeed)
			if utils.PathExists(filepath.Join(configFile, "route.map.json")) {
				rou = utils.GetOldSpeedMap(filepath.Join(configFile, "route.map.json"))
			}
			testRouteMapJSONString := tester.TestSpeed([]string{testURL}, configFile)
			for k, v := range testRouteMapJSONString {
				rou[k] = v
			}
			ddd, _ := json.Marshal(rou)
			if err := ioutil.WriteFile(filepath.Join(configFile, "route.map.json"), ddd, 0644); err != nil {
				log.Fatal(err)
			}
		}

		os.Exit(0)
	}

	if isCmdLs {
		tester := utils.NewSpeedTest()
		if d, err := tester.LsConfig(); err != nil {
			log.Fatal("json:", err)
		} else {
			for ip, loc := range d {
				res := strings.Split(loc, "/")
				if runtime.GOOS == "windows" {
					res = strings.Split(loc, "\\")
				}

				fmt.Println(ip, ":", res[len(res)-1])
			}
		}
		os.Exit(0)
	} else if isCmdShowRoute {
		tester := utils.NewSpeedTest()
		if d, err := tester.LsRoute(); err != nil {
			log.Fatal("json:", err)
		} else {
			fmt.Println(d)
		}
		os.Exit(0)
	} else if isCmdFlowRoute {
		tester := utils.NewSpeedTest()
		o, _ := tester.FlowMode()
		fmt.Println(o)
		os.Exit(0)

	} else if isCmdStop {
		if isServer {
			cmdConfig.Server = "localhost"
			conn := client.NewKcpClient(&cmdConfig, &kcpConfig)
			// g.Println("run client cmd:", bookcmd)
			conn.Numconn = conNum
			conn.CmdString("redirect://kill-my-life")

			os.Exit(0)
		}
		tester := utils.NewSpeedTest()

		tester.LsStop()
		if client.IfProxyStart() {
			client.ProxySet("")
		}
		client.KillKcpee()
		os.Exit(0)
	} else if isCmdSingleRoute != "" {
		tester := utils.NewSpeedTest()
		tester.SetRoute(isCmdSingleRoute)
		os.Exit(0)
	} else if isCmdAutoRoute {
		tester := utils.NewSpeedTest()
		tester.SetRoute()
		os.Exit(0)
	}

	if isCredient {
		var err error
		if configFile, err = utils.Credient(authName, authPasswd); err != nil {
			log.Fatal(err)
		}
	}

	localAddress := fmt.Sprintf("%s:%d", cmdConfig.LocalAddress, cmdConfig.LocalPort)
	utils.ProxyAddr = localAddress
	// var config utils.Config
	// var err error
	g := color.New(color.FgGreen, color.Bold)
	kcpConfig.SetAsDefault()
	kcpConfig.UpdateMode()

	if configFile != "" {
		if utils.IsDir(configFile) {
			utils.BOOK.Scan(configFile)
			if isToUri {
				cs, _ := utils.BOOK.Info()
				for _, cn := range cs {
					fmt.Println(g.Sprint(cn))
				}
				os.RemoveAll("Kcpconfig")
				os.Exit(0)
			}
			cmdConfig = *utils.BOOK.Get()
			if configFile == "Kcpconfig" {
				if utils.BOOK.Count() == 0 {
					utils.FGCOLORS[2]("Error Password!")
					os.Exit(0)
				}
				if !isSync {
					if err := os.RemoveAll("Kcpconfig"); err != nil {
						log.Fatal(err)
					}
				} else {
					utils.ColorL("wait to append new route")
					os.Exit(0)
				}

			}
			utils.ColorL("config dir:", configFile)
		} else {
			if strings.HasPrefix(configFile, "ss://") {
				cmdConfig = *utils.ParseURI(configFile)
			} else {
				if config, err := utils.ParseConfig(configFile); err == nil {
					cmdConfig = *config
				}
			}

		}
		if configToUrl {
			cs, err := utils.BOOK.Info()
			if err != nil {
				log.Fatal(err)
			}
			for _, c := range cs {
				fmt.Println(c)
			}
			os.Exit(0)
		}
	} else {
		if server == "" {
			if utils.IsDir("routes") {
				utils.BOOK.Scan("routes")
				cmdConfig = *utils.BOOK.Get()
			} else {
				utils.ColorL("no config found exit")
				// os.Exit(0)
			}
		}
	}

	if cmdConfig.Server == nil {
		if server != "" {
			cmdConfig.Server = server
		} else {
			cmdConfig.Server = fmt.Sprintf("%s", cmdConfig.LocalAddress)
		}
	}

	if doSomeString != "" {
		operator(doSomeString, &cmdConfig)
		os.Exit(0)
	}
	if isToUri {
		resString := cmdConfig.ToUri()
		fmt.Println(resString)
		os.Exit(0)
	}

	if isChangeConfig {
		tester := utils.NewSpeedTest()
		tester.SetConfig(&cmdConfig)
		os.Exit(0)
	}
	if bookcmd != "" {
		conn := client.NewKcpClient(&cmdConfig, &kcpConfig)
		g.Println("run client cmd:", bookcmd)
		conn.Numconn = conNum
		conn.CmdString(bookcmd)
		return
	}
	if thisnodeproxyto != "" {
		conn := client.NewKcpClient(&cmdConfig, &kcpConfig)
		conn.Numconn = 5
		conn.CmdString("redirect://" + thisnodeproxyto)
		conn = client.NewKcpClient(&cmdConfig, &kcpConfig)
		conn.Numconn = 5
		config := utils.ParseURI(thisnodeproxyto)
		conn.CmdString("redirect://start@" + config.Server.(string))
		os.Exit(0)
	}
	if bookuri != "" {
		tester := utils.NewSpeedTest()
		if d, err := tester.ShowConfig(bookuri); err != nil {
			log.Fatal("json:", err)
		} else {
			fmt.Println(string(d))
		}
		os.Exit(0)
	}
	if isBuild {
		var wait sync.WaitGroup
		for _, confi := range utils.BOOK.Books() {
			wait.Add(1)
			// utils.ColorL("build", confi.Server.(string), "22", confi.ServerPassword)
			go func(conf utils.Config, w *sync.WaitGroup) {
				defer w.Done()
				client.Build("root", conf.Server.(string), "22", conf.ServerPassword, conf)
			}(confi, &wait)
		}
		wait.Wait()
		// client.Build("root", cmdConfig.Server.(string), "22", cmdConfig.ServerPassword, &cmdConfig)
		os.Exit(0)
	} else if isGenerate {
		dst := utils.NORMAL_CONFIG_ROOT

		utils.GenerateConfigJsons(dst)
		os.Exit(0)
	}

	if godaemon {
		args := []string{}
		for _, a := range os.Args[1:] {
			if a == "-d" {
				continue
			}
			args = append(args, a)
		}
		cmd := exec.Command(os.Args[0], args...)
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		// cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

		cmd.Start()

		time.Sleep(2 * time.Second)
		fmt.Printf("%s [PID] %d running...\n", os.Args[0], cmd.Process.Pid)
		os.Exit(0)
	}
	if isSync && !isCredient {
		if ff, err := utils.Sync(configFile); err == nil {
			utils.ColorL("genreate file:", ff)
		}
		os.Exit(0)
	}

	if isServer {
		g.Println("run server mode")
		if runtime.GOOS != "windows" {
			cmd := exec.Command("ulimit", "-n", "4096")
			cmd.Run()
		}
		kcpServe := kcpserver.NewKcpServer(&cmdConfig, &kcpConfig)
		kcpServe.SetRefreshRate(refreshRate)
		kcpServe.Numconn = conNum
		if isRedirect {
			g.Println("start redirect mode")
			utils.BOOK.Scan()
			defaultBook := utils.BOOK.Get()
			kcpServe.Init(defaultBook)
			// kcpServe.StartTunnels(utils.BOOK.GetServers()...)
		}
		// kcpServe.SetConfig(&cmdConfig)
		kcpServe.Listen()
	} else if isTunnel {
		var conn = client.NewKcpTunnel(&cmdConfig, &kcpConfig)
		g.Println("1. >>", tunnelTo)
		conn.SetRefreshRate(refreshRate)
		conn.SetTunMode("map")
		conn.Tunnel(tunnelTo)
	} else if tunnelTo != "" {
		g.Println(">>", tunnelTo)
		if isGBK {
			client.UseGBK = true
		}
		var conn = client.NewKcpTunnel(&cmdConfig, &kcpConfig)
		conn.SetRefreshRate(refreshRate)
		conn.SetTunMode("map")
		conn.Connect(localAddress, tunnelTo)
	} else {
		if isCredient || configFile != "" || server != "" {
			g.Println("run client mode")
			if runtime.GOOS == "darwin" {
				cmd := exec.Command("ulimit", "-n", "4096")
				cmd.Run()
			}
			var conn = client.NewKcpClient(&cmdConfig, &kcpConfig)
			if isStatus {
				conn.Role = "client"
			}
			if isStartTest {
				testConn := client.NewKcpClient(&cmdConfig, &kcpConfig)
				testConn.Numconn = 10
				testConn.Role = "tester"

				go testConn.Listen(utils.TestProxyAddr)
			}

			// conn.Init(nil)
			conn.Numconn = conNum
			if isHttpProxy {
				go func() {
					// client.ProxySet("http://localhost:10091")
					conn.ListenHttpProxy("")
				}()
			}
			conn.Listen(localAddress, ifStartUDPClient)
		} else {
			utils.RunGui(func() {
				if client.IfProxyStart() {
					client.ProxySet("")
				} else {
					client.ProxySet("http://localhost:10091")
				}

			})
		}
	}

}

func operator(some string, cmdConfig *utils.Config) {
	if strings.HasPrefix(some, "tls.") {
		f, _ := utils.CreateCertificate(cmdConfig.Server.(string), false)
		if strings.HasSuffix(some, "json") {
			conf := utils.ParseURI(f)
			fmt.Print(conf.ToJson())
		} else if strings.HasSuffix(some, "file") {
			conf := utils.ParseURI(f)
			if SaveToFile != "" {
				conf.ToFile(SaveToFile)
				utils.ColorL("==> ", SaveToFile)
			}

		} else {
			fmt.Print(f)
		}
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	// `signal.Notify` registers the given channel to
	// receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// This goroutine executes a blocking receive for
	// signals. When it gets one it'll print it out
	// and then notify the program that it can finish.
	go func() {
		<-sigs
		if client.GlobalStatus {
			client.ProxySet("")
		}
		done <- true
		os.Exit(0)
	}()
	DoMain()
	// The program will wait here until it gets the
	// expected signal (as indicated by the goroutine
	// above sending a value on `done`) and then exit.
	// fmt.Println("awaiting signal")
	<-done
	fmt.Println("~ bye")

}
