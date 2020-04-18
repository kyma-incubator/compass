package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricNamespace = "metris"
)

var (
	ReceivedSamplesDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Name:      "samples_received_duration_seconds",
			Help:      "Duration of metrics request to provider in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
	)

	ReceivedSamples = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "samples_received_total",
			Help:      "Total number of received samples.",
		},
	)

	SentSamples = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "samples_sent_total",
			Help:      "Total number of processed samples sent to edp.",
		},
	)

	FailedSamples = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "samples_sent_failed_total",
			Help:      "Total number of processed samples which failed on send to edp.",
		},
	)

	SentSamplesDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Name:      "samples_sent_duration_seconds",
			Help:      "Duration of sample send calls to edp.",
			Buckets:   prometheus.DefBuckets,
		},
	)
)

func init() {
	prometheus.MustRegister(
		ReceivedSamples,
		ReceivedSamplesDuration,
		SentSamples,
		SentSamplesDuration,
		FailedSamples,
	)
}
