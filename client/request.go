package main

import (
	//	"bytes"
	//	"encoding/binary"
	"fmt"
	"net"
)

type reqHello struct {
	ver      uint8
	nmethods uint8
	methods  [255]uint8
}

func (msg *reqHello) read(conn net.Conn) (err error) {
	_, err = recv(msg.methods[:2], 2, conn)
	if nil != err {
		return
	}
	msg.ver, msg.nmethods = msg.methods[0], msg.methods[1]
	_, err = recv(msg.methods[:], int(msg.nmethods), conn)
	if nil != err {
		return
	}
	return
}
func (msg *reqHello) print() {
	Log(fmt.Sprintf("request: ver: %d nmethods: %d, methods: %v", msg.ver, msg.nmethods, msg.methods[:msg.nmethods]))
}

type reqMsg struct {
	ver       uint8     // socks v5: 0x05
	cmd       uint8     // CONNECT: 0x01, BIND:0x02, UDP ASSOCIATE: 0x03
	rsv       uint8     //RESERVED
	atyp      uint8     //IP V4 addr: 0x01, DOMANNAME: 0x03, IP V6 addr: 0x04
	dst_addr  [255]byte //
	dst_port  [2]uint8  //
	dst_port2 uint16    //

	reqtype string
	url     string
}

func (msg *reqMsg) read(conn net.Conn) (err error) {
	buf := make([]byte, 4)
	_, err = recv(buf, 4, conn)
	if nil != err {
		return
	}

	msg.ver, msg.cmd, msg.rsv, msg.atyp = buf[0], buf[1], buf[2], buf[3]

	if 5 != msg.ver || 0 != msg.rsv {
		Log("Request Message VER or RSV error!")
		return
	}
	switch msg.atyp {
	case 1: //ip v4
		_, err = recv(msg.dst_addr[:], 4, conn)
	case 4:
		_, err = recv(msg.dst_addr[:], 16, conn)
	case 3:
		_, err = recv(msg.dst_addr[:1], 1, conn)
		_, err = recv(msg.dst_addr[1:], int(msg.dst_addr[0]), conn)
	}
	if nil != err {
		return
	}
	_, err = recv(msg.dst_port[:], 2, conn)
	if nil != err {
		return
	}
	//bbuf := bytes.NewBuffer(msg.dst_port[:])
	//err = binary.Read(bbuf, binary.BigEndian, msg.dst_port2)
	//if nil != err {
	//	Log.Println(err)
	//	return
	//}
	msg.dst_port2 = (uint16(msg.dst_port[0]) << 8) + uint16(msg.dst_port[1])

	switch msg.cmd {
	case 1:
		msg.reqtype = "tcp"
	case 2:
		Log("BIND not implemented")
	case 3:
		msg.reqtype = "udp"
	}
	switch msg.atyp {
	case 1: // ipv4
		msg.url = fmt.Sprintf("%d.%d.%d.%d:%d", msg.dst_addr[0], msg.dst_addr[1], msg.dst_addr[2], msg.dst_addr[3], msg.dst_port2)
	case 3: //DOMANNAME
		msg.url = string(msg.dst_addr[1 : 1+msg.dst_addr[0]])
		msg.url += fmt.Sprintf(":%d", msg.dst_port2)
	case 4: //ipv6
		Log("IPV6 not implemented")
	}
	return
}
func (msg *reqMsg) print() {
	Log(fmt.Sprintf("request: ver: %d cmd: %d, rsv: %d atyp: %d, dst_addr: %s", msg.ver, msg.cmd, msg.rsv, msg.atyp, msg.url))
}
