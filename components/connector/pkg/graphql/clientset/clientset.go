package clientset

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"strings"

	"github.com/pkg/errors"
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
}

func WithSkipTLSVerify(skipTLSVerify bool) Option {
	return optionFunc(func(o *clientsetOptions) {
		o.skipTLSVerify = skipTLSVerify
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
	return newTokenSecuredClient(baseURL, cs.clientsetOptions.skipTLSVerify)
}

func (cs ConnectorClientSet) CertificateSecuredClient(baseURL string, certificate tls.Certificate) *CertificateSecuredClient {
	return newCertificateSecuredConnectorClient(baseURL, certificate)
}

func (c ConnectorClientSet) GenerateCertificateForToken(token, connectorURL string) (tls.Certificate, error) {
	connectorClient := newTokenSecuredClient(connectorURL, c.skipTLSVerify)

	config, err := connectorClient.Configuration(token)
	if err != nil {
		return tls.Certificate{}, err
	}

	key, csr, err := NewCSR(config.CertificateSigningRequestInfo.Subject, nil)
	if err != nil {
		return tls.Certificate{}, err
	}

	pemCSR := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csr.Raw,
	})

	encodedCSR := base64.StdEncoding.EncodeToString(pemCSR)

	certResult, err := connectorClient.SignCSR(encodedCSR, config.Token.Token)
	if err != nil {
		return tls.Certificate{}, err
	}

	pemCertChain, err := base64.StdEncoding.DecodeString(certResult.CertificateChain)
	if err != nil {
		return tls.Certificate{}, err
	}

	certs, err := decodeCertificates(pemCertChain)
	if err != nil {
		return tls.Certificate{}, err
	}

	return NewTLSCertificate(key, certs...), nil
}

func decodeCertificates(pemCertChain []byte) ([]*x509.Certificate, error) {
	if pemCertChain == nil {
		return nil, errors.New("Certificate data is empty")
	}

	var certificates []*x509.Certificate

	for block, rest := pem.Decode(pemCertChain); block != nil && rest != nil; {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to decode one of the pem blocks")
		}

		certificates = append(certificates, cert)

		block, rest = pem.Decode(rest)
	}

	if len(certificates) == 0 {
		return nil, errors.New("No certificates found in the pem block")
	}

	return certificates, nil
}

func NewTLSCertificate(key *rsa.PrivateKey, certificates ...*x509.Certificate) tls.Certificate {
	rawCerts := make([][]byte, len(certificates))
	for i, c := range certificates {
		rawCerts[i] = c.Raw
	}

	return tls.Certificate{
		Certificate: rawCerts,
		PrivateKey:  key,
	}
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
