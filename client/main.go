package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"

	"github.com/jsimonetti/tlstun/shared"
)

var listenIp string
var listenPort int

var serverIp string
var serverPort int

var registerClient bool

func main() {
	flag.Parse()

	if cpuprofile {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
		/*
			go tool pprof http://localhost:6060/debug/pprof/heap

			go tool pprof http://localhost:6060/debug/pprof/profile

			go tool pprof http://localhost:6060/debug/pprof/block

			wget http://localhost:6060/debug/pprof/trace?seconds=5
		*/
	}

	if registerClient {
		shared.Log("client", "info", "registering client with server")
		register()
		return
	}
	forward(serverIp, serverPort)
}

var cpuprofile bool

func init() {
	flag.BoolVar(&cpuprofile, "cpuprofile", false, "show cpu profile on http://localhost:6060")
	flag.IntVar(&listenPort, "port", 1080, "port to listen on")
	flag.BoolVar(&shared.ShowLog, "log", false, "show logging")
	flag.BoolVar(&registerClient, "register", false, "register client to the server")
	flag.StringVar(&listenIp, "ip", "127.0.0.1", "ip to bind to")
	flag.IntVar(&serverPort, "sport", 12345, "port to listen on")
	flag.StringVar(&serverIp, "sip", "62.148.169.249", "ip to bind to")
}
