# TLSTun

TLSTun is a [Go](http://golang.org/) client server application to tunnel through highly intelligent
firewalls.


The client will bind a local port as a [Socks5](https://en.wikipedia.org/wiki/SOCKS) server. All incomming connections
are tunneled to the server component over
[WebSockets](http://www.rfc-editor.org/rfc/rfc6455.txt).

The server simply forwards the connection from the websocket to the real
destination.

This will punch through all known firewalls that do actual inspection of
traffic.


### TODO:
- add certificate authentication to server and client component
- add passthrough functionality to server to allow running it in front of an existing
webserver
- add proxy support for client component
