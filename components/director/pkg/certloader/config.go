package certloader

// Config holds the configuration available for the certificate loader
type Config struct {
	ExternalClientCert
	IsLocalSetup bool `envconfig:"optional,APP_IS_CERT_LOADER_IN_LOCAL_SETUP"`
}

// ExternalClientCert contains client certificate configuration
type ExternalClientCert struct {
	Secret  string `envconfig:"optional,APP_EXTERNAL_CLIENT_CERT_SECRET"`
	CertKey string `envconfig:"optional,APP_EXTERNAL_CLIENT_CERT_KEY"`
	KeyKey  string `envconfig:"optional,APP_EXTERNAL_CLIENT_KEY_KEY"`
	Cert    string `envconfig:"optional,APP_EXTERNAL_CLIENT_CERT_VALUE"`
	Key     string `envconfig:"optional,APP_EXTERNAL_CLIENT_KEY_VALUE"`
}
