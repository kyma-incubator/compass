package onetimetoken

import "time"

type Config struct {
	Length                int           `envconfig:"default=64"`
	RuntimeExpiration     time.Duration `envconfig:"default=60m"`
	ApplicationExpiration time.Duration `envconfig:"default=5m"`
	CSRExpiration         time.Duration `envconfig:"default=5m"`

	//Connector URL
	ConnectorURL string `envconfig:"APP_CONNECTOR_URL"`
	//Legacy Connector URL
	LegacyConnectorURL string `envconfig:"APP_LEGACY_CONNECTOR_URL"`
}
