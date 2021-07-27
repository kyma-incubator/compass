package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"

	"strings"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"github.com/kyma-incubator/compass/tests/pkg/model"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const (
	rsaKeySize = 2048
	RSAKey     = "rsa2048"
)

// Create Key generates rsa.PrivateKey
func CreateKey(t *testing.T) *rsa.PrivateKey {
	key, err := GenerateKey()
	require.NoError(t, err)

	return key
}

func GenerateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, rsaKeySize)
}

// CreateCsr creates CSR request
func CreateCsr(t *testing.T, strSubject string, keys *rsa.PrivateKey) []byte {
	subject := ParseSubject(strSubject)

	var csrTemplate = x509.CertificateRequest{
		Subject: subject,
	}

	// step: generate the csr request
	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)
	require.NoError(t, err)

	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	return csr
}

// EncodedCertChainToPemBytes decodes certificates chain and return pemBlock's bytes for client cert and ca cert
func EncodedCertChainToPemBytes(t *testing.T, encodedChain string) []byte {
	crtBytes := DecodeBase64Cert(t, encodedChain)

	clientCrtPem, rest := pem.Decode(crtBytes)
	require.NotNil(t, clientCrtPem)
	require.NotEmpty(t, rest)

	caCrtPem, _ := pem.Decode(rest)
	require.NotNil(t, caCrtPem)

	certChainBytes := append(clientCrtPem.Bytes, caCrtPem.Bytes...)

	return certChainBytes
}

// encodedCertToPemBytes decodes certificate and return pemBlock's bytes for it
func encodedCertToPemBytes(t *testing.T, encodedCert string) []byte {
	crtBytes := DecodeBase64Cert(t, encodedCert)

	certificate, _ := pem.Decode(crtBytes)
	require.NotNil(t, certificate)

	return certificate.Bytes
}

// DecodeAndParseCerts decodes base64 encoded certificates chain and parses it
func DecodeAndParseCerts(t *testing.T, crtResponse *model.CrtResponse) model.DecodedCrtResponse {
	certChainBytes := EncodedCertChainToPemBytes(t, crtResponse.CRTChain)
	certificateChain, err := x509.ParseCertificates(certChainBytes)
	require.NoError(t, err)

	clientCertBytes := encodedCertToPemBytes(t, crtResponse.ClientCRT)
	clientCertificate, err := x509.ParseCertificate(clientCertBytes)
	require.NoError(t, err)

	caCertificateBytes := encodedCertToPemBytes(t, crtResponse.CaCRT)
	caCertificate, err := x509.ParseCertificate(caCertificateBytes)
	require.NoError(t, err)

	return model.DecodedCrtResponse{
		CRTChain:  certificateChain,
		ClientCRT: clientCertificate,
		CaCRT:     caCertificate,
	}
}

// CheckIfSubjectEquals verifies that specified subject is equal to this in certificate
func CheckIfSubjectEquals(t *testing.T, expectedSubject string, certificate *x509.Certificate) {
	subjectInfo := extractSubject(expectedSubject)
	actualSubject := certificate.Subject

	require.Equal(t, subjectInfo["CN"], actualSubject.CommonName)
	require.Equal(t, []string{subjectInfo["C"]}, actualSubject.Country)
	require.Equal(t, []string{subjectInfo["O"]}, actualSubject.Organization)
	require.Equal(t, []string{subjectInfo["OU"]}, actualSubject.OrganizationalUnit)
	require.Equal(t, []string{subjectInfo["L"]}, actualSubject.Locality)
	require.Equal(t, []string{subjectInfo["ST"]}, actualSubject.Province)
}

// CheckIfCertIsSigned verifies that client certificate is signed by server certificate
func CheckIfCertIsSigned(t *testing.T, certificates []*x509.Certificate) {
	clientCrt := certificates[0]
	serverCrt := certificates[1]

	err := clientCrt.CheckSignatureFrom(serverCrt)

	require.NoError(t, err)
}

func CheckIfChainContainsTwoCertificates(t *testing.T, certChain string) {
	certificates := DecodeCertChain(t, certChain)
	require.Equal(t, 2, len(certificates))
}

// Certificate chain starts from leaf certificate and ends with a root certificate (https://tools.ietf.org/html/rfc5246#section-7.4.2).
// The correct certificate chain holds the following property: ith certificate in the chain is issued by (i+1)th certificate
func CheckCertificateChainOrder(t *testing.T, chain string) {
	certChain := DecodeCertChain(t, chain)

	for i := 0; i < len(certChain)-1; i++ {
		issuer := certChain[i].Issuer
		nextCertSubject := certChain[i+1].Subject

		require.Equal(t, nextCertSubject, issuer)
	}
}

func GetCertificateHash(t *testing.T, certificateStr string) string {
	cert := DecodeCert(t, certificateStr)
	sha := sha256.Sum256(cert.Raw)
	return hex.EncodeToString(sha[:])
}

func DecodeCert(t *testing.T, certificateStr string) *x509.Certificate {
	crtBytes := DecodeBase64Cert(t, certificateStr)

	clientCrtPem, _ := pem.Decode(crtBytes)
	require.NotNil(t, clientCrtPem)

	certificate, e := x509.ParseCertificate(clientCrtPem.Bytes)
	require.NoError(t, e)
	return certificate
}

func DecodeCertChain(t *testing.T, certificateChain string) []*x509.Certificate {
	crtBytes := DecodeBase64Cert(t, certificateChain)

	clientCrtPem, rest := pem.Decode(crtBytes)
	require.NotNil(t, clientCrtPem)

	caCertPem, _ := pem.Decode(rest)
	require.NotNil(t, caCertPem)

	certificates, e := x509.ParseCertificates(append(clientCrtPem.Bytes, caCertPem.Bytes...))
	require.NoError(t, e)

	return certificates
}

func EncodeBase64(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
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

func DecodeBase64Cert(t *testing.T, certificate string) []byte {
	crtBytes, err := base64.StdEncoding.DecodeString(certificate)
	require.NoError(t, err)
	return crtBytes
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

func AssertConfiguration(t *testing.T, configuration externalschema.Configuration) {
	require.NotEmpty(t, configuration)
	require.NotNil(t, configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL)
	require.NotNil(t, configuration.ManagementPlaneInfo.DirectorURL)

	require.Equal(t, RSAKey, configuration.CertificateSigningRequestInfo.KeyAlgorithm)
}

func AssertCertificate(t *testing.T, expectedSubject string, certificationResult externalschema.CertificationResult) {
	clientCert := certificationResult.ClientCertificate
	certChain := certificationResult.CertificateChain
	caCert := certificationResult.CaCertificate

	require.NotEmpty(t, clientCert)
	require.NotEmpty(t, certChain)
	require.NotEmpty(t, caCert)

	CheckIfSubjectEquals(t, expectedSubject, DecodeCert(t, clientCert))
	CheckIfChainContainsTwoCertificates(t, certChain)
	CheckCertificateChainOrder(t, certChain)
	certificates := []*x509.Certificate{DecodeCert(t, clientCert), DecodeCert(t, caCert)}
	CheckIfCertIsSigned(t, certificates)
}

type ConfigurationResponse struct {
	Result externalschema.Configuration `json:"result"`
}

type CertificationResponse struct {
	Result externalschema.CertificationResult `json:"result"`
}

type RevokeResult struct {
	Result bool `json:"result"`
}

func ChangeCommonName(subject, commonName string) string {
	splitSubject := ParseSubject(subject)

	splitSubject.CommonName = commonName

	return splitSubject.String()
}

func CreateCertDataHeader(subject, hash string) string {
	return fmt.Sprintf(`By=spiffe://cluster.local/ns/kyma-system/sa/default;Hash=%s;Subject="%s";URI=`, hash, subject)
}

func Cleanup(t *testing.T, configmapCleaner *k8s.ConfigmapCleaner, certificationResult externalschema.CertificationResult) {
	hash := GetCertificateHash(t, certificationResult.ClientCertificate)
	err := configmapCleaner.CleanRevocationList(hash)
	assert.NoError(t, err)
}
