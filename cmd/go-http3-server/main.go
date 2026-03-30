// Package main implements the tool.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/quic-go/quic-go/http3"
)

func main() {
	var addr string
	var http3Enabled bool
	flag.StringVar(&addr, "addr", ":8443", "Address to listen on")
	flag.BoolVar(&http3Enabled, "http3", true, "Enable HTTP/3 (QUIC) support")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 1. Discovery: If H3 is enabled, tell the client it exists
		if http3Enabled {
			w.Header().Set("Alt-Svc", `h3="`+addr+`"; ma=86400`)
		}

		// 2. Logic: Return the current time and the protocol used
		currentTime := time.Now().Format(time.RFC3339)
		fmt.Fprintf(w, "Current Time: %s\nProtocol: %s\n", currentTime, r.Proto)
		log.Printf("Served request from %s via %s", r.RemoteAddr, r.Proto)
	})

	certFile := "cert.pem"
	keyFile := "key.pem"

	// START HTTP/3 (UDP)
	if http3Enabled {
		go func() {
			log.Println("Starting HTTP/3 (UDP) on " + addr)
			err := http3.ListenAndServeQUIC(addr, certFile, keyFile, mux)
			if err != nil {
				log.Fatalf("H3 Server Error: %v", err)
			}
		}()
	} else {
		log.Println("HTTP/3 is disabled.")
	}

	// START HTTP/1.1 & HTTP/2 (TCP)
	log.Println("Starting HTTP/1.1 & HTTP/2 (TCP) on " + addr)
	err := http.ListenAndServeTLS(addr, certFile, keyFile, mux)
	if err != nil {
		log.Fatal(err)
	}
}
