package client

import (
	"net"
	"sync/atomic"

	"github.com/hashicorp/yamux"
	"golang.org/x/net/websocket"
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
	session, err := yamux.Client(c.webSocket, nil)
	if err != nil {
		return err
	}

	c.session = session
	return nil
}

func (c *client) openWebsocket() error {
	wsurl := "wss://" + c.serverAddr + "/tlstun/socket/"
	origin := wsurl

	wsConfig, err := websocket.NewConfig(wsurl, origin)
	if err != nil {
		return err
	}
	wsConfig.TlsConfig = c.tlsConfig

	wsconn, err := websocket.DialConfig(wsConfig)
	if err != nil {
		return err
	}

	c.log.Printf("connected to server: %s", wsurl)
	c.webSocket = wsconn
	return nil
}
