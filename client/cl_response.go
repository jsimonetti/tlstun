package main

import (
	"fmt"
	"net"
)

type clEcho struct {
	ver    uint8
	method uint8
	buf    [2]uint8
}

func (msg *clEcho) gen(t uint8) {
	msg.ver, msg.method = 5, t
	msg.buf[0], msg.buf[1] = 5, t
}
func (msg *clEcho) write(conn net.Conn) {
	conn.Write(msg.buf[:])
}
func (msg *clEcho) print() {
	Log(fmt.Sprintf("[c] echo: ver: %d method: %s", msg.ver, msg.method))
}

type clResponse struct {
	ver  uint8
	rep  uint8
	rsv  uint8
	atyp uint8
	//bnd_addr [255]uint8
	//bnd_port [2]uint8
	buf  [300]uint8
	mlen uint16
}

func (msg *clResponse) gen(req *clRequest, rep uint8) {
	msg.ver = 5
	msg.rep = rep //rfc1928
	msg.rsv = 0
	msg.atyp = 1 //req.atyp

	msg.buf[0], msg.buf[1], msg.buf[2], msg.buf[3] = msg.ver, msg.rep, msg.rsv, msg.atyp
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
	Log(fmt.Sprintf("[c] response: %s", msg.buf[:msg.mlen]))
}
