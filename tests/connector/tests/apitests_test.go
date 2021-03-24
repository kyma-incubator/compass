package tests

import (
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestTokens(t *testing.T) {
	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, &graphql.RuntimeInput{
		Name: "test-tokens-runtime",
	})
	runtimeID := runtime.ID
	defer fixtures.UnregisterRuntime(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, runtimeID)

	app, err := fixtures.RegisterApplicationFromInput(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, graphql.ApplicationRegisterInput{
		Name: "test-tokens-app",
	})
	require.NoError(t, err)
	appID := app.ID
	defer fixtures.UnregisterApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, appID)

	t.Run("should return valid response on configuration query for Application token", func(t *testing.T) {
		//when
		token, e := directorClient.GenerateApplicationToken(t, appID)

		//then
		require.NoError(t, e)

		//when
		config, e := connectorClient.Configuration(token.Token)

		//then
		require.NoError(t, e)
		certs.AssertConfiguration(t, config)
	})

	t.Run("should return valid response on configuration query for Runtime token", func(t *testing.T) {
		//when
		token, e := directorClient.GenerateRuntimeToken(t, runtimeID)

		//then
		require.NoError(t, e)

		//when
		config, e := connectorClient.Configuration(token.Token)

		//then
		require.NoError(t, e)
		certs.AssertConfiguration(t, config)
	})

	t.Run("should not accept invalid token on configuration query", func(t *testing.T) {
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

	t.Run("should return error for previously used token on configuration query", func(t *testing.T) {
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
	app, err := fixtures.RegisterApplicationFromInput(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, graphql.ApplicationRegisterInput{
		Name: "test-cert-gen-app",
	})
	require.NoError(t, err)
	appID := app.ID
	defer fixtures.UnregisterApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, appID)

	t.Run("should return client certificate with valid subject and signed with CA certificate", func(t *testing.T) {
		// when
		certResult, configuration := clients.GenerateApplicationCertificate(t, directorClient, connectorClient, appID, clientKey)

		// then
		certs.AssertCertificate(t, configuration.CertificateSigningRequestInfo.Subject, certResult)
	})

	t.Run("should return error when CSR subject is invalid", func(t *testing.T) {
		// given
		configuration := clients.GetConfiguration(t, directorClient, connectorClient, appID)

		certToken := configuration.Token.Token
		wrongSubject := "subject=OU=Test,O=Test,L=Wrong,ST=Wrong,C=PL,CN=Wrong"

		csr := certs.CreateCsr(t, wrongSubject, clientKey)

		// when
		cert, e := connectorClient.SignCSR(certs.EncodeBase64(csr), certToken)

		// then
		require.Error(t, e)
		require.Empty(t, cert)
	})

	t.Run("should return error when different Common Name provided", func(t *testing.T) {
		// given
		configuration := clients.GetConfiguration(t, directorClient, connectorClient, appID)

		certToken := configuration.Token.Token
		differentSubject := certs.ChangeCommonName(configuration.CertificateSigningRequestInfo.Subject, "12y36g45-b340-418d-b653-d95b5e347d74")

		csr := certs.CreateCsr(t, differentSubject, clientKey)

		// when
		cert, e := connectorClient.SignCSR(certs.EncodeBase64(csr), certToken)

		// then
		require.Error(t, e)
		require.Empty(t, cert)
	})

	t.Run("should return error when signing certificate with invalid token", func(t *testing.T) {
		// given
		configuration := clients.GetConfiguration(t, directorClient, connectorClient, appID)

		certInfo := configuration.CertificateSigningRequestInfo

		csr := certs.CreateCsr(t, certInfo.Subject, clientKey)

		wrongToken := "wrongToken"

		// when
		cert, e := connectorClient.SignCSR(certs.EncodeBase64(csr), wrongToken)

		// then
		require.Error(t, e)
		require.Empty(t, cert)
	})

	t.Run("should return error when signing certificate with already used token", func(t *testing.T) {
		// given
		configuration := clients.GetConfiguration(t, directorClient, connectorClient, appID)

		certInfo := configuration.CertificateSigningRequestInfo

		csr := certs.CreateCsr(t, certInfo.Subject, clientKey)

		cert, err := connectorClient.SignCSR(certs.EncodeBase64(csr), configuration.Token.Token)
		require.NoError(t, err)
		certs.AssertCertificate(t, certInfo.Subject, cert)

		// when
		secondCert, err := connectorClient.SignCSR(certs.EncodeBase64(csr), configuration.Token.Token)

		//then
		require.Error(t, err)
		require.Empty(t, secondCert)
	})

	t.Run("should return error when invalid CSR provided for signing", func(t *testing.T) {
		// given
		configuration := clients.GetConfiguration(t, directorClient, connectorClient, appID)

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
	app, err := fixtures.RegisterApplicationFromInput(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, graphql.ApplicationRegisterInput{
		Name: "test-full-flow-app",
	})
	require.NoError(t, err)
	appID := app.ID
	defer fixtures.UnregisterApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, appID)

	t.Log("Generating certificate...")
	certificationResult, configuration := clients.GenerateApplicationCertificate(t, directorClient, connectorClient, appID, clientKey)
	certs.AssertCertificate(t, configuration.CertificateSigningRequestInfo.Subject, certificationResult)

	defer certs.Cleanup(t, configmapCleaner, certificationResult)

	t.Log("Certificate generated. Creating secured client...")
	certChain := certs.DecodeCertChain(t, certificationResult.CertificateChain)
	securedClient := clients.NewCertificateSecuredConnectorClient(*configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL, clientKey, certChain...)

	t.Log("Fetching configuration with certificate...")
	configWithCert, err := securedClient.Configuration()
	require.NoError(t, err)
	require.Equal(t, configuration.ManagementPlaneInfo, configWithCert.ManagementPlaneInfo)
	require.Equal(t, configuration.CertificateSigningRequestInfo, configWithCert.CertificateSigningRequestInfo)

	csr := certs.CreateCsr(t, configWithCert.CertificateSigningRequestInfo.Subject, clientKey)
	require.NoError(t, err)

	renewalResult, err := securedClient.SignCSR(certs.EncodeBase64(csr))
	require.NoError(t, err)
	certs.AssertCertificate(t, configWithCert.CertificateSigningRequestInfo.Subject, renewalResult)

	t.Log("Renewing certificate...")
	renewedCertChain := certs.DecodeCertChain(t, certificationResult.CertificateChain)
	securedClientWithRenewedCert := clients.NewCertificateSecuredConnectorClient(*configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL, clientKey, renewedCertChain...)

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
