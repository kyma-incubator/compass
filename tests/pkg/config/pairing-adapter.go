package config

import "github.com/kyma-incubator/compass/components/director/pkg/certloader"

type PairingAdapterConfig struct {
	DirectorExternalCertSecuredURL string
	CertLoaderConfig               certloader.Config
	SkipSSLValidation              bool `envconfig:"default=true"`
	MTLSPairingAdapterURL          string
	TestTenant                     string
	TestClientUser                 string
	TestApplicationID              string
	TestApplicationName            string
}
