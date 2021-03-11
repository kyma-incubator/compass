package apitests

import (
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"
	director "github.com/kyma-incubator/compass/tests/director/gateway-integration"
	"github.com/stretchr/testify/require"
)

func TestOathkeeperSecurity(t *testing.T) {
	app, err := director.RegisterApplicationWithinTenant(t, ctx, directorClient.DexGraphqlClient, config.Tenant, graphql.ApplicationRegisterInput{
		Name: "test-oathkeeper-security-app",
	})
	require.NoError(t, err)
	appID := app.ID
	defer director.UnregisterApplication(t, ctx, directorClient.DexGraphqlClient, config.Tenant, appID)

	certResult, configuration := connector.GenerateApplicationCertificate(t, directorClient, connectorClient, appID, clientKey)
	certChain := testkit.DecodeCertChain(t, certResult.CertificateChain)
	securedClient := connector.NewCertificateSecuredConnectorClient(*configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL, clientKey, certChain...)

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

		newSubject := connector.ChangeCommonName(configuration.CertificateSigningRequestInfo.Subject, changedAppID)
		certDataHeader := connector.CreateCertDataHeader("df6ab69b34100a1808ddc6211010fa289518f14606d0c8eaa03a0f53ecba578a", newSubject)

		forbiddenHeaders := map[string][]string{
			config.CertificateDataHeader: {certDataHeader},
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
