package config

type ConnectorTestConfig struct {
	ExternalConnectorURL         string `envconfig:"default=http://compass-connector:3000/graphql"`
	InternalConnectorURL         string `envconfig:"default=http://compass-connector:3001/graphql"`
	HydratorURL                  string `envconfig:"default=http://compass-connector:8080"`
	ConnectorURL                 string
	CertificateDataHeader        string `envconfig:"default=Certificate-Data"`
	RevocationConfigMapName      string `envconfig:"default=revocations-config"`
	RevocationConfigMapNamespace string `envconfig:"default=compass-system"`
}
