package e2e

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

//Application
func getApplication(t *testing.T, ctx context.Context, id string) graphql.ApplicationExt {
	appRequest := fixApplicationRequest(id)
	app := graphql.ApplicationExt{}
	require.NoError(t, tc.RunOperation(ctx, appRequest, &app))
	return app
}

func createApplication(t *testing.T, ctx context.Context, name string) graphql.ApplicationExt {
	in := generateSampleApplicationInputWithName("first", name)
	return createApplicationFromInputWithinTenant(t, ctx, in, defaultTenant)
}

func createApplicationFromInputWithinTenant(t *testing.T, ctx context.Context, in graphql.ApplicationInput, tenantID string) graphql.ApplicationExt {
	appInputGQL, err := tc.Graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixCreateApplicationRequest(appInputGQL)

	app := graphql.ApplicationExt{}

	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, tenantID, createRequest, &app))
	require.NotEmpty(t, app.ID)
	return app
}

func deleteApplication(t *testing.T, ctx context.Context, applicationID string) graphql.ApplicationExt {
	deleteRequest := fixDeleteApplicationRequest(t, applicationID)

	app := graphql.ApplicationExt{}

	require.NoError(t, tc.RunOperation(ctx, deleteRequest, &app))
	return app
}

// API Spec
func addApi(t *testing.T, ctx context.Context, in graphql.APIDefinitionInput, applicationID string) *graphql.APIDefinitionExt {
	apiInputGQL, err := tc.Graphqlizer.APIDefinitionInputToGQL(in)
	require.NoError(t, err)

	addApiRequest := fixAddApiRequest(applicationID, apiInputGQL)

	api := graphql.APIDefinitionExt{}

	require.NoError(t, tc.RunOperation(ctx, addApiRequest, &api))
	require.NotEmpty(t, api.ID)
	return &api
}

// Integration System
func getIntegrationSystem(t *testing.T, ctx context.Context, id string) *graphql.IntegrationSystemExt {
	intSysRequest := fixIntegrationSystemRequest(id)
	intSys := graphql.IntegrationSystemExt{}
	require.NoError(t, tc.RunOperation(ctx, intSysRequest, &intSys))
	return &intSys
}

func createIntegrationSystem(t *testing.T, ctx context.Context, name string) *graphql.IntegrationSystemExt {
	input := graphql.IntegrationSystemInput{Name: name}
	in, err := tc.Graphqlizer.IntegrationSystemInputToGQL(input)
	if err != nil {
		return nil
	}

	req := FixCreateIntegrationSystemRequest(in)

	out := &graphql.IntegrationSystemExt{}
	err = tc.RunOperation(ctx, req, out)
	require.NotEmpty(t, out)
	require.NoError(t, err)
	return out
}

func deleteIntegrationSystem(t *testing.T, ctx context.Context, id string) {
	req := fixDeleteIntegrationSystem(id)
	err := tc.RunOperation(ctx, req, nil)
	require.NoError(t, err)
}

func generateClientCredentialsForIntegrationSystem(t *testing.T, ctx context.Context, id string) graphql.SystemAuth {
	req := fixGenerateClientCredentialsForIntegrationSystem(id)
	out := graphql.SystemAuth{}
	err := tc.RunOperation(ctx, req, &out)
	require.NoError(t, err)
	return out
}

func deleteSystemAuthForIntegrationSystem(t *testing.T, ctx context.Context, id string) {
	req := fixDeleteSystemAuthForIntegrationSystem(id)
	err := tc.RunOperation(ctx, req, nil)
	require.NoError(t, err)
}

func generateSampleApplicationInputWithName(placeholder, name string) graphql.ApplicationInput {
	return graphql.ApplicationInput{
		Name: name,
		Documents: []*graphql.DocumentInput{{
			Title:  placeholder,
			Format: graphql.DocumentFormatMarkdown}},
		Apis: []*graphql.APIDefinitionInput{{
			Name:      placeholder,
			TargetURL: placeholder}},
		EventAPIs: []*graphql.EventAPIDefinitionInput{{
			Name: placeholder,
			Spec: &graphql.EventAPISpecInput{
				EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
				Format:        graphql.SpecFormatYaml,
			}}},
		Webhooks: []*graphql.WebhookInput{{
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
			URL:  placeholder},
		},
		Labels: &graphql.Labels{placeholder: []interface{}{placeholder}},
	}
}
