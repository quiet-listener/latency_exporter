package latencymetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/go-kit/kit/log"
	"strings"
	"sync"
	"time"
	urlparse "net/url"
)
const (
	namespace = "latency" // For Prometheus metrics.
	Subsystem = "url_metrics" // For Prometheus metrics
)
var (
	url_label = []string {"url"}
	urlm []*URLMetric
	dns_desc = prometheus.NewDesc(prometheus.BuildFQName(namespace,"url_metric","dns_latency"),"Time taken for DNS resolution to complete.",url_label,nil)
	connect_desc = prometheus.NewDesc(prometheus.BuildFQName(namespace,"url_metric","connect_latency"),"Time taken for TCP connection to complete.",url_label,nil)
	ssl_desc = prometheus.NewDesc(prometheus.BuildFQName(namespace,"url_metric","sslshake_latency"),"Time taken for SSL handshake to complete.",url_label,nil)
	ttfb_desc = prometheus.NewDesc(prometheus.BuildFQName(namespace,"url_metric","ttfb_latency"),"Time taken till the first byte recieved.",url_label,nil)
	rtt_desc = prometheus.NewDesc(prometheus.BuildFQName(namespace,"url_metric","rtt_latency"),"RTT to complete.",url_label,nil)
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
	url_loop := strings.Split(urlsstr,delimiter)
	for _, url := range url_loop{
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
		}, url_label),
		connectlatency : prometheus.NewHistogramVec( prometheus.HistogramOpts {
			Namespace : namespace,
			Name : "connect_latency",
			Help :  "Time taken for TCP connection to complete.",
		}, url_label ),
		sslshake : prometheus.NewHistogramVec( prometheus.HistogramOpts {
			Namespace : namespace,
			Name : "sslshake_latency",
			Help :  "Time taken for SSL handshake to complete.",
		}, url_label ),
		ttfb : prometheus.NewHistogramVec( prometheus.HistogramOpts {
			Namespace : namespace,
			Name : "ttfb_latency",
			Help :  "Time taken till the first byte recieved.",
		}, url_label ),
		rtt : prometheus.NewHistogramVec( prometheus.HistogramOpts {
			Namespace : namespace,
			Name : "rtt_latency",
			Help :  "RTT to complete.",
		}, url_label),
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
		ch <- prometheus.MustNewConstMetric(dns_desc, prometheus.GaugeValue, (float64(url.dns)/float64(time.Millisecond)),url.url)
		ch <- prometheus.MustNewConstMetric(connect_desc, prometheus.GaugeValue,(float64(url.connect)/float64(time.Millisecond)),url.url)
		ch <- prometheus.MustNewConstMetric(ssl_desc, prometheus.GaugeValue,(float64(url.sslshake)/float64(time.Millisecond)),url.url)
		ch <- prometheus.MustNewConstMetric(ttfb_desc, prometheus.GaugeValue,(float64(url.ttfb)/float64(time.Millisecond)),url.url)
		ch <- prometheus.MustNewConstMetric(rtt_desc, prometheus.GaugeValue,(float64(url.rtt)/float64(time.Millisecond)),url.url)
	}
}