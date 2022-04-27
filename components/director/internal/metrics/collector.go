package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector missing godoc
type Collector struct {
	graphQLRequestTotal    *prometheus.CounterVec
	graphQLRequestDuration *prometheus.HistogramVec
	hydraRequestTotal      *prometheus.CounterVec
	hydraRequestDuration   *prometheus.HistogramVec
}

// NewCollector missing godoc
func NewCollector() *Collector {
	return &Collector{
		graphQLRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: DirectorSubsystem,
			Name:      "graphql_request_total",
			Help:      "Total handled GraphQL Requests",
		}, []string{"code", "method"}),
		graphQLRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: DirectorSubsystem,
			Name:      "graphql_request_duration_seconds",
			Help:      "Duration of handling GraphQL requests",
		}, []string{"code", "method"}),
		hydraRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: DirectorSubsystem,
			Name:      "hydra_request_total",
			Help:      "Total HTTP Requests to Hydra",
		}, []string{"code", "method"}),
		hydraRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: DirectorSubsystem,
			Name:      "hydra_request_duration_seconds",
			Help:      "Duration of HTTP Requests to Hydra",
		}, []string{"code", "method"}),
	}
}

// Describe missing godoc
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.graphQLRequestTotal.Describe(ch)
	c.graphQLRequestDuration.Describe(ch)
	c.hydraRequestTotal.Describe(ch)
	c.hydraRequestDuration.Describe(ch)
}

// Collect missing godoc
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.graphQLRequestTotal.Collect(ch)
	c.graphQLRequestDuration.Collect(ch)
	c.hydraRequestTotal.Collect(ch)
	c.hydraRequestDuration.Collect(ch)
}

// GraphQLHandlerWithInstrumentation missing godoc
func (c *Collector) GraphQLHandlerWithInstrumentation(handler http.Handler) http.HandlerFunc {
	return promhttp.InstrumentHandlerCounter(c.graphQLRequestTotal,
		promhttp.InstrumentHandlerDuration(c.graphQLRequestDuration, handler),
	)
}

// InstrumentOAuth20HTTPClient missing godoc
func (c *Collector) InstrumentOAuth20HTTPClient(client *http.Client) {
	client.Transport = promhttp.InstrumentRoundTripperCounter(c.hydraRequestTotal,
		promhttp.InstrumentRoundTripperDuration(c.hydraRequestDuration, client.Transport),
	)
}
