package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/jsimonetti/tlstun/shared"
	"github.com/jsimonetti/tlstun/shared/websocket"
)

func startDaemon() (*Daemon, error) {
	addr := fmt.Sprintf("%s:%d", listenIp, listenPort)

	d := &Daemon{
		listenAddr: addr,
	}

	if err := d.Init(); err != nil {
		return nil, err
	}

	return d, nil
}

type Daemon struct {
	listenAddr string // local ip:port to bind to

	pwd       string      // my location
	tlsConfig *tls.Config // tls configuration for connections to the server
	certf     string      // my client certificate
	keyf      string      // my client key

	server      *http.Server       // the http listening server
	log         log.Logger         // daemon logging
	db          *sql.DB            // database connection
	clientCerts []x509.Certificate // list of client certificates
}

func (d *Daemon) Close() {
}

func (d *Daemon) Run() error {
	http.Handle("/sock/", websocket.Handler(func(w *websocket.Conn) {
		d.log.Debug("handling socket")
		sockHandler(d, w)
	}))
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		d.log.Debug("handling register")
		serveRegister(d, w, r)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d.log.Debug("handling home")
		serveHome(d, w, r)
	})

	server := &http.Server{
		Addr:      d.listenAddr,
		TLSConfig: d.tlsConfig,
		//ReadTimeout:  10 * time.Second,
		//WriteTimeout: 10 * time.Second,
	}

	d.log.Info("Listening", log.Ctx{"address": d.listenAddr})
	err := server.ListenAndServeTLS(d.certf, d.keyf)
	if err != nil {
		d.log.Crit("ListenAndServeTLS", log.Ctx{"Error": err})
		panic("ListenAndServeTLS: " + err.Error())
		return err
	}

	return nil
}

func (d *Daemon) Init() error {

	/* read my path */
	pwd, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return err
	}
	d.pwd = pwd

	/* Setup the TLS authentication */
	certf, keyf, err := shared.ReadMyCert("server.crt", "server.key")
	if err != nil {
		return err
	}
	d.certf = certf
	d.keyf = keyf
	d.tlsConfig, err = shared.GetTLSConfig(d.certf, d.keyf)
	if err != nil {
		return err
	}

	srvlog := log.New(log.Ctx{"module": "server"})
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

	err = initializeDbObject(d, "server.db")
	if err != nil {
		return err
	}
	readSavedClientCAList(d)

	return d.Run()
}
