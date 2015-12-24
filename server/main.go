package main

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  10240,
	WriteBufferSize: 10240,
}

var homeTempl = template.Must(template.ParseFiles("home.html"))

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
	homeTempl.Execute(w, r.Host)
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

// This example demonstrates a trivial echo server.
func main() {
	mux := mux.NewRouter()
	mux.HandleFunc("/sock/{version}/{request}/{parameters}", sockHandler)
	mux.HandleFunc("/", serveHome)

	Log("daemon", "info", "Listening to :12345")
	err := http.ListenAndServe(":12345", mux)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
