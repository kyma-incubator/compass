package director

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

//Application
func createApplication(t *testing.T, ctx context.Context, name string) ApplicationExt {
	in := generateSampleApplicationInputWithName("first", name)
	return createApplicationFromInputWithinTenant(t, ctx, in, defaultTenant)
}

func createApplicationFromInput(t *testing.T, ctx context.Context, in graphql.ApplicationInput) ApplicationExt {
	return createApplicationFromInputWithinTenant(t, ctx, in, defaultTenant)
}

func createApplicationFromInputWithinTenant(t *testing.T, ctx context.Context, in graphql.ApplicationInput, tenantID string) ApplicationExt {
	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixCreateApplicationRequest(appInputGQL)
	createRequest.Header["Tenant"] = []string{tenantID}

	app := ApplicationExt{}

	require.NoError(t, tc.RunQuery(ctx, createRequest, &app))
	require.NotEmpty(t, app.ID)
	return app
}

func deleteApplicationLabel(t *testing.T, ctx context.Context, applicationID, labelKey string) {
	deleteRequest := fixDeleteApplicationLabel(applicationID, labelKey)

	require.NoError(t, tc.RunQuery(ctx, deleteRequest, nil))
}

func setApplicationLabel(t *testing.T, ctx context.Context, applicationID string, labelKey string, labelValue interface{}) graphql.Label {
	setLabelRequest := fixSetApplicationLabelRequest(applicationID, labelKey, labelValue)
	label := graphql.Label{}

	err := tc.RunQuery(ctx, setLabelRequest, &label)
	require.NoError(t, err)

	return label
}

func applications(t *testing.T, ctx context.Context, filter graphql.LabelFilter, first int, after string) graphql.ApplicationPage {
	labelFilterGQL, err := tc.graphqlizer.LabelFilterToGQL(filter)
	require.NoError(t, err)

	applicationRequest := fixApplications(labelFilterGQL, first, after)
	applicationPage := graphql.ApplicationPage{}
	err = tc.RunQuery(ctx, applicationRequest, &applicationPage)
	require.NoError(t, err)

	return applicationPage
}

//Runtime
func getRuntime(t *testing.T, ctx context.Context, runtimeID string) *graphql.Runtime {
	var runtime graphql.Runtime
	runtimeQuery := fixRuntimeQuery(runtimeID)

	err := tc.RunQuery(ctx, runtimeQuery, &runtime)
	require.NoError(t, err)
	return &runtime
}

// Label Definitions
func createLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, input graphql.LabelDefinitionInput, tenantID string) *graphql.LabelDefinition {
	in, err := tc.graphqlizer.LabelDefinitionInputToGQL(input)
	if err != nil {
		return nil
	}

	createRequest := fixCreateLabelDefinitionRequest(in)
	createRequest.Header["Tenant"] = []string{tenantID}

	output := graphql.LabelDefinition{}
	err = tc.RunQuery(ctx, createRequest, &output)
	require.NoError(t, err)

	return &output
}

func deleteLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, labelDefinitionKey string, deleteRelatedResources bool, tenantID string) {
	deleteLabelDefinition(t, ctx, labelDefinitionKey, deleteRelatedResources, tenantID)
}

func deleteLabelDefinitionWithinDefaultTenant(t *testing.T, ctx context.Context, labelDefinitionKey string, deleteRelatedResources bool) {
	deleteLabelDefinition(t, ctx, labelDefinitionKey, deleteRelatedResources, defaultTenant)
}

func deleteLabelDefinition(t *testing.T, ctx context.Context, labelDefinitionKey string, deleteRelatedResources bool, tenantID string) {
	deleteRequest := fixDeleteLabelDefinition(labelDefinitionKey, deleteRelatedResources)
	deleteRequest.Header["Tenant"] = []string{tenantID}

	require.NoError(t, tc.RunQuery(ctx, deleteRequest, nil))
}

func listLabelDefinitionsWithinTenant(t *testing.T, ctx context.Context, tenantID string) ([]*graphql.LabelDefinition, error) {
	labelDefinitionsRequest := fixLabelDefinitionsRequest()
	labelDefinitionsRequest.Header["Tenant"] = []string{tenantID}

	var labelDefinitions []*graphql.LabelDefinition

	err := tc.RunQuery(ctx, labelDefinitionsRequest, &labelDefinitions)
	return labelDefinitions, err
}

func createRuntimeFromInput(t *testing.T, ctx context.Context, input *graphql.RuntimeInput) *graphql.Runtime {
	inputGQL, err := tc.graphqlizer.RuntimeInputToGQL(*input)
	require.NoError(t, err)

	createRequest := fixCreateRuntimeRequest(inputGQL)
	var runtime graphql.Runtime

	err = tc.RunQuery(ctx, createRequest, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)
	return &runtime
}
