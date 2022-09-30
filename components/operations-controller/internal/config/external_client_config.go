package config

type ExternalClientConfig struct {
	CertSecret     string `mapstructure:"cert_secret"`
	CertKey        string `mapstructure:"cert_key"`
	KeyKey         string `mapstructure:"key_key"`
	CertSecretName string `mapstructure:"cert_secret_name"`
}
