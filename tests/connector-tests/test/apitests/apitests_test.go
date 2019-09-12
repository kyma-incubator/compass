package apitests

import (
	"crypto/rsa"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/stretchr/testify/require"
)

func TestTokens(t *testing.T) {
	appID := "54f83a73-b340-418d-b653-d25b5ed47d75"

	t.Run("should return valid response on Configuration query for Application token", func(t *testing.T) {
		//when
		token, e := internalClient.GenerateApplicationToken(appID)

		//then
		require.NoError(t, e)

		//when
		config, e := connectorClient.Configuration(token.Token)

		//then
		require.NoError(t, e)
		require.NotEmpty(t, config)
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
	clientKey := testkit.CreateKey(t)

	t.Run("should return client certificate with valid subject and signed with CA certificate", func(t *testing.T) {
		//when
		certResult, configuration := generateCertificate(t, appID, clientKey)

		// then
		assertCertificate(t, configuration.CertificateSigningRequestInfo.Subject, certResult)
	})

	t.Run("should return error when CSR subject is invalid", func(t *testing.T) {
		//when
		token, e := internalClient.GenerateApplicationToken(appID)

		//then
		require.NoError(t, e)

		//when
		configuration, e := connectorClient.Configuration(token.Token)

		//then
		require.NoError(t, e)

		//given
		certToken := configuration.Token.Token
		wrongSubject := "subject=OU=Test,O=Test,L=Wrong,ST=Wrong,C=PL,CN=Wrong"

		//when
		csr, e := testkit.CreateCsr(wrongSubject, clientKey)

		//then
		require.NoError(t, e)

		//when
		cert, e := connectorClient.SignCSR(csr, certToken)

		//then
		require.Error(t, e)
		require.Empty(t, cert)
	})

	t.Run("should return error on invalid token on Certificate Generation mutation", func(t *testing.T) {
		//when
		token, e := internalClient.GenerateApplicationToken(appID)

		//then
		require.NoError(t, e)

		//when
		configuration, e := connectorClient.Configuration(token.Token)

		//then
		require.NoError(t, e)

		//given
		certInfo := configuration.CertificateSigningRequestInfo

		//when
		csr, e := testkit.CreateCsr(certInfo.Subject, clientKey)

		//then
		require.NoError(t, e)
		require.Equal(t, testkit.RSAKey, certInfo.KeyAlgorithm)

		//given
		wrongToken := "wrongToken"

		//when
		cert, e := connectorClient.SignCSR(csr, wrongToken)

		//then
		require.Error(t, e)
		require.Empty(t, cert)
	})

	t.Run("should return error on invalid CSR on Certificate Generation mutation", func(t *testing.T) {
		//when
		token, e := internalClient.GenerateApplicationToken(appID)

		//then
		require.NoError(t, e)

		//when
		configuration, e := connectorClient.Configuration(token.Token)

		//then
		require.NoError(t, e)

		//given
		certToken := configuration.Token.Token
		wrongCSR := "wrongCSR"

		//when
		cert, e := connectorClient.SignCSR(wrongCSR, certToken)

		//then
		require.Error(t, e)
		require.Empty(t, cert)
	})
}

func TestFullConnectorFlow(t *testing.T) {
	appID := "54f83a73-b340-418d-b653-d95b5e347d74"
	clientKey := testkit.CreateKey(t)

	//when
	certificationResult, configuration := generateCertificate(t, appID, clientKey)

	assertCertificate(t, configuration.CertificateSigningRequestInfo.Subject, certificationResult)

	// when
	certChain := testkit.DecodeCertChain(t, certificationResult.CertificateChain)
	securedClient := connector.NewSecuredConnectorClient(config.SecuredConnectorURL, clientKey, certChain...)

	// then
	configWithCert, err := securedClient.Configuration()
	require.NoError(t, err)
	require.Equal(t, configuration.ManagementPlaneInfo, configWithCert.ManagementPlaneInfo)
	require.Equal(t, configuration.CertificateSigningRequestInfo, configWithCert.CertificateSigningRequestInfo)

	// when
	csr, err := testkit.CreateCsr(configWithCert.CertificateSigningRequestInfo.Subject, clientKey)
	require.NoError(t, err)

	renewalResult, err := securedClient.SignCSR(csr)
	require.NoError(t, err)

	// then
	assertCertificate(t, configWithCert.CertificateSigningRequestInfo.Subject, renewalResult)

	// when
	renewedCertChain := testkit.DecodeCertChain(t, certificationResult.CertificateChain)
	securedClientWithRenewedCert := connector.NewSecuredConnectorClient(config.SecuredConnectorURL, clientKey, renewedCertChain...)

	configWithRenewedCert, err := securedClientWithRenewedCert.Configuration()
	require.NoError(t, err)
	require.Equal(t, configuration.ManagementPlaneInfo, configWithRenewedCert.ManagementPlaneInfo)
	require.Equal(t, configuration.CertificateSigningRequestInfo, configWithRenewedCert.CertificateSigningRequestInfo)
}

func generateCertificate(t *testing.T, appID string, clientKey *rsa.PrivateKey) (gqlschema.CertificationResult, gqlschema.Configuration) {
	token, err := internalClient.GenerateApplicationToken(appID)
	require.NoError(t, err)

	configuration, err := connectorClient.Configuration(token.Token)
	require.NoError(t, err)

	certInfo := configuration.CertificateSigningRequestInfo
	certToken := configuration.Token.Token
	subject := certInfo.Subject

	csr, err := testkit.CreateCsr(subject, clientKey)
	require.NoError(t, err)
	require.Equal(t, testkit.RSAKey, certInfo.KeyAlgorithm)

	result, err := connectorClient.SignCSR(csr, certToken)
	require.NoError(t, err)

	return result, configuration
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

func TestHydrators(t *testing.T) {
	// TODO: test hydrators
}

func TestOathkeeperSecurity(t *testing.T) {
	appID := "54f83a73-b340-418d-b653-d95b5e347d74"
	clientKey := testkit.CreateKey(t)

	certResult, configuration := generateCertificate(t, appID, clientKey)
	certChain := testkit.DecodeCertChain(t, certResult.CertificateChain)
	securedClient := connector.NewSecuredConnectorClient(config.SecuredConnectorURL, clientKey, certChain...)

	t.Run("client id headers should be stripped when calling token-secured api", func(t *testing.T) {
		// given
		forbiddenHeaders := map[string][]string{
			oathkeeper.ClientIdFromTokenHeader:       {appID},
			oathkeeper.ClientIdFromCertificateHeader: {appID},
		}

		csr, err := testkit.CreateCsr(configuration.CertificateSigningRequestInfo.Subject, clientKey)

		// when
		_, err = connectorClient.Configuration("", forbiddenHeaders)

		// then
		require.Error(t, err)

		// when
		_, err = connectorClient.SignCSR(csr, "", forbiddenHeaders)

		// then
		require.Error(t, err)
	})

	t.Run("certificate data header should be stripped", func(t *testing.T) {
		// given
		changedAppID := "aaabbbcc-b340-418d-b653-d95b5e347d74"

		newSubject := changeCommonName(configuration.CertificateSigningRequestInfo.Subject, changedAppID)
		certDataHeader := fmt.Sprintf(`By=spiffe://cluster.local/ns/kyma-system/sa/default;Hash=df6ab69b34100a1808ddc6211010fa289518f14606d0c8eaa03a0f53ecba578a;Subject="%s";URI=`, newSubject)

		forbiddenHeaders := map[string][]string{
			"Certificate-Data": {certDataHeader},
		}

		csr, err := testkit.CreateCsr(newSubject, clientKey)

		t.Run("when calling token-secured API", func(t *testing.T) {
			// when
			_, err = connectorClient.Configuration("", forbiddenHeaders)

			// then
			require.Error(t, err)

			// when
			_, err = connectorClient.SignCSR(csr, "", forbiddenHeaders)

			// then
			require.Error(t, err)
		})

		t.Run("when calling certificate-secured API", func(t *testing.T) {
			// when
			config, err := securedClient.Configuration(forbiddenHeaders)

			// then
			require.NoError(t, err)
			require.Equal(t, configuration.CertificateSigningRequestInfo.Subject, config.CertificateSigningRequestInfo.Subject)

			// when
			_, err = securedClient.SignCSR(csr, forbiddenHeaders)

			// then
			require.Error(t, err)
		})
	})

}

func changeCommonName(subject, commonName string) string {
	splitSubject := testkit.ParseSubject(subject)

	splitSubject.CommonName = commonName

	return splitSubject.String()
}
