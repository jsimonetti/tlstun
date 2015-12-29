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



-- Options for client component:
```
    -log        Show logging
    -ip         Ip address to listen on (This will be your Socks5 ip)
    -port       Local port to bind to (This will be your Socks5 port)
    -sip        Ip of the server component to connect the websockets to
    -sport      Port of the server component to connect the websockets to
    -register   Prompt for server password to register your certificate on the server
```

-- Options for server component:
```
    -log        Show logging
    -ip         Ip address to listen on
    -port       Local port to bind to
    -timeout    Timeout for reading from the websockets
                (defaults to 10 seconds, set to whatever your application needs)
                (SSH can send keepalive packets, so configure that instead of
                 incresing the websocket timeout)
    -regpass    Password to use when registering clients. Only needed during
                client registering.
```

### TODO:
- add passthrough functionality to server to allow running it in front of an existing
webserver


Contributions to this project are welcomed!
