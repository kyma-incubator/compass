package oauth20

import "time"

type Config struct {
	URL                       string        `mapstructure:"url"`
	PublicAccessTokenEndpoint string        `mapstructure:"public_access_token_endpoint"`
	HTTPClientTimeout         time.Duration `mapstructure:"http_client_timeout"`
}

func DefaultConfig() *Config {
	return &Config{
		HTTPClientTimeout: 105 * time.Second,
	}
}
