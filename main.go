package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	lm "github.com/quiet-listener/latency_exporter/latency_metrics"
)

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9101").String()
		metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		urls = kingpin.Flag("web.urls-list","List of urls to for Metrics expose").Default("")
	)
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)
	level.Info(logger).Log("msg", "Starting latency_mettrics_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())
	exporter, err := NewExporter(*haProxyScrapeURI, *haProxySSLVerify, selectedServerMetrics, *haProxyTimeout, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Error creating an exporter", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector("haproxy_exporter"))
	router := gin.New()
	router.Use(gin.Logger())
	url := "https://www.google.com"
	um:= lm.NewLatencyMetricObject(url)
	um.TimeLatency()
	fmt.Println(um)
	router.Run(":" + port)
}