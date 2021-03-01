package apitests

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"
	director "github.com/kyma-incubator/compass/tests/director/gateway-integration"
	"github.com/stretchr/testify/require"
)

func TestTokens(t *testing.T) {
	runtime := director.RegisterRuntimeFromInputWithinTenant(t, ctx, directorClient.DexGraphqlClient, config.Tenant, &graphql.RuntimeInput{
		Name: "test-tokens-runtime",
	})
	runtimeID := runtime.ID
	defer director.UnregisterRuntimeWithinTenant(t, ctx, directorClient.DexGraphqlClient, config.Tenant, runtimeID)

	app, err := director.RegisterApplicationWithinTenant(t, ctx, directorClient.DexGraphqlClient, config.Tenant, graphql.ApplicationRegisterInput{
		Name: "test-tokens-app",
	})
	require.NoError(t, err)
	appID := app.ID
	defer director.UnregisterApplication(t, ctx, directorClient.DexGraphqlClient, config.Tenant, appID)

	t.Run("should return valid response on Configuration query for Application token", func(t *testing.T) {
		//when
		token, e := directorClient.GenerateApplicationToken(t, appID)

		//then
		require.NoError(t, e)

		//when
		config, e := connectorClient.Configuration(token.Token)

		//then
		require.NoError(t, e)
		connector.AssertConfiguration(t, config)
	})

	t.Run("should return valid response on Configuration query for Runtime token", func(t *testing.T) {
		//when
		token, e := directorClient.GenerateRuntimeToken(t, runtimeID)

		//then
		require.NoError(t, e)

		//when
		config, e := connectorClient.Configuration(token.Token)

		//then
		require.NoError(t, e)
		connector.AssertConfiguration(t, config)
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
		token, e := directorClient.GenerateApplicationToken(t, appID)

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
	app, err := director.RegisterApplicationWithinTenant(t, ctx, directorClient.DexGraphqlClient, config.Tenant, graphql.ApplicationRegisterInput{
		Name: "test-cert-gen-app",
	})
	require.NoError(t, err)
	appID := app.ID
	defer director.UnregisterApplication(t, ctx, directorClient.DexGraphqlClient, config.Tenant, appID)

	t.Run("should return client certificate with valid subject and signed with CA certificate", func(t *testing.T) {
		// when
		certResult, configuration := connector.GenerateApplicationCertificate(t, directorClient, connectorClient, appID, clientKey)

		// then
		connector.AssertCertificate(t, configuration.CertificateSigningRequestInfo.Subject, certResult)
	})

	t.Run("should return error when CSR subject is invalid", func(t *testing.T) {
		// given
		configuration := connector.GetConfiguration(t, directorClient, connectorClient, appID)

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
		configuration := connector.GetConfiguration(t, directorClient, connectorClient, appID)

		certToken := configuration.Token.Token
		differentSubject := connector.ChangeCommonName(configuration.CertificateSigningRequestInfo.Subject, "12y36g45-b340-418d-b653-d95b5e347d74")

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
		configuration := connector.GetConfiguration(t, directorClient, connectorClient, appID)
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
		configuration := connector.GetConfiguration(t, directorClient, connectorClient, appID)
		certInfo := configuration.CertificateSigningRequestInfo

		csr, err := testkit.CreateCsr(certInfo.Subject, clientKey)
		require.NoError(t, err)

		cert, err := connectorClient.SignCSR(csr, configuration.Token.Token)
		require.NoError(t, err)
		connector.AssertCertificate(t, certInfo.Subject, cert)

		// when
		secondCert, err := connectorClient.SignCSR(csr, configuration.Token.Token)

		//then
		require.Error(t, err)
		require.Empty(t, secondCert)
	})

	t.Run("should return error when invalid CSR provided for signing", func(t *testing.T) {
		// given
		configuration := connector.GetConfiguration(t, directorClient, connectorClient, appID)
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
	app, err := director.RegisterApplicationWithinTenant(t, ctx, directorClient.DexGraphqlClient, config.Tenant, graphql.ApplicationRegisterInput{
		Name: "test-full-flow-app",
	})
	require.NoError(t, err)
	appID := app.ID
	defer director.UnregisterApplication(t, ctx, directorClient.DexGraphqlClient, config.Tenant, appID)

	t.Log("Generating certificate...")
	certificationResult, configuration := connector.GenerateApplicationCertificate(t, directorClient, connectorClient, appID, clientKey)
	connector.AssertCertificate(t, configuration.CertificateSigningRequestInfo.Subject, certificationResult)

	defer connector.Cleanup(t, configmapCleaner, certificationResult)

	t.Log("Certificate generated. Creating secured client...")
	certChain := testkit.DecodeCertChain(t, certificationResult.CertificateChain)
	securedClient := connector.NewCertificateSecuredConnectorClient(*configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL, clientKey, certChain...)

	t.Log("Fetching configuration with certificate...")
	configWithCert, err := securedClient.Configuration()
	require.NoError(t, err)
	require.Equal(t, configuration.ManagementPlaneInfo, configWithCert.ManagementPlaneInfo)
	require.Equal(t, configuration.CertificateSigningRequestInfo, configWithCert.CertificateSigningRequestInfo)

	csr, err := testkit.CreateCsr(configWithCert.CertificateSigningRequestInfo.Subject, clientKey)
	require.NoError(t, err)

	renewalResult, err := securedClient.SignCSR(csr)
	require.NoError(t, err)
	connector.AssertCertificate(t, configWithCert.CertificateSigningRequestInfo.Subject, renewalResult)

	t.Log("Renewing certificate...")
	renewedCertChain := testkit.DecodeCertChain(t, certificationResult.CertificateChain)
	securedClientWithRenewedCert := connector.NewCertificateSecuredConnectorClient(*configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL, clientKey, renewedCertChain...)

	t.Log("Certificate renewed. Fetching configuration with renewed certificate...")
	configWithRenewedCert, err := securedClientWithRenewedCert.Configuration()
	require.NoError(t, err)
	require.Equal(t, configuration.ManagementPlaneInfo, configWithRenewedCert.ManagementPlaneInfo)
	require.Equal(t, configuration.CertificateSigningRequestInfo, configWithRenewedCert.CertificateSigningRequestInfo)

	t.Logf("Revoking certificate...")
	isRevoked, err := securedClientWithRenewedCert.RevokeCertificate()
	require.NoError(t, err)
	require.True(t, isRevoked)

	t.Logf("Certificate revoked. Failing to fetch configuration with revoked certificate...")
	configWithRevokedCert, err := securedClientWithRenewedCert.Configuration()
	require.Error(t, err)
	require.Nil(t, configWithRevokedCert.Token)
	require.Nil(t, configWithRevokedCert.CertificateSigningRequestInfo)
	require.Nil(t, configWithRevokedCert.ManagementPlaneInfo)
}
