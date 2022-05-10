package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config configures the behaviour of the metrics collector.
type Config struct {
	EnableGraphqlOperationInstrumentation bool `envconfig:"default=false,APP_METRICS_ENABLE_GRAPHQL_OPERATION_INSTRUMENTATION"`
}

// Collector missing godoc
type Collector struct {
	config Config

	graphQLRequestTotal    *prometheus.CounterVec
	graphQLRequestDuration *prometheus.HistogramVec
	hydraRequestTotal      *prometheus.CounterVec
	hydraRequestDuration   *prometheus.HistogramVec
	graphQLOperationCount  *prometheus.CounterVec
}

// NewCollector missing godoc
func NewCollector(config Config) *Collector {
	return &Collector{
		config: config,

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
		graphQLOperationCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: DirectorSubsystem,
			Name:      "graphql_operations_per_endpoint",
			Help:      "Graphql Operations Per Operation",
		}, []string{"operation_name", "operation_type"}),
	}
}

// Describe missing godoc
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.graphQLRequestTotal.Describe(ch)
	c.graphQLRequestDuration.Describe(ch)
	c.hydraRequestTotal.Describe(ch)
	c.hydraRequestDuration.Describe(ch)
	c.graphQLOperationCount.Describe(ch)
}

// Collect missing godoc
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.graphQLRequestTotal.Collect(ch)
	c.graphQLRequestDuration.Collect(ch)
	c.hydraRequestTotal.Collect(ch)
	c.hydraRequestDuration.Collect(ch)
	c.graphQLOperationCount.Collect(ch)
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

// InstrumentGraphqlRequest instruments a graphql request given operationName and operationType
func (c *Collector) InstrumentGraphqlRequest(operationType, operationName string) {
	if !c.config.EnableGraphqlOperationInstrumentation {
		return
	}

	c.graphQLOperationCount.With(prometheus.Labels{
		"operation_name": operationName,
		"operation_type": operationType,
	}).Inc()
}
