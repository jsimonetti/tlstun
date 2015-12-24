package main

import (
	"fmt"
	"io"
	"net"

	"github.com/gorilla/websocket"
)

func print_binary(s []byte) {
	fmt.Printf("print b:")
	for n := 0; n < len(s); n++ {
		fmt.Printf("%d,", s[n])
	}
	fmt.Printf("\n")
}

func pipe(pconn *websocket.Conn, conn net.Conn) {
	go forwardtcp(pconn, conn)
	go forwardws(pconn, conn)

	//	go resendwstoconn(pconn, conn)
	//	go resendconntows(conn, pconn)
}

func forwardtcp(wsconn *websocket.Conn, conn net.Conn) {

	for {
		// Receive and forward pending data from tcp socket to web socket
		tcpbuffer := make([]byte, 1024)

		n, err := conn.Read(tcpbuffer)
		if err == io.EOF {
			fmt.Printf("TCP Read failed")
			break
		}
		if err == nil {
			fmt.Printf("Forwarding from tcp to ws: %d bytes: %s\n", n, tcpbuffer)
			print_binary(tcpbuffer)
			wsconn.WriteMessage(websocket.BinaryMessage, tcpbuffer[:n])
		}
	}
}

func forwardws(wsconn *websocket.Conn, conn net.Conn) {

	for {
		// Send pending data to tcp socket
		n, buffer, err := wsconn.ReadMessage()
		if err == io.EOF {
			fmt.Printf("WS Read Failed")
			break
		}
		if err == nil {
			s := string(buffer[:len(buffer)])
			fmt.Printf("Received (from ws) forwarding to tcp: %d bytes: %s %d\n", len(buffer), s, n)
			print_binary(buffer)
			conn.Write(buffer)
		}
	}
}

func resendwstoconn(in *websocket.Conn, out net.Conn) {
	Log("daemon", "debug", fmt.Sprintf("piping connection %s => %s", out.RemoteAddr(), in.RemoteAddr()))
	_, inreader, err := in.NextReader()

	if err != nil {
		Log("daemon", "debug", fmt.Sprintf("no reader for ws: %s", err))
		in.Close()
		out.Close()
		return
	}
	io.Copy(out, inreader)

	/*
		buf := make([]byte, 10240)
			for {
				n, err := inreader.Read(buf)
				if err == io.EOF {
					Log("daemon", "debug", fmt.Sprintf("connection ws2conn closed %s", out.RemoteAddr()))
					return
				} else if err != nil {
					Log("daemon", "error", fmt.Sprintf("resend ws2conn err: %s", err))
					return
				}
				fmt.Println("DEBUG: write to conn: %v", buf[:n])
				out.Write(buf[:n])
			}
	*/
}

func resendconntows(in net.Conn, out *websocket.Conn) {
	Log("daemon", "debug", fmt.Sprintf("piping connection %s => %s", out.RemoteAddr(), in.RemoteAddr()))
	outwriter, err := out.NextWriter(websocket.BinaryMessage)

	if err != nil {
		Log("daemon", "debug", fmt.Sprintf("no writer for ws: %s", err))
		in.Close()
		out.Close()
		return
	}
	io.Copy(outwriter, in)

	/*
		    	    buf := make([]byte, 10240)
				    	    for {
						n, err := in.Read(buf)
						if err == io.EOF {
							Log("daemon", "debug", fmt.Sprintf("connection conn2ws closed %s", in.RemoteAddr()))
							return
						} else if err != nil {
							Log("daemon", "error", fmt.Sprintf("resend conn2ws err: %s", err))
							return
						}
						fmt.Println("DEBUG: write to ws: %v", buf[:n])
						outwriter.Write(buf[:n])
					}
	*/
}
