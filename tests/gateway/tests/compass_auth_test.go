package tests

import (
	"context"
	"encoding/base64"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/tests/pkg/gql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/require"
)

func TestCompassAuth(t *testing.T) {
	ctx := context.Background()

	t.Log("Register Integration System via Certificate Secured Client")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	t.Log("Request Client Credentials for Integration System")
	intSystemAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, intSys.ID)

	intSysOauthCredentialData, ok := intSystemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")

	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)

	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

	t.Log("Register an application as Integration System")
	appInput := graphql.ApplicationRegisterInput{
		Name:                "app-registered-by-integration-system",
		IntegrationSystemID: &intSys.ID,
	}
	appByIntSys, err := fixtures.RegisterApplicationFromInput(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appInput)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &appByIntSys)
	require.NoError(t, err)
	require.NotEmpty(t, appByIntSys.ID)

	t.Log("Add Bundle with API Spec to Application")
	apiInput := graphql.APIDefinitionInput{
		Name:      "new-api-name",
		TargetURL: "https://kyma-project.io",
	}
	bndl := fixtures.CreateBundle(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appByIntSys.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, bndl.ID)
	fixtures.AddAPIToBundleWithInput(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, bndl.ID, apiInput)

	t.Log("Try removing Integration System")
	fixtures.UnregisterIntegrationSystemWithErr(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, intSys.ID)

	t.Log("Check if SystemAuths are still present in the db")
	auths := fixtures.GetSystemAuthsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, intSys.ID)
	require.NotEmpty(t, auths)
	require.NotNil(t, auths[0])
	credentialDataFromDB, ok := auths[0].Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	assert.Equal(t, intSysOauthCredentialData, credentialDataFromDB)

	t.Log("Remove application to check if the oAuth token is still valid")
	fixtures.UnregisterApplication(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appByIntSys.ID)

	t.Log("Remove Integration System")
	fixtures.UnregisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, intSys.ID)

	t.Log("Check if token granted for Integration System is invalid")
	appByIntSys, err = fixtures.RegisterApplicationFromInput(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appInput)
	require.NoError(t, err)

	require.Empty(t, appByIntSys.BaseEntity)

	t.Log("Check if token can not be fetched with old client credentials")
	oauthCredentials := fmt.Sprintf("%s:%s", intSysOauthCredentialData.ClientID, intSysOauthCredentialData.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	_, err = token.FetchHydraAccessToken(t, encodedCredentials, intSysOauthCredentialData.URL, token.ApplicationScopes)
	require.Error(t, err)
	assert.Equal(t, "response status code is 401", err.Error())
}
