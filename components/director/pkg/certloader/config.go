package certloader

// Config holds external client certificate configuration available for the certificate loader
type Config struct {
	ExternalClientCertSecret    string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET"`
	ExternalClientCertCertKey   string `envconfig:"APP_EXTERNAL_CLIENT_CERT_KEY"`
	ExternalClientCertKeyKey    string `envconfig:"APP_EXTERNAL_CLIENT_KEY_KEY"`
}
