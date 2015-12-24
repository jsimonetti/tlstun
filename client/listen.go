package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"

	"github.com/gorilla/websocket"

	"github.com/jsimonetti/tlstun/shared"
)

var wsServer string = "127.0.0.1:12345"

var tlsConfig *tls.Config
var certf string
var keyf string

func readMyCert() (string, string, error) {
	certf := "client.crt"
	keyf := "client.key"
	shared.Log("daemon", "info", fmt.Sprintf("Looking for existing certificates cert: %s, key: %s", certf, keyf))

	err := shared.FindOrGenCert(certf, keyf)

	return certf, keyf, err
}

func recv(buf []byte, m int, conn net.Conn) (n int, err error) {
	for nn := 0; n < m; {
		nn, err = conn.Read(buf[n:m])
		if nil != err && io.EOF != err {
			shared.Log("daemon", "error", fmt.Sprintf("err: %s", err))
			panic(err)
			return
		}
		n += nn
	}
	return
}

func handleConn(conn net.Conn) {
	shared.Log("client", "info", fmt.Sprintf("accepted connection from: %s", conn.RemoteAddr()))

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

	wsurl := fmt.Sprintf("wss://%s/sock/%d/%s/%s", wsServer, clrequest.version, clrequest.reqtype, parameters)
	shared.Log("daemon", "debug", fmt.Sprintf("dailing: %s", wsurl))

	wsdialer := &websocket.Dialer{TLSClientConfig: tlsConfig}

	pconn, _, err := wsdialer.Dial(wsurl, nil)
	if err != nil {
		shared.Log("daemon", "fatal", fmt.Sprintf("ws dailer failed: %s", err))
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

	shared.Pipe(pconn, conn)
}

func forward(wss string) {
	var err error
	certf, keyf, err = readMyCert()
	if err != nil {
		return
	}
	wsServer = wss

	tlsConfig, err = shared.GetTLSConfig(certf, keyf)
	if err != nil {
		return
	}
	wsServer = wss

	addr := fmt.Sprintf("%s:%d", listenIp, listenPort)
	ln, err := net.Listen("tcp", addr)
	if nil != err {
		shared.Log("daemon", "error", "Bind Error!")
		return
	}
	shared.Log("daemon", "info", fmt.Sprintf("listening for connections on: %s", addr))

	for {
		conn, err := ln.Accept()
		if nil != err {
			shared.Log("daemon", "error", "Accept Error!")
			continue
		}

		go handleConn(conn)
	}
}
