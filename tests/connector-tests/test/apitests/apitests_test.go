package apitests

import (
	"crypto/rsa"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/stretchr/testify/require"
)

func TestTokens(t *testing.T) {
	appID := "54f83a73-b340-418d-b653-d25b5ed47d75"
	runtimeID := "75f42q66-b340-418d-b653-d25b5ed47d75"

	t.Run("should return valid response on Configuration query for Application token", func(t *testing.T) {
		//when
		token, e := internalClient.GenerateApplicationToken(appID)

		//then
		require.NoError(t, e)

		//when
		config, e := connectorClient.Configuration(token.Token)

		//then
		require.NoError(t, e)
		assertConfiguration(t, config)
	})

	t.Run("should return valid response on Configuration query for Runtime token", func(t *testing.T) {
		//when
		token, e := internalClient.GenerateRuntimeToken(runtimeID)

		//then
		require.NoError(t, e)

		//when
		config, e := connectorClient.Configuration(token.Token)

		//then
		require.NoError(t, e)
		assertConfiguration(t, config)
	})

	t.Run("should not accept invalid token on Configuration query", func(t *testing.T) {
		//given
		wrongToken := "incorrectToken"

		//when
		configuration, e := connectorClient.Configuration(wrongToken)

		//then
		require.Empty(t, configuration)
		require.Error(t, e)
	})

	t.Run("should return error when token not provided", func(t *testing.T) {
		//when
		configuration, e := connectorClient.Configuration("")

		//then
		require.Empty(t, configuration)
		require.Error(t, e)
	})

	t.Run("should return error for previously used token on Configuration query", func(t *testing.T) {
		//when
		token, e := internalClient.GenerateApplicationToken(appID)

		//then
		require.NoError(t, e)

		//when
		_, e = connectorClient.Configuration(token.Token)

		//then
		require.NoError(t, e)

		//when
		configuration, e := connectorClient.Configuration(token.Token)

		//then
		require.Empty(t, configuration)
		require.Error(t, e)
	})
}

func TestCertificateGeneration(t *testing.T) {
	appID := "54f83a73-b340-418d-b653-d95b5e347d74"

	t.Run("should return client certificate with valid subject and signed with CA certificate", func(t *testing.T) {
		// when
		certResult, configuration := generateCertificate(t, appID, clientKey)

		// then
		assertCertificate(t, configuration.CertificateSigningRequestInfo.Subject, certResult)
	})

	t.Run("should return error when CSR subject is invalid", func(t *testing.T) {
		// given
		configuration := getConfiguration(t, appID)

		certToken := configuration.Token.Token
		wrongSubject := "subject=OU=Test,O=Test,L=Wrong,ST=Wrong,C=PL,CN=Wrong"

		csr, e := testkit.CreateCsr(wrongSubject, clientKey)
		require.NoError(t, e)

		// when
		cert, e := connectorClient.SignCSR(csr, certToken)

		// then
		require.Error(t, e)
		require.Empty(t, cert)
	})

	t.Run("should return error when different Common Name provided", func(t *testing.T) {
		// given
		configuration := getConfiguration(t, appID)

		certToken := configuration.Token.Token
		differentSubject := changeCommonName(configuration.CertificateSigningRequestInfo.Subject, "12y36g45-b340-418d-b653-d95b5e347d74")

		csr, e := testkit.CreateCsr(differentSubject, clientKey)
		require.NoError(t, e)

		// when
		cert, e := connectorClient.SignCSR(csr, certToken)

		// then
		require.Error(t, e)
		require.Empty(t, cert)
	})

	t.Run("should return error when signing certificate with invalid token", func(t *testing.T) {
		// given
		configuration := getConfiguration(t, appID)
		certInfo := configuration.CertificateSigningRequestInfo

		csr, e := testkit.CreateCsr(certInfo.Subject, clientKey)
		require.NoError(t, e)

		wrongToken := "wrongToken"

		// when
		cert, e := connectorClient.SignCSR(csr, wrongToken)

		// then
		require.Error(t, e)
		require.Empty(t, cert)
	})

	t.Run("should return error when signing certificate with already used token", func(t *testing.T) {
		// given
		configuration := getConfiguration(t, appID)
		certInfo := configuration.CertificateSigningRequestInfo

		csr, err := testkit.CreateCsr(certInfo.Subject, clientKey)
		require.NoError(t, err)

		cert, err := connectorClient.SignCSR(csr, configuration.Token.Token)
		require.NoError(t, err)
		assertCertificate(t, certInfo.Subject, cert)

		// when
		secondCert, err := connectorClient.SignCSR(csr, configuration.Token.Token)

		//then
		require.Error(t, err)
		require.Empty(t, secondCert)
	})

	t.Run("should return error when invalid CSR provided for signing", func(t *testing.T) {
		// given
		configuration := getConfiguration(t, appID)
		certToken := configuration.Token.Token
		wrongCSR := "wrongCSR"

		// when
		cert, e := connectorClient.SignCSR(wrongCSR, certToken)

		// then
		require.Error(t, e)
		require.Empty(t, cert)
	})
}

func TestFullConnectorFlow(t *testing.T) {
	appID := "54f83a73-b340-418d-b653-d95b5e347d76"

	t.Log("Generating certificate...")
	certificationResult, configuration := generateCertificate(t, appID, clientKey)
	assertCertificate(t, configuration.CertificateSigningRequestInfo.Subject, certificationResult)

	t.Log("Certificate generated. Creating secured client...")
	certChain := testkit.DecodeCertChain(t, certificationResult.CertificateChain)
	securedClient := connector.NewCertificateSecuredConnectorClient(config.SecuredConnectorURL, clientKey, certChain...)

	t.Log("Fetching configuration with certificate...")
	configWithCert, err := securedClient.Configuration()
	require.NoError(t, err)
	require.Equal(t, configuration.ManagementPlaneInfo, configWithCert.ManagementPlaneInfo)
	require.Equal(t, configuration.CertificateSigningRequestInfo, configWithCert.CertificateSigningRequestInfo)

	csr, err := testkit.CreateCsr(configWithCert.CertificateSigningRequestInfo.Subject, clientKey)
	require.NoError(t, err)

	renewalResult, err := securedClient.SignCSR(csr)
	require.NoError(t, err)
	assertCertificate(t, configWithCert.CertificateSigningRequestInfo.Subject, renewalResult)

	t.Log("Renewing certificate...")
	renewedCertChain := testkit.DecodeCertChain(t, certificationResult.CertificateChain)
	securedClientWithRenewedCert := connector.NewCertificateSecuredConnectorClient(config.SecuredConnectorURL, clientKey, renewedCertChain...)

	t.Log("Certificate renewed. Fetching configuration with renewed certificate...")
	configWithRenewedCert, err := securedClientWithRenewedCert.Configuration()
	require.NoError(t, err)
	require.Equal(t, configuration.ManagementPlaneInfo, configWithRenewedCert.ManagementPlaneInfo)
	require.Equal(t, configuration.CertificateSigningRequestInfo, configWithRenewedCert.CertificateSigningRequestInfo)

	t.Logf("Revoking certificate...")
	isRevoked, err := securedClientWithRenewedCert.RevokeCertificate()
	require.NoError(t, err)
	require.Equal(t, true, isRevoked)

	t.Logf("Certificate revoked. Failing to fetch configuration with revoked certificate...")
	configWithRevokedCert, err := securedClientWithRenewedCert.Configuration()
	require.Error(t, err)
	require.Equal(t, nil, configWithRevokedCert)

	defer cleanup(t, certificationResult)
}

func getConfiguration(t *testing.T, appID string) gqlschema.Configuration {
	token, err := internalClient.GenerateApplicationToken(appID)
	require.NoError(t, err)

	configuration, err := connectorClient.Configuration(token.Token)
	require.NoError(t, err)
	assertConfiguration(t, configuration)

	return configuration
}

func generateCertificate(t *testing.T, appID string, clientKey *rsa.PrivateKey) (gqlschema.CertificationResult, gqlschema.Configuration) {
	token, err := internalClient.GenerateApplicationToken(appID)
	require.NoError(t, err)

	return generateCertificateForToken(t, token.Token, clientKey)
}

func generateCertificateForToken(t *testing.T, token string, clientKey *rsa.PrivateKey) (gqlschema.CertificationResult, gqlschema.Configuration) {
	configuration, err := connectorClient.Configuration(token)
	require.NoError(t, err)
	assertConfiguration(t, configuration)

	certToken := configuration.Token.Token
	subject := configuration.CertificateSigningRequestInfo.Subject

	csr, err := testkit.CreateCsr(subject, clientKey)
	require.NoError(t, err)

	result, err := connectorClient.SignCSR(csr, certToken)
	require.NoError(t, err)

	return result, configuration
}

func assertConfiguration(t *testing.T, configuration gqlschema.Configuration) {
	require.NotEmpty(t, configuration)
	require.NotEmpty(t, configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL)
	require.NotEmpty(t, configuration.ManagementPlaneInfo.DirectorURL)

	require.Equal(t, testkit.RSAKey, configuration.CertificateSigningRequestInfo.KeyAlgorithm)
}

func assertCertificate(t *testing.T, expectedSubject string, certificationResult gqlschema.CertificationResult) {
	clientCert := certificationResult.ClientCertificate
	certChain := certificationResult.CertificateChain
	caCert := certificationResult.CaCertificate

	require.NotEmpty(t, clientCert)
	require.NotEmpty(t, certChain)
	require.NotEmpty(t, caCert)

	testkit.CheckIfSubjectEquals(t, expectedSubject, clientCert)
	testkit.CheckIfChainContainsTwoCertificates(t, certChain)
	testkit.CheckIfCertIsSigned(t, clientCert, caCert)
}

func changeCommonName(subject, commonName string) string {
	splitSubject := testkit.ParseSubject(subject)

	splitSubject.CommonName = commonName

	return splitSubject.String()
}

func createCertDataHeader(subject, hash string) string {
	return fmt.Sprintf(`By=spiffe://cluster.local/ns/kyma-system/sa/default;Hash=%s;Subject="%s";URI=`, hash, subject)
}

func cleanup(t *testing.T, certificationResult gqlschema.CertificationResult) {
	hash := testkit.GetCertificateHash(t, certificationResult.ClientCertificate)
	_ = configmapCleaner.CleanRevocationList(hash)
}
