package config

import "github.com/kyma-incubator/compass/components/director/pkg/credloader"

type SystemBrokerTestConfig struct {
	Tenant                         string
	SystemBrokerURL                string
	DirectorExternalCertSecuredURL string
	ConnectorURL                   string
	ORDServiceURL                  string
	SkipSSLValidation              bool
	CertLoaderConfig               credloader.CertConfig
	ExternalClientCertSecretName   string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
}
