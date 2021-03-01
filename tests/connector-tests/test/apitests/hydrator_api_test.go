package apitests

import (
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"
	director "github.com/kyma-incubator/compass/tests/director/gateway-integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHydrators(t *testing.T) {
	runtime := director.RegisterRuntimeFromInputWithinTenant(t, ctx, directorClient.DexGraphqlClient, config.Tenant, &graphql.RuntimeInput{
		Name: "test-hydrators-runtime",
	})
	runtimeID := runtime.ID
	defer director.UnregisterRuntimeWithinTenant(t, ctx, directorClient.DexGraphqlClient, config.Tenant, runtimeID)

	app, err := director.RegisterApplicationWithinTenant(t, ctx, directorClient.DexGraphqlClient, config.Tenant, graphql.ApplicationRegisterInput{
		Name: "test-hydrators-app",
	})
	require.NoError(t, err)
	appID := app.ID
	defer director.UnregisterApplication(t, ctx, directorClient.DexGraphqlClient, config.Tenant, appID)

	hash := "df6ab69b34100a1808ddc6211010fa289518f14606d0c8eaa03a0f53ecba578a"

	for _, testCase := range []struct {
		clientType           string
		clientId             string
		tokenGenerationFunc  func(t *testing.T, id string) (externalschema.Token, error)
		expectedCertsHeaders http.Header
	}{
		{
			clientType:          "Application",
			clientId:            appID,
			tokenGenerationFunc: directorClient.GenerateApplicationToken,

			expectedCertsHeaders: http.Header{
				oathkeeper.ClientCertificateHashHeader: []string{hash},
			},
		},
		{
			clientType:          "Runtime",
			clientId:            runtimeID,
			tokenGenerationFunc: directorClient.GenerateRuntimeToken,
			expectedCertsHeaders: http.Header{
				oathkeeper.ClientCertificateHashHeader: []string{hash},
			},
		},
	} {
		t.Run("should resolve one-time token for "+testCase.clientType, func(t *testing.T) {
			//given
			token, err := testCase.tokenGenerationFunc(t, testCase.clientId)
			require.NoError(t, err)
			require.NotEmpty(t, token.Token)

			headers := map[string][]string{
				oathkeeper.ConnectorTokenHeader: {token.Token},
			}

			//when
			authSession := directorHydratorClient.ResolveToken(t, headers)

			var auths []*graphql.SystemAuth
			if testCase.clientType == "Application" {
				auths = director.GetApplication(t, ctx, directorClient.DexGraphqlClient, config.Tenant, testCase.clientId).Auths
			} else {
				auths = director.GetRuntime(t, ctx, directorClient.DexGraphqlClient, config.Tenant, testCase.clientId).Auths
			}

			hasAuth := false
			for _, auth := range auths {
				if auth.ID == authSession.Header.Get(oathkeeper.ClientIdFromTokenHeader) {
					hasAuth = true
					break
				}
			}

			//then
			assert.Equal(t, true, hasAuth)
		})

		t.Run("should resolve certificate for "+testCase.clientType, func(t *testing.T) {
			//given
			token, err := testCase.tokenGenerationFunc(t, testCase.clientId)
			require.NoError(t, err)
			require.NotEmpty(t, token.Token)

			configuration, err := connectorClient.Configuration(token.Token)
			require.NoError(t, err)

			certDataHeader := connector.CreateCertDataHeader(configuration.CertificateSigningRequestInfo.Subject, hash)

			headers := map[string][]string{
				config.CertificateDataHeader: {certDataHeader},
			}

			//when
			authSession := connectorHydratorClient.ResolveCertificateData(t, headers)

			//then
			var auths []*graphql.SystemAuth
			if testCase.clientType == "Application" {
				auths = director.GetApplication(t, ctx, directorClient.DexGraphqlClient, config.Tenant, testCase.clientId).Auths
			} else {
				auths = director.GetRuntime(t, ctx, directorClient.DexGraphqlClient, config.Tenant, testCase.clientId).Auths
			}

			hasAuth := false
			for _, auth := range auths {
				if auth.ID == authSession.Header.Get(oathkeeper.ClientIdFromCertificateHeader) {
					hasAuth = true
					break
				}
			}

			assert.Equal(t, true, hasAuth)
			assert.Equal(t, testCase.expectedCertsHeaders.Get(oathkeeper.ClientCertificateHashHeader), authSession.Header.Get(oathkeeper.ClientCertificateHashHeader))
		})

		t.Run("should return empty Authentication Session when no valid headers found", func(t *testing.T) {
			//given
			token, err := testCase.tokenGenerationFunc(t, testCase.clientId)
			require.NoError(t, err)
			require.NotEmpty(t, token.Token)

			//when
			authSession := directorHydratorClient.ResolveToken(t, nil)

			//then
			assert.Empty(t, authSession)
		})
	}

}
