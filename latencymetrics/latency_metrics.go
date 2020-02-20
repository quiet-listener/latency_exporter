package latencymetrics

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"time"
)

type latencyMetric interface {
	TimeLatency()
}

//URLMetric struct for urlmetric
type URLMetric struct {
	url      string
	dns      time.Duration
	connect  time.Duration
	sslshake time.Duration
	ttfb     time.Duration
	rtt      time.Duration
}

// NewLatencyMetricObject Creates URLMetric Object and returns
func NewLatencyMetricObject(urlStr string) *URLMetric {
	parsedURL, parseErr := url.Parse(urlStr)
	logger := log.NewLogfmtLogger(os.Stderr)
	if parseErr != nil {
		level.Error(logger).Log("msg", "Invalid Url", "err", parseErr)
	} else {
		if parsedURL.Scheme == "" {
			level.Info(logger).Log("msg", "url scheme provided is Empty Falling back to http")
			urlStr = "http://" + urlStr
		}
	}
	return &URLMetric{url: urlStr}
}

func (um *URLMetric) String() string {
	return fmt.Sprintf("url: %s\nDns: %v\nConnect: %v\nSSL Handshake : %v\nTTFB : %v\nRTT: %v", um.url, um.dns, um.connect, um.sslshake, um.ttfb, um.rtt)
}

// TimeLatency collects latency metrics and updats URLMetric
func (um *URLMetric) TimeLatency(e *Exporter) error {
	var start, dns, connect, sslshake time.Time
	req, err := http.NewRequest("GET", um.url, nil)
	if err != nil {
		level.Error(e.logger).Log("msg", "Error Creating New Request", "err", err)
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
				level.Error(e.logger).Log("msg", "Error TLS Handshake", "err", err)
				um.sslshake = time.Duration(0 * time.Millisecond)
			}
			um.sslshake = time.Since(sslshake)
		},
		ConnectStart: func(network, addr string) {
			connect = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				level.Error(e.logger).Log("msg", "Error Conection Time", "err", err)
				um.connect = time.Duration(0 * time.Millisecond)
			}
			um.connect = time.Since(connect)
		},

		GotFirstResponseByte: func() {
			um.ttfb = time.Since(start)
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
		level.Error(e.logger).Log("msg", "Error Completeing RTTs", req.URL, "err", err)
		return err
	}
	um.rtt = time.Since(start)
	return err
}
