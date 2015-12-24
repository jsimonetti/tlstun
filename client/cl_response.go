package main

import (
	"fmt"
	"net"
)

const STATUS_GRANTED uint8 = 0x00
const STATUS_FALURE uint8 = 0x01
const STATUS_UNAUTH uint8 = 0x02
const STATUS_NETUNREACH uint8 = 0x03
const STATUS_HOSTUNREACH uint8 = 0x04

type clEcho struct {
	version uint8
	method  uint8
	buf     [2]uint8
}

func (msg *clEcho) gen(t uint8) {
	msg.version, msg.method = 5, t
	msg.buf[0], msg.buf[1] = 5, t
}
func (msg *clEcho) write(conn net.Conn) {
	conn.Write(msg.buf[:])
}
func (msg *clEcho) print() {
	Log("client", "debug", fmt.Sprintf("sent echo: version: %d method: %d", msg.version, msg.method))
}

type clResponse struct {
	version  uint8
	reply    uint8
	reserved uint8
	addrtype uint8
	//bnd_addr [255]uint8
	//bnd_port [2]uint8
	buf  [300]uint8
	mlen uint16
}

func (msg *clResponse) gen(req *clRequest, reply uint8) {
	msg.version = 5
	msg.reply = reply //rfc1928
	msg.reserved = 0
	msg.addrtype = 1 //req.addrtype

	msg.buf[0], msg.buf[1], msg.buf[2], msg.buf[3] = msg.version, msg.reply, msg.reserved, msg.addrtype
	for i := 5; i < 11; i++ {
		msg.buf[i] = 0
	}
	msg.mlen = 10
	//i := 4
	//for ; i+4 <= int(req.dst_addr[0]); i++ {
	//	msg.buf[i] = req.dst_addr[i-4]
	//}
	//msg.buf[i], msg.buf[i+1] = req.dst_port[0], req.dst_port[1]
	//msg.mlen = uint16(i + 2)
}
func (msg *clResponse) write(conn net.Conn) {
	conn.Write(msg.buf[:msg.mlen])
}
func (msg *clResponse) print() {
	Log("client", "info", fmt.Sprintf("sent response: %v", msg.buf[:msg.mlen]))
}
