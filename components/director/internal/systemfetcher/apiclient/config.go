package apiclient

import (
	"time"
)

// SystemFetcherSyncClientConfig is the configuration needed for system fetcher client
type SystemFetcherSyncClientConfig struct {
	ClientTimeout        time.Duration `envconfig:"APP_SYSTEM_FETCHER_CLIENT_TIMEOUT,default=30s"`
	SystemFetcherSyncAPI string        `envconfig:"APP_SYSTEM_FETCHER_SYNC_API"`
	SkipSSLValidation    bool          `envconfig:"default=false,APP_SYSTEM_FETCHER_SKIP_SSL_VALIDATION"`
}
