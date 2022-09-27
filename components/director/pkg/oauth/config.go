package oauth

import (
	"crypto/tls"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
)

// AuthMode provides a way to select the auth mechanism for fetching an OAuth token
type AuthMode string

const (
	// Standard is used for the standard client-credentials flow with clientId and secret
	Standard AuthMode = "standard"
	// Mtls is used for getting a token using clientId and client certificate
	Mtls AuthMode = "oauth-mtls"
)

// Config is Oauth2 configuration
type Config struct {
	ClientID                     string        `envconfig:"APP_OAUTH_CLIENT_ID"`
	TokenBaseURL                 string        `envconfig:"APP_OAUTH_TOKEN_BASE_URL"`
	TokenPath                    string        `envconfig:"APP_OAUTH_TOKEN_PATH"`
	TokenEndpointProtocol        string        `envconfig:"APP_OAUTH_TOKEN_ENDPOINT_PROTOCOL"`
	TenantHeaderName             string        `envconfig:"APP_OAUTH_TENANT_HEADER_NAME"`
	ScopesClaim                  []string      `envconfig:"APP_OAUTH_SCOPES_CLAIM"`
	TokenRequestTimeout          time.Duration `envconfig:"APP_OAUTH_TOKEN_REQUEST_TIMEOUT"`
	SkipSSLValidation            bool          `envconfig:"APP_OAUTH_SKIP_SSL_VALIDATION"`
}

// X509Config is X509 configuration for getting an OAuth token via mtls
type X509Config struct {
	Cert string `envconfig:"APP_OAUTH_X509_CERT,optional"`
	Key  string `envconfig:"APP_OAUTH_X509_KEY,optional"`
}

// ParseCertificate parses the TLS certificate contained in the X509Config
func (c *X509Config) ParseCertificate() (*tls.Certificate, error) {
	return cert.ParseCertificate(c.Cert, c.Key)
}
