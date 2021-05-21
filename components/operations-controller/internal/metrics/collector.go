package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	operationDuration         *prometheus.HistogramVec
	operationErrorCount       *prometheus.CounterVec
	operationNearTimeoutCount *prometheus.CounterVec
}

const (
	Namespace = "compass"
	Subsystem = "operations_controller"
)

func NewCollector() *Collector {
	return &Collector{
		operationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      "operation_duration_seconds",
			Help:      "Duration of Operation InProgress state",
			Buckets:   []float64{0.5, 1, 1.5, 2.5, 5, 10, 15, 25, 40, 60, 100, 150},
		}, []string{"type"}),
		operationErrorCount: prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      "failed_operations_count",
			Help:      "Count of Operations with error condition",
		}, []string{"name", "correlation_id", "type", "category", "request_object", "error"}),
		operationNearTimeoutCount: prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      "operations_near_reconciliation_timeout_count",
			Help:      "Count of Operations succeeded/failed which were InProgress close to reconciliation timeout.",
		}, []string{"type"}),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.operationDuration.Describe(ch)
	c.operationErrorCount.Describe(ch)
	c.operationNearTimeoutCount.Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.operationDuration.Collect(ch)
	c.operationErrorCount.Collect(ch)
	c.operationNearTimeoutCount.Collect(ch)
}

func (c *Collector) RecordLatency(operationType string, inProgressTime time.Duration) {
	c.operationDuration.WithLabelValues(operationType).Observe(inProgressTime.Seconds())
}

func (c *Collector) RecordError(name, correlationID, operationType, category, requestObject, error string) {
	c.operationErrorCount.WithLabelValues(name, correlationID, operationType, category, requestObject, error).Inc()
}

func (c *Collector) RecordOperationInProgressNearTimeout(operationType string) {
	c.operationNearTimeoutCount.WithLabelValues(operationType).Inc()
}
