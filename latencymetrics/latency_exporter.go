package latencymetrics

import (
	urlparse "net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "latency"     // For Prometheus metrics.
	subsystem = "url_metrics" // For Prometheus metrics
)

var (
	urlLabel = []string{"url"}
	urlm     []*URLMetric
)

type Exporter struct {
	mutex          sync.RWMutex
	dnslatency     *prometheus.HistogramVec
	connectlatency *prometheus.HistogramVec
	sslshake       *prometheus.HistogramVec
	ttfb           *prometheus.HistogramVec
	rtt            *prometheus.HistogramVec
	logger         log.Logger
}

// Describe describes all the metrics ever exported. It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.dnslatency.Describe(ch)
	e.connectlatency.Describe(ch)
	e.connectlatency.Describe(ch)
	e.sslshake.Describe(ch)
	e.ttfb.Describe(ch)
	e.rtt.Describe(ch)
}

func NewExporter(urlsstr string, delimiter string, logger log.Logger) (*Exporter, error) {
	url_loop := strings.Split(urlsstr, delimiter)
	for _, url := range url_loop {
		if _, err := urlparse.Parse(url); err != nil {
			return nil, err
		}
		urlm = append(urlm, NewLatencyMetricObject(url))
	}
	return &Exporter{
		dnslatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "dns_latency",
			Subsystem: "url_metric",
			Help:      "Time taken for DNS resolution to complete.",
			Buckets:   prometheus.LinearBuckets(0, 10, 100),
		}, urlLabel),
		connectlatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "connect_latency",
			Help:      "Time taken for TCP connection to complete.",
			Buckets:   prometheus.LinearBuckets(0, 10, 100),
		}, urlLabel),
		sslshake: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "sslshake_latency",
			Help:      "Time taken for SSL handshake to complete.",
			Buckets:   prometheus.LinearBuckets(0, 20, 50),
		}, urlLabel),
		ttfb: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "ttfb_latency",
			Help:      "Time taken till the first byte recieved.",
			Buckets:   prometheus.LinearBuckets(0, 20, 50),
		}, urlLabel),
		rtt: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "rtt_latency",
			Help:      "RTT to complete.",
			Buckets:   prometheus.LinearBuckets(0, 20, 50),
		}, urlLabel),
		logger: logger,
	}, nil
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	var wg sync.WaitGroup
	for _, url := range urlm {
		wg.Add(1)
		go e.scrape(url, ch, &wg)
	}
	wg.Wait()
	e.dnslatency.Collect(ch)
	e.connectlatency.Collect(ch)
	e.sslshake.Collect(ch)
	e.ttfb.Collect(ch)
	e.rtt.Collect(ch)
}

func (e *Exporter) scrape(url *URLMetric, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := url.TimeLatency(e); err != nil {
		level.Error(e.logger).Log("msg", "Error starting HTTP server", "err", err)
	}
	e.dnslatency.WithLabelValues(url.url).Observe((float64(url.dns) / float64(time.Millisecond)))
	e.connectlatency.WithLabelValues(url.url).Observe((float64(url.connect) / float64(time.Millisecond)))
	e.sslshake.WithLabelValues(url.url).Observe((float64(url.sslshake) / float64(time.Millisecond)))
	e.ttfb.WithLabelValues(url.url).Observe((float64(url.ttfb) / float64(time.Millisecond)))
	e.rtt.WithLabelValues(url.url).Observe((float64(url.rtt) / float64(time.Millisecond)))
}
