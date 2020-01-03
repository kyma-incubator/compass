package auth_related_tests

import (
	"context"
	"testing"

	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

//Application
func registerApplicationFromInputWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, in graphql.ApplicationRegisterInput) graphql.ApplicationExt {
	app, err := registerApplicationWithinTenant(t, ctx, gqlClient, tenant, in)
	require.NoError(t, err)
	return app
}

func registerApplicationWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, in graphql.ApplicationRegisterInput) (graphql.ApplicationExt, error) {
	appInputGQL, err := tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixRegisterApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}
	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, createRequest, &app)
	return app, err
}

func requestClientCredentialsForApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) graphql.SystemAuth {
	req := fixRequestClientCredentialsForApplication(id)
	systemAuth := graphql.SystemAuth{}

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &systemAuth)
	require.NoError(t, err)
	return systemAuth
}

func unregisterApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, applicationID string) graphql.ApplicationExt {
	deleteRequest := fixDeleteApplicationRequest(t, applicationID)
	app := graphql.ApplicationExt{}

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, deleteRequest, &app)
	require.NoError(t, err)
	return app
}

func getApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) graphql.ApplicationExt {
	appRequest := fixGetApplicationRequest(id)
	app := graphql.ApplicationExt{}

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, appRequest, &app)
	require.NoError(t, err)
	return app
}

// Runtime
func registerRuntimeFromInputWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, input *graphql.RuntimeInput) graphql.RuntimeExt {
	inputGQL, err := tc.Graphqlizer.RuntimeInputToGQL(*input)
	require.NoError(t, err)

	registerRuntimeRequest := fixRegisterRuntimeRequest(inputGQL)
	var runtime graphql.RuntimeExt

	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, registerRuntimeRequest, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)
	return runtime
}

func requestClientCredentialsForRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) graphql.SystemAuth {
	req := fixRequestClientCredentialsForRuntime(id)
	systemAuth := graphql.SystemAuth{}

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &systemAuth)
	require.NoError(t, err)
	return systemAuth
}

func unregisterRuntimeWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) {
	delReq := fixUnregisterRuntime(id)

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, delReq, nil)
	require.NoError(t, err)
}

func getRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) graphql.RuntimeExt {
	req := fixRuntimeRequest(id)
	runtime := graphql.RuntimeExt{}

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &runtime)
	require.NoError(t, err)
	return runtime
}

// API Spec
func addAPIWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, in graphql.APIDefinitionInput, applicationID string) graphql.APIDefinitionExt {
	apiInputGQL, err := tc.Graphqlizer.APIDefinitionInputToGQL(in)
	require.NoError(t, err)

	addApiRequest := fixAddApiRequest(applicationID, apiInputGQL)
	api := graphql.APIDefinitionExt{}

	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, addApiRequest, &api)
	require.NoError(t, err)
	require.NotEmpty(t, api.ID)
	return api
}

// Integration System
func registerIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, name string) *graphql.IntegrationSystemExt {
	input := graphql.IntegrationSystemInput{Name: name}
	in, err := tc.Graphqlizer.IntegrationSystemInputToGQL(input)
	if err != nil {
		return nil
	}

	req := fixCreateIntegrationSystemRequest(in)
	intSys := &graphql.IntegrationSystemExt{}

	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys)
	return intSys
}

func unregisterIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) {
	req := fixUnregisterIntegrationSystem(id)
	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	require.NoError(t, err)
}

func unregisterIntegrationSystemWithErr(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) {
	req := fixUnregisterIntegrationSystem(id)
	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "referenced by it")
}

func getSystemAuthsForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) []*graphql.SystemAuth {
	req := fixGetIntegrationSystemRequest(id)
	intSys := graphql.IntegrationSystemExt{}
	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &intSys)
	require.NoError(t, err)
	return intSys.Auths
}

func requestClientCredentialsForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) *graphql.OAuthCredentialData {
	req := fixGenerateClientCredentialsForIntegrationSystem(id)
	systemAuth := graphql.SystemAuth{}

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &systemAuth)
	require.NoError(t, err)
	intSysOauthCredentialData, ok := systemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, intSysOauthCredentialData.ClientSecret)
	require.NotEmpty(t, intSysOauthCredentialData.ClientID)
	return intSysOauthCredentialData
}

func generateOneTimeTokenForApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) graphql.OneTimeTokenExt {
	req := fixGenerateOneTimeTokenForApplication(id)
	oneTimeToken := graphql.OneTimeTokenExt{}

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &oneTimeToken)
	require.NoError(t, err)

	require.NotEmpty(t, oneTimeToken.ConnectorURL)
	require.NotEmpty(t, oneTimeToken.Token)
	require.NotEmpty(t, oneTimeToken.Raw)
	require.NotEmpty(t, oneTimeToken.RawEncoded)
	return oneTimeToken
}

//Application Template
func createApplicationTemplate(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, input graphql.ApplicationTemplateInput) graphql.ApplicationTemplate {
	appTemplate, err := tc.Graphqlizer.ApplicationTemplateInputToGQL(input)
	require.NoError(t, err)

	req := fixCreateApplicationTemplateRequest(appTemplate)
	appTpl := graphql.ApplicationTemplate{}
	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &appTpl)
	require.NoError(t, err)
	return appTpl
}

func getApplicationTemplate(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) graphql.ApplicationTemplate {
	req := fixApplicationTemplateRequest(id)
	appTpl := graphql.ApplicationTemplate{}

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &appTpl)
	require.NoError(t, err)
	return appTpl
}

func deleteApplicationTemplate(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, applicationTemplateID string) {
	req := fixDeleteApplicationTemplateRequest(applicationTemplateID)

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	require.NoError(t, err)
}
