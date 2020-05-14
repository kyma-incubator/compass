package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/util/workqueue"
)

const (
	metricsNamespace          = "metris"
	workqueueMetricsNamespace = metricsNamespace + "_workqueue"
)

var (
	ClusterSyncFailureVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "cluster_sync_failure_total",
			Help:      "Total number of failed cluster syncs.",
		},
		[]string{"reason"},
	)

	StoredAccounts = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "account_total",
			Help:      "Total number of account added to the storage.",
		},
	)

	ReceivedSamplesDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "samples_received_duration_seconds",
			Help:      "Duration of metrics request to provider in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
	)

	ReceivedSamples = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "samples_received_total",
			Help:      "Total number of received samples.",
		},
	)

	SentSamples = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "samples_sent_total",
			Help:      "Total number of processed samples sent to edp.",
		},
	)

	FailedSamples = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "samples_sent_failed_total",
			Help:      "Total number of processed samples which failed on send to edp.",
		},
	)

	SentSamplesDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "samples_sent_duration_seconds",
			Help:      "Duration of sample send calls to edp.",
			Buckets:   prometheus.DefBuckets,
		},
	)

	// Definition of metrics for provider queue
	workqueueDepthMetricVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: workqueueMetricsNamespace,
			Name:      "depth",
			Help:      "Current depth of the work queue.",
		},
		[]string{"queue_name"},
	)
	workqueueAddsMetricVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: workqueueMetricsNamespace,
			Name:      "items_total",
			Help:      "Total number of items added to the work queue.",
		},
		[]string{"queue_name"},
	)
	workqueueLatencyMetricVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  workqueueMetricsNamespace,
			Name:       "latency_seconds",
			Help:       "How long an item stays in the work queue.",
			Objectives: map[float64]float64{},
		},
		[]string{"queue_name"},
	)
	workqueueUnfinishedWorkSecondsMetricVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: workqueueMetricsNamespace,
			Name:      "unfinished_work_seconds",
			Help:      "How long an item has remained unfinished in the work queue.",
		},
		[]string{"queue_name"},
	)
	workqueueLongestRunningProcessorMetricVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: workqueueMetricsNamespace,
			Name:      "longest_running_processor_seconds",
			Help:      "Duration of the longest running processor in the work queue.",
		},
		[]string{"queue_name"},
	)
	workqueueWorkDurationMetricVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  workqueueMetricsNamespace,
			Name:       "work_duration_seconds",
			Help:       "How long processing an item from the work queue takes.",
			Objectives: map[float64]float64{},
		},
		[]string{"queue_name"},
	)
)

// Definition of dummy metric used as a placeholder if we don't want to observe some data.
type noopMetric struct{}

func (noopMetric) Inc()            {}
func (noopMetric) Dec()            {}
func (noopMetric) Observe(float64) {}
func (noopMetric) Set(float64)     {}

// Definition of workqueue metrics provider definition
type WorkqueueMetricsProvider struct{}

func (f *WorkqueueMetricsProvider) NewDepthMetric(name string) workqueue.GaugeMetric {
	return workqueueDepthMetricVec.WithLabelValues(name)
}

func (f *WorkqueueMetricsProvider) NewAddsMetric(name string) workqueue.CounterMetric {
	return workqueueAddsMetricVec.WithLabelValues(name)
}

func (f *WorkqueueMetricsProvider) NewLatencyMetric(name string) workqueue.HistogramMetric {
	return workqueueLatencyMetricVec.WithLabelValues(name)
}

func (f *WorkqueueMetricsProvider) NewWorkDurationMetric(name string) workqueue.HistogramMetric {
	return workqueueWorkDurationMetricVec.WithLabelValues(name)
}

func (f *WorkqueueMetricsProvider) NewUnfinishedWorkSecondsMetric(name string) workqueue.SettableGaugeMetric {
	return workqueueUnfinishedWorkSecondsMetricVec.WithLabelValues(name)
}

func (f *WorkqueueMetricsProvider) NewLongestRunningProcessorSecondsMetric(name string) workqueue.SettableGaugeMetric {
	return workqueueLongestRunningProcessorMetricVec.WithLabelValues(name)
}

func (WorkqueueMetricsProvider) NewRetriesMetric(name string) workqueue.CounterMetric {
	// Retries are not used so the metric is omitted.
	return noopMetric{}
}

func init() {
	prometheus.MustRegister(
		ClusterSyncFailureVec,
		StoredAccounts,
		ReceivedSamples,
		ReceivedSamplesDuration,
		SentSamples,
		SentSamplesDuration,
		FailedSamples,
		workqueueDepthMetricVec,
		workqueueAddsMetricVec,
		workqueueLatencyMetricVec,
		workqueueWorkDurationMetricVec,
		workqueueUnfinishedWorkSecondsMetricVec,
		workqueueLongestRunningProcessorMetricVec,
	)
}
