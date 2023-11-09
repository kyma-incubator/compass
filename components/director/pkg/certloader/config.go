package certloader

// Config holds external client certificate configuration available for the certificate loader
type Config struct {
	ExternalClientCertSecret  string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET"`
	ExternalClientCertCertKey string `envconfig:"APP_EXTERNAL_CLIENT_CERT_KEY"`
	ExternalClientCertKeyKey  string `envconfig:"APP_EXTERNAL_CLIENT_KEY_KEY"`

	ExtSvcClientCertSecret  string `envconfig:"APP_EXT_SVC_CLIENT_CERT_SECRET"`
	ExtSvcClientCertCertKey string `envconfig:"APP_EXT_SVC_CLIENT_CERT_KEY"`
	ExtSvcClientCertKeyKey  string `envconfig:"APP_EXT_SVC_CLIENT_KEY_KEY"`
}

// KeysConfig holds keys configuration available for the key loader
type KeysConfig struct {
	KeysSecret string `envconfig:"APP_SYSTEM_FETCHER_EXTERNAL_KEYS_SECRET_NAME"`
	KeysData   string `envconfig:"APP_SYSTEM_FETCHER_EXTERNAL_KEYS_SECRET"`
}
