package shared

import (
	//	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var WsTimeOut int

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
			defer wsconn.Close()
			defer conn.Close()
			break
		} else if err != nil {
			//Log("daemon", "info", fmt.Sprintf("error reading net: %s", err))
			defer wsconn.Close()
			defer conn.Close()
			break
		}
		wsconn.WriteMessage(websocket.BinaryMessage, tcpbuffer[:n])
	}
}

func forwardws(wsconn *websocket.Conn, conn net.Conn) {

	timeout := time.Duration(WsTimeOut) * time.Second

	for {
		wsconn.SetReadDeadline(time.Now().Add(timeout))
		// Send pending data to tcp socket
		_, buffer, err := wsconn.ReadMessage()
		if err == io.EOF {
			Log("daemon", "info", "ws connection closed")
			defer wsconn.Close()
			defer conn.Close()
			break
		} else if err != nil {
			//Log("daemon", "info", fmt.Sprintf("error reading websocket: %s", err))
			defer wsconn.Close()
			defer conn.Close()
			break
		}
		conn.Write(buffer)
	}
}
