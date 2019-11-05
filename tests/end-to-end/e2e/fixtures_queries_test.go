package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

//Application
func createApplicationFromInputWithinTenant(t *testing.T, ctx context.Context, in graphql.ApplicationCreateInput, tc TestContext) graphql.ApplicationExt {
	appInputGQL, err := tc.Graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixCreateApplicationRequest(appInputGQL)
	createRequest.Header = http.Header{"Tenant": {"3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"}}
	app := graphql.ApplicationExt{}
	m := resultMapperFor(&app)

	err = tc.withRetryOnTemporaryConnectionProblems(func() error { return tc.cli.Run(ctx, createRequest, &m) })
	require.NoError(t, err)
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
	addApiRequest.Header = http.Header{"Tenant": {"3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"}}

	api := graphql.APIDefinitionExt{}
	m := resultMapperFor(&api)

	err = tc.withRetryOnTemporaryConnectionProblems(func() error { return tc.cli.Run(ctx, addApiRequest, &m) })
	require.NoError(t, err)
	require.NotEmpty(t, api.ID)
	return &api
}

// Integration System
func createIntegrationSystem(t *testing.T, ctx context.Context, name string) *graphql.IntegrationSystemExt {
	input := graphql.IntegrationSystemInput{Name: name}
	in, err := tc.Graphqlizer.IntegrationSystemInputToGQL(input)
	if err != nil {
		return nil
	}

	req := fixCreateIntegrationSystemRequest(in)

	out := &graphql.IntegrationSystemExt{}
	err = tc.RunOperation(ctx, req, out)
	require.NoError(t, err)
	require.NotEmpty(t, out)
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

func generateSampleApplicationInputWithName(placeholder, name string) graphql.ApplicationCreateInput {
	return graphql.ApplicationCreateInput{
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
