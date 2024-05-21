package apiclient

import (
	"time"
)

// SystemFieldDiscoveryEngineClientConfig is the configuration needed for system field discovery engine client
type SystemFieldDiscoveryEngineClientConfig struct {
	ClientTimeout                             time.Duration `envconfig:"APP_SYSTEM_FIELD_DISCOVERY_ENGINE_CLIENT_TIMEOUT,default=30s"`
	SystemFieldDiscoveryEngineSaaSRegistryAPI string        `envconfig:"APP_SYSTEM_FIELD_DISCOVERY_ENGINE_SAAS_REGISTRY_API"`
	SkipSSLValidation                         bool          `envconfig:"default=false,APP_SYSTEM_FIELD_DISCOVERY_ENGINE_SKIP_SSL_VALIDATION"`
}
