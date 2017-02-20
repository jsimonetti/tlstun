package main

import (
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/websocket"

	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/hashicorp/yamux"
	"github.com/jsimonetti/tlstun/shared"
)

func sockHandler(d *Daemon, w *websocket.Conn) {
	r := w.Request()
	raddr := r.RemoteAddr
	if !d.isTrustedClient(r) {
		d.log.Warn("untrusted client connected", log.Ctx{"raddr": raddr})
		return
	}

	d.log.Debug("handing over client connection", log.Ctx{"raddr": raddr})

	clog := log.New(log.Ctx{"module": "socks", "raddr": raddr})
	handler := log.StdoutHandler

	if quiet {
		clog.SetHandler(log.DiscardHandler())
	} else if verbose {
		clog.SetHandler(log.LvlFilterHandler(log.LvlInfo, handler))
	} else if debug {
		clog.SetHandler(log.LvlFilterHandler(log.LvlDebug, handler))
	} else {
		clog.SetHandler(log.LvlFilterHandler(log.LvlError, handler))
	}

	client := &clientConnection{
		log:   clog,
		raddr: raddr,
	}

	client.log.Debug("starting yamux on ws")

	session, err := yamux.Server(w, nil)
	if err != nil {
		client.log.Crit("could not initialise yamux session", log.Ctx{"error": err})
		return
	}

	client.log.Debug("yamux session started")
	client.session = session

	client.log.Debug("listening for streams")
	// Accept a stream
	for {
		stream, id, err := client.acceptStream()
		stream.SetDeadline(time.Now().Add(shared.TimeOut))
		if err != nil {
			if err != io.EOF {
				client.log.Error("error acception stream", log.Ctx{"error": err})
			}
			w.Close()
			client.session.Close()
			return
		}
		client.log.Debug("accepted stream", log.Ctx{"streamid": id})
		go client.handleStream(stream, id)
	}

}

func serveRegister(d *Daemon, w http.ResponseWriter, r *http.Request) {
	d.log.Debug("handled register")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var cert *x509.Certificate
	var name string

	if r.TLS != nil {

		if len(r.TLS.PeerCertificates) < 1 {
			d.log.Debug("no client cert found")
			return
		}
		cert = r.TLS.PeerCertificates[len(r.TLS.PeerCertificates)-1]

		remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			d.log.Debug("internal error", log.Ctx{"error": err})
			return
		}

		name = remoteHost
	} else {
		return
	}

	fingerprint := certGenerateFingerprint(cert)
	for _, existingCert := range d.clientCerts {
		if fingerprint == certGenerateFingerprint(&existingCert) {
			return
		}
	}

	password := r.FormValue("password")
	if !d.isTrustedClient(r) && !PasswordCheck(d, password) {
		w.Write([]byte("Failed"))
		return
	}

	err := saveCert(d, name, cert)
	if err != nil {
		d.log.Warn("cannot save cert", log.Ctx{"error": err})
		return
	}

	d.clientCerts = append(d.clientCerts, *cert)
	w.Write([]byte("OK"))
}

func serveHome(d *Daemon, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		d.log.Debug("404 not found", log.Ctx{"path": r.URL.Path})
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		d.log.Debug("Method not allowed", log.Ctx{"path": r.URL.Path, "method": r.Method})
		return
	}
	d.log.Debug("served page", log.Ctx{"url": r.URL.Path})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !d.isTrustedClient(r) {
		w.Write([]byte("It Works!"))
		return
	}
	w.Write([]byte("It Works and you have a trusted cert!"))
}
