package apitests

import (
	"github.com/stretchr/testify/require"
	"github.wdf.sap/compass/tests/connector-tests/test/testkit"
	"testing"
)

func TestTokens(t *testing.T) {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	appID := "54f83a73-b340-418d-b653-d25b5ed47d75"
	client := testkit.NewConnectorClient(config.APIUrl)

	t.Run("should return valid response on Configuration query", func(t *testing.T) {
		//when
		token, e := client.GenerateToken(appID)

		//then
		require.NoError(t, e)

		//when
		config, e := client.Configuration(token.Token)

		//then
		require.NoError(t, e)
		require.NotNil(t, config)
	})

	t.Run("should not accept invalid token on Configuration query", func(t *testing.T) {
		//given
		wrongToken := "incorrectToken"

		//when
		configuration, e := client.Configuration(wrongToken)

		//then
		require.Nil(t, configuration)
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
		require.Nil(t, configuration)
		require.NotNil(t, e)
	})
}

func TestCertificateGeneration(t *testing.T) {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	appID := "54f83a73-b340-418d-b653-d95b5e347d74"
	clientKey := testkit.CreateKey(t)
	client := testkit.NewConnectorClient(config.APIUrl)

	t.Run("should create client certificate", func(t *testing.T) {
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

		//when
		csr, e := testkit.CreateCsr(certInfo.Subject, clientKey)

		//then
		require.NoError(t, e)
		require.Equal(t, testkit.RSAKeySize, certInfo.Subject)

		//when
		result, e := client.GenerateCert(csr, certToken)

		//then
		require.NoError(t, e)

		clientCert := result.ClientCertificate
		require.NotEmpty(t, clientCert)

		testkit.CheckIfSubjectEquals(t, clientCert, certInfo.Subject)
	})

	t.Run("should return chain with two certificates", func(t *testing.T) {
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

		//when
		csr, e := testkit.CreateCsr(certInfo.Subject, clientKey)

		//then
		require.NoError(t, e)
		require.Equal(t, testkit.RSAKeySize, certInfo.Subject)

		//when
		result, e := client.GenerateCert(csr, certToken)

		//then
		require.NoError(t, e)

		certChain := result.Certificate
		require.NotEmpty(t, certChain)

		testkit.CheckIfChainContainsTwoCertificates(t, certChain)
	})

	t.Run("client certificate should be signed by CA certificate", func(t *testing.T) {
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

		//when
		csr, e := testkit.CreateCsr(certInfo.Subject, clientKey)

		//then
		require.NoError(t, e)
		require.Equal(t, testkit.RSAKeySize, certInfo.Subject)

		//when
		result, e := client.GenerateCert(csr, certToken)

		//then
		require.NoError(t, e)

		clientCert := result.Certificate
		require.NotEmpty(t, clientCert)

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
		require.Nil(t, cert)
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
		require.Equal(t, testkit.RSAKeySize, certInfo.Subject)

		//given
		wrongToken := "wrongToken"

		//when
		cert, e := client.GenerateCert(csr, wrongToken)

		//then
		require.Error(t, e)
		require.Nil(t, cert)
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
		require.Nil(t, cert)
	})
}
