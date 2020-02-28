package latencymetrics

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
	um.dns = time.Duration(0 * time.Second)
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
				um.sslshake = time.Duration(0 * time.Second)
			}
			um.sslshake = time.Since(sslshake)
		},
		ConnectStart: func(network, addr string) {
			connect = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				level.Error(e.logger).Log("msg", "Error Conection Time", "err", err)
				um.connect = time.Duration(0 * time.Second)
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
	// customTransport can be updated for differrent Values as per requirement
	var customTransport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 2 * time.Second,
			DualStack: true,
		}).DialContext,
		DisableKeepAlives:     true,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          5,
		IdleConnTimeout:       5 * time.Second,
		TLSHandshakeTimeout:   1 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if _, err := customTransport.RoundTrip(req); err != nil {
		level.Error(e.logger).Log("msg", "Error Completeing RTTs", req.URL, "err", err)
		return err
	}
	um.rtt = time.Since(start)
	return err
}
