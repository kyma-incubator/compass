//+build !ignore_external_dependencies

package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateClientCredentialsToApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app"

	t.Log("Create application")
	app := registerApplication(t, ctx, name)
	require.NotEmpty(t, app)
	defer unregisterApplication(t, app.ID)

	generateApplicationClientCredentialsRequest := fixRequestClientCredentialsForApplication(app.ID)
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
	rtm := registerRuntime(t, ctx, name)
	require.NotEmpty(t, rtm)
	defer unregisterRuntime(t, rtm.ID)

	generateRuntimeClientCredentialsRequest := fixRequestClientCredentialsForRuntime(rtm.ID)
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

func TestGenerateClientCredentialsToIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	t.Log("Create integration system")
	intSys := registerIntegrationSystem(t, ctx, name)
	require.NotEmpty(t, intSys)
	defer unregisterIntegrationSystem(t, ctx, intSys.ID)

	generateIntSysAuthRequest := fixRequestClientCredentialsForIntegrationSystem(intSys.ID)
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

func TestDeleteSystemAuthFromApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app"

	t.Log("Create application")
	app := registerApplication(t, ctx, name)
	require.NotEmpty(t, app)
	defer unregisterApplication(t, app.ID)

	appAuth := generateClientCredentialsForApplication(t, ctx, app.ID)
	require.NotEmpty(t, appAuth)

	deleteSystemAuthForApplicationRequest := fixDeleteSystemAuthForApplicationRequest(appAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for application")
	err := tc.RunOperation(ctx, deleteSystemAuthForApplicationRequest, &deleteOutput)
	require.NoError(t, err)
	require.NotEmpty(t, deleteOutput)

	//THEN
	t.Log("Check if system auth was deleted")
	appAfterDelete := getApplication(t, ctx, app.ID)
	require.Empty(t, appAfterDelete.Auths)
}

func TestDeleteSystemAuthFromApplicationUsingRuntimeMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app"

	t.Log("Create application")
	app := registerApplication(t, ctx, name)
	require.NotEmpty(t, app)
	defer unregisterApplication(t, app.ID)

	appAuth := generateClientCredentialsForApplication(t, ctx, app.ID)
	require.NotEmpty(t, appAuth)
	defer deleteClientCredentialsForApplication(t, appAuth.ID)

	deleteSystemAuthForRuntimeRequest := fixDeleteSystemAuthForRuntimeRequest(appAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for application using runtime mutation")
	err := tc.RunOperation(ctx, deleteSystemAuthForRuntimeRequest, &deleteOutput)

	// THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromApplicationUsingIntegrationSystemMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app"

	t.Log("Create application")
	app := registerApplication(t, ctx, name)
	require.NotEmpty(t, app)
	defer unregisterApplication(t, app.ID)

	appAuth := generateClientCredentialsForApplication(t, ctx, app.ID)
	require.NotEmpty(t, appAuth)
	defer deleteClientCredentialsForApplication(t, appAuth.ID)

	deleteSystemAuthForIntegrationSystemRequest := fixDeleteSystemAuthForIntegrationSystemRequest(appAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for application using runtime mutation")
	err := tc.RunOperation(ctx, deleteSystemAuthForIntegrationSystemRequest, &deleteOutput)

	// THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "rtm"

	t.Log("Create runtime")
	rtm := registerRuntime(t, ctx, name)
	require.NotEmpty(t, rtm)
	defer unregisterRuntime(t, rtm.ID)

	rtmAuth := generateClientCredentialsForRuntime(t, ctx, rtm.ID)
	require.NotEmpty(t, rtmAuth)

	deleteSystemAuthForRuntimeRequest := fixDeleteSystemAuthForRuntimeRequest(rtmAuth.ID)
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

func TestDeleteSystemAuthFromRuntimeUsingApplicationMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "rtm"

	t.Log("Create runtime")
	rtm := registerRuntime(t, ctx, name)
	require.NotEmpty(t, rtm)
	defer unregisterRuntime(t, rtm.ID)

	rtmAuth := generateClientCredentialsForRuntime(t, ctx, rtm.ID)
	require.NotEmpty(t, rtmAuth)
	defer deleteClientCredentialsForRuntime(t, rtmAuth.ID)

	deleteSystemAuthForApplicationRequest := fixDeleteSystemAuthForApplicationRequest(rtmAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for runtime using application mutation")
	err := tc.RunOperation(ctx, deleteSystemAuthForApplicationRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromRuntimeUsingIntegrationSystemMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "rtm"

	t.Log("Create runtime")
	rtm := registerRuntime(t, ctx, name)
	require.NotEmpty(t, rtm)
	defer unregisterRuntime(t, rtm.ID)

	rtmAuth := generateClientCredentialsForRuntime(t, ctx, rtm.ID)
	require.NotEmpty(t, rtmAuth)
	defer deleteClientCredentialsForRuntime(t, rtmAuth.ID)

	deleteSystemAuthForIntegrationSystemRequest := fixDeleteSystemAuthForIntegrationSystemRequest(rtmAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for runtime using integration system mutation")
	err := tc.RunOperation(ctx, deleteSystemAuthForIntegrationSystemRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	t.Log("Create integration system")
	intSys := registerIntegrationSystem(t, ctx, name)
	require.NotEmpty(t, intSys)
	defer unregisterIntegrationSystem(t, ctx, intSys.ID)

	intSysAuth := requestClientCredentialsForIntegrationSystem(t, ctx, intSys.ID)
	require.NotEmpty(t, intSysAuth)

	deleteSystemAuthForIntegrationSystemRequest := fixDeleteSystemAuthForIntegrationSystemRequest(intSysAuth.ID)
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

func TestDeleteSystemAuthFromIntegrationSystemUsingApplicationMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	t.Log("Create integration system")
	intSys := registerIntegrationSystem(t, ctx, name)
	require.NotEmpty(t, intSys)
	defer unregisterIntegrationSystem(t, ctx, intSys.ID)

	intSysAuth := requestClientCredentialsForIntegrationSystem(t, ctx, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer deleteClientCredentialsForIntegrationSystem(t, intSysAuth.ID)

	deleteSystemAuthForApplicationRequest := fixDeleteSystemAuthForApplicationRequest(intSysAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for integration system using application mutation")
	err := tc.RunOperation(ctx, deleteSystemAuthForApplicationRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromIntegrationSystemUsingRuntimeMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	t.Log("Create integration system")
	intSys := registerIntegrationSystem(t, ctx, name)
	require.NotEmpty(t, intSys)
	defer unregisterIntegrationSystem(t, ctx, intSys.ID)

	intSysAuth := requestClientCredentialsForIntegrationSystem(t, ctx, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer deleteClientCredentialsForIntegrationSystem(t, intSysAuth.ID)

	deleteSystemAuthForRuntimeRequest := fixDeleteSystemAuthForRuntimeRequest(intSysAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for integration system using runtime mutation")
	err := tc.RunOperation(ctx, deleteSystemAuthForRuntimeRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}

func deleteClientCredentialsForApplication(t *testing.T, authID string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			deleteSystemAuthForApplication(authID: "%s") {
			id
		}	
	}`, authID))
	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}

func deleteClientCredentialsForRuntime(t *testing.T, authID string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			deleteSystemAuthForRuntime(authID: "%s") {
			id
		}	
	}`, authID))
	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}

func deleteClientCredentialsForIntegrationSystem(t *testing.T, authID string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			deleteSystemAuthForIntegrationSystem(authID: "%s") {
			id
		}	
	}`, authID))
	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}
