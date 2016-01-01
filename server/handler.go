package main

import (
	"crypto/x509"
	"net"
	"net/http"

	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/jsimonetti/tlstun/shared/websocket"
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
		websocket: w,
		log:       clog,
		raddr:     raddr,
	}
	go client.run()
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
	if !d.isTrustedClient(r) && !PasswordCheck(password) {
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
