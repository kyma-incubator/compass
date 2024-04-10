package outbound

import (
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	logCtx "github.com/kyma-incubator/compass/components/ias-adapter/internal/logger/context"
)

type cert struct {
	Crt string `yaml:"cert"`
	Key string `yaml:"key"`
}

func LoadClientCert(certPath string) (tls.Certificate, error) {
	bytes, err := os.ReadFile(certPath)
	if err != nil {
		return tls.Certificate{}, errors.Newf("failed to read outbound certificate file: %w", err)
	}

	var certificate cert
	err = yaml.Unmarshal(bytes, &certificate)
	if err != nil {
		return tls.Certificate{}, errors.Newf("failed to unmarshal outbound certificate: %w", err)
	}

	clientCert, err := tls.X509KeyPair([]byte(certificate.Crt), []byte(certificate.Key))
	if err != nil {
		return tls.Certificate{}, errors.Newf("failed to load outbound certificate: %w", err)
	}

	return clientCert, nil
}

type ClientConfig struct {
	Certificate tls.Certificate
	Timeout     time.Duration
}

func NewClient(cfg ClientConfig) *http.Client {
	transport := &headerTransport{clientTransport: &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cfg.Certificate},
		},
	}}
	return &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}
}

type headerTransport struct {
	clientTransport http.RoundTripper
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	requestID := req.Context().Value(logCtx.RequestIDCtxKey).(string)
	if requestID != "" {
		req.Header.Add(logCtx.RequestIDHeader, requestID)
	}
	req.Header.Add("Content-Type", "application/json")
	return t.clientTransport.RoundTrip(req)
}
