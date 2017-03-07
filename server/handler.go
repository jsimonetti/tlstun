package server

import (
	"crypto/x509"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/jsimonetti/tlstun/protocol"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
)

var upgrader = websocket.Upgrader{}

func (s *server) sockHandler(w http.ResponseWriter, r *http.Request) {
	if !s.isTrusted(r) {
		s.log.Printf("untrusted client connected from %s", r.RemoteAddr)
		return
	}
	atomic.AddInt32(&s.connections, 1)

	s.log.Printf("serving client connection, raddr: %s, connections: %d", r.RemoteAddr, atomic.LoadInt32(&s.connections))

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Printf("could not upgrade to websocket: %s", err)
	}

	wrc := protocol.NewWsWRC(ws, s.log)
	session, err := yamux.Server(wrc, nil)
	if err != nil {
		s.log.Printf("could not initialise yamux session: %s", err)
		return
	}

	var streams int32

	for {
		stream, err := session.AcceptStream()
		if err != nil {
			if err != io.EOF {
				s.log.Printf("error acception stream: %s, connections: %d, ", err, atomic.LoadInt32(&s.connections))
			}
			break
		}
		atomic.AddInt32(&streams, 1)
		id := stream.StreamID()
		s.log.Printf("accepted stream for id: %d, connections: %d, streams: %d", id, atomic.LoadInt32(&s.connections), streams)

		go func() {
			//no error handling needed, socks package allready logs errors
			s.socksServer.ServeConn(stream)
			atomic.AddInt32(&streams, -1)
			s.log.Printf("ended stream for id: %d, connections: %d, streams: %d", id, atomic.LoadInt32(&s.connections), atomic.LoadInt32(&streams))
			stream.Close()
		}()
	}

	atomic.AddInt32(&s.connections, -1)
	s.log.Printf("connection closed, raddr: %s, connections: %d", r.RemoteAddr, atomic.LoadInt32(&s.connections))
	wrc.Close()
	return
}

func (s *server) serveRegister(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/tlstun/register" || r.Method != "POST" {
		http.Error(w, "Not found", 404)
		s.log.Printf("404 not found: %s %s", r.Method, r.URL.Path)
		return
	}
	s.log.Printf("handled register for %s", r.RemoteAddr)

	var cert *x509.Certificate

	if r.TLS != nil {
		if len(r.TLS.PeerCertificates) < 1 {
			s.log.Printf("no client cert found registering for %s", r.RemoteAddr)
			http.Error(w, "Not found", 404)
			return
		}
		cert = r.TLS.PeerCertificates[len(r.TLS.PeerCertificates)-1]
	} else {
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	password := r.FormValue("password")
	if s.isTrusted(r) {
		w.Write([]byte("Allready trusted"))
		return
	}

	if !s.passwordCheck(password) {
		w.Write([]byte("Failed"))
		return
	}

	err := s.saveCert(cert)
	if err != nil {
		s.log.Printf("cannot save cert: %s", err)
		return
	}

	w.Write([]byte("OK"))
}

func (s *server) servePoison(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" || !s.isTrusted(r) {
		http.Error(w, "Not found", 404)
		s.log.Printf("404 not found: %s %s", r.Method, r.URL.Path)
		return
	}
	s.log.Printf("served poison: %s", r.URL.Path)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(strings.TrimPrefix(r.URL.Path, "/tlstun/poison/")))
	return
}

func (s *server) serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/tlstun/status" || r.Method != "GET" {
		http.Error(w, "Not found", 404)
		s.log.Printf("404 not found: %s %s", r.Method, r.URL.Path)
		return
	}

	s.log.Printf("served page: %s", r.URL.Path)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !s.isTrusted(r) {
		w.Write([]byte(UnTrustedResponse()))
		return
	}
	w.Write([]byte(TrustedResponse()))
	return
}
