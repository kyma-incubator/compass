package metrics

import "github.com/prometheus/client_golang/prometheus"

type AuditlogCollector struct {
	channelLength prometheus.Gauge
}

func NewAuditlogMetricCollector() *AuditlogCollector {
	return &AuditlogCollector{
		channelLength: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "compass",
			Subsystem: "gateway",
			Name:      "auditlog_channel_length",
			Help:      "current audit log async channel size",
		}),
	}
}

func (c *AuditlogCollector) Describe(ch chan<- *prometheus.Desc) {
	c.channelLength.Describe(ch)
}

func (c *AuditlogCollector) Collect(ch chan<- prometheus.Metric) {
	c.channelLength.Collect(ch)
}

func (c *AuditlogCollector) SetChannelSize(size int) {
	c.channelLength.Set(float64(size))
}
