package e2e

import (
	"context"
	"testing"

	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

//Application
func createApplicationFromInputWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, in graphql.ApplicationCreateInput) graphql.ApplicationExt {
	appInputGQL, err := tc.Graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixCreateApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}
	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, createRequest, &app)
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
func createIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, name string) *graphql.IntegrationSystemExt {
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

func deleteIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) {
	req := fixDeleteIntegrationSystem(id)
	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	require.NoError(t, err)
}

func deleteIntegrationSystemWithErr(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) {
	req := fixDeleteIntegrationSystem(id)
	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	require.Contains(t, err.Error(), "delete")
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
