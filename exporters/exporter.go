package exporters

import (
	"fmt"
	"log"
	"net"

	"github.com/prometheus/client_golang/prometheus"
)

type OpenVPNExporter struct {
	statusPath, updated         string
	totalBytesIn, totalBytesOut float64
	clients                     map[string]client
}

type client struct {
	commonName               string
	virtAddr, realAddr       net.IP
	connSince, lastRef       string
	bytesReceived, bytesSent float64
}

var ovpn OpenVPNExporter

var (
	statusDescr = prometheus.NewDesc(
		prometheus.BuildFQName("openvpn", "server", "status"),
		"Whether scraping OpenVPNExporter's metrics was successful.",
		[]string{"status_path"}, nil)

	clCountDescr = prometheus.NewDesc(
		prometheus.BuildFQName("openvpn", "clients", "count"),
		"OpenVPNExporter's clients count.",
		nil, nil)

	rxDescr = prometheus.NewDesc(
		prometheus.BuildFQName("openvpn", "bytes", "received_total"),
		"OpenVPNExporter's bytes received summary.",
		nil, nil)

	txDescr = prometheus.NewDesc(
		prometheus.BuildFQName("openvpn", "bytes", "sent_total"),
		"OpenVPNExporter's bytes sent summary.",
		nil, nil)

	clientLabels = []string{"common_name", "virt_addr", "real_addr", "conn_since", "bytes_received", "bytes_sent", "last_req"}
	clDescr      = prometheus.NewDesc(
		prometheus.BuildFQName("openvpn", "clients", "info"),
		"Client's info.",
		clientLabels, nil)
)

func NewOVPNExporter(path string) *OpenVPNExporter {
	ovpn.statusPath = path
	ovpn.clients = make(map[string]client, 128)

	return &ovpn
}

func (ov *OpenVPNExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- statusDescr
	ch <- clCountDescr
	ch <- rxDescr
	ch <- txDescr
	ch <- clDescr
}

func (ov *OpenVPNExporter) Collect(ch chan<- prometheus.Metric) {
	err := parseStatusFile()
	if err != nil {
		log.Println(err)
		ch <- prometheus.MustNewConstMetric(
			statusDescr, prometheus.GaugeValue, 0, ov.statusPath,
		)
		return
	}
	ch <- prometheus.MustNewConstMetric(
		statusDescr, prometheus.GaugeValue, 1, ov.statusPath,
	)
	ch <- prometheus.MustNewConstMetric(
		clCountDescr, prometheus.GaugeValue, float64(len(ov.clients)),
	)
	ch <- prometheus.MustNewConstMetric(
		rxDescr, prometheus.GaugeValue, ov.totalBytesIn,
	)
	ch <- prometheus.MustNewConstMetric(
		txDescr, prometheus.GaugeValue, ov.totalBytesOut,
	)

	if len(ov.clients) == 0 {
		ch <- prometheus.MustNewConstMetric(
			clDescr, prometheus.GaugeValue, 0, "", "", "", "", "", "", "",
		)
	} else {
		for _, cl := range ov.clients {
			ch <- prometheus.MustNewConstMetric(
				clDescr, prometheus.GaugeValue, 1, cl.commonName, cl.virtAddr.String(), cl.realAddr.String(), cl.connSince, fmt.Sprintf("%f", cl.bytesReceived), fmt.Sprintf("%f", cl.bytesSent), cl.lastRef,
			)
		}
	}
}
