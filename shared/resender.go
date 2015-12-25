package shared

import (
	"io"
	"net"
	"sync"

	"github.com/gorilla/websocket"
)

func NewPipe(src io.ReadWriteCloser, dst io.ReadWriteCloser) (int64, int64) {

	var sent, received int64
	var c = make(chan bool)
	var o sync.Once

	close := func() {
		src.Close()
		dst.Close()
		close(c)
	}

	go func() {
		received, _ = io.Copy(src, dst)
		o.Do(close)
	}()

	go func() {
		sent, _ = io.Copy(dst, src)
		o.Do(close)
	}()

	<-c
	return received, sent
}

func OldPipe(wsconn *websocket.Conn, conn net.Conn) {
	go forwardconn(wsconn, conn)
	go forwardws(wsconn, conn)
}

func forwardconn(wsconn *websocket.Conn, conn net.Conn) {

	for {
		// Receive and forward pending data from tcp socket to web socket
		tcpbuffer := make([]byte, 1024*1024)

		n, err := conn.Read(tcpbuffer)
		if err == io.EOF {
			Log("daemon", "info", "net connection closed")
			wsconn.Close()
			break
		}
		if err == nil {
			//			Log("daemon", "info", fmt.Sprintf("Forwarding from tcp to ws: %d bytes", n))
			wsconn.WriteMessage(websocket.BinaryMessage, tcpbuffer[:n])
		}
	}
}

func forwardws(wsconn *websocket.Conn, conn net.Conn) {

	for {
		// Send pending data to tcp socket
		_, buffer, err := wsconn.ReadMessage()
		if err == io.EOF {
			Log("daemon", "info", "ws connection closed")
			conn.Close()
			break
		}
		if err == nil {
			//			Log("daemon", "info", fmt.Sprintf("Forwarding from tcp to ws: %d bytes", n))
			conn.Write(buffer)
		}
	}
}
