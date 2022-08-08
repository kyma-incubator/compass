package config

import "github.com/kyma-incubator/compass/components/director/pkg/certloader"

type SystemBrokerTestConfig struct {
	Tenant                         string
	SystemBrokerURL                string
	DirectorExternalCertSecuredURL string
	ConnectorURL                   string
	ORDServiceURL                  string
	SkipSSLValidation              bool
	CertLoaderConfig               certloader.Config
}
