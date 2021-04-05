package oauth20

import "time"

type Config struct {
	URL                       string        `envconfig:"APP_OAUTH20_URL"`
	PublicAccessTokenEndpoint string        `envconfig:"APP_OAUTH20_PUBLIC_ACCESS_TOKEN_ENDPOINT"`
	HTTPClientTimeout         time.Duration `envconfig:"default=105s,APP_OAUTH20_HTTP_CLIENT_TIMEOUT"`
}
