package oauth20

import "time"

type Config struct {
	URL                       string        `mapstructure:"OAUTH20_URL"`
	PublicAccessTokenEndpoint string        `mapstructure:"OAUTH20_PUBLIC_ACCESS_TOKEN_ENDPOINT"`
	HTTPClientTimeout         time.Duration `mapstructure:"OAUTH20_HTTP_CLIENT_TIMEOUT"`
}

func DefaultConfig() *Config {
	return &Config{
		HTTPClientTimeout: 105 * time.Second,
	}
}
