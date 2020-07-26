package utils

import (
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/fatih/color"
)

type LeakyBuf struct {
	bufSize  int // size of each buffer
	freeList chan []byte
}

// const leakyBufSize = 1348 // kcp len is 1348
const leakyBufSize = 4108 // kcp len is 1348
const maxNBuf = 2048

var GloballeakyBuf = NewLeakyBuf(maxNBuf, leakyBufSize)
var END_BYTES = []byte{91, 91, 69, 79, 70, 93, 93}

// NewLeakyBuf creates a leaky buffer which can hold at most n buffer, each
// with bufSize bytes.
func NewLeakyBuf(n, bufSize int) *LeakyBuf {
	return &LeakyBuf{
		bufSize:  bufSize,
		freeList: make(chan []byte, n),
	}
}

// Get returns a buffer from the leaky buffer or create a new buffer.
func (lb *LeakyBuf) Get() (b []byte) {
	select {
	case b = <-lb.freeList:
	default:
		b = make([]byte, lb.bufSize)
	}
	return
}

// Put add the buffer into the free buffer pool for reuse. Panic if the buffer
// size is not the same with the leaky buffer's. This is intended to expose
// error usage of leaky buffer.
func (lb *LeakyBuf) Put(b []byte) {
	if len(b) != lb.bufSize {
		panic("invalid buffer size that's put into leaky buffer")
	}
	select {
	case lb.freeList <- b:
	default:
	}
	return
}

type CanCopy interface {
	io.Reader
	io.Writer
	io.Closer
}

func PipeThenClose(src, dst CanCopy) {
	defer func() {
		dst.Close()
		// io.Close(dst)
	}()
	io.Copy(dst, src)
}

func Tcp2Kcp(src, dst io.ReadWriteCloser) {
	s := 0
	start_at := time.Now()
	defer func() {
		// time.Sleep(1)
		// log.Println("tcp2kcp closed")

		dst.Write(END_BYTES)
		end_at := time.Now()
		log.Println("sock :[", s, "]", end_at.Sub(start_at))
		end_buf := make([]byte, 2)
		dst.Read(end_buf)

		dst.Close()
		src.Close()
		log.Println("-- END --")

	}()

	// EOF_DELAY_TIMEOUT := 3000
	buf := GloballeakyBuf.Get()
	defer GloballeakyBuf.Put(buf)
	for {
		n, err := src.Read(buf)
		s += n
		if n > 0 {
			// Note: avoid overwrite err returned by Read.

			// log.Println(src.LocalAddr(), "tcp -> kcp ", dst.RemoteAddr().String(), " : [", n, "]", "md5:", Md5Str(buf[0:n]))

			// ensure write all buf to dst
			var wn int
			var wnTmp int
			var err error
			if wn, err = dst.Write(buf[0:n]); err != nil {
				Debug.Println("write:", err)
				break
			}
			for {

				if wn >= n {
					break
				}
				wnTmp, err = dst.Write(buf[wn:n])
				if err != nil {
					Debug.Println("write:", err)
					break
				}
				wn += wnTmp
			}
			// if Debug {
			log.Println(time.Now().Sub(start_at), "T -> K : [", n, "] md5:", Md5Str(buf[0:n]))
			// }

		}
		if err != nil {
			// Always "use of closed network connection", but no easy way to
			// identify this specific error. So just leave the error along for now.
			// More info here: https://code.google.com/p/go/issues/detail?id=4373
			/*
				if bool(Debug) && err != io.EOF {
					Debug.Println("read:", err)
				}
			*/

			// time.Sleep(time.Duration(EOF_DELAY_TIMEOUT))
			// log.Println(err)
			break
		}
	}

}

// Copy write by my self
func LogCopy(src, dst net.Conn) {
	closed := false
	s := 0
	startAt := time.Now()
	defer func() {
		endAt := time.Now()
		log.Println(src.RemoteAddr(), "-->", dst.RemoteAddr(), " pass : ", s, "used:", endAt.Sub(startAt))
		dst.Close()
	}()

	buf := GloballeakyBuf.Get()
	defer GloballeakyBuf.Put(buf)
	for {
		n, err := src.Read(buf)
		s += n
		if n > 0 {
			log.Println(FGCOLORS[1](time.Now().Sub(startAt)), FGCOLORS[0](src.RemoteAddr()), "->", FGCOLORS[0](dst.RemoteAddr()), " : [", n, "]", "md5:", FGCOLORS[2](Md5Str(buf[0:n])))
			if _, err = dst.Write(buf[0:n]); err != nil {
				break
			}
		} else {
			closed = true
		}

		if closed {
			log.Println("END")
			break
		}
	}
}

// const bufSize = 4096
const bufSize = 8192

// Memory optimized io.Copy function specified for this library
func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}

	// fallback to standard io.CopyBuffer
	buf := make([]byte, bufSize)
	return io.CopyBuffer(dst, src, buf)
}

// func CopyLog(dst io.Writer, src io.Reader) (written int64, err error) {
// 	defer dst.Close()
// 	buf := make([]byte, 4096)
// 	if n, err := src.ReadFull(buf){

// 	}

// }

func Pipe(p1, p2 net.Conn) (err error) {
	// start tunnel & wait for tunnel termination
	streamCopy := func(dst io.Writer, src io.ReadCloser, fr, to net.Addr) error {
		// startAt := time.Now()
		_, err := Copy(dst, src)

		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
			} else if strings.Contains(err.Error(), "EOF") {
			} else if strings.Contains(err.Error(), "read/write on closed pipe") {
			} else {
				r := color.New(color.FgRed)
				r.Println("error : ", err)
			}

		}
		// endAt := time.Now().Sub(startAt)
		// log.Println("passed:", FGCOLORS[1](n), FGCOLORS[0](p1.RemoteAddr()), "->", FGCOLORS[0](p2.RemoteAddr()), "Used:", endAt)
		p1.Close()
		p2.Close()
		return err
		// }()
	}
	go streamCopy(p1, p2, p2.RemoteAddr(), p1.RemoteAddr())
	err = streamCopy(p2, p1, p1.RemoteAddr(), p2.RemoteAddr())
	return
}
