package credloader

// CertConfig holds external client certificate configuration available for the certificate loader
type CertConfig struct {
	ExternalClientCertSecret  string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET"`
	ExternalClientCertCertKey string `envconfig:"APP_EXTERNAL_CLIENT_CERT_KEY"`
	ExternalClientCertKeyKey  string `envconfig:"APP_EXTERNAL_CLIENT_KEY_KEY"`
}

// KeysConfig holds keys configuration available for the key loader
type KeysConfig struct {
	KeysSecretName string `envconfig:"APP_SYSTEM_FETCHER_EXTERNAL_KEYS_SECRET_NAME"`
	KeysSecret     string `envconfig:"APP_SYSTEM_FETCHER_EXTERNAL_KEYS_SECRET"`
	KeysData       string `envconfig:"APP_SYSTEM_FETCHER_EXTERNAL_KEYS_SECRET_DATA_KEY"`
}
