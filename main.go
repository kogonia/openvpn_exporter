package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/kogonia/openvpn_exporter/exporters"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var (
		listenAddress      = flag.String("port", ":9119", "port to listen.")
		metricsPath        = flag.String("path", "/metrics", "path under which to expose metrics.")
		openvpnStatusPaths = flag.String("dir", "examples/example.openvpn-status.log", "paths at which openVPN places its status files.")
	)
	flag.Parse()

	log.Printf("Starting openVPN Exporter\n")
	log.Printf("Listen address: %v\n", *listenAddress)
	log.Printf("Metrics path: %v\n", *metricsPath)
	log.Printf("openvpn status path: %v\n", *openvpnStatusPaths)

	exporter := exporters.NewOVPNExporter(*openvpnStatusPaths)
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
					<html>
					<head><title>openVPN Exporter</title></head>
					<body>
					<h1>openVPN Exporter</h1>
					<p><a href='` + *metricsPath + `'>Metrics</a></p>
					</body>
					</html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
