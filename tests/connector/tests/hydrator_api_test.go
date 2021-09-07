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
	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, &graphql.RuntimeInput{
		Name: "test-hydrators-runtime",
	})
	defer fixtures.CleanupRuntime(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)
	runtimeID := runtime.ID

	hash := "df6ab69b34100a1808ddc6211010fa289518f14606d0c8eaa03a0f53ecba578a"

	for _, testCase := range []struct {
		clientType           string
		tokenGenerationFunc  func(t *testing.T, id string) (externalschema.Token, error)
		expectedCertsHeaders http.Header
	}{
		{
			clientType:          "Application",
			tokenGenerationFunc: directorClient.GenerateApplicationToken,
			expectedCertsHeaders: http.Header{
				oathkeeper.ClientCertificateHashHeader: []string{hash},
			},
		},
		{
			clientType:          "Runtime",
			tokenGenerationFunc: directorClient.GenerateRuntimeToken,
			expectedCertsHeaders: http.Header{
				oathkeeper.ClientCertificateHashHeader: []string{hash},
			},
		},
	} {
		t.Run("should resolve one-time token for "+testCase.clientType, func(t *testing.T) {
			//given
			var err error
			app, err := fixtures.RegisterApplicationFromInput(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, graphql.ApplicationRegisterInput{
				Name: "test-hydrators-app",
			})
			defer fixtures.CleanupApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, &app)
			require.NoError(t, err)
			require.NotEmpty(t, app.ID)

			var token externalschema.Token

			if testCase.clientType == "Application" {
				token, err = testCase.tokenGenerationFunc(t, app.ID)
			} else {
				token, err = testCase.tokenGenerationFunc(t, runtimeID)
			}
			require.NoError(t, err)
			require.NotEmpty(t, token.Token)

			headers := map[string][]string{
				oathkeeper.ConnectorTokenHeader: {token.Token},
			}

			var appSystemAuths []*graphql.AppSystemAuth
			var runtimeSystemAuths []*graphql.RuntimeSystemAuth

			if testCase.clientType == "Application" {
				appSystemAuths = fixtures.GetApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, app.ID).Auths
			} else {
				runtimeSystemAuths = fixtures.GetRuntime(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, runtimeID).Auths
			}

			//when
			authSession := directorHydratorClient.ResolveToken(t, headers)

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

			assert.Equal(t, true, hasAuth)

			// check if the gql resolver is not sending used/expired auths
			if testCase.clientType == "Application" {
				appSystemAuths = fixtures.GetApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, app.ID).Auths
				assert.Len(t, appSystemAuths, 0)
			}
		})

		t.Run("should resolve certificate for "+testCase.clientType, func(t *testing.T) {
			//given
			var err error
			var token externalschema.Token

			app, err := fixtures.RegisterApplicationFromInput(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, graphql.ApplicationRegisterInput{
				Name: "test-hydrators-app",
			})
			defer fixtures.CleanupApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, &app)
			require.NoError(t, err)
			require.NotEmpty(t, app.ID)

			if testCase.clientType == "Application" {
				token, err = testCase.tokenGenerationFunc(t, app.ID)
			} else {
				token, err = testCase.tokenGenerationFunc(t, runtimeID)
			}

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
				appSystemAuths = fixtures.GetApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, app.ID).Auths
			} else {
				runtimeSystemAuths = fixtures.GetRuntime(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, runtimeID).Auths
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
			var err error

			var token externalschema.Token
			if testCase.clientType == "Application" {
				app, err := fixtures.RegisterApplicationFromInput(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, graphql.ApplicationRegisterInput{
					Name: "test-hydrators-app",
				})
				defer fixtures.CleanupApplication(t, ctx, directorClient.DexGraphqlClient, cfg.Tenant, &app)
				require.NoError(t, err)
				require.NotEmpty(t, app.ID)

				token, err = testCase.tokenGenerationFunc(t, app.ID)
			} else {
				token, err = testCase.tokenGenerationFunc(t, runtimeID)
			}

			require.NoError(t, err)
			require.NotEmpty(t, token.Token)

			//when
			authSession := directorHydratorClient.ResolveToken(t, nil)

			//then
			assert.Empty(t, authSession)
		})
	}
}
