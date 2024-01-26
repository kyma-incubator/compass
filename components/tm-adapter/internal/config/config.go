package config

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/credloader"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/server"
	"github.com/vrischmann/envconfig"
)

const envPrefix = "APP"

type Config struct {
	Server                       *server.Config
	Log                          *log.Config
	OAuthProvider                OAuthConfig
	HTTPClient                   HTTPClient
	CertLoaderConfig             credloader.CertConfig
	ExternalClientCertSecretName string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
}

type OAuthConfig struct {
	ClientID       string `envconfig:"APP_SM_SVC_CLIENT_ID"`
	ClientSecret   string `envconfig:"APP_SM_SVC_CLIENT_SECRET"`
	OAuthURL       string `envconfig:"APP_SM_SVC_OAUTH_URL"`
	OAuthTokenPath string `envconfig:"APP_SM_SVC_OAUTH_TOKEN_PATH"`
}

type HTTPClient struct {
	Timeout           time.Duration `envconfig:"APP_TM_ADAPTER_CLIENT_TIMEOUT"`
	SkipSSLValidation bool          `envconfig:"APP_TM_ADAPTER_CLIENT_SKIP_SSL_VALIDATION"`
}

func New() (Config, error) {
	cfg := Config{}
	return cfg, envconfig.InitWithPrefix(&cfg, envPrefix)
}
