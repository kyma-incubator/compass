package config

import "github.com/kyma-incubator/compass/components/director/pkg/credloader"

type PairingAdapterConfig struct {
	FQDNPairingAdapterURL          string
	TestTenant                     string
	TestClientUser                 string
	TestApplicationID              string
	TestApplicationName            string
	ClientIDHeader                 string `envconfig:"APP_CLIENT_ID_HTTP_HEADER"`
	DirectorExternalCertSecuredURL string
	CertLoaderConfig               credloader.CertConfig
	SkipSSLValidation              bool `envconfig:"default=true"`
	IsLocalEnv                     bool
	TemplateName                   string
	ConfigMapName                  string `envconfig:"optional"`
	ConfigMapNamespace             string `envconfig:"optional"`
	ConfigMapKey                   string `envconfig:"optional"`
	IntegrationSystemID            string `envconfig:"optional"`
	LocalAdapterFQDN               string `envconfig:"optional"`
	SelfRegDistinguishLabelKey     string
	SelfRegDistinguishLabelValue   string
	SelfRegRegion                  string
	SelfRegLabelKey                string
	ExternalClientCertSecretName   string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
	GatewayOauth                   string `envconfig:"APP_GATEWAY_OAUTH"`
}
