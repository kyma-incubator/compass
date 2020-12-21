package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type AuditlogCollector struct {
	channelLength           prometheus.Gauge
	auditlogRequestDuration *prometheus.HistogramVec
}

func NewAuditlogMetricCollector() *AuditlogCollector {
	return &AuditlogCollector{
		channelLength: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "compass",
			Subsystem: "gateway",
			Name:      "auditlog_channel_length",
			Help:      "current audit log async channel size",
		}),
		auditlogRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "compass",
			Subsystem: "gateway",
			Name:      "auditlog_request_duration_seconds",
			Help:      "Duration of HTTP Requests to Auditlog",
		}, []string{"code", "method"}),
	}
}

func (c *AuditlogCollector) Describe(ch chan<- *prometheus.Desc) {
	c.channelLength.Describe(ch)
	c.auditlogRequestDuration.Describe(ch)
}

func (c *AuditlogCollector) Collect(ch chan<- prometheus.Metric) {
	c.channelLength.Collect(ch)
	c.auditlogRequestDuration.Collect(ch)
}

func (c *AuditlogCollector) SetChannelSize(size int) {
	c.channelLength.Set(float64(size))
}

func (c *AuditlogCollector) InstrumentAuditlogHTTPClient(client *http.Client) {
	client.Transport = promhttp.InstrumentRoundTripperDuration(c.auditlogRequestDuration, client.Transport)
}
