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
	namespace = "latency_exporter" // For Prometheus metrics.
	subsystem = "url_metrics"      // For Prometheus metrics
)

var (
	urlLabel = []string{"url"}
	urlm     []*URLMetric
)

// Exporter struc for the latency Exporter
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

// NewExporter creates urlm objects and return exporter
func NewExporter(urlsstr string, delimiter string, logger log.Logger) (*Exporter, error) {
	urlLoop := strings.Split(urlsstr, delimiter)
	for _, url := range urlLoop {
		if _, err := urlparse.Parse(url); err != nil {
			return nil, err
		}
		urlm = append(urlm, NewLatencyMetricObject(url))
	}
	return &Exporter{
		dnslatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "dns_latency_sec",
			Subsystem: "url_metric",
			Help:      "Time taken for DNS resolution to complete in Seconds",
		}, urlLabel),
		connectlatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "connect_latency_sec",
			Subsystem: "url_metric",
			Help:      "Time taken for TCP connection to complete in Seconds",
		}, urlLabel),
		sslshake: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "sslshake_latency_sec",
			Subsystem: "url_metric",
			Help:      "Time taken for SSL handshake to complete in Seconds",
		}, urlLabel),
		ttfb: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "ttfb_latency_sec",
			Subsystem: "url_metric",
			Help:      "Time taken till the first byte recieved in Seconds",
		}, urlLabel),
		rtt: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "rtt_latency_sec",
			Subsystem: "url_metric",
			Help:      "RTT to complete in Seconds",
		}, urlLabel),
		logger: logger,
	}, nil
}

// Collect implementation for prometheus exporter
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
	e.dnslatency.WithLabelValues(url.url).Observe((float64(url.dns) / float64(time.Second)))
	e.connectlatency.WithLabelValues(url.url).Observe((float64(url.connect) / float64(time.Second)))
	e.sslshake.WithLabelValues(url.url).Observe((float64(url.sslshake) / float64(time.Second)))
	e.ttfb.WithLabelValues(url.url).Observe((float64(url.ttfb) / float64(time.Second)))
	e.rtt.WithLabelValues(url.url).Observe((float64(url.rtt) / float64(time.Second)))
}
