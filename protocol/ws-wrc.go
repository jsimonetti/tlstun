package protocol

import (
	"sync"
	"time"

	"github.com/jsimonetti/tlstun/log"

	"github.com/gorilla/websocket"
)

type ReadWriteCloser struct {
	mWrite  sync.Mutex
	ws      *websocket.Conn
	mBuffer sync.Mutex
	wrclog  *log.Logger
	buffer  []byte
}

func NewWsWRC(ws *websocket.Conn, log *log.Logger) *ReadWriteCloser {
	c := &ReadWriteCloser{
		ws: ws,
	}

	ws.SetReadLimit(64 * 1024)
	ws.SetReadDeadline(time.Now().Add(30 * time.Second))
	ws.SetPongHandler(c.pongHandler)

	c.wrclog = log
	go c.sendPings()

	return c
}

func (c *ReadWriteCloser) Read(p []byte) (int, error) {
	// Guard access to c.buffer
	c.mBuffer.Lock()
	defer c.mBuffer.Unlock()

	// Read message until we can populate c.buffer
	for len(c.buffer) == 0 {
		// Read a message
		t, m, err := c.ws.ReadMessage()
		if err != nil {
			return 0, err
		}

		// Skip anything that isn't a binary message
		if t != websocket.BinaryMessage {
			continue
		}

		// Set buffer
		c.buffer = m
	}

	// Copy out from buffer
	n := copy(p, c.buffer)
	c.buffer = c.buffer[n:]

	return n, nil
}

func (c *ReadWriteCloser) Write(p []byte) (int, error) {
	c.mWrite.Lock()
	c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err := c.ws.WriteMessage(websocket.BinaryMessage, p)
	c.mWrite.Unlock()

	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close will close the underlying websocket
func (c *ReadWriteCloser) Close() error {
	// Attempt a graceful close
	c.mWrite.Lock()
	err := c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.mWrite.Unlock()

	// Always make sure we close properly
	cerr := c.ws.Close()

	// Prefer error sending the close message over any error from closing the
	// websocket.
	if err != nil {
		return err
	}
	return cerr
}

func (c *ReadWriteCloser) sendPings() {
	for {
		// Sleep for ping interval time
		time.Sleep(5 * time.Second)

		// Write a ping message, and reset the write deadline
		c.mWrite.Lock()
		c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		err := c.ws.WriteMessage(websocket.PingMessage, []byte{})
		c.mWrite.Unlock()

		// If there is an error we resolve with internal error
		if err != nil {
			c.wrclog.Printf("Ping failed, probably the connection was closed, error: %s", err)
			return
		}
	}
}

func (c *ReadWriteCloser) pongHandler(string) error {
	// Reset the read deadline
	c.ws.SetReadDeadline(time.Now().Add(30 * time.Second))
	c.wrclog.Printf("Received pong, now extending the read-deadline")
	return nil
}
