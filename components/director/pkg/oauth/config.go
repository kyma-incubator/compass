package oauth

import "time"

// Config is Oauth2 configuration
type Config struct {
	ClientID               string        `envconfig:"APP_OAUTH_CLIENT_ID"`
	TokenBaseURL           string        `envconfig:"APP_OAUTH_TOKEN_BASE_URL"`
	TokenPath              string        `envconfig:"APP_OAUTH_TOKEN_PATH"`
	TokenEndpointProtocol  string        `envconfig:"APP_OAUTH_TOKEN_ENDPOINT_PROTOCOL"`
	ClientSecret           string        `envconfig:"APP_OAUTH_CLIENT_SECRET"`
	TenantHeaderName       string        `envconfig:"APP_OAUTH_TENANT_HEADER_NAME"`
	ScopesClaim            []string      `envconfig:"APP_OAUTH_SCOPES_CLAIM"`
	TokenExpirationTimeout time.Duration `envconfig:"APP_OAUTH_TOKEN_EXPIRATION_TIMEOUT"`
	TokenRequestTimeout    time.Duration `envconfig:"APP_OAUTH_TOKEN_REQUEST_TIMEOUT"`
}
