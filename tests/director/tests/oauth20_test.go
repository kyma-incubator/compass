//+build !ignore_external_dependencies

package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateClientCredentialsToApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "app"

	t.Log("Create application")
	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, name, tenantId)
	require.NotEmpty(t, app)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

	generateApplicationClientCredentialsRequest := fixtures.FixRequestClientCredentialsForApplication(app.ID)
	appAuth := graphql.AppSystemAuth{}

	// WHEN
	t.Log("Generate client credentials for application")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, generateApplicationClientCredentialsRequest, &appAuth)
	require.NoError(t, err)

	//THEN
	t.Log("Get updated application")
	app = fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
	require.NotEmpty(t, app.Auths)
	defer fixtures.DeleteSystemAuthForApplication(t, ctx, dexGraphQLClient, appAuth.ID)

	t.Log("Check if client credentials were generated")
	assert.NotEmpty(t, appAuth.Auth.Credential)
	oauthCredentialData, ok := appAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	assert.NotEmpty(t, oauthCredentialData.ClientID)
	assert.NotEmpty(t, oauthCredentialData.ClientSecret)
	assert.Equal(t, appAuth.ID, oauthCredentialData.ClientID)
}

func TestGenerateClientCredentialsToRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "runtime"
	input := fixtures.FixRuntimeInput(name)

	t.Log("Create runtime")
	rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	require.NotEmpty(t, rtm)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)

	generateRuntimeClientCredentialsRequest := fixtures.FixRequestClientCredentialsForRuntime(rtm.ID)
	rtmAuth := graphql.RuntimeSystemAuth{}

	// WHEN
	t.Log("Generate client credentials for runtime")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, generateRuntimeClientCredentialsRequest, &rtmAuth)
	require.NoError(t, err)

	//THEN
	t.Log("Get updated runtime")
	rtm = fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)
	require.NotEmpty(t, rtm.Auths)
	defer fixtures.DeleteSystemAuthForRuntime(t, ctx, dexGraphQLClient, rtm.Auths[0].ID)

	t.Log("Check if client credentials were generated")
	assert.NotEmpty(t, rtm.Auths[0])
	oauthCredentialData, ok := rtm.Auths[0].Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	assert.NotEmpty(t, oauthCredentialData.ClientID)
	assert.NotEmpty(t, oauthCredentialData.ClientSecret)
	assert.Equal(t, rtmAuth.ID, oauthCredentialData.ClientID)
}

func TestGenerateClientCredentialsToIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Create integration system")
	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, name)
	require.NotEmpty(t, intSys)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys.ID)

	generateIntSysAuthRequest := fixtures.FixRequestClientCredentialsForIntegrationSystem(intSys.ID)
	intSysAuth := graphql.IntSysSystemAuth{}

	// WHEN
	t.Log("Generate client credentials for integration system")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, generateIntSysAuthRequest, &intSysAuth)
	require.NoError(t, err)
	require.NotEmpty(t, intSysAuth.Auth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSysAuth.ID)

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

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "app"

	t.Log("Create application")
	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, name, tenantId)
	require.NotEmpty(t, app)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

	appAuth := fixtures.GenerateClientCredentialsForApplication(t, ctx, dexGraphQLClient, app.ID)
	require.NotEmpty(t, appAuth)

	deleteSystemAuthForApplicationRequest := fixtures.FixDeleteSystemAuthForApplicationRequest(appAuth.ID)
	deleteOutput := graphql.AppSystemAuth{}

	// WHEN
	t.Log("Delete system auth for application")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForApplicationRequest, &deleteOutput)
	require.NoError(t, err)
	require.NotEmpty(t, deleteOutput)

	//THEN
	t.Log("Check if system auth was deleted")
	appAfterDelete := fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
	require.Empty(t, appAfterDelete.Auths)
}

func TestDeleteSystemAuthFromApplicationUsingRuntimeMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "app"

	t.Log("Create application")
	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, name, tenantId)
	require.NotEmpty(t, app)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

	appAuth := fixtures.GenerateClientCredentialsForApplication(t, ctx, dexGraphQLClient, app.ID)
	require.NotEmpty(t, appAuth)
	defer fixtures.DeleteSystemAuthForApplication(t, ctx, dexGraphQLClient, appAuth.ID)

	deleteSystemAuthForRuntimeRequest := fixtures.FixDeleteSystemAuthForRuntimeRequest(appAuth.ID)
	deleteOutput := graphql.RuntimeSystemAuth{}

	// WHEN
	t.Log("Delete system auth for application using runtime mutation")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForRuntimeRequest, &deleteOutput)

	// THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromApplicationUsingIntegrationSystemMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "app"

	t.Log("Create application")
	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, name, tenantId)
	require.NotEmpty(t, app)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

	appAuth := fixtures.GenerateClientCredentialsForApplication(t, ctx, dexGraphQLClient, app.ID)
	require.NotEmpty(t, appAuth)
	defer fixtures.DeleteSystemAuthForApplication(t, ctx, dexGraphQLClient, appAuth.ID)

	deleteSystemAuthForIntegrationSystemRequest := fixtures.FixDeleteSystemAuthForIntegrationSystemRequest(appAuth.ID)
	deleteOutput := graphql.IntSysSystemAuth{}

	// WHEN
	t.Log("Delete system auth for application using runtime mutation")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForIntegrationSystemRequest, &deleteOutput)

	// THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "rtm"
	input := fixtures.FixRuntimeInput(name)

	t.Log("Create runtime")
	rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	require.NotEmpty(t, rtm)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)
	require.NotEmpty(t, rtmAuth)

	deleteSystemAuthForRuntimeRequest := fixtures.FixDeleteSystemAuthForRuntimeRequest(rtmAuth.ID)
	deleteOutput := graphql.RuntimeSystemAuth{}

	// WHEN
	t.Log("Delete system auth for runtime")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForRuntimeRequest, &deleteOutput)
	require.NoError(t, err)
	require.NotEmpty(t, deleteOutput)

	//THEN
	t.Log("Check if system auth was deleted")
	rtmAfterDelete := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)
	require.Empty(t, rtmAfterDelete.Auths)
}

func TestDeleteSystemAuthFromRuntimeUsingApplicationMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "rtm"
	input := fixtures.FixRuntimeInput(name)
	t.Log("Create runtime")
	rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	require.NotEmpty(t, rtm)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)
	require.NotEmpty(t, rtmAuth)
	defer fixtures.DeleteSystemAuthForRuntime(t, ctx, dexGraphQLClient, rtmAuth.ID)

	deleteSystemAuthForApplicationRequest := fixtures.FixDeleteSystemAuthForApplicationRequest(rtmAuth.ID)
	deleteOutput := graphql.AppSystemAuth{}

	// WHEN
	t.Log("Delete system auth for runtime using application mutation")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForApplicationRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromRuntimeUsingIntegrationSystemMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "rtm"
	input := fixtures.FixRuntimeInput(name)
	t.Log("Create runtime")
	rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	require.NotEmpty(t, rtm)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)
	require.NotEmpty(t, rtmAuth)
	defer fixtures.DeleteSystemAuthForRuntime(t, ctx, dexGraphQLClient, rtmAuth.ID)

	deleteSystemAuthForIntegrationSystemRequest := fixtures.FixDeleteSystemAuthForIntegrationSystemRequest(rtmAuth.ID)
	deleteOutput := graphql.IntSysSystemAuth{}

	// WHEN
	t.Log("Delete system auth for runtime using integration system mutation")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForIntegrationSystemRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Create integration system")
	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, name)
	require.NotEmpty(t, intSys)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)

	deleteSystemAuthForIntegrationSystemRequest := fixtures.FixDeleteSystemAuthForIntegrationSystemRequest(intSysAuth.ID)
	deleteOutput := graphql.IntSysSystemAuth{}

	// WHEN
	t.Log("Delete system auth for integration system")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForIntegrationSystemRequest, &deleteOutput)
	require.NoError(t, err)
	require.NotEmpty(t, deleteOutput)

	//THEN
	t.Log("Check if system auth was deleted")
	intSysAfterDelete := fixtures.GetIntegrationSystem(t, ctx, dexGraphQLClient, intSys.ID)
	require.Empty(t, intSysAfterDelete.Auths)
}

func TestDeleteSystemAuthFromIntegrationSystemUsingApplicationMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Create integration system")
	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, name)
	require.NotEmpty(t, intSys)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSysAuth.ID)

	deleteSystemAuthForApplicationRequest := fixtures.FixDeleteSystemAuthForApplicationRequest(intSysAuth.ID)
	deleteOutput := graphql.AppSystemAuth{}

	// WHEN
	t.Log("Delete system auth for integration system using application mutation")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForApplicationRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}

func TestDeleteSystemAuthFromIntegrationSystemUsingRuntimeMutationShouldReportError(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Create integration system")
	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, name)
	require.NotEmpty(t, intSys)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSysAuth.ID)

	deleteSystemAuthForRuntimeRequest := fixtures.FixDeleteSystemAuthForRuntimeRequest(intSysAuth.ID)
	deleteOutput := graphql.RuntimeSystemAuth{}

	// WHEN
	t.Log("Delete system auth for integration system using runtime mutation")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteSystemAuthForRuntimeRequest, &deleteOutput)

	//THEN
	require.Error(t, err)
}
