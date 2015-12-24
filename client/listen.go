package main

import (
	//	"bytes"
	//	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

func recv(buf []byte, m int, conn net.Conn) (n int, err error) {
	for nn := 0; n < m; {
		nn, err = conn.Read(buf[n:m])
		if nil != err && io.EOF != err {
			log.Println("err:", err)
			panic(err)
			return
		}
		n += nn
	}
	return
}

func handleConn(conn net.Conn) {
	//defer conn.Close()
	log.Println("remote addr:", conn.RemoteAddr())

	var reqhello reqHello
	var ansecho ansEcho
	var reqmsg reqMsg
	var ansmsg ansMsg

	//recv hello
	var err error = nil
	err = reqhello.read(conn)
	if nil != err {
		return
	}
	reqhello.print()

	//send echo
	ansecho.gen(0)
	ansecho.write(conn)
	ansecho.print()

	//recv request
	err = reqmsg.read(conn)
	if nil != err {
		return
	}
	reqmsg.print()
	//connect
	var pconn net.Conn
	pconn, err = net.Dial(reqmsg.reqtype, reqmsg.url)
	//defer pconn.Close()

	//reply
	//error occur
	if nil != err {
		ansmsg.gen(&reqmsg, 4)
		ansmsg.write(conn)
		ansmsg.print()
		return
	}
	//success
	ansmsg.gen(&reqmsg, 0)
	ansmsg.write(conn)
	ansmsg.print()
	pipe(conn, pconn)
}

func resend(in net.Conn, out net.Conn) {
	buf := make([]byte, 10240)
	for {
		n, err := in.Read(buf)
		if io.EOF == err {
			log.Printf("io.EOF")
			return
		} else if nil != err {
			log.Printf("resend err\n", err)
			return
		}
		out.Write(buf[:n])
	}
}

func pipe(a net.Conn, b net.Conn) {
	go resend(a, b)
	go resend(b, a)
}

func socks5proxy() {
	ln, err := net.Listen("tcp", ":8000")
	if nil != err {
		fmt.Println("Bind Error!")
		return
	}

	for {
		conn, err := ln.Accept()
		if nil != err {
			fmt.Println("Accept Error!")
			continue
		}

		go handleConn(conn)
	}
}
