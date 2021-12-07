package clientset

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
)

const (
	keyLength = 4096
)

type ConnectorClientSet struct {
	*clientsetOptions
}

type Option interface {
	apply(*clientsetOptions)
}

type optionFunc func(*clientsetOptions)

func (f optionFunc) apply(o *clientsetOptions) {
	f(o)
}

type clientsetOptions struct {
	skipTLSVerify bool
	timeout       time.Duration
}

func WithSkipTLSVerify(skipTLSVerify bool) Option {
	return optionFunc(func(o *clientsetOptions) {
		o.skipTLSVerify = skipTLSVerify
	})
}

func WithTimeout(timeout time.Duration) Option {
	return optionFunc(func(o *clientsetOptions) {
		o.timeout = timeout
	})
}

func NewConnectorClientSet(options ...Option) *ConnectorClientSet {
	opts := &clientsetOptions{
		skipTLSVerify: false,
	}

	for _, opt := range options {
		opt.apply(opts)
	}

	return &ConnectorClientSet{
		clientsetOptions: opts,
	}
}

func (cs ConnectorClientSet) TokenSecuredClient(baseURL string) *TokenSecuredClient {
	return newTokenSecuredClient(baseURL, cs.clientsetOptions)
}

func (cs ConnectorClientSet) CertificateSecuredClient(baseURL string, certificate tls.Certificate) *CertificateSecuredClient {
	return newCertificateSecuredConnectorClient(baseURL, certificate, cs.clientsetOptions)
}

func (cs ConnectorClientSet) GenerateCertificateForToken(ctx context.Context, token, connectorURL string) (tls.Certificate, error) {
	connectorClient := newTokenSecuredClient(connectorURL, cs.clientsetOptions)

	config, err := connectorClient.Configuration(ctx, token)
	if err != nil {
		return tls.Certificate{}, err
	}

	key, csr, err := NewCSR(config.CertificateSigningRequestInfo.Subject, nil)
	if err != nil {
		return tls.Certificate{}, err
	}

	encodedCSR := encodeCSR(csr)

	certResult, err := connectorClient.SignCSR(ctx, encodedCSR, config.Token.Token)
	if err != nil {
		return tls.Certificate{}, err
	}

	pemCertChain, err := base64.StdEncoding.DecodeString(certResult.CertificateChain)
	if err != nil {
		return tls.Certificate{}, err
	}

	certs, err := cert.DecodeCertificates(pemCertChain)
	if err != nil {
		return tls.Certificate{}, err
	}

	return cert.NewTLSCertificate(key, certs...), nil
}

func encodeCSR(csr *x509.CertificateRequest) string {
	pemCSR := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csr.Raw,
	})

	return base64.StdEncoding.EncodeToString(pemCSR)
}

func NewCSR(subject string, key *rsa.PrivateKey) (*rsa.PrivateKey, *x509.CertificateRequest, error) {
	var err error

	if key == nil {
		key, err = rsa.GenerateKey(rand.Reader, keyLength)
		if err != nil {
			return nil, nil, err
		}
	}

	parsedSubject := parseSubject(subject)

	var csrTemplate = x509.CertificateRequest{
		Subject: parsedSubject,
	}

	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, key)
	if err != nil {
		return nil, nil, err
	}

	csr, err := x509.ParseCertificateRequest(csrCertificate)
	if err != nil {
		return nil, nil, err
	}

	return key, csr, nil
}

func parseSubject(subject string) pkix.Name {
	subjectInfo := extractSubject(subject)

	return pkix.Name{
		CommonName:         subjectInfo["CN"],
		Country:            []string{subjectInfo["C"]},
		Organization:       []string{subjectInfo["O"]},
		OrganizationalUnit: []string{subjectInfo["OU"]},
		Locality:           []string{subjectInfo["L"]},
		Province:           []string{subjectInfo["ST"]},
	}
}

func extractSubject(subject string) map[string]string {
	result := map[string]string{}

	segments := strings.Split(subject, ",")

	for _, segment := range segments {
		parts := strings.Split(segment, "=")
		result[parts[0]] = parts[1]
	}

	return result
}
