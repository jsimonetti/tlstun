package main

import (
	"flag"

	"github.com/jsimonetti/tlstun/shared"
)

var listenIp string
var listenPort int

var serverIp string
var serverPort int

func main() {
	flag.Parse()
	forward(serverIp, serverPort)
}

func init() {
	flag.IntVar(&listenPort, "port", 1080, "port to listen on")
	flag.BoolVar(&shared.ShowLog, "log", false, "show logging")
	flag.StringVar(&listenIp, "ip", "127.0.0.1", "ip to bind to")
	flag.IntVar(&serverPort, "sport", 12345, "port to listen on")
	flag.StringVar(&serverIp, "sip", "62.148.169.249", "ip to bind to")
}
