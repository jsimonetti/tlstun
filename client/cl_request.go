package main

import (
	//	"bytes"
	//	"encoding/binary"
	"fmt"
	"net"
)

const SOCKS_VERSION5 uint8 = 0x05

const CMD_CONNECT uint8 = 0x01
const CMD_BIND uint8 = 0x02
const CMD_UDPASSOC uint8 = 0x03

const ADDRTYPE_IPV4 uint8 = 0x01
const ADDRTYPE_DOMAIN uint8 = 0x03
const ADDRTYPE_IPV6 uint8 = 0x04

const AUTH_NONE uint8 = 0x00
const AUTH_GSSAPI uint8 = 0x01
const AUTH_USERPASS uint8 = 0x02

type clHello struct {
	version         uint8
	authmethodcount uint8
	authmethods     [255]uint8
}

func (msg *clHello) read(conn net.Conn) (err error) {
	_, err = recv(msg.authmethods[:2], 2, conn)
	if nil != err {
		return
	}
	msg.version, msg.authmethodcount = msg.authmethods[0], msg.authmethods[1]
	_, err = recv(msg.authmethods[:], int(msg.authmethodcount), conn)
	if nil != err {
		return
	}
	return
}
func (msg *clHello) print() {
	Log("client", "debug", fmt.Sprintf("received hello: version: %d authmethodcount: %d, authmethods: %+v", msg.version, msg.authmethodcount, msg.authmethods[:msg.authmethodcount]))
}

type clRequest struct {
	version    uint8     // socks v5: 0x05
	command    uint8     // CONNECT: 0x01, BIND:0x02, UDP ASSOCIATE: 0x03
	reserved   uint8     //RESERVED
	addrestype uint8     //IP V4 addr: 0x01, DOMANNAME: 0x03, IP V6 addr: 0x04
	dst_addr   [255]byte //
	dst_port   [2]uint8  //
	dst_port2  uint16    //

	reqtype string
	url     string
}

func (msg *clRequest) read(conn net.Conn) (err error) {
	buf := make([]byte, 4)
	_, err = recv(buf, 4, conn)
	if nil != err {
		return
	}

	msg.version, msg.command, msg.reserved, msg.addrestype = buf[0], buf[1], buf[2], buf[3]

	if 5 != msg.version || 0 != msg.reserved {
		Log("client", "error", "Request Message VER or RSV error!")
		return
	}
	switch msg.addrestype {
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

	switch msg.command {
	case 1:
		msg.reqtype = "tcp"
	case 2:
		Log("client", "error", "BIND not implemented")
	case 3:
		msg.reqtype = "udp"
	}
	switch msg.addrestype {
	case 1: // ipv4
		msg.url = fmt.Sprintf("%d.%d.%d.%d:%d", msg.dst_addr[0], msg.dst_addr[1], msg.dst_addr[2], msg.dst_addr[3], msg.dst_port2)
	case 3: //DOMANNAME
		msg.url = string(msg.dst_addr[1 : 1+msg.dst_addr[0]])
		msg.url += fmt.Sprintf(":%d", msg.dst_port2)
	case 4: //ipv6
		Log("client", "error", "IPV6 not implemented")
	}
	return
}
func (msg *clRequest) print() {
	Log("client", "info", fmt.Sprintf("received request: version: %d command: %d, reserved: %d addrestype: %d, dst_addr: %s", msg.version, msg.command, msg.reserved, msg.addrestype, msg.url))
}
