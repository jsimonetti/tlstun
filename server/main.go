package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"golang.org/x/net/http2"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  10240,
	WriteBufferSize: 10240,
}

var listenIp string
var listenPort int

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

	Log("daemon", "debug", fmt.Sprintf("handled connection: version: %d, request: %s, parameters: %s", c.version, c.request, c.parameters))

	var err error
	c.conn, err = net.Dial(c.request, c.parameters)

	if err != nil {
		Log("daemon", "debug", fmt.Sprintf("error dialing %s - %s, err: %s", c.request, c.parameters, err))
		c.ws.Close()
		return
	}
	pipe(c.ws, c.conn)
}

// serveWs handles websocket requests from the peer.
func sockHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version, _ := strconv.Atoi(vars["version"])
	request := vars["request"]
	parameters, err := base64.StdEncoding.DecodeString(vars["parameters"])
	if err != nil {
		Log("daemon", "error", fmt.Sprintf("base64decode failed: %s", err))
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Log("daemon", "error", fmt.Sprintf("ws upgrade failed: %s", err))
		return
	}
	c := &connection{version: version, request: request, parameters: fmt.Sprintf("%s", parameters), ws: ws}
	go c.handle()
}
func listen() {
	addr := fmt.Sprintf("%s:%d", listenIp, listenPort)
	mux := mux.NewRouter()
	mux.HandleFunc("/sock/{version}/{request}/{parameters}", sockHandler)
	mux.HandleFunc("/", serveHome)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	Log("daemon", "info", "attemting upgrade of server to http/2")
	if err := http2.ConfigureServer(server, nil); err != nil {
		Log("daemon", "info", fmt.Sprintf("upgrade to http/2 failed: %s", err))
	}

	Log("daemon", "info", fmt.Sprintf("Listening on %s", addr))
	err := server.ListenAndServe()
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func main() {
	Log("daemon", "info", "starting proxy")
	flag.Parse()
	listen()
}

func init() {
	flag.IntVar(&listenPort, "port", 443, "port to listen on")
	flag.StringVar(&listenIp, "ip", "", "ip to bind to")
}
