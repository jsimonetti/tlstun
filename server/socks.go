package main

import (
	"fmt"
	"io"
	"net"
	"time"

	log "gopkg.in/inconshreveable/log15.v2"
)

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
func (msg *clHello) log() (string, log.Ctx) {
	return "received hello", log.Ctx{"version": msg.version, "authmethodcount": msg.authmethodcount, "authmethods": msg.authmethods[:msg.authmethodcount]}
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
		fmt.Println("Request Message VER or RSV error!")
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

	msg.dst_port2 = (uint16(msg.dst_port[0]) << 8) + uint16(msg.dst_port[1])

	switch msg.command {
	case 1:
		msg.reqtype = "tcp"
	case 2:
		fmt.Println("BIND not implemented")
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
		fmt.Println("IPV6 not implemented")
	}
	return
}
func (msg *clRequest) log() (string, log.Ctx) {
	return "received request", log.Ctx{"version": msg.version, "command": msg.command, "rsvd": msg.reserved, "addrtype": msg.addrestype, "dst_addr": msg.url}
}

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

func (msg *clEcho) log() (string, log.Ctx) {
	return "sent echo", log.Ctx{"version": msg.version, "method": msg.method}
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

func (msg *clResponse) log() (string, log.Ctx) {
	return "sent response", log.Ctx{"response": msg.buf[:msg.mlen]}
}

func recv(buf []byte, m int, conn net.Conn) (n int, err error) {
	ch := make(chan int)
	eCh := make(chan error)
	var tick time.Time
	go func(ch chan int, eCh chan error) {
		var num int
		var merr error
		for nn := 0; num < m; {
			// try to read the data
			nn, merr = conn.Read(buf[num:m])
			if nil != merr && io.EOF != merr {
				eCh <- merr
				return
			}
			num += nn
		}
		ch <- num
	}(ch, eCh)

	for {
		ticker := time.NewTicker(3 * time.Second)
		select {
		case n = <-ch:
			return
		case err = <-eCh:
			return
		case tick = <-ticker.C:
			ticker.Stop()
			break
		}
		ticker.Stop()
	}
	err = fmt.Errorf("recv timeout: %v", tick)
	return
}
