package gateway_integration

import (
	"context"
	"encoding/base64"

	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/tests/director/pkg/gql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"
	"github.com/stretchr/testify/require"
)

func TestCompassAuth(t *testing.T) {

	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Register Integration System with Dex id token")
	intSys := registerIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, "integration-system")

	t.Log("Request Client Credentials for Integration System")
	intSysOauthCredentialData := requestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

	t.Log("Issue a Hydra token with Client Credentials")

	accessToken := getAccessToken(t, intSysOauthCredentialData, applicationScopes)

	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

	t.Log("Register an application as Integration System")
	appInput := graphql.ApplicationRegisterInput{
		Name:                "app-registered-by-integration-system",
		IntegrationSystemID: &intSys.ID,
	}
	appByIntSys := registerApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appInput)
	require.NotEmpty(t, appByIntSys.ID)

	t.Log("Add API Spec to Application")
	apiInput := graphql.APIDefinitionInput{
		Name:      "new-api-name",
		TargetURL: "https://kyma-project.io",
	}
	addAPIWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, apiInput, appByIntSys.ID)
	t.Log("Try removing Integration System")
	unregisterIntegrationSystemWithErr(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

	t.Log("Check if SystemAuths are still present in the db")
	auths := getSystemAuthsForIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)
	require.NotEmpty(t, auths)
	require.NotNil(t, auths[0])
	credentialDataFromDB, ok := auths[0].Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	assert.Equal(t, intSysOauthCredentialData, credentialDataFromDB)

	t.Log("Remove application to check if the oAuth token is still valid")
	unregisterApplication(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appByIntSys.ID)

	t.Log("Remove Integration System")
	unregisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

	t.Log("Check if token granted for Integration System is invalid")
	appByIntSys = registerApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appInput)
	require.Empty(t, appByIntSys.ID)

	t.Log("Check if token can not be fetched with old client credentials")
	oauthCredentials := fmt.Sprintf("%s:%s", intSysOauthCredentialData.ClientID, intSysOauthCredentialData.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	_, err = fetchHydraAccessToken(t, encodedCredentials, intSysOauthCredentialData.URL, applicationScopes)
	require.Error(t, err)
	assert.Equal(t, "response status code is 401", err.Error())
}
