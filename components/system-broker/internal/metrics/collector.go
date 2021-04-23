package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	Namespace             = "compass"
	SystemBrokerSubsystem = "system_broker"
)

type Collector struct {
	catalogRequestDuration *prometheus.HistogramVec
	catalogResponseSize    *prometheus.HistogramVec
	catalogRequestTotal    *prometheus.CounterVec

	provisionRequestDuration *prometheus.HistogramVec
	provisionResponseSize    *prometheus.HistogramVec
	provisionRequestTotal    *prometheus.CounterVec

	deprovisionRequestDuration *prometheus.HistogramVec
	deprovisionResponseSize    *prometheus.HistogramVec
	deprovisionRequestTotal    *prometheus.CounterVec

	bindRequestDuration *prometheus.HistogramVec
	bindResponseSize    *prometheus.HistogramVec
	bindRequestTotal    *prometheus.CounterVec

	unbindRequestDuration *prometheus.HistogramVec
	unbindResponseSize    *prometheus.HistogramVec
	unbindRequestTotal    *prometheus.CounterVec
}

func NewCollector() *Collector {
	return &Collector{
		catalogRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "catalog_request_duration_seconds",
			Help:      "Duration of handling Catalog requests",
			Buckets:   []float64{0.1, 0.3, 0.5, 0.7, 1, 1.5, 2.5, 5, 10, 15},
		}, []string{"code", "method"}),
		catalogResponseSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "catalog_response_size",
			Help:      "Size of Catalog responses",
		}, []string{"code", "method"}),
		catalogRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "catalog_request_total",
			Help:      "Total handled Catalog requests",
		}, []string{"code", "method"}),

		provisionRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "provision_request_duration_seconds",
			Help:      "Duration of handling Provision requests",
		}, []string{"code", "method"}),
		provisionResponseSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "provision_response_size",
			Help:      "Size of Provision responses",
		}, []string{"code", "method"}),
		provisionRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "provision_request_total",
			Help:      "Total handled Provision requests",
		}, []string{"code", "method"}),

		deprovisionRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "deprovision_request_duration_seconds",
			Help:      "Duration of handling Deprovision requests",
		}, []string{"code", "method"}),
		deprovisionResponseSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "deprovision_response_size",
			Help:      "Size of Deprovision responses",
		}, []string{"code", "method"}),
		deprovisionRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "deprovision_request_total",
			Help:      "Total handled Deprovision requests",
		}, []string{"code", "method"}),

		bindRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "bind_request_duration_seconds",
			Help:      "Duration of handling Bind requests",
		}, []string{"code", "method"}),
		bindResponseSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "bind_response_size",
			Help:      "Size of Bind responses",
		}, []string{"code", "method"}),
		bindRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "bind_request_total",
			Help:      "Total handled Bind requests",
		}, []string{"code", "method"}),

		unbindRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "unbind_request_duration_seconds",
			Help:      "Duration of handling Unbind requests",
		}, []string{"code", "method"}),
		unbindResponseSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "unbind_response_size",
			Help:      "Size of Unbind responses",
		}, []string{"code", "method"}),
		unbindRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: SystemBrokerSubsystem,
			Name:      "unbind_request_total",
			Help:      "Total handled Unbind requests",
		}, []string{"code", "method"}),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.catalogRequestDuration.Describe(ch)
	c.catalogResponseSize.Describe(ch)
	c.catalogRequestTotal.Describe(ch)

	c.provisionRequestDuration.Describe(ch)
	c.provisionResponseSize.Describe(ch)
	c.provisionRequestTotal.Describe(ch)

	c.deprovisionRequestDuration.Describe(ch)
	c.deprovisionResponseSize.Describe(ch)
	c.deprovisionRequestTotal.Describe(ch)

	c.bindRequestDuration.Describe(ch)
	c.bindResponseSize.Describe(ch)
	c.bindRequestTotal.Describe(ch)

	c.unbindRequestDuration.Describe(ch)
	c.unbindResponseSize.Describe(ch)
	c.unbindRequestTotal.Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.catalogRequestDuration.Collect(ch)
	c.catalogResponseSize.Collect(ch)
	c.catalogRequestTotal.Collect(ch)

	c.provisionRequestDuration.Collect(ch)
	c.provisionResponseSize.Collect(ch)
	c.provisionRequestTotal.Collect(ch)

	c.deprovisionRequestDuration.Collect(ch)
	c.deprovisionResponseSize.Collect(ch)
	c.deprovisionRequestTotal.Collect(ch)

	c.bindRequestDuration.Collect(ch)
	c.bindResponseSize.Collect(ch)
	c.bindRequestTotal.Collect(ch)

	c.unbindRequestDuration.Collect(ch)
	c.unbindResponseSize.Collect(ch)
	c.unbindRequestTotal.Collect(ch)
}

func (c *Collector) CatalogHandlerWithInstrumentation(handler http.Handler) http.HandlerFunc {
	return promhttp.InstrumentHandlerDuration(c.catalogRequestDuration,
		promhttp.InstrumentHandlerResponseSize(c.catalogResponseSize,
			promhttp.InstrumentHandlerCounter(c.catalogRequestTotal, handler),
		),
	)
}

func (c *Collector) ProvisionHandlerWithInstrumentation(handler http.Handler) http.HandlerFunc {
	return promhttp.InstrumentHandlerDuration(c.provisionRequestDuration,
		promhttp.InstrumentHandlerResponseSize(c.provisionResponseSize,
			promhttp.InstrumentHandlerCounter(c.provisionRequestTotal, handler),
		),
	)
}

func (c *Collector) DeprovosionHandlerWithInstrumentation(handler http.Handler) http.HandlerFunc {
	return promhttp.InstrumentHandlerDuration(c.deprovisionRequestDuration,
		promhttp.InstrumentHandlerResponseSize(c.deprovisionResponseSize,
			promhttp.InstrumentHandlerCounter(c.deprovisionRequestTotal, handler),
		),
	)
}

func (c *Collector) BindHandlerWithInstrumentation(handler http.Handler) http.HandlerFunc {
	return promhttp.InstrumentHandlerDuration(c.bindRequestDuration,
		promhttp.InstrumentHandlerResponseSize(c.bindResponseSize,
			promhttp.InstrumentHandlerCounter(c.bindRequestTotal, handler),
		),
	)
}

func (c *Collector) UnbindHandlerWithInstrumentation(handler http.Handler) http.HandlerFunc {
	return promhttp.InstrumentHandlerDuration(c.unbindRequestDuration,
		promhttp.InstrumentHandlerResponseSize(c.unbindResponseSize,
			promhttp.InstrumentHandlerCounter(c.unbindRequestTotal, handler),
		),
	)
}
