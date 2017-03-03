package client

import (
	"net"
	"sync"
	"time"
)

func Pipe(src, dst net.Conn) (sent, received int64) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		sent = PipeThenClose(src, dst)
		wg.Done()
	}()
	go func() {
		received = PipeThenClose(dst, src)
		wg.Done()
	}()
	wg.Wait()
	return
}

// below is taken from shadowsocks-go (github.com/shadowsocks/shadowsocks-go)
// slightly modified to return values and use a writetimeout

var writeTimeout = 10 * time.Second

func SetWriteTimeout(c net.Conn) {
	if writeTimeout != 0 {
		c.SetWriteDeadline(time.Now().Add(writeTimeout))
	}
}

// PipeThenClose copies data from src to dst, closes dst when done.
func PipeThenClose(src, dst net.Conn) (copied int64) {

	defer dst.Close()
	buf := leakyBuf.Get()
	defer leakyBuf.Put(buf)
	for {
		n, err := src.Read(buf)
		copied += int64(n)
		// read may return EOF with n > 0
		// should always process n > 0 bytes before handling error
		if n > 0 {
			SetWriteTimeout(dst)
			// Note: avoid overwrite err returned by Read.
			if _, err := dst.Write(buf[0:n]); err != nil {
				break
			}
		}
		if err != nil {
			// Always "use of closed network connection", but no easy way to
			// identify this specific error. So just leave the error along for now.
			// More info here: https://code.google.com/p/go/issues/detail?id=4373
			break
		}
	}
	return
}

type LeakyBuf struct {
	bufSize  int // size of each buffer
	freeList chan []byte
}

const leakyBufSize = 4108 // data.len(2) + hmacsha1(10) + data(4096)
const maxNBuf = 2048

var leakyBuf = NewLeakyBuf(maxNBuf, leakyBufSize)

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
