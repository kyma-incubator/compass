//+build !ignore_external_dependencies

package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateClientCredentialsToApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "app"

	t.Log("Create application")
	app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, name, tenant)
	require.NotEmpty(t, app)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)

	generateApplicationClientCredentialsRequest := pkg.FixRequestClientCredentialsForApplication(app.ID)
	appAuth := graphql.SystemAuth{}

	// WHEN
	t.Log("Generate client credentials for application")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, generateApplicationClientCredentialsRequest, &appAuth)
	require.NoError(t, err)

	//THEN
	t.Log("Get updated application")
	app = pkg.GetApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	require.NotEmpty(t, app.Auths)
	defer pkg.DeleteSystemAuthForApplication(t, ctx, dexGraphQLClient, app.Auths[0].ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "runtime"
	input := pkg.FixRuntimeInput(name)

	t.Log("Create runtime")
	rtm := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	require.NotEmpty(t, rtm)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)

	generateRuntimeClientCredentialsRequest := pkg.FixRequestClientCredentialsForRuntime(rtm.ID)
	rtmAuth := graphql.SystemAuth{}

	// WHEN
	t.Log("Generate client credentials for runtime")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, generateRuntimeClientCredentialsRequest, &rtmAuth)
	require.NoError(t, err)

	//THEN
	t.Log("Get updated runtime")
	rtm = pkg.GetRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)
	require.NotEmpty(t, rtm.Auths)
	defer pkg.DeleteSystemAuthForRuntime(t, ctx, dexGraphQLClient, rtm.Auths[0].ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Create integration system")
	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, name)
	require.NotEmpty(t, intSys)
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)

	generateIntSysAuthRequest := pkg.FixRequestClientCredentialsForIntegrationSystem(intSys.ID)
	intSysAuth := graphql.SystemAuth{}

	// WHEN
	t.Log("Generate client credentials for integration system")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, generateIntSysAuthRequest, &intSysAuth)
	require.NoError(t, err)
	require.NotEmpty(t, intSysAuth.Auth)
	defer pkg.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSysAuth.ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "app"

	t.Log("Create application")
	app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, name, tenant)
	require.NotEmpty(t, app)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)

	appAuth := pkg.GenerateClientCredentialsForApplication(t, ctx, dexGraphQLClient, app.ID)
	require.NotEmpty(t, appAuth)

	deleteSystemAuthForApplicationRequest := pkg.FixDeleteSystemAuthForApplicationRequest(appAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for application")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForApplicationRequest, &deleteOutput)
	require.NoError(t, err)
	require.NotEmpty(t, deleteOutput)

	//THEN
	t.Log("Check if system auth was deleted")
	appAfterDelete := pkg.GetApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	require.Empty(t, appAfterDelete.Auths)
}

func TestDeleteSystemAuthFromApplicationUsingRuntimeMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "app"

	t.Log("Create application")
	app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, name, tenant)
	require.NotEmpty(t, app)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)

	appAuth := pkg.GenerateClientCredentialsForApplication(t, ctx, dexGraphQLClient, app.ID)
	require.NotEmpty(t, appAuth)
	defer pkg.DeleteSystemAuthForApplication(t, ctx, dexGraphQLClient, appAuth.ID)

	deleteSystemAuthForRuntimeRequest := pkg.FixDeleteSystemAuthForRuntimeRequest(appAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for application using runtime mutation")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForRuntimeRequest, &deleteOutput)

	// THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromApplicationUsingIntegrationSystemMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "app"

	t.Log("Create application")
	app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, name, tenant)
	require.NotEmpty(t, app)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)

	appAuth := pkg.GenerateClientCredentialsForApplication(t, ctx, dexGraphQLClient, app.ID)
	require.NotEmpty(t, appAuth)
	defer pkg.DeleteSystemAuthForApplication(t, ctx, dexGraphQLClient, appAuth.ID)

	deleteSystemAuthForIntegrationSystemRequest := pkg.FixDeleteSystemAuthForIntegrationSystemRequest(appAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for application using runtime mutation")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForIntegrationSystemRequest, &deleteOutput)

	// THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "rtm"
	input := pkg.FixRuntimeInput(name)

	t.Log("Create runtime")
	rtm := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	require.NotEmpty(t, rtm)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)

	rtmAuth := pkg.RequestClientCredentialsForRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)
	require.NotEmpty(t, rtmAuth)

	deleteSystemAuthForRuntimeRequest := pkg.FixDeleteSystemAuthForRuntimeRequest(rtmAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for runtime")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForRuntimeRequest, &deleteOutput)
	require.NoError(t, err)
	require.NotEmpty(t, deleteOutput)

	//THEN
	t.Log("Check if system auth was deleted")
	rtmAfterDelete := pkg.GetRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)
	require.Empty(t, rtmAfterDelete.Auths)
}

func TestDeleteSystemAuthFromRuntimeUsingApplicationMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "rtm"
	input := pkg.FixRuntimeInput(name)
	t.Log("Create runtime")
	rtm := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	require.NotEmpty(t, rtm)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)

	rtmAuth := pkg.RequestClientCredentialsForRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)
	require.NotEmpty(t, rtmAuth)
	defer pkg.DeleteSystemAuthForRuntime(t, ctx, dexGraphQLClient, rtmAuth.ID)

	deleteSystemAuthForApplicationRequest := pkg.FixDeleteSystemAuthForApplicationRequest(rtmAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for runtime using application mutation")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForApplicationRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromRuntimeUsingIntegrationSystemMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "rtm"
	input := pkg.FixRuntimeInput(name)
	t.Log("Create runtime")
	rtm := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	require.NotEmpty(t, rtm)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)

	rtmAuth := pkg.RequestClientCredentialsForRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)
	require.NotEmpty(t, rtmAuth)
	defer pkg.DeleteSystemAuthForRuntime(t, ctx, dexGraphQLClient, rtmAuth.ID)

	deleteSystemAuthForIntegrationSystemRequest := pkg.FixDeleteSystemAuthForIntegrationSystemRequest(rtmAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for runtime using integration system mutation")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForIntegrationSystemRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Create integration system")
	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, name)
	require.NotEmpty(t, intSys)
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)

	intSysAuth := pkg.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)
	require.NotEmpty(t, intSysAuth)

	deleteSystemAuthForIntegrationSystemRequest := pkg.FixDeleteSystemAuthForIntegrationSystemRequest(intSysAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for integration system")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForIntegrationSystemRequest, &deleteOutput)
	require.NoError(t, err)
	require.NotEmpty(t, deleteOutput)

	//THEN
	t.Log("Check if system auth was deleted")
	intSysAfterDelete := pkg.GetIntegrationSystem(t, ctx, dexGraphQLClient, intSys.ID)
	require.Empty(t, intSysAfterDelete.Auths)
}

func TestDeleteSystemAuthFromIntegrationSystemUsingApplicationMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Create integration system")
	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, name)
	require.NotEmpty(t, intSys)
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)

	intSysAuth := pkg.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer pkg.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSysAuth.ID)

	deleteSystemAuthForApplicationRequest := pkg.FixDeleteSystemAuthForApplicationRequest(intSysAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for integration system using application mutation")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForApplicationRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromIntegrationSystemUsingRuntimeMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Create integration system")
	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, name)
	require.NotEmpty(t, intSys)
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)

	intSysAuth := pkg.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer pkg.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSysAuth.ID)

	deleteSystemAuthForRuntimeRequest := pkg.FixDeleteSystemAuthForRuntimeRequest(intSysAuth.ID)
	deleteOutput := graphql.SystemAuth{}

	// WHEN
	t.Log("Delete system auth for integration system using runtime mutation")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForRuntimeRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}
