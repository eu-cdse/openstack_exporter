package main

import (
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	lib "github.com/eu-cdse/openstack_exporter/internal"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
)

var (
	config      = promlog.Config{}
	volumeLimit = kingpin.Flag("volume.limit", "Max number of volumes when on OTC").Default("-1").Float64()
)

func main() {
	kingpin.CommandLine.UsageWriter(os.Stdout)
	flag.AddFlags(kingpin.CommandLine, &config)
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Print("openstack_exporter"))
	kingpin.Parse()

	// Initialize the logger
	logger := promlog.New(&config)

	lib.SetLogger(logger)

	openStack := lib.NewOpenStackCollector(*volumeLimit)
	// Custom registry to not collect all go low-level metrics
	promRegistry := prometheus.NewRegistry()
	promRegistry.MustRegister(openStack)
	handler := promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{})

	http.Handle("/metrics", handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			 <head><title>OpenStack Exporter</title></head>
			 <body>
			 <h1>OpenStack Exporter</h1>
			 <p><a href='/metrics'>Metrics</a></p>
			 </body>
			 </html>`))
	})
	level.Info(logger).Log("msg", "Starting exporter", "address", ":9595")

	if err := http.ListenAndServe(":9595", nil); err != nil {
		level.Error(logger).Log("msg", "Failed to start HTTP server", "err", err)
	}
}
