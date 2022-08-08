package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHydrators(t *testing.T) {
	app, err := fixtures.RegisterApplicationFromInput(t, ctx, directorClient.CertSecuredGraphqlClient, cfg.Tenant, graphql.ApplicationRegisterInput{
		Name: "test-hydrators-app",
	})
	defer fixtures.CleanupApplication(t, ctx, directorClient.CertSecuredGraphqlClient, cfg.Tenant, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)
	appID := app.ID

	input := fixtures.FixRuntimeRegisterInput("test-hydrators-runtime")
	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorClient.CertSecuredGraphqlClient, cfg.Tenant, &input)
	defer fixtures.CleanupRuntime(t, ctx, directorClient.CertSecuredGraphqlClient, cfg.Tenant, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)
	runtimeID := runtime.ID

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

			var appSystemAuths []*graphql.AppSystemAuth
			var runtimeSystemAuths []*graphql.RuntimeSystemAuth

			if testCase.clientType == "Application" {
				appSystemAuths = fixtures.GetApplication(t, ctx, directorClient.CertSecuredGraphqlClient, cfg.Tenant, testCase.clientId).Auths
			} else {
				runtimeSystemAuths = fixtures.GetRuntime(t, ctx, directorClient.CertSecuredGraphqlClient, cfg.Tenant, testCase.clientId).Auths
			}

			//when
			authSession := hydratorClient.ResolveToken(t, headers)

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

			// check that gql resolver is sending used/expired auths, but there is valid property
			if testCase.clientType == "Application" {
				appSystemAuths = fixtures.GetApplication(t, ctx, directorClient.CertSecuredGraphqlClient, cfg.Tenant, appID).Auths
				assert.Len(t, appSystemAuths, 1)
				assert.True(t, appSystemAuths[0].Auth.OneTimeToken.(*graphql.OneTimeTokenForApplication).Used)
				assert.True(t, time.Time(*appSystemAuths[0].Auth.OneTimeToken.(*graphql.OneTimeTokenForApplication).ExpiresAt).After(time.Now()))

			}
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
			authSession := hydratorClient.ResolveCertificateData(t, headers)

			var appSystemAuths []*graphql.AppSystemAuth
			var runtimeSystemAuths []*graphql.RuntimeSystemAuth

			if testCase.clientType == "Application" {
				appSystemAuths = fixtures.GetApplication(t, ctx, directorClient.CertSecuredGraphqlClient, cfg.Tenant, testCase.clientId).Auths
			} else {
				runtimeSystemAuths = fixtures.GetRuntime(t, ctx, directorClient.CertSecuredGraphqlClient, cfg.Tenant, testCase.clientId).Auths
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
			authSession := hydratorClient.ResolveToken(t, nil)

			//then
			assert.Empty(t, authSession)
		})
	}
}
