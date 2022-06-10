package config

// PairingAdapterConfig contains configuration for the pairing adapters
type PairingAdapterConfig struct {
	ConfigmapName        string `envconfig:"APP_PAIRING_ADAPTER_CM_NAME"`
	ConfigmapNamespace   string `envconfig:"APP_PAIRING_ADAPTER_CM_NAMESPACE"`
	ConfigmapKey         string `envconfig:"APP_PAIRING_ADAPTER_CM_KEY"`
	WatcherCorrelationID string `envconfig:"APP_PAIRING_ADAPTER_WATCHER_ID"`
}
