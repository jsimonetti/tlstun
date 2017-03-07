package client

import (
	"net"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"

	"github.com/jsimonetti/tlstun/protocol"
)

func (c *client) handleSession(conn net.Conn) {
	var sent, received int64

	defer func(sent, received *int64) {
		atomic.AddInt32(&c.connections, -1)
		c.log.Printf("closed connection from: %s, sent: %d, received: %d, open connections: %d\n", conn.RemoteAddr(), *sent, *received, atomic.LoadInt32(&c.connections))
		conn.Close()
	}(&sent, &received)

	var err error
	var stream net.Conn

	c.lock.Lock()
	// Open a new stream
	if c.session == nil {
		err = c.openSession()
		if err != nil {
			c.lock.Unlock()
			c.log.Printf("error opening session: %s", err)
			return
		}
	}
	c.lock.Unlock()

	stream, err = c.session.Open()
	if err != nil {
		if err == yamux.ErrSessionShutdown {
			c.lock.Lock()
			err = c.openSession()
			c.lock.Unlock()
			if err != nil {
				c.log.Printf("error opening session: %s", err)
				return
			}
		}
		stream, err = c.session.Open()
		if err != nil {
			c.log.Printf("error opening stream: %s", err)
			return
		}
	}

	received, sent = Pipe(stream, conn)
}

func (c *client) openSession() error {
	err := c.openWebsocket()
	if err != nil {
		return err
	}

	// Setup client side of yamux
	wrc := protocol.NewWsWRC(c.webSocket, c.log)
	session, err := yamux.Client(wrc, nil)
	if err != nil {
		return err
	}

	c.session = session
	return nil
}

func (c *client) openWebsocket() error {
	wsurl := "wss://" + c.serverAddr + "/tlstun/socket/"
	origin := wsurl

	wsDialer := &websocket.Dialer{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: c.tlsConfig,
	}

	wsHeader := make(http.Header)
	wsHeader.Add("Origin", origin)
	wsHeader.Add("Sec-WebSocket-Protocol", protocol.ProtoVersion())

	wsconn, _, err := wsDialer.Dial(wsurl, wsHeader)
	if err != nil {
		c.log.Printf("error opening websocket: %s", err)
		return err
	}

	c.log.Printf("connected to server: %s", wsurl)
	c.webSocket = wsconn
	return nil
}
