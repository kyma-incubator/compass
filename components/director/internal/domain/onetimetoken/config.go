package onetimetoken

import "time"

type Config struct {
	Length                int           `mapstructure:"length"`
	RuntimeExpiration     time.Duration `mapstructure:"runtime_expiration"`
	ApplicationExpiration time.Duration `mapstructure:"application_expiration"`
	CSRExpiration         time.Duration `mapstructure:"csr_expiration"`

	//Connector URL
	ConnectorURL string `mapstructure:"CONNECTOR_URL"`
	//Legacy Connector URL
	LegacyConnectorURL string `mapstructure:"LEGACY_CONNECTOR_URL"`
}

func DefaultConfig() *Config {
	return &Config{
		Length:                64,
		RuntimeExpiration:     60 * time.Minute,
		ApplicationExpiration: 5 * time.Minute,
		CSRExpiration:         5 * time.Minute,
		ConnectorURL:          "",
		LegacyConnectorURL:    "",
	}
}
