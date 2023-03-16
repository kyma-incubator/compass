package ias

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
)

type iasCockpit struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
	CA   string `yaml:"ca"`
}

func NewClient(cfg config.IAS) (*http.Client, error) {
	bytes, err := os.ReadFile(cfg.CockpitSecretPath)
	if err != nil {
		return nil, errors.Newf("failed to read IAS cockpit secret: %w", err)
	}

	var iasCockpit iasCockpit
	err = yaml.Unmarshal(bytes, &iasCockpit)
	if err != nil {
		return nil, errors.Newf("failed to unmarshal IAS cockpit secret: %w", err)
	}

	clientCert, err := tls.X509KeyPair([]byte(iasCockpit.Cert), []byte(iasCockpit.Key))
	if err != nil {
		return nil, errors.Newf("failed to load IAS client cert: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(iasCockpit.CA))

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{clientCert},
				RootCAs:      caCertPool,
			},
		},
		Timeout: cfg.RequestTimeout,
	}, nil
}
