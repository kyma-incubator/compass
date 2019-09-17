package testkit

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	RSAKeySize = 2048
	RSAKey     = "rsa2048"
)

func CreateKey(t *testing.T) *rsa.PrivateKey {
	key, err := GenerateKey()
	require.NoError(t, err)

	return key
}

func GenerateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, RSAKeySize)
}

func CreateCsr(strSubject string, keys *rsa.PrivateKey) (string, error) {
	subject := ParseSubject(strSubject)

	var csrTemplate = x509.CertificateRequest{
		Subject: subject,
	}

	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)

	if err != nil {
		return "", err
	}

	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	encodedCsr := encodeBase64Cert(csr)

	return encodedCsr, nil
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

func CheckIfSubjectEquals(t *testing.T, expectedSubject, certificateStr string) {
	certificate := decodeCert(t, certificateStr)

	actualSubject := certificate.Subject
	subjectInfo := extractSubject(expectedSubject)

	require.Equal(t, subjectInfo["CN"], actualSubject.CommonName)
	require.Equal(t, []string{subjectInfo["C"]}, actualSubject.Country)
	require.Equal(t, []string{subjectInfo["O"]}, actualSubject.Organization)
	require.Equal(t, []string{subjectInfo["OU"]}, actualSubject.OrganizationalUnit)
	require.Equal(t, []string{subjectInfo["L"]}, actualSubject.Locality)
	require.Equal(t, []string{subjectInfo["ST"]}, actualSubject.Province)
}

func CheckIfChainContainsTwoCertificates(t *testing.T, certChain string) {
	certificates := DecodeCertChain(t, certChain)
	require.Equal(t, 2, len(certificates))
}

func CheckIfCertIsSigned(t *testing.T, clientCertStr, caCertStr string) {
	clientCert := decodeCert(t, clientCertStr)

	caCert := decodeCert(t, caCertStr)

	err := clientCert.CheckSignatureFrom(caCert)
	require.NoError(t, err)
}

func GetCertificateHash(t *testing.T, certificateStr string) string {
	cert := decodeCert(t, certificateStr)
	sha := sha256.Sum256(cert.Raw)
	return hex.EncodeToString(sha[:])
}

func decodeCert(t *testing.T, certificateStr string) *x509.Certificate {
	crtBytes := decodeBase64Cert(t, certificateStr)

	clientCrtPem, _ := pem.Decode(crtBytes)
	require.NotNil(t, clientCrtPem)

	certificate, e := x509.ParseCertificate(clientCrtPem.Bytes)
	require.NoError(t, e)
	return certificate
}

func DecodeCertChain(t *testing.T, certificateChain string) []*x509.Certificate {
	crtBytes := decodeBase64Cert(t, certificateChain)

	clientCrtPem, rest := pem.Decode(crtBytes)
	require.NotNil(t, clientCrtPem)

	caCertPem, _ := pem.Decode(rest)
	require.NotNil(t, caCertPem)

	certificates, e := x509.ParseCertificates(append(clientCrtPem.Bytes, caCertPem.Bytes...))
	require.NoError(t, e)

	return certificates
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

func encodeBase64Cert(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func decodeBase64Cert(t *testing.T, certificate string) []byte {
	crtBytes, err := base64.StdEncoding.DecodeString(certificate)
	require.NoError(t, err)
	return crtBytes
}
