package director

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

//Application
func getApplication(t *testing.T, ctx context.Context, id string) ApplicationExt {
	appRequest := fixApplicationRequest(id)
	app := ApplicationExt{}
	require.NoError(t, tc.RunQuery(ctx, appRequest, &app))
	return app
}

func createApplication(t *testing.T, ctx context.Context, name string) ApplicationExt {
	in := generateSampleApplicationInputWithName("first", name)
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

//Runtime
func createRuntime(t *testing.T, ctx context.Context, placeholder string) *graphql.Runtime {
	input := fixRuntimeInput(placeholder)
	return createRuntimeFromInput(t, ctx, &input)
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

func getRuntime(t *testing.T, ctx context.Context, runtimeID string) *RuntimeExt {
	var runtime RuntimeExt
	runtimeQuery := fixRuntimeQuery(runtimeID)

	err := tc.RunQuery(ctx, runtimeQuery, &runtime)
	require.NoError(t, err)
	return &runtime
}

func setRuntimeLabel(t *testing.T, ctx context.Context, runtimeID string, labelKey string, labelValue interface{}) *graphql.Label {
	setLabelRequest := fixSetRuntimeLabelRequest(runtimeID, labelKey, labelValue)
	label := graphql.Label{}

	err := tc.RunQuery(ctx, setLabelRequest, &label)
	require.NoError(t, err)

	return &label
}

func deleteRuntime(t *testing.T, id string) {
	deleteRuntimeInTenant(t, id, defaultTenant)
}

func deleteRuntimeInTenant(t *testing.T, id string, tenantID string) {
	delReq := fixDeleteRuntime(id)

	delReq.Header["Tenant"] = []string{tenantID}
	err := tc.RunQuery(context.Background(), delReq, nil)
	require.NoError(t, err)
}

// Label Definitions
func createLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, key string, schema interface{}, tenantID string) *graphql.LabelDefinition {
	input := graphql.LabelDefinitionInput{
		Key:    key,
		Schema: &schema,
	}

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

func deleteLabelDefinition(t *testing.T, ctx context.Context, labelDefinitionKey string, deleteRelatedResources bool) {
	deleteLabelDefinitionWithinTenant(t, ctx, labelDefinitionKey, deleteRelatedResources, defaultTenant)
}

func deleteLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, labelDefinitionKey string, deleteRelatedResources bool, tenantID string) {
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

// API Definition
func setAPIAuth(t *testing.T, ctx context.Context, apiID string, rtmID string, auth graphql.AuthInput) *graphql.RuntimeAuth {
	authInStr, err := tc.graphqlizer.AuthInputToGQL(&auth)
	require.NoError(t, err)

	var rtmAuth graphql.RuntimeAuth

	request := fixSetAPIAuthRequest(apiID, rtmID, authInStr)
	err = tc.RunQuery(ctx, request, &rtmAuth)
	require.NoError(t, err)

	return &rtmAuth
}

func deleteAPIAuth(t *testing.T, ctx context.Context, apiID string, rtmID string) {
	request := fixDeleteAPIAuthRequest(apiID, rtmID)

	err := tc.RunQuery(ctx, request, nil)

	require.NoError(t, err)
}
