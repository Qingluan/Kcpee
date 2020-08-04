package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	// "log"
	"os"
	"reflect"
	"time"

	"github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
	"golang.org/x/crypto/pbkdf2"
)

const (
	NORMAL_SCAN_DIR = "/etc/shadowsocks/:/tmp/_configs/"
)

var (
	readTimeout time.Duration
	TODAY       time.Time = time.Now()
	BOOK        Book      = Book{
		books:  make(map[string]Config),
		lastId: 0,
	}
	HOME, _              = os.UserHomeDir()
	NORMAL_OPTIMISE_FILE = filepath.Join(HOME, "Desktop", "route.map.json")
	NORMAL_CONFIG_ROOT   = filepath.Join(HOME, "Desktop", "routes")
)

// Config can use file to load
type Config struct {
	Server       interface{} `json:"server"`
	ServerPort   int         `json:"server_port"`
	LocalPort    int         `json:"local_port"`
	LocalAddress string      `json:"local_address"`
	Password     string      `json:"password"`
	Method       string      `json:"method"` // encryption method

	// following options are only used by server
	PortPassword map[string]string `json:"port_password"`
	Timeout      int               `json:"timeout"`

	// following options are only used by client

	// The order of servers in the client config is significant, so use array
	// instead of map to preserve the order.
	ServerPassword string `json:"server_password"`
}

// GetBook reutrn book
func (config *Config) GetBook(key string) (book *Config) {
	if b, ok := BOOK.books[key]; ok {
		book = &b
	}
	return
}

type TlsConfig struct {
	Ca     x509.CertPool
	Cert   tls.Certificate
	Server string
	priKey rsa.PrivateKey
}

func (tlsConfig *TlsConfig) GenerateConfig() (config tls.Config) {
	// tlsConfig.Ca.AppendCertsFromPEM(tlsConfig.Cert.)
	config = tls.Config{
		Certificates: []tls.Certificate{tlsConfig.Cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    &tlsConfig.Ca,
	}
	config.Rand = rand.Reader
	return
}

func (tlsConfig *TlsConfig) WithConn() (conn *tls.Conn, err error) {
	config := tls.Config{
		Certificates:       []tls.Certificate{tlsConfig.Cert},
		InsecureSkipVerify: true,
	}
	serverAddress := tlsConfig.Server
	conn, err = tls.Dial("tcp", serverAddress, &config)
	if err != nil {
		log.Println("tls connect:", serverAddress)
		return
	}
	state := conn.ConnectionState()
	// for _, v := range state.PeerCertificates {
	// 	log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
	// }
	if !state.HandshakeComplete {
		return nil, errors.New("Not TLS Handleshare finished!!")
	}

	return
}

func (tlsConfig *TlsConfig) WithTlsListener() (listenr net.Listener, err error) {
	config := tlsConfig.GenerateConfig()
	ColorL(tlsConfig.Server)
	listenr, err = tls.Listen("tcp", tlsConfig.Server, &config)
	return
}

func (config *Config) ToTlsConfig() (tlsConfig *TlsConfig, err error) {
	if config.Method != "tls" {
		return
	}
	tlsConfig = new(TlsConfig)
	tlsConfig.Server = fmt.Sprintf("%s:%d", config.Server.(string), config.ServerPort)

	// ColorL("raw:", config.Password)
	pems := strings.SplitN(config.Password, "<SEP>", 2)

	pemBlock := []byte(strings.TrimSpace(pems[0]))
	keyBlock := []byte(strings.TrimSpace(pems[1]))

	// preName := ".tmp." + strconv.Itoa(random.Int())

	// ioutil.WriteFile(preName+".pem", pemBlock, os.ModePerm)
	// ioutil.WriteFile(preName+".key", keyBlock, os.ModePerm)
	// defer os.Remove(preName + ".pem")
	// defer os.Remove(preName + ".key")
	// crtx, err2 := x509.ParseCertificate(pemBlock.Bytes)
	// crt, err2 := tls.LoadX509KeyPair(preName+".pem", preName+".key")
	crt, err2 := tls.X509KeyPair(pemBlock, keyBlock)
	if err2 != nil {
		ColorE("parir error:", err2)
		return nil, err2
	}

	tlsConfig.Cert = crt
	// tlsConfig.priKey = *key
	tlsConfig.Ca = *x509.NewCertPool()
	tlsConfig.Ca.AppendCertsFromPEM(pemBlock)

	return
}

func (config *Config) ToString() string {
	return fmt.Sprintf("%s:%d", config.Server.(string), config.ServerPort)
}

func (config *Config) ToFile(dst string) (err error) {
	if f, err := json.Marshal(config); err == nil {
		if err := ioutil.WriteFile(dst, f, 0644); err != nil {
			return err
		}
	} else {
		return err
	}
	return
}

func (config *Config) ToJson() string {
	if f, err := json.Marshal(config); err == nil {
		return string(f)
	}
	return ""
}

func (config *Config) ToUri() string {
	base := fmt.Sprintf("%s:%s@%s:%d", config.Method, config.Password, config.Server.(string), config.ServerPort)
	encoder := base64.StdEncoding.EncodeToString([]byte(base))
	return fmt.Sprintf("ss://%s", encoder)
}

// GetBookByID return a book from BOOK
func (config *Config) GetBookByID(id uint16) (book *Config) {
	if len(BOOK.books) > 0 {

		for _, _book := range BOOK.books {
			if id == BOOK.lastId {
				book = &_book
			}
		}

	}
	return
}

// GeneratePassword by config
func (config *Config) GeneratePassword() (en kcp.BlockCrypt) {
	klen := 32
	if strings.Contains(config.Method, "128") {
		klen = 16
	}
	mainMethod := strings.Split(config.Method, "-")[0]
	keyData := pbkdf2.Key([]byte(config.Password), []byte("demo salt"), 1024, klen, sha1.New)

	switch mainMethod {

	case "des":
		en, _ = kcp.NewTripleDESBlockCrypt(keyData[:klen])
	case "tea":
		en, _ = kcp.NewTEABlockCrypt(keyData[:klen])
	case "simple":
		en, _ = kcp.NewSimpleXORBlockCrypt(keyData[:klen])
	case "xtea":
		en, _ = kcp.NewXTEABlockCrypt(keyData[:klen])
	default:
		en, _ = kcp.NewAESBlockCrypt(keyData[:klen])
	}

	return
}

// GetServerArray get server
func (config *Config) GetServerArray() []string {
	// Specifying multiple servers in the "server" options is deprecated.
	// But for backward compatibility, keep this.
	if config.Server == nil {
		return nil
	}
	single, ok := config.Server.(string)
	if ok {
		return []string{single}
	}
	arr, ok := config.Server.([]interface{})
	if ok {
		serverArr := make([]string, len(arr), len(arr))
		for i, s := range arr {
			serverArr[i], ok = s.(string)
			if !ok {
				goto typeError
			}
		}
		return serverArr
	}
typeError:
	panic(fmt.Sprintf("Config.Server type error %v", reflect.TypeOf(config.Server)))
}

// ParseConfig parse path to json
func ParseConfig(path string) (config *Config, err error) {
	file, err := os.Open(path) // For read access.
	if err != nil {
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	config = &Config{}
	if err = json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	readTimeout = time.Duration(config.Timeout) * time.Second
	return
}

// SetDebug if ok
func SetDebug(d DebugLog) {
	Debug = d
}

// UpdateConfig : Useful for command line to override options specified in config file  Debug is not updated.
func UpdateConfig(old, new *Config) {
	// Using reflection here is not necessary, but it's a good exercise.
	// For more information on reflections in Go, read "The Laws of Reflection"
	// http://golang.org/doc/articles/laws_of_reflection.html
	newVal := reflect.ValueOf(new).Elem()
	oldVal := reflect.ValueOf(old).Elem()

	// typeOfT := newVal.Type()
	for i := 0; i < newVal.NumField(); i++ {
		newField := newVal.Field(i)
		oldField := oldVal.Field(i)
		// log.Printf("%d: %s %s = %v\n", i,
		// typeOfT.Field(i).Name, newField.Type(), newField.Interface())
		switch newField.Kind() {
		case reflect.Interface:
			if fmt.Sprintf("%v", newField.Interface()) != "" {
				oldField.Set(newField)
			}
		case reflect.String:
			s := newField.String()
			if s != "" {
				oldField.SetString(s)
			}
		case reflect.Int:
			i := newField.Int()
			if i != 0 {
				oldField.SetInt(i)
			}
		}
	}

	old.Timeout = new.Timeout
	readTimeout = time.Duration(old.Timeout) * time.Second
}

func GetMainDomain(urlOrHost string) string {
	host := urlOrHost
	if strings.HasPrefix(urlOrHost, "http") {
		u, _ := url.Parse(urlOrHost)
		host = u.Host
	}
	dotCount := strings.Count(host, ".")
	if dotCount > 1 {
		return strings.Join(strings.Split(host, ".")[dotCount-1:], ".")
	} else {
		return host
	}
}

func ParseURI(u string) (config *Config) {
	config = new(Config)
	parseURI(u, config)
	return
}

func parseURI(u string, cfg *Config) (string, error) {
	if u == "" {
		return "", nil
	}
	invalidURI := errors.New("invalid URI")
	// ss://base64(method:password)@host:port
	// ss://base64(method:password@host:port)
	u = strings.TrimLeft(u, "ss://")
	i := strings.IndexRune(u, '@')
	var headParts, tailParts [][]byte
	if i == -1 {
		dat, err := base64.StdEncoding.DecodeString(u)
		if err != nil {
			return "", err
		}
		parts := bytes.Split(dat, []byte("@"))
		if len(parts) != 2 {
			return "", invalidURI
		}
		headParts = bytes.SplitN(parts[0], []byte(":"), 2)
		tailParts = bytes.SplitN(parts[1], []byte(":"), 2)

	} else {
		if i+1 >= len(u) {
			return "", invalidURI
		}
		tailParts = bytes.SplitN([]byte(u[i+1:]), []byte(":"), 2)
		dat, err := base64.StdEncoding.DecodeString(u[:i])
		if err != nil {
			return "", err
		}
		headParts = bytes.SplitN(dat, []byte(":"), 2)
	}
	if len(headParts) != 2 {
		return "", invalidURI
	}

	if len(tailParts) != 2 {
		return "", invalidURI
	}
	cfg.Method = string(headParts[0])

	cfg.Password = string(headParts[1])
	p, e := strconv.Atoi(string(tailParts[1]))
	if e != nil {
		return "", e
	}
	cfg.Server = string(tailParts[0])
	cfg.ServerPort = p
	return string(tailParts[0]), nil

}

type KcpConfig struct {
	Mode         string `json:"mode"`
	NoDelay      int    `json:"nodelay"`
	Interval     int    `json:"interval"`
	Resend       int    `json:"resend"`
	NoCongestion int    `json:"nocongeestion"`
	AutoExpire   int    `json:"autoexpire"`
	ScavengeTTL  int    `json:"scavengettl"`
	MTU          int    `json:"mtu"`
	SndWnd       int    `json:"sndwnd"`
	RcvWnd       int    `json:"rcvwnd"`
	DataShard    int    `json:"datashard"`
	ParityShard  int    `json:"parityshard"`
	KeepAlive    int    `json:"keepalive"`
	SmuxBuf      int    `json:"smuxbuf"`
	StreamBuf    int    `json:"streambuf"`
	AckNodelay   bool   `json:"acknodelay"`
}

func (kconfig *KcpConfig) SetAsDefault() {
	kconfig.Mode = "fast2"
	kconfig.KeepAlive = 10
	kconfig.MTU = 1400
	kconfig.DataShard = 10
	kconfig.ParityShard = 3
	kconfig.SndWnd = 4096
	kconfig.RcvWnd = 4096
	kconfig.ScavengeTTL = 600
	kconfig.AutoExpire = 7
	kconfig.SmuxBuf = 32777217
	kconfig.StreamBuf = 2097152
	kconfig.AckNodelay = false
}

func (kconfig *KcpConfig) UpdateMode() {
	// kconfig.Mode = mode
	switch kconfig.Mode {
	case "normal":
		kconfig.NoDelay, kconfig.Interval, kconfig.Resend, kconfig.NoCongestion = 0, 40, 2, 1
	case "fast":
		kconfig.NoDelay, kconfig.Interval, kconfig.Resend, kconfig.NoCongestion = 0, 30, 2, 1
	case "fast2":
		kconfig.NoDelay, kconfig.Interval, kconfig.Resend, kconfig.NoCongestion = 1, 20, 2, 1
	case "fast3":
		kconfig.NoDelay, kconfig.Interval, kconfig.Resend, kconfig.NoCongestion = 1, 10, 2, 1
	}
	ColorL("kcp mode", kconfig.Mode)
}

func (kconfig *KcpConfig) GenerateConfig() *smux.Config {
	smuxConfig := smux.DefaultConfig()
	kconfig.UpdateMode()
	smuxConfig.MaxReceiveBuffer = kconfig.SmuxBuf
	// smuxConfig.MaxStreamBuffer = kconfig.StreamBuf
	smuxConfig.KeepAliveInterval = time.Duration(kconfig.KeepAlive) * time.Second
	if err := smux.VerifyConfig(smuxConfig); err != nil {
		log.Fatalf("%+v", err)
	}
	return smuxConfig
}
