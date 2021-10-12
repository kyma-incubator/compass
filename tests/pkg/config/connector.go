package config

type ConnectorTestConfig struct {
	Tenant               string `envconfig:"default=3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"`
	AppsForRuntimeTenant string `envconfig:"default=5984a414-1eed-4972-af2c-b2b6a415c7d7"`
	ConnectorURL         string `envconfig:"default=http://compass-connector:3000/graphql"`
	DirectorURL          string `envconfig:"default=http://compass-director:3000/graphql"`
	DirectorMtlsURL      string
	ConnectorHydratorURL string `envconfig:"default=http://compass-connector:8080"`
	DirectorHydratorURL  string `envconfig:"default=http://compass-director:8080"`

	CertificateDataHeader        string `envconfig:"default=Certificate-Data"`
	RevocationConfigMapName      string `envconfig:"default=revocations-config"`
	RevocationConfigMapNamespace string `envconfig:"default=compass-system"`
}
