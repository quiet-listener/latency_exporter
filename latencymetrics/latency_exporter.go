package latencymetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/go-kit/kit/log"
	"strings"
	"sync"
	urlparse "net/url"
)
const (
	namespace = "latency" // For Prometheus metrics.
	Subsystem = "url_metrics" // For Prometheus metrics
)
var (
	urlm []*URLMetric
	dns_desc = prometheus.NewDesc(prometheus.BuildFQName(namespace,"url_metric","dns_latency"),"Time taken for DNS resolution to complete.",nil,nil)
	connect_desc = prometheus.NewDesc(prometheus.BuildFQName(namespace,"url_metric","connect_latency"),"Time taken for TCP connection to complete.",nil,nil)
	ssl_desc = prometheus.NewDesc(prometheus.BuildFQName(namespace,"url_metric","sslshake_latency"),"Time taken for SSL handshake to complete.",nil,nil)
	ttfb_desc = prometheus.NewDesc(prometheus.BuildFQName(namespace,"url_metric","ttfb_latency"),"Time taken till the first byte recieved.",nil,nil)
	rtt_desc = prometheus.NewDesc(prometheus.BuildFQName(namespace,"url_metric","rtt_latency"),"RTT to complete.",nil,nil)
)

type Exporter struct {
	mutex sync.RWMutex
	dnslatency	*prometheus.HistogramVec
	connectlatency *prometheus.HistogramVec
	sslshake *prometheus.HistogramVec
	ttfb *prometheus.HistogramVec
	rtt *prometheus.HistogramVec
	logger log.Logger
}

// Describe describes all the metrics ever exported. It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- dns_desc
	ch <- connect_desc
	ch <- ssl_desc
	ch <- ttfb_desc
	ch <- rtt_desc
}


func NewExporter(urlsstr string, delimiter string, logger log.Logger) (*Exporter, error){
	urls:=strings.Split(urlsstr,delimiter)
	for _, url := range urls {
		if _,err:=urlparse.Parse(url); err != nil {
			return nil,err
		}
		urlm=append(urlm, NewLatencyMetricObject(url))
	}
	return &Exporter{
		dnslatency : prometheus.NewHistogramVec( prometheus.HistogramOpts {
			Namespace : namespace,
			Name : "dns_latency",
			Help :  "Time taken for DNS resolution to complete.",
		}, urls ),
		connectlatency : prometheus.NewHistogramVec( prometheus.HistogramOpts {
			Namespace : namespace,
			Name : "connect_latency",
			Help :  "Time taken for TCP connection to complete.",
		}, urls ),
		sslshake : prometheus.NewHistogramVec( prometheus.HistogramOpts {
			Namespace : namespace,
			Name : "sslshake_latency",
			Help :  "Time taken for SSL handshake to complete.",
		}, urls ),
		ttfb : prometheus.NewHistogramVec( prometheus.HistogramOpts {
			Namespace : namespace,
			Name : "ttfb_latency",
			Help :  "Time taken till the first byte recieved.",
		}, urls ),
		rtt : prometheus.NewHistogramVec( prometheus.HistogramOpts {
			Namespace : namespace,
			Name : "rtt_latency",
			Help :  "RTT to complete.",
		}, urls ),
		logger: logger,
	}, nil
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for _,url := range urlm {
		go url.TimeLatency()
	}
	for _, url := range urlm {
		ch <- prometheus.MustNewConstMetric(dns_desc, prometheus.UntypedValue, float64(url.dns),url.url)
		ch <- prometheus.MustNewConstMetric(connect_desc, prometheus.UntypedValue,float64(url.connect),url.url)
		ch <- prometheus.MustNewConstMetric(ssl_desc, prometheus.UntypedValue,float64(url.sslshake),url.url)
		ch <- prometheus.MustNewConstMetric(ttfb_desc, prometheus.UntypedValue,float64(url.ttfb),url.url)
		ch <- prometheus.MustNewConstMetric(rtt_desc, prometheus.UntypedValue,float64(url.rtt),url.url)
		// e.dnslatency.WithLabelValues(url.url, ObserveVec(float64(url.dns)))
		// e.connectlatency.WithLabelValues(url.url, Observe(float64(url.connect)))
		// e.sslshake.WithLabelValues(url.url, Observe(float64(url.sslshake)))
		// e.ttfb.WithLabelValues(url.url, Observe(float64(url.ttfb)))
		// e.rtt.WithLabelValues(url.url, Observe(float64(url.ttfb)))
	}
}