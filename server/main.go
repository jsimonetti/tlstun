package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	//	"encoding/pem"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/net/http2"

	"github.com/jsimonetti/tlstun/shared"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  10240,
	WriteBufferSize: 10240,
}

var listenIp string
var listenPort int

var clientCerts []x509.Certificate
var tlsConfig *tls.Config

var certf string
var keyf string

func readSavedClientCAList() {
	return
	/*
		clientCerts = []x509.Certificate{}

		dbCerts, err := dbCertsGet()
		if err != nil {
			shared.Log("daemon", "error", fmt.Sprintf("Error reading certificates from database: %s", err))
			return
		}

		for _, dbCert := range dbCerts {
			certBlock, _ := pem.Decode([]byte(dbCert.Certificate))
			cert, err := x509.ParseCertificate(certBlock.Bytes)
			if err != nil {
				shared.Log("daemon", "error", fmt.Sprintf("Error reading certificate for %s: %s", dbCert.Name, err))
				continue
			}
			clientCerts = append(clientCerts, *cert)
		}
	*/
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("It Works!"))
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
	c.conn, err = net.Dial(c.request, c.parameters)

	if err != nil {
		shared.Log("daemon", "debug", fmt.Sprintf("error dialing %s - %s, err: %s", c.request, c.parameters, err))
		c.ws.Close()
		return
	}
	shared.Pipe(c.ws, c.conn)
}

// serveWs handles websocket requests from the peer.
func sockHandler(w http.ResponseWriter, r *http.Request) {
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

func isTrustedClient(r *http.Request) bool {
	if r.TLS == nil {
		return false
	}
	for i := range r.TLS.PeerCertificates {
		if checkTrustState(*r.TLS.PeerCertificates[i]) {
			return true
		}
	}
	return false
}

func checkTrustState(cert x509.Certificate) bool {
	for _, v := range clientCerts {
		if bytes.Compare(cert.Raw, v.Raw) == 0 {
			shared.Log("daemon", "debug", "Found cert")
			return true
		}
		shared.Log("daemon", "debug", "Client cert != key")
	}
	return false
}

func readMyCert() (string, string, error) {
	certf := "server.crt"
	keyf := "server.key"
	shared.Log("daemon", "info", fmt.Sprintf("Looking for existing certificates cert: %s, key: %s", certf, keyf))

	err := shared.FindOrGenCert(certf, keyf)

	return certf, keyf, err
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
	mux.HandleFunc("/", serveHome)

	server := &http.Server{
		Addr:      addr,
		Handler:   mux,
		TLSConfig: tlsConfig,
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
	listen()
}

func init() {
	flag.IntVar(&listenPort, "port", 443, "port to listen on")
	flag.StringVar(&listenIp, "ip", "", "ip to bind to")
}
