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
	createRequest.Header.Set("Tenant", tenant)
	app := graphql.ApplicationExt{}
	m := resultMapperFor(&app)

	err = gqlClient.Run(ctx, createRequest, &m)
	require.NoError(t, err)
	//err = tc.withRetryOnTemporaryConnectionProblems(func() error { return tc.cli.Run(ctx, createRequest, &m) })
	//require.NoError(t, err)
	require.NotEmpty(t, app.ID)
	return app
}

func deleteApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, applicationID string) graphql.ApplicationExt {
	deleteRequest := fixDeleteApplicationRequest(t, applicationID)
	deleteRequest.Header.Set("Tenant", tenant)

	app := graphql.ApplicationExt{}
	m := resultMapperFor(&app)

	err := tc.withRetryOnTemporaryConnectionProblems(func() error { return gqlClient.Run(ctx, deleteRequest, &m) })
	require.NoError(t, err)
	return app
}

// API Spec
func addApiWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, in graphql.APIDefinitionInput, applicationID string) *graphql.APIDefinitionExt {
	apiInputGQL, err := tc.Graphqlizer.APIDefinitionInputToGQL(in)
	require.NoError(t, err)

	addApiRequest := fixAddApiRequest(applicationID, apiInputGQL)
	addApiRequest.Header.Set("Tenant", tenant)

	api := graphql.APIDefinitionExt{}
	m := resultMapperFor(&api)

	err = tc.withRetryOnTemporaryConnectionProblems(func() error { return gqlClient.Run(ctx, addApiRequest, &m) })
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
	req.Header.Set("Tenant", tenant)
	out := &graphql.IntegrationSystemExt{}

	m := resultMapperFor(&out)
	err = tc.withRetryOnTemporaryConnectionProblems(func() error { return gqlClient.Run(ctx, req, &m) })
	require.NoError(t, err)
	require.NotEmpty(t, out)
	return out
}

func deleteIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) {
	req := fixDeleteIntegrationSystem(id)
	req.Header.Set("Tenant", tenant)
	err := tc.withRetryOnTemporaryConnectionProblems(func() error { return gqlClient.Run(ctx, req, nil) })
	require.NoError(t, err)
}

func generateClientCredentialsForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) graphql.SystemAuth {
	req := fixGenerateClientCredentialsForIntegrationSystem(id)
	req.Header.Set("Tenant", tenant)
	out := graphql.SystemAuth{}

	m := resultMapperFor(&out)
	err := gqlClient.Run(ctx, req, &m)
	require.NoError(t, err)
	return out
}
