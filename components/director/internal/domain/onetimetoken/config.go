package onetimetoken

import "time"

// Config missing godoc
type Config struct {
	Length                int           `envconfig:"default=64"`
	RuntimeExpiration     time.Duration `envconfig:"default=60m"`
	ApplicationExpiration time.Duration `envconfig:"default=5m"`
	CSRExpiration         time.Duration `envconfig:"default=5m"`

	ConnectorURL          string `envconfig:"APP_CONNECTOR_URL"`
	LegacyConnectorURL    string `envconfig:"APP_LEGACY_CONNECTOR_URL"`
	SuggestTokenHeaderKey string `envconfig:"APP_SUGGEST_TOKEN_HTTP_HEADER"`
}
