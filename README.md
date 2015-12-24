# TLSTun

TLSTun is a [Go](http://golang.org/) client server application to tunnel through highly intelligent
firewalls.


The client will bind a local port as a Socks5 server. All incomming connections
are tunneled to the server component over websockets.

The server simply forwards the connection from the websocket to the read
destination.

This will punch through all known firewalls that do actual inspection of
traffic.


### TODO:
- add TLS listener to server component
- add certificate authentication to server and client component
