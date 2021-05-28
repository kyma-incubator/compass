package tests

import (
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHydrators(t *testing.T) {
	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, &graphql.RuntimeInput{
		Name: "test-hydrators-runtime",
	})
	runtimeID := runtime.ID
	defer fixtures.UnregisterRuntime(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, runtimeID)

	app, err := fixtures.RegisterApplicationFromInput(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, graphql.ApplicationRegisterInput{
		Name: "test-hydrators-app",
	})
	require.NoError(t, err)
	appID := app.ID
	defer fixtures.UnregisterApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, appID)

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

			var appSystemAuths []*graphql.AppSystemAuth
			var runtimeSystemAuths []*graphql.RuntimeSystemAuth

			if testCase.clientType == "Application" {
				appSystemAuths = fixtures.GetApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, testCase.clientId).Auths
			} else {
				runtimeSystemAuths = fixtures.GetRuntime(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, testCase.clientId).Auths
			}

			hasAuth := false
			for _, auth := range appSystemAuths {
				if auth.ID == authSession.Header.Get(oathkeeper.ClientIdFromTokenHeader) {
					hasAuth = true
					break
				}
			}

			for _, auth := range runtimeSystemAuths {
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

			certDataHeader := certs.CreateCertDataHeader(configuration.CertificateSigningRequestInfo.Subject, hash)

			headers := map[string][]string{
				cfg.CertificateDataHeader: {certDataHeader},
			}

			//when
			authSession := connectorHydratorClient.ResolveCertificateData(t, headers)

			var appSystemAuths []*graphql.AppSystemAuth
			var runtimeSystemAuths []*graphql.RuntimeSystemAuth

			if testCase.clientType == "Application" {
				appSystemAuths = fixtures.GetApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, testCase.clientId).Auths
			} else {
				runtimeSystemAuths = fixtures.GetRuntime(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, testCase.clientId).Auths
			}

			hasAuth := false
			for _, auth := range appSystemAuths {
				if auth.ID == authSession.Header.Get(oathkeeper.ClientIdFromCertificateHeader) {
					hasAuth = true
					break
				}
			}

			for _, auth := range runtimeSystemAuths {
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
