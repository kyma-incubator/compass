package metrics

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config configures the behaviour of the metrics collector.
type Config struct {
	EnableClientIDInstrumentation         bool     `envconfig:"default=true,APP_METRICS_ENABLE_CLIENT_ID_INSTRUMENTATION"`
	EnableGraphqlOperationInstrumentation bool     `envconfig:"default=false,APP_METRICS_ENABLE_GRAPHQL_OPERATION_INSTRUMENTATION"`
	CensoredFlows                         []string `envconfig:"optional,APP_METRICS_CENSORED_FLOWS"`
}

// Collector missing godoc
type Collector struct {
	config Config

	graphQLRequestTotal    *prometheus.CounterVec
	graphQLRequestDuration *prometheus.HistogramVec
	hydraRequestTotal      *prometheus.CounterVec
	hydraRequestDuration   *prometheus.HistogramVec
	clientTotal            *prometheus.CounterVec
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
		hydraRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: DirectorSubsystem,
			Name:      "hydra_request_duration_seconds",
			Help:      "Duration of HTTP Requests to Hydra",
		}, []string{"code", "method"}),
		hydraRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: DirectorSubsystem,
			Name:      "hydra_request_total",
			Help:      "Total HTTP Requests to Hydra",
		}, []string{"code", "method"}),
		clientTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: DirectorSubsystem,
			Name:      "total_requests_per_client",
			Help:      "Total requests per client",
		}, []string{"client_id", "auth_flow", "details"}),
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
	c.clientTotal.Describe(ch)
	c.graphQLOperationCount.Describe(ch)
}

// Collect missing godoc
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.graphQLRequestTotal.Collect(ch)
	c.graphQLRequestDuration.Collect(ch)
	c.hydraRequestTotal.Collect(ch)
	c.hydraRequestDuration.Collect(ch)
	c.clientTotal.Collect(ch)
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

// InstrumentClient instruments a given client caller.
func (c *Collector) InstrumentClient(clientID, authFlow, details string) {
	if !c.config.EnableClientIDInstrumentation {
		return
	}

	if len(c.config.CensoredFlows) > 0 {
		for _, censoredFlow := range c.config.CensoredFlows {
			if authFlow == censoredFlow {
				clientIDHash := sha256.Sum256([]byte(clientID))
				clientID = fmt.Sprintf("%x", clientIDHash)
				break
			}
		}
	}

	c.clientTotal.With(prometheus.Labels{
		"client_id": clientID,
		"auth_flow": authFlow,
		"details":   details,
	}).Inc()
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
