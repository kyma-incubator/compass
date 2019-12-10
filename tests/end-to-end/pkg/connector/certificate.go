package connector

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"strings"
)

const (
	RSAKeySize = 2048
	RSAKey     = "rsa2048"
)

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

func DecodeCertChain(certificateChain string) ([]*x509.Certificate, error) {
	crtBytes, err := decodeBase64Cert(certificateChain)
	if err != nil {
		return nil, err
	}

	clientCrtPem, rest := pem.Decode(crtBytes)

	caCertPem, _ := pem.Decode(rest)

	certificates, err := x509.ParseCertificates(append(clientCrtPem.Bytes, caCertPem.Bytes...))
	if err != nil {
		return nil, err
	}

	return certificates, nil
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

func decodeBase64Cert(certificate string) ([]byte, error) {
	crtBytes, err := base64.StdEncoding.DecodeString(certificate)
	if err != nil {
		return nil, err
	}
	return crtBytes, nil
}
