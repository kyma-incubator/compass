package director

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
	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixCreateApplicationRequest(appInputGQL)

	app := graphql.ApplicationExt{}

	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, tenantID, createRequest, &app))
	require.NotEmpty(t, app.ID)
	return app
}

func deleteApplicationLabel(t *testing.T, ctx context.Context, applicationID, labelKey string) {
	deleteRequest := fixDeleteApplicationLabel(applicationID, labelKey)

	require.NoError(t, tc.RunOperation(ctx, deleteRequest, nil))
}

func setApplicationLabel(t *testing.T, ctx context.Context, applicationID string, labelKey string, labelValue interface{}) graphql.Label {
	setLabelRequest := fixSetApplicationLabelRequest(applicationID, labelKey, labelValue)
	label := graphql.Label{}

	err := tc.RunOperation(ctx, setLabelRequest, &label)
	require.NoError(t, err)

	return label
}

//Runtime
func createRuntime(t *testing.T, ctx context.Context, placeholder string) *graphql.Runtime {
	input := fixRuntimeInput(placeholder)
	return createRuntimeFromInput(t, ctx, &input)
}

func createRuntimeFromInput(t *testing.T, ctx context.Context, input *graphql.RuntimeInput) *graphql.Runtime {
	return createRuntimeFromInputWithinTenant(t, ctx, input, defaultTenant)
}

func createRuntimeFromInputWithinTenant(t *testing.T, ctx context.Context, input *graphql.RuntimeInput, tenant string) *graphql.Runtime {
	inputGQL, err := tc.graphqlizer.RuntimeInputToGQL(*input)
	require.NoError(t, err)

	createRequest := fixCreateRuntimeRequest(inputGQL)
	var runtime graphql.Runtime

	err = tc.RunOperationWithCustomTenant(ctx, tenant, createRequest, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)
	return &runtime
}

func getRuntime(t *testing.T, ctx context.Context, runtimeID string) *graphql.RuntimeExt {
	return getRuntimeWithinTenant(t, ctx, runtimeID, defaultTenant)
}

func getRuntimeWithinTenant(t *testing.T, ctx context.Context, runtimeID string, tenant string) *graphql.RuntimeExt {
	var runtime graphql.RuntimeExt
	runtimeQuery := fixRuntimeQuery(runtimeID)

	err := tc.RunOperationWithCustomTenant(ctx, tenant, runtimeQuery, &runtime)
	require.NoError(t, err)
	return &runtime
}

func setRuntimeLabel(t *testing.T, ctx context.Context, runtimeID string, labelKey string, labelValue interface{}) *graphql.Label {
	setLabelRequest := fixSetRuntimeLabelRequest(runtimeID, labelKey, labelValue)
	label := graphql.Label{}

	err := tc.RunOperation(ctx, setLabelRequest, &label)
	require.NoError(t, err)

	return &label
}

func deleteRuntime(t *testing.T, id string) {
	deleteRuntimeWithinTenant(t, id, defaultTenant)
}

func deleteRuntimeWithinTenant(t *testing.T, id string, tenantID string) {
	delReq := fixDeleteRuntime(id)

	err := tc.RunOperationWithCustomTenant(context.Background(), tenantID, delReq, nil)
	require.NoError(t, err)
}

// Label Definitions
func createLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, key string, schema interface{}, tenantID string) *graphql.LabelDefinition {
	input := graphql.LabelDefinitionInput{
		Key:    key,
		Schema: marshallJSONSchema(t, schema),
	}

	in, err := tc.graphqlizer.LabelDefinitionInputToGQL(input)
	if err != nil {
		return nil
	}

	createRequest := fixCreateLabelDefinitionRequest(in)

	output := graphql.LabelDefinition{}
	err = tc.RunOperationWithCustomTenant(ctx, tenantID, createRequest, &output)
	require.NoError(t, err)

	return &output
}

func deleteLabelDefinition(t *testing.T, ctx context.Context, labelDefinitionKey string, deleteRelatedResources bool) {
	deleteLabelDefinitionWithinTenant(t, ctx, labelDefinitionKey, deleteRelatedResources, defaultTenant)
}

func deleteLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, labelDefinitionKey string, deleteRelatedResources bool, tenantID string) {
	deleteRequest := fixDeleteLabelDefinition(labelDefinitionKey, deleteRelatedResources)

	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, tenantID, deleteRequest, nil))
}

func listLabelDefinitionsWithinTenant(t *testing.T, ctx context.Context, tenantID string) ([]*graphql.LabelDefinition, error) {
	labelDefinitionsRequest := fixLabelDefinitionsRequest()

	var labelDefinitions []*graphql.LabelDefinition

	err := tc.RunOperationWithCustomTenant(ctx, tenantID, labelDefinitionsRequest, &labelDefinitions)
	return labelDefinitions, err
}

// API Definition
func setAPIAuth(t *testing.T, ctx context.Context, apiID string, rtmID string, auth graphql.AuthInput) *graphql.RuntimeAuth {
	authInStr, err := tc.graphqlizer.AuthInputToGQL(&auth)
	require.NoError(t, err)

	var rtmAuth graphql.RuntimeAuth

	request := fixSetAPIAuthRequest(apiID, rtmID, authInStr)
	err = tc.RunOperation(ctx, request, &rtmAuth)
	require.NoError(t, err)

	return &rtmAuth
}

func deleteAPIAuth(t *testing.T, ctx context.Context, apiID string, rtmID string) {
	request := fixDeleteAPIAuthRequest(apiID, rtmID)

	err := tc.RunOperation(ctx, request, nil)

	require.NoError(t, err)
}

//OneTimeToken

func generateOneTimeTokenForApplication(t *testing.T, ctx context.Context, id string) graphql.OneTimeToken {
	tokenRequest := fixGenerateOneTimeTokenForApp(id)
	token := graphql.OneTimeToken{}
	err := tc.RunOperation(ctx, tokenRequest, &token)
	require.NoError(t, err)
	return token
}

func generateOneTimeTokenForRuntime(t *testing.T, ctx context.Context, id string) graphql.OneTimeToken {
	tokenRequest := fixGenerateOneTimeTokenForRuntime(id)
	token := graphql.OneTimeToken{}
	err := tc.RunOperation(ctx, tokenRequest, &token)
	require.NoError(t, err)
	return token
}
