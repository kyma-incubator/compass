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
	in := fixSampleApplicationCreateInputWithName("first", name)
	return createApplicationFromInputWithinTenant(t, ctx, in, defaultTenant)
}

func createApplicationFromInputWithinTenant(t *testing.T, ctx context.Context, in graphql.ApplicationCreateInput, tenantID string) graphql.ApplicationExt {
	appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixCreateApplicationRequest(appInputGQL)

	app := graphql.ApplicationExt{}

	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, tenantID, createRequest, &app))
	require.NotEmpty(t, app.ID)
	return app
}

func deleteApplicationLabel(t *testing.T, ctx context.Context, applicationID, labelKey string) {
	deleteRequest := fixDeleteApplicationLabelRequest(applicationID, labelKey)

	require.NoError(t, tc.RunOperation(ctx, deleteRequest, nil))
}

func setApplicationLabel(t *testing.T, ctx context.Context, applicationID string, labelKey string, labelValue interface{}) graphql.Label {
	setLabelRequest := fixSetApplicationLabelRequest(applicationID, labelKey, labelValue)
	label := graphql.Label{}

	err := tc.RunOperation(ctx, setLabelRequest, &label)
	require.NoError(t, err)

	return label
}

func generateClientCredentialsForApplication(t *testing.T, ctx context.Context, id string) graphql.SystemAuth {
	req := fixGenerateClientCredentialsForApplicationRequest(id)
	out := graphql.SystemAuth{}
	err := tc.RunOperation(ctx, req, &out)
	require.NoError(t, err)
	return out
}

func deleteSystemAuthForApplication(t *testing.T, ctx context.Context, id string) {
	req := fixDeleteSystemAuthForApplicationRequest(id)
	err := tc.RunOperation(ctx, req, nil)
	require.NoError(t, err)
}

//Application Template
func getApplicationTemplate(t *testing.T, ctx context.Context, id string) *graphql.ApplicationTemplate {
	appTemplateRequest := fixApplicationTemplateRequest(id)
	appTemplate := graphql.ApplicationTemplate{}
	require.NoError(t, tc.RunOperation(ctx, appTemplateRequest, &appTemplate))
	return &appTemplate
}

func createApplicationTemplate(t *testing.T, ctx context.Context, name string) graphql.ApplicationTemplate {
	appTemplateInput := fixApplicationTemplate(name)
	appTemplate, err := tc.graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	createApplicationTemplateRequest := fixCreateApplicationTemplateRequest(appTemplate)
	output := graphql.ApplicationTemplate{}

	err = tc.RunOperation(ctx, createApplicationTemplateRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	return output
}

func deleteApplicationTemplate(t *testing.T, ctx context.Context, id string) {
	req := fixDeleteApplicationTemplate(id)
	err := tc.RunOperation(ctx, req, nil)
	require.NoError(t, err)
}

//Runtime
func createRuntime(t *testing.T, ctx context.Context, placeholder string) *graphql.RuntimeExt {
	input := fixRuntimeInput(placeholder)
	return createRuntimeFromInput(t, ctx, &input)
}

func createRuntimeFromInput(t *testing.T, ctx context.Context, input *graphql.RuntimeInput) *graphql.RuntimeExt {
	return createRuntimeFromInputWithinTenant(t, ctx, input, defaultTenant)
}

func createRuntimeFromInputWithinTenant(t *testing.T, ctx context.Context, input *graphql.RuntimeInput, tenant string) *graphql.RuntimeExt {
	inputGQL, err := tc.graphqlizer.RuntimeInputToGQL(*input)
	require.NoError(t, err)

	createRequest := fixCreateRuntimeRequest(inputGQL)
	var runtime graphql.RuntimeExt

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
	runtimeQuery := fixRuntimeRequest(runtimeID)

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
	delReq := fixDeleteRuntimeRequest(id)

	err := tc.RunOperationWithCustomTenant(context.Background(), tenantID, delReq, nil)
	require.NoError(t, err)
}

func generateClientCredentialsForRuntime(t *testing.T, ctx context.Context, id string) graphql.SystemAuth {
	req := fixGenerateClientCredentialsForRuntimeRequest(id)
	out := graphql.SystemAuth{}
	err := tc.RunOperation(ctx, req, &out)
	require.NoError(t, err)
	return out
}

func deleteSystemAuthForRuntime(t *testing.T, ctx context.Context, id string) {
	req := fixDeleteSystemAuthForRuntimeRequest(id)
	err := tc.RunOperation(ctx, req, nil)
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
	deleteRequest := fixDeleteLabelDefinitionRequest(labelDefinitionKey, deleteRelatedResources)

	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, tenantID, deleteRequest, nil))
}

func listLabelDefinitionsWithinTenant(t *testing.T, ctx context.Context, tenantID string) ([]*graphql.LabelDefinition, error) {
	labelDefinitionsRequest := fixLabelDefinitionsRequest()

	var labelDefinitions []*graphql.LabelDefinition

	err := tc.RunOperationWithCustomTenant(ctx, tenantID, labelDefinitionsRequest, &labelDefinitions)
	return labelDefinitions, err
}

// API Definition
func setAPIAuth(t *testing.T, ctx context.Context, apiID string, rtmID string, auth graphql.AuthInput) *graphql.APIRuntimeAuth {
	authInStr, err := tc.graphqlizer.AuthInputToGQL(&auth)
	require.NoError(t, err)

	var apiRtmAuth graphql.APIRuntimeAuth

	request := fixSetAPIAuthRequest(apiID, rtmID, authInStr)
	err = tc.RunOperation(ctx, request, &apiRtmAuth)
	require.NoError(t, err)

	return &apiRtmAuth
}

func deleteAPIAuth(t *testing.T, ctx context.Context, apiID string, rtmID string) {
	request := fixDeleteAPIAuthRequestRequest(apiID, rtmID)

	err := tc.RunOperation(ctx, request, nil)

	require.NoError(t, err)
}

func addAPI(t *testing.T, ctx context.Context, appID string, apiInput graphql.APIDefinitionInput) graphql.APIDefinition {
	apiGQL, err := tc.graphqlizer.APIDefinitionInputToGQL(apiInput)
	require.NoError(t, err)
	addAPIRequest := fixAddAPIRequest(appID, apiGQL)
	//WHEN
	outAPI := graphql.APIDefinition{}
	err = tc.RunOperation(ctx, addAPIRequest, &outAPI)
	require.NoError(t, err)
	return outAPI
}

func addEventAPI(t *testing.T, ctx context.Context, appID string, apiInput graphql.EventAPIDefinitionInput) graphql.EventAPIDefinition {
	eventAPIGQL, err := tc.graphqlizer.EventAPIDefinitionInputToGQL(apiInput)
	require.NoError(t, err)
	addAPIRequest := fixAddEventAPIRequest(appID, eventAPIGQL)
	//WHEN
	outEventAPI := graphql.EventAPIDefinition{}
	err = tc.RunOperation(ctx, addAPIRequest, &outEventAPI)
	require.NoError(t, err)
	return outEventAPI
}

//OneTimeToken

func generateOneTimeTokenForApplication(t *testing.T, ctx context.Context, id string) graphql.OneTimeToken {
	tokenRequest := fixGenerateOneTimeTokenForAppRequest(id)
	token := graphql.OneTimeToken{}
	err := tc.RunOperation(ctx, tokenRequest, &token)
	require.NoError(t, err)
	return token
}

func generateOneTimeTokenForRuntime(t *testing.T, ctx context.Context, id string) graphql.OneTimeToken {
	tokenRequest := fixGenerateOneTimeTokenForRuntimeRequest(id)
	token := graphql.OneTimeToken{}
	err := tc.RunOperation(ctx, tokenRequest, &token)
	require.NoError(t, err)
	return token
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
	in, err := tc.graphqlizer.IntegrationSystemInputToGQL(input)
	if err != nil {
		return nil
	}
	req := fixCreateIntegrationSystemRequest(in)
	out := &graphql.IntegrationSystemExt{}
	err = tc.RunOperation(ctx, req, out)
	require.NotEmpty(t, out)
	require.NoError(t, err)
	return out
}

func deleteIntegrationSystem(t *testing.T, ctx context.Context, id string) {
	req := fixDeleteIntegrationSystemRequest(id)
	err := tc.RunOperation(ctx, req, nil)
	require.NoError(t, err)
}

func generateClientCredentialsForIntegrationSystem(t *testing.T, ctx context.Context, id string) graphql.SystemAuth {
	req := fixGenerateClientCredentialsForIntegrationSystemRequest(id)
	out := graphql.SystemAuth{}
	err := tc.RunOperation(ctx, req, &out)
	require.NoError(t, err)
	return out
}

func deleteSystemAuthForIntegrationSystem(t *testing.T, ctx context.Context, id string) {
	req := fixDeleteSystemAuthForIntegrationSystemRequest(id)
	err := tc.RunOperation(ctx, req, nil)
	require.NoError(t, err)
}
