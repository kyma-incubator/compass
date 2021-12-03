package certs

type CAConfig struct {
	Certificate          []byte `envconfig:"-"`
	Key                  []byte `envconfig:"-"`
	SecretName           string
	SecretNamespace      string
	SecretCertificateKey string
	SecretKeyKey         string
}
