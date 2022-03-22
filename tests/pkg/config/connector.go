package config

import "github.com/kyma-incubator/compass/components/director/pkg/certloader"

type ConnectorTestConfig struct {
	Tenant                         string `envconfig:"default=3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"`
	AppsForRuntimeTenant           string `envconfig:"default=2263cc13-5698-4a2d-9257-e8e76b543e33"`
	ConnectorURL                   string `envconfig:"default=http://compass-connector:3000/graphql"`
	DirectorExternalCertSecuredURL string
	SkipSSLValidation              bool `envconfig:"default=false"`
	DirectorMtlsURL                string
	HydratorURL                    string `envconfig:"default=http://compass-hydrator:3000"`

	CertificateDataHeader        string `envconfig:"default=Certificate-Data"`
	RevocationConfigMapName      string `envconfig:"default=revocations-config"`
	RevocationConfigMapNamespace string `envconfig:"default=compass-system"`

	CertLoaderConfig certloader.Config
}
