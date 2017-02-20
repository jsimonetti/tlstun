package main

import (
	"crypto/tls"
	"fmt"
	"net"

	"golang.org/x/net/websocket"

	"github.com/hashicorp/yamux"
	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/jsimonetti/tlstun/shared"
)

func startDaemon() (*Daemon, error) {
	caddr := fmt.Sprintf("%s:%d", listenIp, listenPort)
	saddr := fmt.Sprintf("%s:%d", serverIp, serverPort)

	d := &Daemon{
		listenAddr: caddr,
		serverAddr: saddr,
	}

	if err := d.Init(); err != nil {
		return nil, err
	}

	return d, nil
}

type Daemon struct {
	listenAddr string // local ip:port to bind to
	serverAddr string // server ip:port running the tls websocket proxy

	tlsConfig *tls.Config // tls configuration for connections to the server
	certf     string      // my client certificate
	keyf      string      // my client key

	log     log.Logger
	wsconn  *websocket.Conn
	session *yamux.Session
}

func (d *Daemon) Close() {
}

func (d *Daemon) Run() error {
	var err error

	ln, err := net.Listen("tcp", d.listenAddr)
	if nil != err {
		d.log.Error("Bind Error!", log.Ctx{"address": d.listenAddr, "error": err})
		return err
	}
	d.log.Info("Listening for connections", log.Ctx{"address": d.listenAddr})

	for {
		conn, err := ln.Accept()
		if nil != err {
			d.log.Error("Accept error!", log.Ctx{"error": err})
			continue
		}
		go handleSession(d, conn)
	}

	return nil
}

func (d *Daemon) Init() error {

	/* Setup the TLS authentication */
	certf, keyf, err := shared.ReadMyCert("client.crt", "client.key")
	if err != nil {
		return err
	}
	d.certf = certf
	d.keyf = keyf
	d.tlsConfig, err = shared.GetTLSConfig(d.certf, d.keyf)
	if err != nil {
		return err
	}
	srvlog := log.New(log.Ctx{"module": "client"})
	d.log = srvlog

	handler := log.StdoutHandler

	if quiet {
		d.log.SetHandler(log.DiscardHandler())
	} else if verbose {
		d.log.SetHandler(log.LvlFilterHandler(log.LvlInfo, handler))
	} else if debug {
		d.log.SetHandler(log.LvlFilterHandler(log.LvlDebug, handler))
	} else {
		d.log.SetHandler(log.LvlFilterHandler(log.LvlError, handler))
	}

	d.Run()
	return nil
}
