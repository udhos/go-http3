# go-http3

[go-http3](https://github.com/udhos/go-http3) shows how to serve both HTTP3 and HTTP2 and how to make the client to use HTTP3 when available and fallback to HTTP2 when not.

# Build

```bash
./build.sh
```

# Serving http3

```bash
$ go-http3-server 
2026/03/29 22:29:49 Starting HTTP/1.1 & HTTP/2 (TCP) on :8443
2026/03/29 22:29:49 Starting HTTP/3 (UDP) on :8443
2026/03/29 22:29:53 Served request from 127.0.0.1:48892 via HTTP/2.0
2026/03/29 22:29:53 Served request from 127.0.0.1:53448 via HTTP/3.0
2026/03/29 22:29:53 Served request from 127.0.0.1:53448 via HTTP/3.0
```

# Serving http2

```bash
$ go-http3-server -http3=false
2026/03/29 22:30:08 HTTP/3 is disabled.
2026/03/29 22:30:08 Starting HTTP/1.1 & HTTP/2 (TCP) on :8443
2026/03/29 22:30:11 Served request from 127.0.0.1:59688 via HTTP/2.0
2026/03/29 22:30:11 Served request from 127.0.0.1:59688 via HTTP/2.0
2026/03/29 22:30:11 Served request from 127.0.0.1:59688 via HTTP/2.0
```

# Running the client

```bash
$ go-http3-client 

--- Request #1 ---
Protocol: HTTP/2.0
Response: Current Time: 2026-03-29T22:30:11-03:00
Protocol: HTTP/2.0

--- Request #2 ---
Protocol: HTTP/2.0
Response: Current Time: 2026-03-29T22:30:11-03:00
Protocol: HTTP/2.0

--- Request #3 ---
Protocol: HTTP/2.0
Response: Current Time: 2026-03-29T22:30:11-03:00
Protocol: HTTP/2.0
```

# Creating certificates

```bash
openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes \
  -keyout key.pem -out cert.pem \
  -subj "/CN=localhost" \
  -addext "subjectAltName = DNS:localhost, IP:127.0.0.1"
```

# Curl for http2

```bash
$ curl -k https://localhost:8443
Current Time: 2026-03-29T22:06:33-03:00
Protocol: HTTP/2.0
```

# Error on receive buffer size

```bash
$ go-http3-server 
2026/03/29 22:11:58 Starting HTTP/1.1 & HTTP/2 (TCP) on :8443
2026/03/29 22:11:58 Starting HTTP/3 (UDP) on :8443
2026/03/29 22:11:58 failed to sufficiently increase receive buffer size (was: 208 kiB, wanted: 7168 kiB, got: 416 kiB). See https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes for details.
2026/03/29 22:12:03 Served request from 127.0.0.1:53816 via HTTP/3.0
2026/03/29 22:12:03 Served request from 127.0.0.1:53816 via HTTP/3.0
```

Fix:

```bash
sudo sysctl -w net.core.rmem_max=7500000
sudo sysctl -w net.core.wmem_max=7500000
```

See: https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes
