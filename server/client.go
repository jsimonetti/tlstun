package main

import (
	"net"
	"time"

	"github.com/hashicorp/yamux"
	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/jsimonetti/tlstun/shared"
)

type clientConnection struct {
	log     log.Logger
	session *yamux.Session
	raddr   string
}

// Accept is used to block until the next available stream
// // is ready to be accepted.
func (c *clientConnection) acceptStream() (net.Conn, uint32, error) {
	conn, err := c.session.AcceptStream()
	if err != nil {
		time.Sleep(10 * time.Nanosecond)
		return nil, 0, err
	}
	return conn, conn.StreamID(), err
}

func (c *clientConnection) handleStream(stream net.Conn, streamid uint32) {

	clog := c.log.New("stream", streamid)

	var clhello clHello
	var clecho clEcho
	var clrequest clRequest
	var clresponse clResponse

	//recv hello
	var err error = nil
	err = clhello.read(stream)
	if nil != err {
		clog.Error("error reading socks hello", log.Ctx{"error": err})
		stream.Close()
		return
	}
	clog.Debug(clhello.print())

	//send echo
	clecho.gen(0)
	clecho.write(stream)
	clog.Debug(clecho.print())

	//recv request
	err = clrequest.read(stream)
	if nil != err {
		clog.Error("error reading socks request", log.Ctx{"error": err})
		stream.Close()
		return
	}
	clog.Debug(clrequest.print())
	//connect

	clog.Debug("accepted socks request", log.Ctx{"request": clrequest.reqtype, "dst": clrequest.url})
	conn, err := net.DialTimeout(clrequest.reqtype, clrequest.url, time.Duration(500)*time.Millisecond)

	if err != nil {
		clog.Error("error dialing upstream", log.Ctx{"error": err})
		clresponse.gen(&clrequest, 4)
		clresponse.write(stream)
		clog.Debug(clresponse.print())
		stream.Close()
		return
	}
	//success
	clresponse.gen(&clrequest, 0)
	clresponse.write(stream)
	clog.Debug(clresponse.print())

	inbytes, outbytes := shared.Pipe(conn, stream)
	clog.Info("upstream connection closed", log.Ctx{"request": clrequest.reqtype, "dst": clrequest.url, "inbytes": inbytes, "outbytes": outbytes})
	defer func() {
		conn.Close()
		stream.Close()
	}()

}
