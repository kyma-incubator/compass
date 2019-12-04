package clientset

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"strings"
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

func NewCSR(subject string, keyLength int) (*rsa.PrivateKey, *x509.CertificateRequest, error) {
	key, err := rsa.GenerateKey(rand.Reader, keyLength)
	if err != nil {
		return nil, nil, err
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
