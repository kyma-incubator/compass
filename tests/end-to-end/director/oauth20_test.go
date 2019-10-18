//+build !ignore_external_dependencies

package director

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateClientCredentialsToIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	t.Log("Create integration system")
	intSys := createIntegrationSystem(t, ctx, name)
	require.NotEmpty(t, intSys)
	defer deleteIntegrationSystem(t, ctx, intSys.ID)

	generateIntSysAuthRequest := fixGenerateClientCredentialsForIntegrationSystem(intSys.ID)
	intSysAuth := graphql.SystemAuth{}

	// WHEN
	t.Log("Generate client credentials for integration system")
	err := tc.RunOperation(ctx, generateIntSysAuthRequest, &intSysAuth)
	require.NoError(t, err)
	require.NotEmpty(t, intSysAuth.Auth)
	defer deleteSystemAuthForIntegrationSystem(t, ctx, intSysAuth.ID)

	//THEN
	t.Log("Check if client credentials were generated")
	assert.NotEmpty(t, intSysAuth.Auth.Credential)
	oauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	assert.NotEmpty(t, oauthCredentialData.ClientID)
	assert.NotEmpty(t, oauthCredentialData.ClientSecret)
	assert.Equal(t, intSysAuth.ID, oauthCredentialData.ClientID)
}

func TestGenerateClientCredentialsToApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app"

	t.Log("Create application")
	app := createApplication(t, ctx, name)
	require.NotEmpty(t, app)
	defer deleteApplication(t, app.ID)

	generateApplicationClientCredentialsRequest := fixGenerateClientCredentialsForApplication(app.ID)
	appAuth := graphql.SystemAuth{}

	// WHEN
	t.Log("Generate client credentials for application")
	err := tc.RunOperation(ctx, generateApplicationClientCredentialsRequest, &appAuth)
	require.NoError(t, err)

	//THEN
	t.Log("Get updated application")
	app = getApplication(t, ctx, app.ID)
	require.NotEmpty(t, app.Auths)
	defer deleteSystemAuthForApplication(t, ctx, app.Auths[0].ID)

	t.Log("Check if client credentials were generated")
	assert.NotEmpty(t, app.Auths[0].Auth.Credential)
	oauthCredentialData, ok := app.Auths[0].Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	assert.NotEmpty(t, oauthCredentialData.ClientID)
	assert.NotEmpty(t, oauthCredentialData.ClientSecret)
	assert.Equal(t, appAuth.ID, oauthCredentialData.ClientID)
}

func TestGenerateClientCredentialsToRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "runtime"

	t.Log("Create runtime")
	rtm := createRuntime(t, ctx, name)
	require.NotEmpty(t, rtm)
	defer deleteRuntime(t, rtm.ID)

	generateRuntimeClientCredentialsRequest := fixGenerateClientCredentialsForRuntime(rtm.ID)
	rtmAuth := graphql.SystemAuth{}

	// WHEN
	t.Log("Generate client credentials for runtime")
	err := tc.RunOperation(ctx, generateRuntimeClientCredentialsRequest, &rtmAuth)
	require.NoError(t, err)

	//THEN
	t.Log("Get updated runtime")
	rtm = getRuntime(t, ctx, rtm.ID)
	require.NotEmpty(t, rtm.Auths)
	defer deleteSystemAuthForRuntime(t, ctx, rtm.Auths[0].ID)

	t.Log("Check if client credentials were generated")
	assert.NotEmpty(t, rtm.Auths[0].Auth)
	oauthCredentialData, ok := rtm.Auths[0].Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	assert.NotEmpty(t, oauthCredentialData.ClientID)
	assert.NotEmpty(t, oauthCredentialData.ClientSecret)
	assert.Equal(t, rtmAuth.ID, oauthCredentialData.ClientID)
}

func TestDeleteSystemAuthFromApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app"

	t.Log("Create application")
	app := createApplication(t, ctx, name)
	require.NotEmpty(t, app)
	defer deleteApplication(t, app.ID)

	appAuth := fixGenerateClientCredentialsForApplication(app.ID)
	require.NotEmpty(t, appAuth)

	deleteSystemAuthForApplicationRequest := fixDeleteSystemAuthForApplication(app.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for application")
	err := tc.RunOperation(ctx, deleteSystemAuthForApplicationRequest, &deleteOutput)
	require.NoError(t, err)
	require.NotEmpty(t, deleteOutput)

	//THEN
	t.Log("Check if system auth was deleted")
	appAfterDelete := getApp(ctx, t, app.ID)
	require.Empty(t, appAfterDelete.Auths)
}

func TestDeleteSystemAuthFromRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "rtm"

	t.Log("Create runtime")
	rtm := createRuntime(t, ctx, name)
	require.NotEmpty(t, rtm)
	defer deleteRuntime(t, rtm.ID)

	rtmAuth := fixGenerateClientCredentialsForRuntime(rtm.ID)
	require.NotEmpty(t, rtmAuth)

	deleteSystemAuthForRuntimeRequest := fixDeleteSystemAuthForRuntime(rtm.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for runtime")
	err := tc.RunOperation(ctx, deleteSystemAuthForRuntimeRequest, &deleteOutput)
	require.NoError(t, err)
	require.NotEmpty(t, deleteOutput)

	//THEN
	t.Log("Check if system auth was deleted")
	rtmAfterDelete := getRuntime(t, ctx, rtm.ID)
	require.Empty(t, rtmAfterDelete.Auths)
}

func TestDeleteSystemAuthFromIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	t.Log("Create integration system")
	intSys := createIntegrationSystem(t, ctx, name)
	require.NotEmpty(t, intSys)
	defer deleteIntegrationSystem(t, ctx, intSys.ID)

	intSysAuth := generateClientCredentialsForIntegrationSystem(t, ctx, intSys.ID)
	require.NotEmpty(t, intSysAuth)

	deleteSystemAuthForIntegrationSystemRequest := fixDeleteSystemAuthForIntegrationSystem(intSysAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for integration system")
	err := tc.RunOperation(ctx, deleteSystemAuthForIntegrationSystemRequest, &deleteOutput)
	require.NoError(t, err)
	require.NotEmpty(t, deleteOutput)

	//THEN
	t.Log("Check if system auth was deleted")
	intSysAfterDelete := getIntegrationSystem(t, ctx, intSys.ID)
	require.Empty(t, intSysAfterDelete.Auths)
}
