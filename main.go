package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log/level"
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	lm "github.com/quiet-listener/latency_exporter/latencymetrics"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		port        = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default("9101").String()
		metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		urls        = kingpin.Flag("web.urls-list", "List of urls to for Metrics expose").Required().String()
		delimiter   = kingpin.Flag("web.url-delimiter", "Delimiter used to split url").Default(",").String()
	)
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)
	version.Branch = "master"
	version.Version = "v1"
	version.Revision = "1"
	level.Info(logger).Log("msg", "Starting latency_metrics_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())
	exporter, err := lm.NewExporter(*urls, *delimiter, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Error creating an exporter", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector("latency_exporter"))
	router := gin.New()
	router.LoadHTMLGlob("templates/*")
	router.StaticFile("/favicon.ico", "statics/favicon.ico")
	router.Use(gin.Logger())
	router.GET(*metricsPath, gin.WrapH(promhttp.Handler()))
	router.GET("/", startPage)
	level.Info(logger).Log("msg", "Listening on port", "port", *port)
	if err := router.Run(":" + *port); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}

func startPage(c *gin.Context) {
	c.HTML(http.StatusOK, "startPage.tmpl", gin.H{
		"title":       "Latency Exporter",
		"metric_path": "/metrics",
	})
}
