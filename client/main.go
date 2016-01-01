package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

// Global variables
var debug bool
var verbose bool
var quiet bool
var help bool

var listenIp string
var listenPort int
var serverIp string
var serverPort int
var cpuprofile bool
var registerClient bool

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	flag.BoolVar(&cpuprofile, "cpuprofile", false, "show cpu profile on http://localhost:6060")
	flag.IntVar(&listenPort, "port", 1080, "port to listen on")
	flag.BoolVar(&debug, "debug", false, "show debug logging")
	flag.BoolVar(&verbose, "verbose", false, "show verbose logging")
	flag.BoolVar(&quiet, "quiet", false, "suppress logging")
	flag.BoolVar(&help, "help", false, "show usage")
	flag.BoolVar(&registerClient, "register", false, "register client to the server")
	flag.StringVar(&listenIp, "ip", "127.0.0.1", "ip to bind to")
	flag.IntVar(&serverPort, "sport", 443, "port to listen on")
	flag.StringVar(&serverIp, "sip", "127.0.0.1", "ip to bind to")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()
	if help {
		flag.Usage()
		return nil
	}

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
		fmt.Println("Registering client with server...")
		return register()
	}
	d, err := startDaemon()
	if err != nil {
		return err
	}
	return d.Run()
}
