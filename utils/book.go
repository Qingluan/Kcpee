package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const DEFAULT_TIMEOUT = 86400

type Route struct {
	config *Config
	ttl    time.Time
	mode   string
}

var AutoMap map[string][]ConfigSpeed

func (route *Route) GetConfig() *Config {
	return route.config
}

func (route *Route) SetConfig(config *Config) {
	route.config = config
}

func NewRoute(cmd string) (route *Route, err error) {
	route = new(Route)
	route.mode = "proxy"
	if cmd == "start" {
		config := BOOK.Get()
		route.config = config
		route.ttl = time.Now().Add(time.Duration(DEFAULT_TIMEOUT) * time.Second)
	} else if strings.Contains(cmd, "@") {
		parts := strings.Split(cmd, "@")
		switch len(parts) {
		case 2:
			if config, ok := BOOK.books[parts[1]]; ok {
				route.config = &config
			} else {
				err = errors.New(fmt.Sprintf("not found %s", cmd))
			}
			route.ttl = time.Now().Add(time.Duration(86400) * time.Second)
		case 3:
			if config, ok := BOOK.books[parts[1]]; ok {
				route.config = &config
			} else {
				err = errors.New(fmt.Sprintf("not found %s", cmd))
			}
			if timeout, ierr := strconv.Atoi(parts[2]); err == nil {
				route.ttl = time.Now().Add(time.Duration(timeout) * time.Second)
			} else {
				err = ierr
			}
		}
	} else if cmd == "TUNNEL" {
		route.mode = "tunnel"
		route.ttl = time.Now().Add(time.Duration(DEFAULT_TIMEOUT) * time.Second)

	} else if strings.HasPrefix(cmd, "cc://") {
		route.config = &Config{
			Server: strings.SplitN(cmd, "cc://", 2)[1],
		}
		route.mode = "connect"
		route.ttl = time.Now().Add(time.Duration(DEFAULT_TIMEOUT) * time.Second)

	}
	return
}

func (route *Route) SetMode(mode string) {
	route.mode = mode
}

func (route *Route) Mode() string {
	return route.mode
}

func (route *Route) SetExireTime(sec int) {
	route.ttl = time.Now().Add(time.Duration(sec) * time.Second)
}

func (route *Route) Host() string {
	return fmt.Sprintf("%s:%d", route.config.Server.(string), route.config.ServerPort)
}

func (route *Route) Left() int {
	now := time.Now()
	d := route.ttl.Sub(now)
	return int(d / time.Second)
}

func (route *Route) IfNoExpired() bool {
	return !time.Now().After(route.ttl)
}

func (book *Book) Books() (cons []Config) {
	for _, v := range book.books {
		cons = append(cons, v)
	}
	return
}

type Book struct {
	books  map[string]Config
	lastId uint16
}

func NewBook() *Book {

	book := Book{
		books:  make(map[string]Config),
		lastId: 0,
	}
	return &book
}
func (book *Book) GetServers() (s []string) {
	for _, v := range book.books {
		s = append(s, fmt.Sprintf("%s:%d", v.Server.(string), v.ServerPort))
	}
	return
}

func (book *Book) Ls() (cs []string, err error) {
	for _, v := range book.books {
		cs = append(cs, v.ToUri())
	}
	return
}

func (book *Book) Info() (cs []string, err error) {
	for _, v := range book.books {
		_cts := strings.Split(v.LocalAddress, "/")
		cs = append(cs, "["+_cts[len(_cts)-1]+"] "+v.Server.(string)+" : "+v.ToUri())
	}
	return
}

func (book *Book) Add(ssuri string) (s string, err error) {
	config := new(Config)
	if s, err = parseURI(ssuri, config); err != nil {
		log.Fatal("add:", err)
	} else {
		// s = fmt.Sprintf("%s:%d", s, config.ServerPort)
		ColorL("add route:", s, config)
		book.books[s] = *config
	}
	return
}

func (book *Book) FlowGet() (config *Config) {
	lastUse := book.lastId
	c := 0
	AllLen := len(book.books)
	for _, v := range book.books {
		if c == (int(lastUse)+1)%AllLen {
			config = &v
			book.lastId = uint16((int(book.lastId) + 1) % AllLen)
			return
		}
		c++

	}
	return book.Get()
}

func (book *Book) Get(server ...string) (config *Config) {
	if len(server) > 0 {
		if s, ok := book.books[server[0]]; ok {
			config = &s
		} else {
			return
		}
	} else {
		if len(book.books) == 0 {
			return
		}
		id := TODAY.Day() % len(book.books)
		c := 0
		for _, v := range book.books {
			if c == id {
				config = &v
				break
			}
			c++
		}
	}
	return
}

func (book *Book) RandGet(server ...string) (config *Config) {
	c := 0
	id := randomInt(0, len(book.books)-1)
	for _, v := range book.books {
		if c == id {
			config = &v
			break
		}
		c++
	}
	return
}

func (book *Book) Count() int {
	return len(book.books)
}

// Scan config in dir try to found correct json files
func (book *Book) Scan(dirs ...string) (dirErr error) {
	if len(dirs) == 0 {
		dirs = strings.Split(NORMAL_SCAN_DIR, ":")
	}
	var routeMapFile string
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		dirErr = filepath.Walk(dir, func(path string, f os.FileInfo, ierr error) (err error) {
			if f == nil {
				return ierr
			}
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
				if strings.Contains(path, "route.map.json") {
					routeMapFile = path
				} else {
					if c, err := ParseConfig(path); err == nil {
						// key := fmt.Sprintf("%s:%d", c.Server.(string), c.ServerPort)
						key := c.Server.(string)
						if strings.Contains(path, "-") {
							ps := strings.SplitN(path, "-", 2)
							c.LocalAddress = ps[0]
						}
						if _, ok := book.books[key]; !ok {
							book.books[key] = *c
							ColorL("found config: ", key, "Pwd:", c.Password, "Port:", c.ServerPort)
						} else {
							log.Println("exists : ", key)
						}
					}
				}

			}
			return
		})
	}
	if PathExists(routeMapFile) {
		if db, err := os.Open(routeMapFile); err == nil {
			defer db.Close()
			// buffer := bufio.NewReader(db)
			if data, err := ioutil.ReadAll(db); err == nil {
				json.Unmarshal(data, &AutoMap)
				ColorL("use optimise file to auto route multi server", fmt.Sprint("load: ", len(AutoMap)))
			}
		}
	} else if PathExists(NORMAL_OPTIMISE_FILE) {
		if db, err := os.Open(NORMAL_OPTIMISE_FILE); err == nil {
			// buffer := bufio.NewReader(db)
			defer db.Close()
			if data, err := ioutil.ReadAll(db); err == nil {
				json.Unmarshal(data, &AutoMap)
				ColorL("use optimise file to auto route multi server", fmt.Sprint("load: ", len(AutoMap)))
			}
		}
	} else if PathExists("route.map.json") {
		if db, err := os.Open("route.map.json"); err == nil {
			// buffer := bufio.NewReader(db)
			defer db.Close()
			if data, err := ioutil.ReadAll(db); err == nil {
				json.Unmarshal(data, &AutoMap)
				ColorL("use optimise file to auto route multi server", fmt.Sprint("load: ", len(AutoMap)))
			}
		}
	}

	return
}

func DeepCopy(dst interface{}, src interface{}) {
	data, _ := json.Marshal(src)
	json.Unmarshal(data, dst)
}
