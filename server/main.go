package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/net/http2"

	"github.com/jsimonetti/tlstun/shared"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 1024,
	WriteBufferSize: 1024 * 1024,
}

var listenIp string
var listenPort int

var tlsConfig *tls.Config

var certf string
var keyf string

var registerPass string

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		shared.Log("daemon", "debug", fmt.Sprintf("404 not found: %s", r.URL.Path))
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !isTrustedClient(r) {
		w.Write([]byte("It Works!"))
		return
	}
	w.Write([]byte("It Works and you have a trusted cert!"))
}

func serveRegister(w http.ResponseWriter, r *http.Request) {
	shared.Log("daemon", "debug", fmt.Sprintf("handled register"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var cert *x509.Certificate
	var name string

	if r.TLS != nil {

		if len(r.TLS.PeerCertificates) < 1 {
			shared.Log("daemon", "debug", "no client cert found")
			return
		}
		cert = r.TLS.PeerCertificates[len(r.TLS.PeerCertificates)-1]

		remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			shared.Log("daemon", "debug", fmt.Sprintf("internal error: %s", err))
			return
		}

		name = remoteHost
	} else {
		return
	}

	fingerprint := certGenerateFingerprint(cert)
	for _, existingCert := range clientCerts {
		if fingerprint == certGenerateFingerprint(&existingCert) {
			return
		}
	}

	password := r.FormValue("password")
	if !isTrustedClient(r) && !PasswordCheck(password) {
		w.Write([]byte("Failed"))
		return
	}

	err := saveCert(name, cert)
	if err != nil {
		shared.Log("daemon", "debug", fmt.Sprintf("cannot save cert: %s", err))
		return
	}

	clientCerts = append(clientCerts, *cert)
	w.Write([]byte("OK"))
}

type connection struct {
	ws   *websocket.Conn
	conn net.Conn

	version    int
	request    string
	parameters string
}

func (c *connection) handle() {

	shared.Log("daemon", "debug", fmt.Sprintf("handled connection: version: %d, request: %s, parameters: %s", c.version, c.request, c.parameters))

	var err error
	c.conn, err = net.DialTimeout(c.request, c.parameters, time.Duration(500)*time.Millisecond)

	if err != nil {
		shared.Log("daemon", "debug", fmt.Sprintf("error dialing %s - %s, err: %s", c.request, c.parameters, err))
		c.ws.Close()
		return
	}
	shared.OldPipe(c.ws, c.conn)
}

// serveWs handles websocket requests from the peer.
func sockHandler(w http.ResponseWriter, r *http.Request) {
	if !isTrustedClient(r) {
		shared.Log("daemon", "warn", "untrusted client connected")
		return
	}
	vars := mux.Vars(r)
	version, _ := strconv.Atoi(vars["version"])
	request := vars["request"]
	parameters, err := base64.StdEncoding.DecodeString(vars["parameters"])
	if err != nil {
		shared.Log("daemon", "error", fmt.Sprintf("base64decode failed: %s", err))
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		shared.Log("daemon", "error", fmt.Sprintf("ws upgrade failed: %s", err))
		return
	}
	c := &connection{version: version, request: request, parameters: fmt.Sprintf("%s", parameters), ws: ws}
	go c.handle()
}

func listen() {
	/* Setup the TLS authentication */
	certf, keyf, err := readMyCert()
	if err != nil {
		return
	}
	readSavedClientCAList()

	tlsConfig, err = shared.GetTLSConfig(certf, keyf)
	if err != nil {
		return
	}

	addr := fmt.Sprintf("%s:%d", listenIp, listenPort)
	mux := mux.NewRouter()
	mux.HandleFunc("/sock/{version}/{request}/{parameters}", sockHandler)
	mux.HandleFunc("/register", serveRegister)
	mux.HandleFunc("/", serveHome)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		TLSConfig:    tlsConfig,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shared.Log("daemon", "info", "attemting upgrade of server to http/2")
	if err := http2.ConfigureServer(server, nil); err != nil {
		shared.Log("daemon", "info", fmt.Sprintf("upgrade to http/2 failed: %s", err))
	}

	shared.Log("daemon", "info", fmt.Sprintf("Listening on %s", addr))
	err = server.ListenAndServeTLS(certf, keyf)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func main() {

	flag.Parse()

	if cpuprofile {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}
	if err := initializeDbObject("./tlstun.sqlite3"); err != nil {
		shared.Log("daemon", "error", "Could not init database")
	}
	listen()
}

var cpuprofile bool

func init() {
	flag.BoolVar(&cpuprofile, "cpuprofile", false, "show cpu profile on http://localhost:6060")
	flag.BoolVar(&shared.ShowLog, "log", false, "show logging")
	flag.IntVar(&listenPort, "port", 443, "port to listen on")
	flag.StringVar(&listenIp, "ip", "", "ip to bind to")
	flag.StringVar(&registerPass, "regpass", "", "password to use for registration")
	flag.IntVar(&shared.WsTimeOut, "timeout", 10, "timeout for websocket connections")
}
