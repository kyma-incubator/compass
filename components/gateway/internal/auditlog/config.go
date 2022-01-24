package auditlog

import (
	"crypto/tls"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
)

type AuthMode string

const (
	Basic     AuthMode = "basic"
	OAuth     AuthMode = "oauth"
	OAuthMtls AuthMode = "oauth-mtls"
)

type Config struct {
	URL               string        `envconfig:"APP_AUDITLOG_URL"`
	ConfigPath        string        `envconfig:"APP_AUDITLOG_CONFIG_PATH"`
	SecurityPath      string        `envconfig:"APP_AUDITLOG_SECURITY_PATH"`
	AuthMode          AuthMode      `envconfig:"APP_AUDITLOG_AUTH_MODE"`
	ClientTimeout     time.Duration `envconfig:"APP_AUDITLOG_CLIENT_TIMEOUT,default=30s"`
	MsgChannelSize    int           `envconfig:"APP_AUDITLOG_CHANNEL_SIZE,default=100"`
	MsgChannelTimeout time.Duration `envconfig:"APP_AUDITLOG_CHANNEL_TIMEOUT,default=5s"`
	WriteWorkers      int           `envconfig:"APP_AUDITLOG_WRITE_WORKERS,default=5"`
}

type BasicAuthConfig struct {
	User     string `envconfig:"APP_AUDITLOG_USER"`
	Password string `envconfig:"APP_AUDITLOG_PASSWORD"`
	Tenant   string `envconfig:"APP_AUDITLOG_TENANT"`
}

type OAuthConfig struct {
	ClientID     string `envconfig:"APP_AUDITLOG_CLIENT_ID"`
	ClientSecret string `envconfig:"APP_AUDITLOG_CLIENT_SECRET"`
	OAuthURL     string `envconfig:"APP_AUDITLOG_OAUTH_URL"`
	User         string `envconfig:"APP_AUDITLOG_OAUTH_USER,default=$USER"`
	Tenant       string `envconfig:"APP_AUDITLOG_OAUTH_TENANT,default=$PROVIDER"`
	TokenPath    string `envconfig:"APP_AUDITLOG_TOKEN_PATH"`
}

type OAuthMtlsConfig struct {
	ClientID  string `envconfig:"APP_AUDITLOG_CLIENT_ID"`
	OAuthURL  string `envconfig:"APP_AUDITLOG_OAUTH_URL"`
	User      string `envconfig:"APP_AUDITLOG_OAUTH_USER,default=$USER"`
	Tenant    string `envconfig:"APP_AUDITLOG_OAUTH_TENANT,default=$PROVIDER"`
	TokenPath string `envconfig:"APP_AUDITLOG_TOKEN_PATH"`
	X509Cert  string `envconfig:"APP_AUDITLOG_X509_CERT"`
	X509Key   string `envconfig:"APP_AUDITLOG_X509_KEY"`
}

func (c *OAuthMtlsConfig) ParseCertificate() (*tls.Certificate, error) {
	return cert.ParseCertificate(c.X509Cert, c.X509Key)
}
