package config

type ExternalSystemConfig struct {
	ConnectorURL string `envconfig:"default=localhost,APP_CONNECTOR_URL"`
}
