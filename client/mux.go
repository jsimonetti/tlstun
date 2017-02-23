package main

import (
	"fmt"
	"net"

	"golang.org/x/net/websocket"

	"github.com/hashicorp/yamux"
	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/jsimonetti/tlstun/shared"
)

func openWebsocket(d *Daemon) error {
	// open websocket
	wsurl := fmt.Sprintf("wss://%s/sock/", d.serverAddr)
	origin := wsurl

	d.log.Debug("dialing server", log.Ctx{"wsurl": wsurl})

	wsConfig, err := websocket.NewConfig(wsurl, origin)
	if err != nil {
		d.log.Error("ws config failed", log.Ctx{"error": err})
		return err
	}

	wsConfig.TlsConfig = d.tlsConfig

	wsconn, err := websocket.DialConfig(wsConfig)
	if err != nil {
		d.log.Error("ws dial failed", log.Ctx{"error": err})
		return err
	}
	d.wsconn = wsconn
	//
	return nil
}

func openSession(d *Daemon) error {

	err := openWebsocket(d)
	if err != nil {
		d.log.Error("could not open websocket", log.Ctx{"error": err})
		return err
	}

	// Setup client side of yamux
	session, err := yamux.Client(d.wsconn, nil)
	if err != nil {
		d.log.Error("Yamux session failed!", log.Ctx{"error": err})
	}

	d.session = session

	return nil
}

func handleSession(d *Daemon, conn net.Conn) {
	var err error
	var stream net.Conn

	connlog := d.log.New("raddr", conn.RemoteAddr())
	connlog.Debug("opening yamux stream")

	// Open a new stream
	if d.session == nil {
		d.log.Debug("session was not open, opening...")
		err = openSession(d)
		if err != nil {
			d.log.Error("could not open session", log.Ctx{"error": err})
			return
		}

	}

	stream, err = d.session.Open()
	if err != nil {
		if err == yamux.ErrSessionShutdown {
			d.log.Debug("session was shut down, reopening...")
			err = openSession(d)
			if err != nil {
				d.log.Error("could not open session", log.Ctx{"error": err})
				return
			}
		}
		stream, err = d.session.Open()
		if err != nil {
			d.log.Error("Stream open failed!", log.Ctx{"error": err})
			conn.Close()
			return
		}
	}

	connlog.Info("connection copy started", log.Ctx{"open": d.session.NumStreams()})
	received, sent := shared.Pipe(stream, conn)
	connlog.Info("connection closed", log.Ctx{"inbytes": received, "outbytes": sent, "open": d.session.NumStreams()})

}
