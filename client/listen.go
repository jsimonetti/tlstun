package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"

	"github.com/gorilla/websocket"
)

var wsServer string = "127.0.0.1:12345"

func recv(buf []byte, m int, conn net.Conn) (n int, err error) {
	for nn := 0; n < m; {
		nn, err = conn.Read(buf[n:m])
		if nil != err && io.EOF != err {
			Log("daemon", "error", fmt.Sprintf("err: %s", err))
			panic(err)
			return
		}
		n += nn
	}
	return
}

func handleConn(conn net.Conn) {
	Log("client", "info", fmt.Sprintf("accepted connection from: %s", conn.RemoteAddr()))

	var clhello clHello
	var clecho clEcho
	var clrequest clRequest
	var clresponse clResponse

	//recv hello
	var err error = nil
	err = clhello.read(conn)
	if nil != err {
		return
	}
	clhello.print()

	//send echo
	clecho.gen(0)
	clecho.write(conn)
	clecho.print()

	//recv request
	err = clrequest.read(conn)
	if nil != err {
		return
	}
	clrequest.print()
	//connect
	durl := []byte(clrequest.url)
	parameters := base64.StdEncoding.EncodeToString(durl)

	wsurl := fmt.Sprintf("ws://%s/sock/%d/%s/%s", wsServer, clrequest.version, clrequest.reqtype, parameters)
	Log("daemon", "debug", fmt.Sprintf("dailing: %s", wsurl))

	pconn, _, err := websocket.DefaultDialer.Dial(wsurl, nil)
	if err != nil {
		Log("daemon", "fatal", fmt.Sprintf("ws dailer failed: %s", err))
	}

	//reply
	//error occur
	if nil != err {
		clresponse.gen(&clrequest, 4)
		clresponse.write(conn)
		clresponse.print()
		return
	}
	//success
	clresponse.gen(&clrequest, 0)
	clresponse.write(conn)
	clresponse.print()

	pipe(pconn, conn)
}

func forward(wss string) {
	wsServer = wss
	addr := fmt.Sprintf("%s:%d", listenIp, listenPort)
	ln, err := net.Listen("tcp", addr)
	if nil != err {
		Log("daemon", "error", "Bind Error!")
		return
	}
	Log("daemon", "info", fmt.Sprintf("listening for connections on: %s", addr))

	for {
		conn, err := ln.Accept()
		if nil != err {
			Log("daemon", "error", "Accept Error!")
			continue
		}

		go handleConn(conn)
	}
}
