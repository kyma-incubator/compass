package e2e

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/idtokenprovider"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/gql"
	"github.com/stretchr/testify/require"
)

const integrationSystemScopes = "application:read application:write application_template:read application_template:write runtime:read runtime:write"

func TestIntegrationSystemScenario(t *testing.T) {
	domain := os.Getenv("DOMAIN")
	require.NotEmpty(t, domain)
	tenant := os.Getenv("DEFAULT_TENANT")
	require.NotEmpty(t, tenant)
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexTokenFromEnv()
	require.NoError(t,err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Request Integration System with Dex id token")
	intSys := registerIntegrationSystem(t, ctx, dexGraphQLClient, tenant, "integration-system")

	t.Log("Generate Client Credentials for Integration System")
	intSysOauthCredentialData := generateOAuthClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)


	t.Log("Issue a token with Client Credentials")
	oauthCredentials := fmt.Sprintf("%s:%s", intSysOauthCredentialData.ClientID, intSysOauthCredentialData.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	token, err := getAccessToken(t, encodedCredentials, intSysOauthCredentialData.URL, integrationSystemScopes)
	require.NoError(t, err)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(token.AccessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", domain))

	t.Log("Register an application")
	appInput := graphql.ApplicationRegisterInput{
		Name:                "app",
		ProviderName:        "compass",
		IntegrationSystemID: &intSys.ID,
	}
	appByIntSys := registerApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, tenant, appInput)
	require.NotEmpty(t, appByIntSys.ID)

	t.Log("Get application")
	app := getApplication(t, ctx, oauthGraphQLClient, tenant, appByIntSys.ID)
	require.NotEmpty(t, app.ID)
	require.Equal(t, appByIntSys.ID, app.ID)

	t.Log("Remove application")
	deleteApplication(t, ctx, oauthGraphQLClient, tenant, appByIntSys.ID)

	t.Log("Create an application template")
	appTplInput := graphql.ApplicationTemplateInput{
		Name:        "test",
		Description: nil,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:         "test",
			ProviderName: "test",
		},
		Placeholders: nil,
		AccessLevel:  "GLOBAL",
	}
	appTpl := createApplicationTemplate(t, ctx, oauthGraphQLClient, tenant, appTplInput)
	require.NotEmpty(t, appTpl.ID)

	t.Log("Get application template")
	gqlAppTpl := getApplicationTemplate(t, ctx, oauthGraphQLClient, tenant, appTpl.ID)
	require.NotEmpty(t, gqlAppTpl.ID)
	require.Equal(t, appTpl.ID, gqlAppTpl.ID)

	t.Log("Remove application template")
	deleteApplicationTemplate(t, ctx, oauthGraphQLClient, tenant, appTpl.ID)

	t.Log("Create runtime")
	runtimeInput := graphql.RuntimeInput{
		Name: "test",
	}
	runtime := registerRuntimeFromInputWithinTenant(t, ctx, oauthGraphQLClient, tenant, runtimeInput)
	require.NotEmpty(t, runtime.ID)

	t.Log("Get runtime")
	gqlRuntime := getRuntime(t, ctx, oauthGraphQLClient, tenant, runtime.ID)
	require.NotEmpty(t, gqlRuntime.ID)
	require.Equal(t, runtime.ID, gqlRuntime.ID)

	t.Log("Remove runtime")
	unregisterRuntime(t, ctx, oauthGraphQLClient, tenant, runtime.ID)

	t.Log("Remove Integration System")
	unregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)
}
