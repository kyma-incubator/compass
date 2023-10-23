package apiclient

import (
	"time"
)

// OrdAggregatorClientConfig is the configuration needed for ord aggregator client
type OrdAggregatorClientConfig struct {
	ClientTimeout             time.Duration `envconfig:"APP_ORD_AGGREGATOR_CLIENT_TIMEOUT,default=30s"`
	OrdAggregatorAggregateAPI string        `envconfig:"APP_ORD_AGGREGATOR_AGGREGATE_API"`
	SkipSSLValidation         bool          `envconfig:"default=false,APP_ORD_AGGREGATOR_SKIP_SSL_VALIDATION"`
}
