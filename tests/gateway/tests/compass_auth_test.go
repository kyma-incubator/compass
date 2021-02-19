package tests

import (
	"context"
	"encoding/base64"
	"github.com/kyma-incubator/compass/tests/pkg"

	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/tests/pkg/gql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/stretchr/testify/require"
)

func TestCompassAuth(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Register Integration System with Dex id token")
	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, "integration-system")

	t.Log("Request Client Credentials for Integration System")
	intSystemAuth := pkg.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

	intSysOauthCredentialData, ok:= intSystemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")

	accessToken := pkg.GetAccessToken(t, intSysOauthCredentialData, pkg.ApplicationScopes)

	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

	t.Log("Register an application as Integration System")
	appInput := graphql.ApplicationRegisterInput{
		Name:                "app-registered-by-integration-system",
		IntegrationSystemID: &intSys.ID,
	}
	appByIntSys := pkg.RegisterApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appInput)
	require.NotEmpty(t, appByIntSys.ID)

	t.Log("Add Bundle with API Spec to Application")
	apiInput := graphql.APIDefinitionInput{
		Name:      "new-api-name",
		TargetURL: "https://kyma-project.io",
	}
	bndl := pkg.CreateBundle(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appByIntSys.ID, "bndl")
	defer pkg.DeleteBundle(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, bndl.ID)
	pkg.AddAPIToBundleWithInput(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, bndl.ID, apiInput)

	t.Log("Try removing Integration System")
	pkg.UnregisterIntegrationSystemWithErr(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

	t.Log("Check if SystemAuths are still present in the db")
	auths := pkg.GetSystemAuthsForIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)
	require.NotEmpty(t, auths)
	require.NotNil(t, auths[0])
	credentialDataFromDB, ok := auths[0].Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	assert.Equal(t, intSysOauthCredentialData, credentialDataFromDB)

	t.Log("Remove application to check if the oAuth token is still valid")
	pkg.UnregisterApplication(t, ctx, oauthGraphQLClient, appByIntSys.ID, testConfig.DefaultTenant)

	t.Log("Remove Integration System")
	pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

	t.Log("Check if token granted for Integration System is invalid")
	appByIntSys = pkg.RegisterApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appInput)
	require.Empty(t, appByIntSys.ID)

	t.Log("Check if token can not be fetched with old client credentials")
	oauthCredentials := fmt.Sprintf("%s:%s", intSysOauthCredentialData.ClientID, intSysOauthCredentialData.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	_, err = pkg.FetchHydraAccessToken(t, encodedCredentials, intSysOauthCredentialData.URL, pkg.ApplicationScopes)
	require.Error(t, err)
	assert.Equal(t, "response status code is 401", err.Error())
}
