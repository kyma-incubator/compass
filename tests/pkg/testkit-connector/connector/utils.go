package connector

import (
	"crypto/rsa"
	"fmt"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/testkit-connector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GenerateRuntimeCertificate(t *testing.T, token *externalschema.Token, connectorClient *TokenSecuredClient, clientKey *rsa.PrivateKey) (externalschema.CertificationResult, externalschema.Configuration) {
	return generateCertificateForToken(t, connectorClient, token.Token, clientKey)
}

func GetConfiguration(t *testing.T, internalClient *InternalClient, connectorClient *TokenSecuredClient, appID string) externalschema.Configuration {
	token, err := internalClient.GenerateApplicationToken(appID)
	require.NoError(t, err)

	configuration, err := connectorClient.Configuration(token.Token)
	require.NoError(t, err)
	AssertConfiguration(t, configuration)

	return configuration
}

func GenerateApplicationCertificate(t *testing.T, internalClient *InternalClient, connectorClient *TokenSecuredClient, appID string, clientKey *rsa.PrivateKey) (externalschema.CertificationResult, externalschema.Configuration) {
	token, err := internalClient.GenerateApplicationToken(appID)
	require.NoError(t, err)

	return generateCertificateForToken(t, connectorClient, token.Token, clientKey)
}

func generateCertificateForToken(t *testing.T, connectorClient *TokenSecuredClient, token string, clientKey *rsa.PrivateKey) (externalschema.CertificationResult, externalschema.Configuration) {
	configuration, err := connectorClient.Configuration(token)
	require.NoError(t, err)
	AssertConfiguration(t, configuration)

	certToken := configuration.Token.Token
	subject := configuration.CertificateSigningRequestInfo.Subject

	csr, err := testkit_connector.CreateCsr(subject, clientKey)
	require.NoError(t, err)

	result, err := connectorClient.SignCSR(csr, certToken)
	require.NoError(t, err)

	return result, configuration
}

func AssertConfiguration(t *testing.T, configuration externalschema.Configuration) {
	require.NotEmpty(t, configuration)
	require.NotNil(t, configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL)
	require.NotNil(t, configuration.ManagementPlaneInfo.DirectorURL)

	require.Equal(t, testkit_connector.RSAKey, configuration.CertificateSigningRequestInfo.KeyAlgorithm)
}

func AssertCertificate(t *testing.T, expectedSubject string, certificationResult externalschema.CertificationResult) {
	clientCert := certificationResult.ClientCertificate
	certChain := certificationResult.CertificateChain
	caCert := certificationResult.CaCertificate

	require.NotEmpty(t, clientCert)
	require.NotEmpty(t, certChain)
	require.NotEmpty(t, caCert)

	testkit_connector.CheckIfSubjectEquals(t, expectedSubject, clientCert)
	testkit_connector.CheckIfChainContainsTwoCertificates(t, certChain)
	testkit_connector.CheckCertificateChainOrder(t, certChain)
	testkit_connector.CheckIfCertIsSigned(t, clientCert, caCert)
}

func ChangeCommonName(subject, commonName string) string {
	splitSubject := testkit_connector.ParseSubject(subject)

	splitSubject.CommonName = commonName

	return splitSubject.String()
}

func CreateCertDataHeader(subject, hash string) string {
	return fmt.Sprintf(`By=spiffe://cluster.local/ns/kyma-system/sa/default;Hash=%s;Subject="%s";URI=`, hash, subject)
}

func Cleanup(t *testing.T, configmapCleaner *testkit_connector.ConfigmapCleaner, certificationResult externalschema.CertificationResult) {
	hash := testkit_connector.GetCertificateHash(t, certificationResult.ClientCertificate)
	err := configmapCleaner.CleanRevocationList(hash)
	assert.NoError(t, err)
}
