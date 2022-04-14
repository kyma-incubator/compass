package metrics

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// Namespace missing godoc
	Namespace = "compass"
	// HydratorSubsystem missing godoc
	HydratorSubsystem = "hydrator"
)

// Collector missing godoc
type Collector struct {
	config Config

	clientTotal            *prometheus.CounterVec
	graphQLRequestTotal    *prometheus.CounterVec
	graphQLRequestDuration *prometheus.HistogramVec
}

// NewCollector missing godoc
func NewCollector(config Config) *Collector {
	return &Collector{
		config: config,

		clientTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: HydratorSubsystem,
			Name:      "total_requests_per_client",
			Help:      "Total requests per client",
		}, []string{"client_id", "auth_flow", "details"}),
		graphQLRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: HydratorSubsystem,
			Name:      "graphql_request_total",
			Help:      "Total handled GraphQL Requests",
		}, []string{"code", "method"}),
		graphQLRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: HydratorSubsystem,
			Name:      "graphql_request_duration_seconds",
			Help:      "Duration of handling GraphQL requests",
		}, []string{"code", "method"}),
	}
}

// Describe missing godoc
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.clientTotal.Describe(ch)
	c.graphQLRequestTotal.Describe(ch)
	c.graphQLRequestDuration.Describe(ch)
}

// Collect missing godoc
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.clientTotal.Collect(ch)
	c.graphQLRequestTotal.Collect(ch)
	c.graphQLRequestDuration.Collect(ch)
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

// GraphQLHandlerWithInstrumentation missing godoc
func (c *Collector) GraphQLHandlerWithInstrumentation(handler http.Handler) http.HandlerFunc {
	return promhttp.InstrumentHandlerCounter(c.graphQLRequestTotal,
		promhttp.InstrumentHandlerDuration(c.graphQLRequestDuration, handler),
	)
}
