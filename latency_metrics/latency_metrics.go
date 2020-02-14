package latency_metrics

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"
)

type latencyMetric interface {
	TimeLatency()
}

//UrlMetric struct for urlmetric
type UrlMetric struct {
	url      string
	dns      time.Duration
	connect  time.Duration
	sslshake time.Duration
	ttfb     time.Duration
	rtt      time.Duration
}

// NewLatencyMetricObject Creates UrlMetric Object and returns
func NewLatencyMetricObject(url string) *UrlMetric {
	return &UrlMetric{url: url}
}

func (um *UrlMetric) String() string {
	return fmt.Sprintf("url: %s\nDns: %v\nConnect: %v\nSSL Handshake : %v\nTTFB : %v\nRTT: %v",um.url,um.dns,um.connect,um.sslshake,um.ttfb,um.rtt)
}

func (um *UrlMetric) TimeLatency() {
	var start , dns, connect, sslshake time.Time
	req, err := http.NewRequest("GET", um.url, nil)
	if err != nil {
		log.Fatal(err)
	}
	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) {
			dns = time.Now()
		},
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			um.dns = time.Since(dns)
		},
		TLSHandshakeStart: func() {
			sslshake = time.Now()
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			if err != nil {
				log.Fatal(err)
			}
			um.sslshake = time.Since(sslshake)
		},
		ConnectStart: func(network, addr string) {
			connect = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				log.Fatal(err)
			}
			um.connect = time.Since(connect)
		},

		GotFirstResponseByte: func() {
			um.ttfb = time.Since(start)
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
		log.Fatal(err)
	}
	um.rtt = time.Since(start)
}