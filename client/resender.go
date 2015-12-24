package main

import (
	"io"
	"net"

	"github.com/gorilla/websocket"
)

func pipe(pconn *websocket.Conn, conn net.Conn) {
	go forwardconn(pconn, conn)
	go forwardws(pconn, conn)
}

func forwardconn(wsconn *websocket.Conn, conn net.Conn) {

	for {
		// Receive and forward pending data from tcp socket to web socket
		tcpbuffer := make([]byte, 1024)

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
