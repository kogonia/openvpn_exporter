package exporters

import "github.com/prometheus/client_golang/prometheus"

// 	updated                     string
//	totalBytesIn, totalBytesOut int
//	clients                     map[string]client

func prom() {
	var desc = prometheus.NewDesc(
		"openvpn_server",
		"Curent status of server",
		nil, nil,
	)
	s := prometheus.NewMetricWithTimestamp(
		ov.updated,
		prometheus.MustNewConstMetric(
			desc, prometheus.GaugeValue))
}
