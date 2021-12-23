package utils

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"encoding/json"

	"golang.org/x/net/proxy"
)

var (
	ProxyAddr          = "127.0.0.1:1091"
	TestProxyAddr      = "127.0.0.1:40086"
	RouteIsBroken      = errors.New("Route is broken. can not connect to outer world!")
	SocksConnectBroken = errors.New("Socks5 connect Error!")
	globalTransport    *http.Transport
	tmpCacheTestURL    = make(chan string, 10)
)

type SpeedTest struct {
	domainMap   map[string]*Config
	message     chan ConfigSpeed
	httpClient  *http.Client
	SortResults map[string][]ConfigSpeed
	wait        chan int
}

type UrlSpeed struct {
	host string
	urls map[string]float32
}

type ConfigSpeed struct {
	Server string
	Url    string
	Used   float32
}

func NewSpeedTest() (tester *SpeedTest) {
	tester = new(SpeedTest)
	dialer, err := proxy.SOCKS5("tcp", ProxyAddr, nil, proxy.Direct)
	if err != nil {
		fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
		os.Exit(1)
	}
	httpTransport := &http.Transport{ResponseHeaderTimeout: 12 * time.Second}
	httpTransport.Dial = dialer.Dial
	globalTransport = httpTransport
	// tester.httpClient = &http.Client{Transport: httpTransport}
	tester.domainMap = make(map[string]*Config)
	tester.message = make(chan ConfigSpeed, 20)
	tester.SortResults = make(map[string][]ConfigSpeed)
	tester.wait = make(chan int, 1)
	return
}

func (test *SpeedTest) SetConfig(config *Config) (err error) {
	ssUri := config.ToUri()
	ColorL("Set", config.Server.(string))
	data := []byte{0x5, 0x1, 0x0, 0x5, uint8(len(ssUri))}
	buffer := bytes.NewBuffer(data)
	buffer.Write([]byte(ssUri))
	if con, ierr := net.Dial("tcp", ProxyAddr); ierr == nil {
		defer con.Close()
		con.Write([]byte{0x5, 0x1, 0x0})
		buf := make([]byte, 10)
		con.Read(buf)
		d := buffer.Bytes()
		con.Write(d)
		con.Read(buf)
	} else {
		err = ierr
	}
	return
}

func SetConfigS(config *Config, addr string) (err error) {
	ssUri := config.ToUri()
	// ColorL("Set", config.Server.(string))
	data := []byte{0x5, 0x1, 0x0, 0x5, uint8(len(ssUri))}
	buffer := bytes.NewBuffer(data)
	buffer.Write([]byte(ssUri))
	if con, ierr := net.Dial("tcp", addr); ierr == nil {
		defer con.Close()
		con.Write([]byte{0x5, 0x1, 0x0})
		buf := make([]byte, 10)
		con.Read(buf)
		d := buffer.Bytes()
		con.Write(d)
		con.Read(buf)
	} else {
		err = ierr
	}
	return
}

func (test *SpeedTest) baseCmd(cmdStr string) (outputdata []byte, err error) {
	data := []byte{0x5, 0x1, 0x0, 0x5, uint8(len(cmdStr))}
	buffer := bytes.NewBuffer(data)
	buffer.Write([]byte(cmdStr))
	if con, ierr := net.Dial("tcp", ProxyAddr); ierr == nil {
		defer con.Close()
		con.Write([]byte{0x5, 0x1, 0x0})
		buf := make([]byte, 40)
		con.Read(buf)
		//fmt.Println("C:",string(buf))
		//fmt.Println(" ------ ")
		d := buffer.Bytes()
		con.Write(d)
		// buf3 := make([]byte, 10)
		// con.Read(buf3)
		// ColorL(buf)
		buf2 := make([]byte, 2048)
		for {
			if n, ierr := con.Read(buf2); ierr == nil {
				// ColorL(buf2[:n])
				outputdata = buf2[:n]
				return
			} else {
				fmt.Println(buf2)
				err = ierr
			}
		}
	} else {
		ColorL("dial err:", ierr)
		err = ierr
	}
	return
}

func (test *SpeedTest) ShowConfig(ip string) (output string, err error) {
	ssUri := "ss://show/" + ip
	if buf, ierr := test.baseCmd(ssUri); err == nil {
		// fmt.Printf("%v", buf)
		output = string(buf)
	} else {
		err = ierr
	}
	return
}

func (test *SpeedTest) LsConfig() (output map[string]string, err error) {
	ssUri := "ss://ls"
	// data := []byte{0x5, 0x1, 0x0, 0x5, uint8(len(ssUri))}
	// buffer := bytes.NewBuffer(data)
	// buffer.Write([]byte(ssUri))
	// if con, ierr := net.Dial("tcp", ProxyAddr); ierr == nil {
	// 	defer con.Close()
	// 	con.Write([]byte{0x5, 0x1, 0x0})
	// 	buf := make([]byte, 30)
	// 	con.Read(buf)
	// 	d := buffer.Bytes()
	// 	con.Write(d)
	// 	con.Read(buf)
	// 	// ColorL(buf)
	// 	buf2 := make([]byte, 1024)
	// 	for {
	// 		if n, ierr := con.Read(buf2); ierr == nil {
	// 			// ColorL(buf2[:n])
	// 			err = json.Unmarshal(buf2[:n], &output)
	// 			return
	// 		} else {
	// 			// panic(err)
	// 			log.Fatal("b json:", err)
	// 			// log.Print("wait \r")
	// 		}
	// 	}
	// } else {
	// 	err = ierr
	// }
	if buf, ierr := test.baseCmd(ssUri); err == nil {
		// fmt.Printf("%s", string(buf))
		err = json.Unmarshal(buf, &output)
	} else {
		err = ierr
	}
	return
}

func (test *SpeedTest) GetNow() (out map[string]string, err error) {
	var buf []byte
	buf, err = test.baseCmd("ss://now")
	d := make(map[string]string)
	json.Unmarshal(buf, &d)
	return d, err
}

func (test *SpeedTest) FlowMode() (output string, err error) {
	var buf []byte
	buf, err = test.baseCmd("ss://flow")
	output = string(buf)
	return
}

func (test *SpeedTest) SetRoute(uri ...string) (output string, err error) {
	var buf []byte
	if len(uri) > 0 {
		buf, err = test.baseCmd("ss://use" + strings.TrimSpace(uri[0]))
	} else {
		buf, err = test.baseCmd("ss://auto")
	}
	if err != nil {
		return
	}
	output = string(buf)
	return
}

func (test *SpeedTest) LsRoute() (out string, err error) {
	ssUri := "ss://route"
	data := []byte{0x5, 0x1, 0x0, 0x5, uint8(len(ssUri))}
	buffer := bytes.NewBuffer(data)
	buffer.Write([]byte(ssUri))
	if con, ierr := net.Dial("tcp", ProxyAddr); ierr == nil {
		defer con.Close()
		con.Write([]byte{0x5, 0x1, 0x0})
		buf := make([]byte, 30)
		con.Read(buf)
		d := buffer.Bytes()
		con.Write(d)
		// con.Read(buf)
		// ColorL(buf)
		buf2 := make([]byte, 14096)
		for {
			if n, ierr := con.Read(buf2); ierr == nil {
				// ColorL(buf2[:n])
				out = string(buf2[:n])
				return
			} else {
				// panic(err)
				log.Fatal("b json:", ierr)
				// log.Print("wait \r")
			}
		}

	} else {
		err = ierr
	}
	return
}

func (test *SpeedTest) LsStop() (out string, err error) {
	ssUri := "ss://stop"
	data := []byte{0x5, 0x1, 0x0, 0x5, uint8(len(ssUri))}
	buffer := bytes.NewBuffer(data)
	buffer.Write([]byte(ssUri))
	if con, ierr := net.Dial("tcp", ProxyAddr); ierr == nil {
		defer con.Close()
		con.Write([]byte{0x5, 0x1, 0x0})
		buf := make([]byte, 30)
		con.Read(buf)
		d := buffer.Bytes()
		con.Write(d)
	} else {
		err = ierr
	}
	return
}

func (test *SpeedTest) sortWithResult() (out map[string][]ConfigSpeed) {
	results := test.SortResults
	start := time.Now()
	for {
		speed := <-test.message
		if speed.Server == "end" {
			break
		}
		if speed.Used == 33333 {
			continue
		}
		// ColorL(speed.Url, speed.Used, "in", speed.Server)
		if _, ok := results[speed.Url]; ok {
			results[speed.Url] = append(results[speed.Url], speed)
		} else {
			results[speed.Url] = []ConfigSpeed{speed}
		}
		if start.Sub(time.Now())/time.Second > 10 {
			for _, arr := range results {
				sort.Slice(arr, func(i, j int) bool {
					return arr[i].Used > arr[j].Used
				})
				// ColorL(u, arr[0], arr[len(arr)-1])
			}
			start = time.Now()
		}

	}
	for _, arr := range results {
		sort.Slice(arr, func(i, j int) bool {
			return arr[i].Used < arr[j].Used
		})
		// ColorL(u, "Min", arr[0], "Max", arr[len(arr)-1])
	}
	out = results
	// if jsonStr, err := json.Marshal(results); err == nil {
	// 	out = string(jsonStr)
	// }
	return
}

func (test *SpeedTest) testSpeed(urls []string, dirs ...string) {
	// var urls []string
	var err error
	if len(urls) == 0 {
		urls, err = test.GetHistoryUrls()
	}

	book := Book{
		books: make(map[string]Config),
	}
	book.Scan(dirs...)
	ssuris, err := book.Ls()

	if err != nil {
		return
	}
	for _, ssuri := range ssuris {
		config := new(Config)
		parseURI(ssuri, config)
		ColorL("test", len(urls), config.Server.(string))
		if config != nil {
			test.testSingleConfig(urls, config)
		}

	}
	test.message <- ConfigSpeed{
		Server: "end",
	}
}

func GetOldSpeedMap(filepath string) (out map[string][]ConfigSpeed) {
	if PathExists(filepath) {
		db, _ := os.Open(filepath)
		defer db.Close()
		dd, _ := ioutil.ReadAll(db)
		json.Unmarshal(dd, &out)
		return
	}
	return
}

func (test *SpeedTest) TestSpeed(urls []string, dirs ...string) (out map[string][]ConfigSpeed) {

	go test.testSpeed(urls, dirs...)
	ColorL(len(urls), dirs)
	return test.sortWithResult()
}

func (test *SpeedTest) testSingleConfig(urls []string, config *Config) (err error) {

	// setup a http client

	// set client to config

	var wait sync.WaitGroup
	ColorL("ready set config")
	test.SetConfig(config)

	for i, reqURL := range urls {
		wait.Add(1)
		go test.testUrl(reqURL, config.Server.(string), &wait)
		if i%20 == 0 && i > 0 {
			ColorL("ulimite", i)
			wait.Wait()
		}
	}
	wait.Wait()
	// test.message <- result
	return

}

// Returns an int >= min, < max
func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// Generate a random string of A-Z chars with len = l
func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return string(bytes)
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		return true
	default:
		return false
	}
}

func GenerateConfigJsons(dstRoot string) (err error) {
	user, _ := user.Current()
	home := user.HomeDir
	dbFile := filepath.Join(home, ".config", "cache.db")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter filter tag [like tokyo]: ")
	text, _ := reader.ReadString('\n')
	sqlStmt := `SELECT host,passwd,location FROM Host;`
	allwolrd := false
	if strings.TrimSpace(text) != "" && strings.TrimSpace(text) != "world" {
		if strings.HasPrefix(text, "!") {
			sqlStmt = `SELECT host,passwd,location FROM Host WHERE tag not like "%` + strings.TrimSpace(text) + `%" ;`
		} else {
			sqlStmt = `SELECT host,passwd,location FROM Host WHERE tag like "%` + strings.TrimSpace(text) + `%" ;`
		}
	} else if strings.TrimSpace(text) == "world" {
		allwolrd = true
		sqlStmt = `SELECT host,passwd,location FROM Host WHERE tag not like "%search%" ;`
	}
	cmd := exec.Command("sqlite3", dbFile, sqlStmt)
	var out bytes.Buffer
	cmd.Stdout = &out
	if !PathExists(dstRoot) {
		os.Mkdir(dstRoot, 0777)
	}
	filter := make(map[string]string)

	// oldExists := make(map[string]int)
	useChangeTime := 0

	ColorL("Used which turn as result : [0]")
	tmp, _ := reader.ReadString('\n')
	if strings.TrimSpace(tmp) != "" {
		useChangeTime, _ = strconv.Atoi(strings.TrimSpace(tmp))
	}

	useChangeMap := make(map[string]int)
	if err = cmd.Run(); err == nil {
		outstr := strings.TrimSpace(out.String())
		for _, r := range strings.Split(outstr, "\n") {
			rs := strings.SplitN(r, "|", 3)
			if rs[0] == "" {
				continue
			}
			target := filepath.Join(dstRoot, strings.ReplaceAll(strings.TrimSpace(rs[2]), " ", "_")+"-"+rs[0]+".json")
			// fmt.Println(target)
			if len(rs) < 2 {
				log.Fatal(sqlStmt)
			}
			if allwolrd {
				if _, ok := filter[rs[2]]; ok {
					if useChangeMap[rs[2]] == useChangeTime {
						continue
					} else {
						ColorL(filter[rs[2]], "=>", target)
						os.Remove(filter[rs[2]])
						filter[rs[2]] = target
						useChangeMap[rs[2]]++

					}
				} else {
					filter[rs[2]] = target
					useChangeMap[rs[2]] = 0
				}
			}

			if !PathExists(target) {
				c := Config{
					Server:         rs[0],
					ServerPassword: rs[1],
					ServerPort:     10443,
					Password:       randomString(9),
					Method:         "aes-256-cfb",
					Timeout:        600,
				}
				if err := c.ToFile(target); err != nil {
					log.Fatal(err)
				}
				if allwolrd {
					ColorL(rs[0], c.ToUri(), strings.TrimSpace(rs[2]))
				} else {
					fmt.Println(c.ToUri())
				}

			}

		}
		for coun, c := range useChangeMap {
			if c < useChangeTime {
				tar := filter[coun]
				os.Remove(tar)
				ColorL("Remove Old Route:", coun, tar)
			}
		}
	}
	return

}

// GetHistoryUrls s
// GetHistoryUrls s
func (test *SpeedTest) GetHistoryUrls() (history []string, err error) {
	res := make(map[string]int)
	user, _ := user.Current()
	home := user.HomeDir
	var chromePath string
	var dst string
	if runtime.GOOS == "darwin" {
		chromePath = filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "Default", "History")
		dst = "/tmp/history.url"
	} else if runtime.GOOS == "linux" {

	} else {
		chromePath = filepath.Join(home, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "History")
		dst = filepath.Join(home, "Desktop", "history.url")
	}
	if !PathExists(chromePath) {
		ColorL("chrome history not found", chromePath)
		return
	}
	ColorL("sqlite3", chromePath)

	sqlStmt := `SELECT url FROM urls;`
	// sqlStmt := `SELECT url AS last_visit_time FROM urls ORDER BY last_visit_time DESC;`

	cmd := exec.Command("sqlite3", chromePath, sqlStmt)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		return
	} else {
		for _, r := range strings.Split(out.String(), "\n") {
			// ColorL(r)
			if !strings.HasPrefix(r, "http") {
				continue
			}
			if u, err := url.Parse(strings.TrimSpace(r)); err == nil {
				if !strings.HasPrefix(u.Host, "192.168") {
					h := u.Scheme + "://" + u.Host
					if _, ok := res[h]; ok {
						continue
					} else {
						res[h] = 1
						history = append(history, h)
						ColorL(h)
					}
				}

			} else {
				log.Fatal("parse url", err)
			}

		}
		if err := ioutil.WriteFile(dst, []byte(strings.Join(history, "\n")), 0644); err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (test *SpeedTest) _test(addr string) (usedTime time.Duration, err error) {
	var dialer proxy.Dialer
	st := time.Now()
	dialer, err = proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
	httpTransport := &http.Transport{}
	httpTransport.Dial = dialer.Dial
	if err != nil {
		// fmt.Println("1")
		log.Println("create socks5 error:", err)
		return time.Duration(-1) * time.Second, err
	}
	httpClient := &http.Client{Transport: httpTransport, Timeout: 12 * time.Second}
	req, err := http.NewRequest("GET", "https://www.google.com", nil)

	if err != nil && !strings.Contains(err.Error(), "EOF") {
		// fmt.Println("1")

		log.Println("new req error:", err)
		return time.Duration(-1) * time.Second, err
	}
	resp, err := httpClient.Do(req)
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		// fmt.Println("2")

		// log.Println("send socks5 error:", err)
		return time.Duration(-1) * time.Second, err
	}
	if resp == nil {

		log.Println("eof error:", err)
		return time.Duration(-1) * time.Second, err
	}
	if resp.StatusCode/200 == 1 {
		err = nil
		return time.Now().Sub(st), err
	} else {
		// fmt.Println("3")

		log.Println("fuck error:", err)
		out, _ := ioutil.ReadAll(resp.Body)
		return time.Now().Sub(st), errors.New(string(out) + ":code:" + resp.Status)
	}

}

func (test *SpeedTest) testUrl(reqUrl, host string, wg *sync.WaitGroup) {
	var speed float32

	defer func() {
		wg.Done()

	}()
	dialer, err := proxy.SOCKS5("tcp", ProxyAddr, nil, proxy.Direct)
	httpTransport := &http.Transport{}
	httpTransport.Dial = dialer.Dial

	httpClient := &http.Client{Transport: globalTransport, Timeout: 8 * time.Second}
	for i := 0; i < 2; i++ {
		startAt := time.Now()
		var speedAt float32
		if err != nil {
			fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
			break
		}
		if req, err := http.NewRequest("GET", reqUrl, nil); err == nil {

			if resp, err := httpClient.Do(req); err == nil {
				if _, err := ioutil.ReadAll(resp.Body); err == nil {
				}
				speedAt = float32(time.Now().Sub(startAt)) / float32(time.Millisecond)
				resp.Body.Close()

			} else {
				// ColorL("-", "-", "err", err)
			}

		} else {
			// ColorL("-", "-", "err", err)
		}

		if speedAt == 0 {
			speed = 99999
			break
		} else {
			speed += speedAt
		}
	}
	u, _ := url.Parse(reqUrl)
	c := ConfigSpeed{
		Url:    GetMainDomain(u.Host),
		Used:   speed / float32(3),
		Server: host,
	}
	// ColorL(url, c.used)
	test.message <- c

}

func (test *SpeedTest) TestAllConfigs(howTest func(proaxyaddr string, config Config)) {
	// s, err := test.LsRoute()
	// if err != nil {
	// 	log.Fatal("ls route error:", err)
	// }
	// out := make(map[string]Config)
	// err = json.Unmarshal([]byte(s), &out)
	// if err != nil {
	// 	log.Fatal("ls route error json unmarshal:", err)
	// }
	var wait sync.WaitGroup
	num := 0
	for _, c := range BOOK.books {
		go func(paddr string, conf Config) {
			defer wait.Done()
			howTest(paddr, conf)
		}(fmt.Sprintf("localhost:%d", 30303+num), c)
		wait.Add(1)
		num++
	}
	wait.Wait()

}

func (test *SpeedTest) TestOneConfig(config *Config, addr string) {
	// test.SetConfig(config)
	count := 7
	type S struct {
		t    time.Duration
		addr string
		err  error
	}
	ones := make(chan S, count)

	for i := 0; i < count; i++ {
		go func() {
			out, err := test._test(addr)
			ones <- S{
				t:    out,
				addr: addr,
				err:  err,
			}
		}()
	}
	var all int64
	for i := 0; i < count; i++ {
		out := <-ones
		if out.err != nil {
			count--
		}
		all += out.t.Nanoseconds()
	}
	if count > 0 {
		ColorL(config.Server.(string), "avt:", time.Duration(all/int64(count)), "loss:", 7-count)
	} else {
		ColorL(config.Server.(string), "X")
	}
	// if err != nil && !strings.Contains(err.Error(), "EOF") {
	// 	ColorL(config.Server.(string), "err:", err)

	// } else {
	// 	ColorL(config.Server.(string), out)

	// }
	return
}

func TestURLUsedTime(requrl, host string) ConfigSpeed {
	var err error
	var speed float32
	if globalTransport == nil {

		dialer, _ := proxy.SOCKS5("tcp", TestProxyAddr, nil, proxy.Direct)
		httpTransport := &http.Transport{ResponseHeaderTimeout: 12 * time.Second}
		httpTransport.Dial = dialer.Dial

		globalTransport = httpTransport

	}
	httpClient := &http.Client{Transport: globalTransport, Timeout: 8 * time.Second}
	for i := 0; i < 2; i++ {
		startAt := time.Now()
		var speedAt float32
		if err != nil {
			fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
			break
		}
		if req, err := http.NewRequest("HEAD", requrl, nil); err == nil {

			if resp, err := httpClient.Do(req); err == nil {
				if _, err := ioutil.ReadAll(resp.Body); err == nil {
				}
				speedAt = float32(time.Now().Sub(startAt)) / float32(time.Millisecond)
				resp.Body.Close()

			} else {
				// ColorL("-", "-", "err", err)
			}
		} else {
			// ColorL("-", "-", "err", err)
		}

		if speedAt == 0 {
			speed = 99999
			break
		} else {
			speed += speedAt
		}
	}
	u, _ := url.Parse(requrl)
	return ConfigSpeed{
		Url:    GetMainDomain(u.Host),
		Used:   speed / float32(3),
		Server: host,
	}

}

// func httpTest(url) {
// 	c := &fasthttp.Client{
// 		Dial: fasthttpproxy.FasthttpSocksDialer(utils.ProxyAddr),
// 	}
// }
