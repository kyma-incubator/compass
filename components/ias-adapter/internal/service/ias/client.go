package ias

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
)

func NewClient(cfg config.IAS) (*http.Client, error) {
	clientCert, err := tls.LoadX509KeyPair(cfg.CockpitClientCertPath, cfg.CockpitClientKeyPath)
	if err != nil {
		return nil, errors.Newf("failed to load IAS client cert: %w", err)
	}

	caCert, err := os.ReadFile(cfg.CockpitCAPath)
	if err != nil {
		return nil, errors.Newf("failed to load IAS CA: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

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
