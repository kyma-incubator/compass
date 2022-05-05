package config

import "github.com/kyma-incubator/compass/components/director/pkg/certloader"

type PairingAdapterConfig struct {
	DirectorExternalCertSecuredURL string
	CertLoaderConfig               certloader.Config
	SkipSSLValidation              bool `envconfig:"default=true"`
	IsLocalEnv                     bool
	IntegrationSystemID            string
	TemplateName                   string
	LocalAdapterFQDN               string
	ConfigMapKey                   string
	ConfigMapName                  string
	ConfigMapNamespace             string
	FQDNPairingAdapterURL          string
	TestTenant                     string
	TestClientUser                 string
	TestApplicationID              string
	TestApplicationName            string
}
