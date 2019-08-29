package apitests

import (
	"testing"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/stretchr/testify/require"
)

func TestTokens(t *testing.T) {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	appID := "54f83a73-b340-418d-b653-d25b5ed47d75"
	client := testkit.NewConnectorClient(config.InternalConnectorUrl)

	t.Run("should return valid response on Configuration query", func(t *testing.T) {
		//when
		token, e := client.GenerateToken(appID)

		//then
		require.NoError(t, e)

		//when
		config, e := client.Configuration(token.Token)

		//then
		require.NoError(t, e)
		require.NotEmpty(t, config)
	})

	t.Run("should not accept invalid token on Configuration query", func(t *testing.T) {
		//given
		wrongToken := "incorrectToken"

		//when
		configuration, e := client.Configuration(wrongToken)

		//then
		require.Empty(t, configuration)
		require.NotNil(t, e)
	})

	t.Run("should return error for previously used token on Configuration query", func(t *testing.T) {
		//when
		token, e := client.GenerateToken(appID)

		//then
		require.NoError(t, e)

		//when
		_, e = client.Configuration(token.Token)

		//then
		require.NoError(t, e)

		//when
		configuration, e := client.Configuration(token.Token)

		//then
		require.Empty(t, configuration)
		require.NotNil(t, e)
	})
}

func TestCertificateGeneration(t *testing.T) {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	appID := "54f83a73-b340-418d-b653-d95b5e347d74"
	clientKey := testkit.CreateKey(t)
	client := testkit.NewConnectorClient(config.InternalConnectorUrl)

	t.Run("should return client certificate with valid subject and signed with CA certificate", func(t *testing.T) {
		//when
		token, e := client.GenerateToken(appID)

		//then
		require.NoError(t, e)

		//when
		configuration, e := client.Configuration(token.Token)

		//then
		require.NoError(t, e)

		//given
		certInfo := configuration.CertificateSigningRequestInfo
		certToken := configuration.Token.Token
		subject := certInfo.Subject

		//when
		csr, e := testkit.CreateCsr(subject, clientKey)

		//then
		require.NoError(t, e)
		require.Equal(t, testkit.RSAKey, certInfo.KeyAlgorithm)

		//when
		result, e := client.GenerateCert(csr, certToken)

		//then
		require.NoError(t, e)

		clientCert := result.ClientCertificate
		require.NotEmpty(t, clientCert)
		testkit.CheckIfSubjectEquals(t, clientCert, certInfo.Subject)

		certChain := result.CertificateChain
		require.NotEmpty(t, certChain)
		testkit.CheckIfChainContainsTwoCertificates(t, certChain)

		caCert := result.CaCertificate
		require.NotEmpty(t, caCert)
		testkit.CheckIfCertIsSigned(t, clientCert, caCert)
	})

	t.Run("should verify CSR subject", func(t *testing.T) {
		//when
		token, e := client.GenerateToken(appID)

		//then
		require.NoError(t, e)

		//when
		configuration, e := client.Configuration(token.Token)

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
		cert, e := client.GenerateCert(csr, certToken)

		//then
		require.Error(t, e)
		require.Empty(t, cert)
	})

	t.Run("should return error on invalid token on Certificate Generation mutation", func(t *testing.T) {
		//when
		token, e := client.GenerateToken(appID)

		//then
		require.NoError(t, e)

		//when
		configuration, e := client.Configuration(token.Token)

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
		cert, e := client.GenerateCert(csr, wrongToken)

		//then
		require.Error(t, e)
		require.Empty(t, cert)
	})

	t.Run("should return error on invalid CSR on Certificate Generation mutation", func(t *testing.T) {
		//when
		token, e := client.GenerateToken(appID)

		//then
		require.NoError(t, e)

		//when
		configuration, e := client.Configuration(token.Token)

		//then
		require.NoError(t, e)

		//given
		certToken := configuration.Token.Token
		wrongCSR := "wrongCSR"

		//when
		cert, e := client.GenerateCert(wrongCSR, certToken)

		//then
		require.Error(t, e)
		require.Empty(t, cert)
	})
}
