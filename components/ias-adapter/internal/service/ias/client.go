package ias

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	logCtx "github.com/kyma-incubator/compass/components/ias-adapter/internal/logger/context"
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

	transport := &headerTransport{clientTransport: &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      caCertPool,
		},
	}}
	return &http.Client{
		Transport: transport,
		Timeout:   cfg.RequestTimeout,
	}, nil
}

type headerTransport struct {
	clientTransport http.RoundTripper
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	requestID := req.Context().Value(logCtx.RequestIDCtxKey).(string)
	if requestID != "" {
		req.Header.Add(logCtx.RequestIDHeader, requestID)
	}
	return t.clientTransport.RoundTrip(req)
}
