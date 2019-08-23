package testkit

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const (
	rsaKeySize = 2048
)

func CreateKey(t *testing.T) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	require.NoError(t, err)

	return key
}

func CreateCsr(t *testing.T, strSubject string, keys *rsa.PrivateKey) string {
	subject := ParseSubject(strSubject)

	var csrTemplate = x509.CertificateRequest{
		Subject: subject,
	}

	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)
	require.NoError(t, err)

	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	encodedCsr := EncodeBase64Cert(csr)

	return encodedCsr
}

func ParseSubject(subject string) pkix.Name {
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

func EncodeBase64Cert(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func DecodeBase64Cert(certificate string, t *testing.T) []byte {
	crtBytes, err := base64.StdEncoding.DecodeString(certificate)
	require.NoError(t, err)
	return crtBytes
}
