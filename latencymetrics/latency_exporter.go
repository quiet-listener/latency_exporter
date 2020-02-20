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
	Subsystem = "url_metrics" // For Prometheus metrics
)

var (
	url_label    = []string{"url"}
	urlm         []*URLMetric
	dns_desc     = prometheus.NewDesc(prometheus.BuildFQName(namespace, "url_metric", "dns_latency"), "Time taken for DNS resolution to complete.", url_label, nil)
	connect_desc = prometheus.NewDesc(prometheus.BuildFQName(namespace, "url_metric", "connect_latency"), "Time taken for TCP connection to complete.", url_label, nil)
	ssl_desc     = prometheus.NewDesc(prometheus.BuildFQName(namespace, "url_metric", "sslshake_latency"), "Time taken for SSL handshake to complete.", url_label, nil)
	ttfb_desc    = prometheus.NewDesc(prometheus.BuildFQName(namespace, "url_metric", "ttfb_latency"), "Time taken till the first byte recieved.", url_label, nil)
	rtt_desc     = prometheus.NewDesc(prometheus.BuildFQName(namespace, "url_metric", "rtt_latency"), "RTT to complete.", url_label, nil)
)

type Exporter struct {
	mutex          sync.RWMutex
	dnslatency     prometheus.Histogram
	connectlatency prometheus.Histogram
	sslshake       prometheus.Histogram
	ttfb           prometheus.Histogram
	rtt            prometheus.Histogram
	logger         log.Logger
}

// Describe describes all the metrics ever exported. It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- dns_desc
	ch <- connect_desc
	ch <- ssl_desc
	ch <- ttfb_desc
	ch <- rtt_desc
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
		dnslatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "dns_latency",
			Help:      "Time taken for DNS resolution to complete.",
		}),
		connectlatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "connect_latency",
			Help:      "Time taken for TCP connection to complete.",
		}),
		sslshake: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "sslshake_latency",
			Help:      "Time taken for SSL handshake to complete.",
		}),
		ttfb: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "ttfb_latency",
			Help:      "Time taken till the first byte recieved.",
		}),
		rtt: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "rtt_latency",
			Help:      "RTT to complete.",
		}),
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
}

func (e *Exporter) scrape(url *URLMetric, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := url.TimeLatency(e); err != nil {
		level.Error(e.logger).Log("msg", "Error starting HTTP server", "err", err)
	}
	ch <- prometheus.MustNewConstMetric(dns_desc, prometheus.GaugeValue, (float64(url.dns) / float64(time.Millisecond)), url.url)
	ch <- prometheus.MustNewConstMetric(connect_desc, prometheus.GaugeValue, (float64(url.connect) / float64(time.Millisecond)), url.url)
	ch <- prometheus.MustNewConstMetric(ssl_desc, prometheus.GaugeValue, (float64(url.sslshake) / float64(time.Millisecond)), url.url)
	ch <- prometheus.MustNewConstMetric(ttfb_desc, prometheus.GaugeValue, (float64(url.ttfb) / float64(time.Millisecond)), url.url)
	ch <- prometheus.MustNewConstMetric(rtt_desc, prometheus.GaugeValue, (float64(url.rtt) / float64(time.Millisecond)), url.url)
}
