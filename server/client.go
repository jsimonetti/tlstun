package main

import (
	"io"
	"net"
	"time"

	"github.com/hashicorp/yamux"
	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/jsimonetti/tlstun/shared"
	"github.com/jsimonetti/tlstun/shared/websocket"
)

type clientConnection struct {
	log       log.Logger
	websocket *websocket.Conn
	session   *yamux.Session
	raddr     string
}

// Accept is used to block until the next available stream
// // is ready to be accepted.
func (c *clientConnection) acceptStream() (net.Conn, uint32, error) {
	conn, err := c.session.AcceptStream()
	if err != nil {
		return nil, 0, err
	}
	return conn, conn.StreamID(), err
}

func (c *clientConnection) run() {
	c.log.Debug("starting yamux on ws")

	session, err := yamux.Server(c.websocket, nil)
	if err != nil {
		c.log.Crit("could not initialise yamux session", log.Ctx{"error": err})
	}
	c.log.Debug("yamux session started")
	c.session = session

	c.log.Debug("listening for streams")
	// Accept a stream
	for {
		stream, id, err := c.acceptStream()
		if err != nil {
			if err != io.EOF {
				c.log.Error("error acception stream", log.Ctx{"error": err})
			}
			c.websocket.Close()
			c.session.Close()
			return
		}
		c.log.Debug("accepted stream", log.Ctx{"streamid": id})
		go c.handleStream(stream, id)
	}
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
}
