package e2e

import (
	"context"
	"encoding/base64"
	"os"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/idtokenprovider"

	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/gql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/idtokenprovider"
	"github.com/stretchr/testify/require"
)

func TestCompassAuth(t *testing.T) {
	domain := os.Getenv("DOMAIN")
	require.NotEmpty(t, domain)
	tenant := os.Getenv("DEFAULT_TENANT")
	require.NotEmpty(t, tenant)
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexTokenFromEnv()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Register Integration System with Dex id token")
	intSys := registerIntegrationSystem(t, ctx, dexGraphQLClient, tenant, "integration-system")

	t.Log("Request Client Credentials for Integration System")
	intSysOauthCredentialData := requestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)

	t.Log("Issue a Hydra token with Client Credentials")

	accessToken := getAccessToken(t, intSysOauthCredentialData, applicationScopes)

	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", domain))

	t.Log("Register an application as Integration System")
	appInput := graphql.ApplicationRegisterInput{
		Name:                "app-registered-by-integration-system",
		ProviderName:        "compass",
		IntegrationSystemID: &intSys.ID,
	}
	appByIntSys := registerApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, tenant, appInput)
	require.NotEmpty(t, appByIntSys.ID)

	t.Log("Add API Spec to Application")
	apiInput := graphql.APIDefinitionInput{
		Name:      "new-api-name",
		TargetURL: "https://kyma-project.io",
	}
	addAPIWithinTenant(t, ctx, oauthGraphQLClient, tenant, apiInput, appByIntSys.ID)
	t.Log("Try removing Integration System")
	unregisterIntegrationSystemWithErr(t, ctx, dexGraphQLClient, tenant, intSys.ID)

	t.Log("Check if SystemAuths are still present in the db")
	auths := getSystemAuthsForIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)
	require.NotEmpty(t, auths)
	require.NotNil(t, auths[0])
	credentialDataFromDB, ok := auths[0].Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	assert.Equal(t, intSysOauthCredentialData, credentialDataFromDB)

	t.Log("Remove application to check if the oAuth token is still valid")
	deleteApplication(t, ctx, oauthGraphQLClient, tenant, appByIntSys.ID)

	t.Log("Remove Integration System")
	unregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)

	t.Log("Check if token granted for Integration System is invalid")
	appByIntSys = registerApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, tenant, appInput)
	require.Empty(t, appByIntSys.ID)

	t.Log("Check if token can not be fetched with old client credentials")
	oauthCredentials := fmt.Sprintf("%s:%s", intSysOauthCredentialData.ClientID, intSysOauthCredentialData.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	_, err = fetchHydraAccessToken(t, encodedCredentials, intSysOauthCredentialData.URL, applicationScopes)
	require.Error(t, err)
	assert.Equal(t, "response status code is 401", err.Error())
}
