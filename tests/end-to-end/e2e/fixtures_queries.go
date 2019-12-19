package e2e

import (
	"context"
	"testing"

	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

//Application
func registerApplicationFromInputWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, in graphql.ApplicationRegisterInput) graphql.ApplicationExt {
	app, err := createApplicationWithinTenant(t, ctx, gqlClient, tenant, in)
	require.NoError(t, err)
	return app
}

func deleteApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, applicationID string) graphql.ApplicationExt {
	deleteRequest := fixDeleteApplicationRequest(t, applicationID)
	app := graphql.ApplicationExt{}

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, deleteRequest, &app)
	require.NoError(t, err)
	return app
}

// API Spec
func addAPIWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, in graphql.APIDefinitionInput, applicationID string) *graphql.APIDefinitionExt {
	apiInputGQL, err := tc.Graphqlizer.APIDefinitionInputToGQL(in)
	require.NoError(t, err)

	addApiRequest := fixAddApiRequest(applicationID, apiInputGQL)
	api := graphql.APIDefinitionExt{}

	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, addApiRequest, &api)
	require.NoError(t, err)
	require.NotEmpty(t, api.ID)
	return &api
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

func generateClientCredentialsForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) graphql.SystemAuth {
	req := fixGenerateClientCredentialsForIntegrationSystem(id)
	systemAuth := graphql.SystemAuth{}

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &systemAuth)
	require.NoError(t, err)
	return systemAuth
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

func createApplicationWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, in graphql.ApplicationRegisterInput) (graphql.ApplicationExt, error) {
	appInputGQL, err := tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixCreateApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}
	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, createRequest, &app)
	return app, err
}
