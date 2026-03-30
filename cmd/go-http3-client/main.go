// Package main implements the tool.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/quic-go/http3"
)

// SmartTransport manages the selection between H3 and TCP
type SmartTransport struct {
	h3      *http3.Transport
	tcp     *http.Transport
	h3Hosts sync.Map // Remembers which hosts support H3
	upgrade bool     // upgrade=true means detect http3 and upgrade from http2 attempt. upgrade=false means first attempt http3 and if it fails, fallback to http2.
}

func (s *SmartTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	host := req.URL.Host

	// MODE: AGGRESSIVE (Attempt H3 First)
	if !s.upgrade {
		// Use a context timeout so we don't hang forever if UDP is black-holed
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()

		resp, err := s.h3.RoundTrip(req.WithContext(ctx))
		if err == nil {
			return resp, nil
		}
		log.Printf("Immediate H3 attempt failed, falling back to TCP: %v", err)
		return s.tcp.RoundTrip(req)
	}

	// MODE: DISCOVERY (Browser-like)
	// 1. Check if we already discovered H3 for this host
	if _, supported := s.h3Hosts.Load(host); supported {
		resp, err := s.h3.RoundTrip(req)
		if err == nil {
			return resp, nil
		}
		log.Printf("H3 failed for %s, falling back to TCP: %v", host, err)
	}

	// 2. Use TCP (H1/H2)
	resp, err := s.tcp.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// 3. Discovery: If the server advertises H3, remember it for next time
	if altSvc := resp.Header.Get("Alt-Svc"); altSvc != "" {
		log.Printf("Discovered H3 support for %s via Alt-Svc", host)
		s.h3Hosts.Store(host, true)
	}

	return resp, nil
}

func main() {
	var serverURL string
	var insecureSkipVerifyTLS bool
	var upgrade bool
	flag.StringVar(&serverURL, "url", "https://localhost:8443", "Server URL to fetch")
	flag.BoolVar(&insecureSkipVerifyTLS, "insecureSkipVerifyTLS", true, "Skip TLS certificate verification")
	flag.BoolVar(&upgrade, "upgrade", false, "upgrade=true means detect http3 and upgrade from http2 attempt. upgrade=false means first attempt http3 and if it fails, fallback to http2.")
	flag.Parse()

	fmt.Printf("upgrade: %t\n", upgrade)

	tlsConf := &tls.Config{
		InsecureSkipVerify: insecureSkipVerifyTLS,
	}

	st := &SmartTransport{
		upgrade: upgrade,
		h3:      &http3.Transport{TLSClientConfig: tlsConf},
		tcp: &http.Transport{
			TLSClientConfig:   tlsConf,
			ForceAttemptHTTP2: true,
		},
	}

	client := &http.Client{
		Transport: st,
	}

	// Test it multiple times to see the switch
	for i := 1; i <= 3; i++ {
		fmt.Printf("\n--- Request #%d ---\n", i)
		resp, err := client.Get(serverURL)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Protocol: %s\nResponse: %s", resp.Proto, string(body))
		resp.Body.Close()
	}
}
