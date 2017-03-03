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
		sent = PipeAndClose(src, dst)
		wg.Done()
	}()
	go func() {
		received = PipeAndClose(dst, src)
		wg.Done()
	}()
	wg.Wait()
	return
}

func PipeAndClose(src, dst net.Conn) (copied int64) {
	timeout := 10 * time.Second
	defer dst.Close()
	buf := make([]byte, 4096)
	for {
		n, err := src.Read(buf)
		copied += int64(n)
		if n > 0 {
			dst.SetWriteDeadline(time.Now().Add(timeout))
			if _, err := dst.Write(buf[0:n]); err != nil {
				break
			}
		}
		if err != nil {
			break
		}
	}
	return
}
