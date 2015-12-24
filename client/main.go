package main

import (
	"flag"
)

var listenIp string
var listenPort int

func main() {
	Log("daemon", "info", "starting proxy")
	flag.Parse()
	listen()
}

func init() {
	flag.IntVar(&listenPort, "port", 1080, "port to listen on")
	flag.StringVar(&listenIp, "ip", "127.0.0.1", "ip to bind to")
}
