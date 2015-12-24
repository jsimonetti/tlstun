package main

import (
	"fmt"
	"net"

	"github.com/jsimonetti/tlstun/shared"
)

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
	shared.Log("client", "debug", fmt.Sprintf("sent echo: version: %d method: %d", msg.version, msg.method))
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
}
func (msg *clResponse) write(conn net.Conn) {
	conn.Write(msg.buf[:msg.mlen])
}
func (msg *clResponse) print() {
	shared.Log("client", "info", fmt.Sprintf("sent response: %v", msg.buf[:msg.mlen]))
}
