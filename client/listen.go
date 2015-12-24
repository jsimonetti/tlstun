package main

import (
	"fmt"
	"io"
	"net"
)

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
	var pconn net.Conn
	pconn, err = net.Dial(clrequest.reqtype, clrequest.url)

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
	pipe(conn, pconn)
}

func resend(in net.Conn, out net.Conn) {
	Log("daemon", "debug", fmt.Sprintf("piping connection %s => %s", out.RemoteAddr(), in.RemoteAddr()))
	buf := make([]byte, 10240)
	for {
		n, err := in.Read(buf)
		if err == io.EOF {
			Log("daemon", "debug", fmt.Sprintf("connection closed %s => %s", out.RemoteAddr(), in.RemoteAddr()))
			return
		} else if err != nil {
			Log("daemon", "error", fmt.Sprintf("resend err: %s", err))
			return
		}
		out.Write(buf[:n])
	}
}

func pipe(a net.Conn, b net.Conn) {
	go resend(a, b)
	go resend(b, a)
}

func listen() {
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
