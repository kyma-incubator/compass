package tests

import (
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/stretchr/testify/require"
)

func TestOathkeeperSecurity(t *testing.T) {
	app, err := fixtures.RegisterApplicationFromInput(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, graphql.ApplicationRegisterInput{
		Name: "test-oathkeeper-security-app",
	})
	require.NoError(t, err)
	appID := app.ID
	defer fixtures.UnregisterApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, appID)

	certResult, configuration := clients.GenerateApplicationCertificate(t, directorClient, connectorClient, appID, clientKey)
	certChain := certs.DecodeCertChain(t, certResult.CertificateChain)
	securedClient := clients.NewCertificateSecuredConnectorClient(*configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL, clientKey, certChain...)

	t.Run("client id headers should be stripped when calling token-secured api", func(t *testing.T) {
		// given
		forbiddenHeaders := map[string][]string{
			oathkeeper.ClientIdFromTokenHeader:       {appID},
			oathkeeper.ClientIdFromCertificateHeader: {appID},
		}

		csr := certs.CreateCsr(t, configuration.CertificateSigningRequestInfo.Subject, clientKey)

		// when
		_, err := connectorClient.Configuration("", forbiddenHeaders)

		// then
		require.Error(t, err)

		// when
		_, err = connectorClient.SignCSR(certs.EncodeBase64(csr), "", forbiddenHeaders)

		// then
		require.Error(t, err)
	})

	t.Run("certificate data header should be stripped", func(t *testing.T) {
		// given
		changedAppID := "aaabbbcc-b340-418d-b653-d95b5e347d74"

		newSubject := certs.ChangeCommonName(configuration.CertificateSigningRequestInfo.Subject, changedAppID)
		certDataHeader := certs.CreateCertDataHeader("df6ab69b34100a1808ddc6211010fa289518f14606d0c8eaa03a0f53ecba578a", newSubject)

		forbiddenHeaders := map[string][]string{
			cfg.CertificateDataHeader: {certDataHeader},
		}

		csr := certs.CreateCsr(t, newSubject, clientKey)

		t.Run("when calling token-secured API", func(t *testing.T) {
			// when
			_, err := connectorClient.Configuration("", forbiddenHeaders)

			// then
			require.Error(t, err)

			// when
			_, err = connectorClient.SignCSR(certs.EncodeBase64(csr), "", forbiddenHeaders)

			// then
			require.Error(t, err)
		})

		t.Run("when calling certificate-secured API", func(t *testing.T) {
			// when
			cfg, err := securedClient.Configuration(forbiddenHeaders)

			// then
			require.NoError(t, err)
			require.Equal(t, configuration.CertificateSigningRequestInfo.Subject, cfg.CertificateSigningRequestInfo.Subject)

			// when
			_, err = securedClient.SignCSR(certs.EncodeBase64(csr), forbiddenHeaders)

			// then
			require.Error(t, err)
		})
	})

}
