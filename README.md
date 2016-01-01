# TLSTun

TLSTun is a [Go](http://golang.org/) client server application to tunnel through highly intelligent
firewalls.


The client will connect to the server component over a [WebSockets](http://www.rfc-editor.org/rfc/rfc6455.txt).
All client connections are then muxed over the WebSocket
to the server which connects the to a socks [Socks](https://en.wikipedia.org/wiki/SOCKS) proxy.
The server then proxies the connection from the mux to the real
destination.

This will punch through firewalls that do actual inspection of
traffic.



-- Options for client component:
```
    -ip         Ip address to listen on (This will be your Socks5 ip)
    -port       Local port to bind to (This will be your Socks5 port)
    -sip        Ip of the server component to connect the websockets to
    -sport      Port of the server component to connect the websockets to
    -register   Prompt for server password to register your certificate on the server
    -help       Show the usage
    -cpuprofile Add cpuprofiling to the webserver (see source for more detail)

Log level settings:
  (By default only errors lvl Error and up are shown)
    -debug      Show debug logging
    -verbose    Show info logging
    -quiet      Do not show any logging
```

-- Options for server component:
```
    -ip         Ip address to listen on
    -port       Local port to bind to
    -regpass    Password to use when registering clients. Only needed during
                client registering.
    -help       Show the usage
    -cpuprofile Add cpuprofiling to the webserver (see source for more detail)

Log level settings:
  (By default only errors lvl Error and up are shown)
    -debug      Show debug logging
    -verbose    Show info logging
    -quiet      Do not show any logging
```

### TODO:
- add passthrough functionality to server to allow running it in front of an existing
webserver


Contributions to this project are welcomed!
