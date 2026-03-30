// Package main implements the tool.
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/quic-go/quic-go/http3"
)

// SmartTransport manages the selection between H3 and TCP
type SmartTransport struct {
	h3      *http3.Transport
	tcp     *http.Transport
	h3Hosts sync.Map // Remembers which hosts support H3
}

func (s *SmartTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	host := req.URL.Host

	// 1. If we've seen this host support H3 before, try H3 first
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
		log.Printf("Discovered H3 support for %s", host)
		s.h3Hosts.Store(host, true)
	}

	return resp, nil
}

func main() {
	var serverURL string
	var insecureSkipVerifyTLS bool
	flag.StringVar(&serverURL, "url", "https://localhost:8443", "Server URL to fetch")
	flag.BoolVar(&insecureSkipVerifyTLS, "insecureSkipVerifyTLS", true, "Skip TLS certificate verification")
	flag.Parse()

	tlsConf := &tls.Config{
		InsecureSkipVerify: insecureSkipVerifyTLS,
	}

	st := &SmartTransport{
		h3: &http3.Transport{TLSClientConfig: tlsConf},
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
