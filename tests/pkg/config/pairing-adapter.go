package config

import "github.com/kyma-incubator/compass/components/director/pkg/certloader"

type PairingAdapterConfig struct {
	DirectorExternalCertSecuredURL string
	CertLoaderConfig               certloader.Config
	SkipSSLValidation              bool `envconfig:"default=true"`
	IsLocalEnv                     bool
	TemplateName                   string
	LocalAdapterFQDN               string `envconfig:"optional"`
	ConfigMapKey                   string `envconfig:"optional"`
	ConfigMapName                  string `envconfig:"optional"`
	ConfigMapNamespace             string `envconfig:"optional"`
	FQDNPairingAdapterURL          string
	TestTenant                     string
	TestClientUser                 string
	TestApplicationID              string
	TestApplicationName            string
}
