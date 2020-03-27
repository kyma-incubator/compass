package auditlog

import "time"

type Config struct {
	URL               string        `envconfig:"APP_AUDITLOG_URL"`
	ConfigPath        string        `envconfig:"APP_AUDITLOG_CONFIG_PATH"`
	SecurityPath      string        `envconfig:"APP_AUDITLOG_SECURITY_PATH"`
	AuthMode          AuthMode      `envconfig:"APP_AUDITLOG_AUTH_MODE"`
	MsgChannelSize    int           `envconfig:"APP_AUDITLOG_CHANNEL_SIZE,default=100"`
	MsgChannelTimeout time.Duration `envconfig:"APP_AUDITLOG_CHANNEL_TIMEOUT,default=5s"`
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
}

type AuthMode string

const (
	Basic AuthMode = "basic"
	OAuth AuthMode = "oauth"
)
