package apiclient

import (
	"time"
)

type OrdAggregatorClientConfig struct {
	ClientTimeout             time.Duration `envconfig:"default=30s"`
	OrdAggregatorAggregateAPI string        `envconfig:"APP_ORD_AGGREGATOR_AGGREGATE_API"`
	SkipSSLValidation         bool          `envconfig:"default=false,APP_ORD_AGGREGATOR_SKIP_SSL_VALIDATION"`
}
