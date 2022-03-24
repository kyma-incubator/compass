package config

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
)

// SelfRegConfig is configuration for the runtime self-registration flow
type SelfRegConfig struct {
	SelfRegisterDistinguishLabelKey string `envconfig:"APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY"`
	SelfRegisterLabelKey            string `envconfig:"APP_SELF_REGISTER_LABEL_KEY,optional"`
	SelfRegisterLabelValuePrefix    string `envconfig:"APP_SELF_REGISTER_LABEL_VALUE_PREFIX,optional"`
	SelfRegisterResponseKey         string `envconfig:"APP_SELF_REGISTER_RESPONSE_KEY,optional"`
	SelfRegisterPath                string `envconfig:"APP_SELF_REGISTER_PATH,optional"`
	SelfRegisterNameQueryParam      string `envconfig:"APP_SELF_REGISTER_NAME_QUERY_PARAM,optional"`
	SelfRegisterTenantQueryParam    string `envconfig:"APP_SELF_REGISTER_TENANT_QUERY_PARAM,optional"`
	SelfRegisterRequestBodyPattern  string `envconfig:"APP_SELF_REGISTER_REQUEST_BODY_PATTERN,optional"`

	ClientID       string         `envconfig:"APP_SELF_REGISTER_CLIENT_ID,optional"`
	ClientSecret   string         `envconfig:"APP_SELF_REGISTER_CLIENT_SECRET,optional"`
	OAuthMode      oauth.AuthMode `envconfig:"APP_SELF_REGISTER_OAUTH_MODE,default=oauth-mtls"`
	URL            string         `envconfig:"APP_SELF_REGISTER_URL,optional"`
	TokenURL       string         `envconfig:"APP_SELF_REGISTER_TOKEN_URL,optional"`
	OauthTokenPath string         `envconfig:"APP_SELF_REGISTER_OAUTH_TOKEN_PATH,optional"`

	SkipSSLValidation bool `envconfig:"APP_SELF_REGISTER_SKIP_SSL_VALIDATION,default=false"`

	ClientTimeout time.Duration `envconfig:"default=30s"`

	Cert string `envconfig:"APP_SELF_REGISTER_OAUTH_X509_CERT,optional"`
	Key  string `envconfig:"APP_SELF_REGISTER_OAUTH_X509_KEY,optional"`
}
