package kcpserver

import (
	"crypto/md5"
	"crypto/sha1"
	"errors"
	"io"
	"net"
	"strings"

	"github.com/Qingluan/Kcpee/shadowaead"
	"golang.org/x/crypto/hkdf"
)

type Cipher interface {
	StreamConnCipher
	PacketConnCipher
}

type StreamConnCipher interface {
	StreamConn(net.Conn) net.Conn
	NewStreamConn(c net.Conn) *shadowaead.StreamConn
	// StreamConn2(net.Conn) shadowaead.StreamConn
}

type PacketConnCipher interface {
	PacketConn(net.PacketConn) net.PacketConn
}

// ErrCipherNotSupported occurs when a cipher is not supported (likely because of security concerns).
var ErrCipherNotSupported = errors.New("cipher not supported")

const (
	aeadAes128Gcm        = "AEAD_AES_128_GCM"
	aeadAes256Gcm        = "AEAD_AES_256_GCM"
	aeadChacha20Poly1305 = "AEAD_CHACHA20_POLY1305"
)

// List of AEAD ciphers: key size in bytes and constructor
var aeadList = map[string]struct {
	KeySize int
	New     func([]byte) (shadowaead.Cipher, error)
}{
	aeadAes128Gcm:        {16, shadowaead.AESGCM},
	aeadAes256Gcm:        {32, shadowaead.AESGCM},
	aeadChacha20Poly1305: {32, shadowaead.Chacha20Poly1305},
}

// PickCipher returns a Cipher of the given name. Derive key from password if given key is empty.
func PickCipher(name string, key []byte, password string) (Cipher, error) {
	name = strings.ToUpper(name)

	switch name {
	case "DUMMY":
		return &dummy{}, nil
	case "CHACHA20-IETF-POLY1305":
		name = aeadChacha20Poly1305
	case "AES-128-GCM":
		name = aeadAes128Gcm
	case "AES-256-GCM":
		name = aeadAes256Gcm
	}

	if choice, ok := aeadList[name]; ok {
		if len(key) == 0 {
			key = kdf(password, choice.KeySize)
		}
		if len(key) != choice.KeySize {
			return nil, shadowaead.KeySizeError(choice.KeySize)
		}
		aead, err := choice.New(key)
		return &aeadCipher{aead}, err
	}

	return nil, ErrCipherNotSupported
}

type aeadCipher struct{ shadowaead.Cipher }

func (aead *aeadCipher) StreamConn2(c net.Conn) *shadowaead.StreamConn {
	return shadowaead.NewConn(c, aead).(*shadowaead.StreamConn)
}
func (aead *aeadCipher) StreamConn(c net.Conn) net.Conn { return shadowaead.NewConn(c, aead) }
func (aead *aeadCipher) PacketConn(c net.PacketConn) net.PacketConn {
	return shadowaead.NewPacketConn(c, aead)
}

func (aead *aeadCipher) NewStreamConn(c net.Conn) *shadowaead.StreamConn {
	return shadowaead.NewStreamConn(c, aead)
}

// dummy cipher does not encrypt
type dummy struct{}

func (dummy) StreamConn(c net.Conn) net.Conn                  { return c }
func (dummy) PacketConn(c net.PacketConn) net.PacketConn      { return c }
func (dummy) NewStreamConn(c net.Conn) *shadowaead.StreamConn { return nil }

// key-derivation function from original Shadowsocks
func kdf(password string, keyLen int) []byte {
	var b, prev []byte
	h := md5.New()
	for len(b) < keyLen {
		h.Write(prev)
		h.Write([]byte(password))
		b = h.Sum(b)
		prev = b[len(b)-h.Size():]
		h.Reset()
	}
	return b[:keyLen]
}

func hkdfSHA1(secret, salt, info, outkey []byte) {
	r := hkdf.New(sha1.New, secret, salt, info)
	if _, err := io.ReadFull(r, outkey); err != nil {
		panic(err) // should never happen
	}
}

// func Decrypter(psk []byte, salt []byte) (cipher.AEAD, error) {
// 	subkey := make([]byte, len(psk))
// 	hkdfSHA1(psk, salt, []byte("ss-subkey"), subkey)
// 	return a.makeAEAD(subkey)
// }
