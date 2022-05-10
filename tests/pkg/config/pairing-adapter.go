package config

import "github.com/kyma-incubator/compass/components/director/pkg/certloader"

type PairingAdapterConfig struct {
	FQDNPairingAdapterURL          string
	TestTenant                     string
	TestClientUser                 string
	TestApplicationID              string
	TestApplicationName            string
	ClientIDHeader                 string `envconfig:"APP_CLIENT_ID_HTTP_HEADER"`
	DirectorExternalCertSecuredURL string
	CertLoaderConfig               certloader.Config
	SkipSSLValidation              bool `envconfig:"default=true"`
	IsLocalEnv                     bool
	TemplateName                   string
	ConfigMapName                  string `envconfig:"optional"`
	ConfigMapNamespace             string `envconfig:"optional"`
	ConfigMapKey                   string `envconfig:"optional"`
	IntegrationSystemID            string `envconfig:"optional"`
	LocalAdapterFQDN               string `envconfig:"optional"`
}
