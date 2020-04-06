package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func SetAutomaticScenarioAssignmentWithinTenant(t *testing.T, ctx context.Context, tenantID, automaticScenario, selecterKey, selectorValue string) {
	input := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: automaticScenario,
		Selector: &graphql.LabelSelectorInput{
			Key:   selecterKey,
			Value: selectorValue,
		},
	}
	payload, err := tc.graphqlizer.AutomaticScenarioAssignmentSetInputToGQL(input)
	require.NoError(t, err)
	request := fixCreateAutomaticScenarioAssignmentRequest(payload)
	err = tc.RunOperationWithCustomTenant(ctx, tenantID, request, nil)
	require.NoError(t, err)
}

//Application
func getApplication(t *testing.T, ctx context.Context, id string) graphql.ApplicationExt {
	appRequest := fixApplicationRequest(id)
	app := graphql.ApplicationExt{}
	require.NoError(t, tc.RunOperation(ctx, appRequest, &app))
	return app
}

func registerApplication(t *testing.T, ctx context.Context, name string) graphql.ApplicationExt {
	in := fixSampleApplicationRegisterInputWithName("first", name)
	return registerApplicationFromInputWithinTenant(t, ctx, in, testTenants.GetDefaultTenantID())
}

func registerApplicationFromInputWithinTenant(t *testing.T, ctx context.Context, in graphql.ApplicationRegisterInput, tenantID string) graphql.ApplicationExt {
	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixRegisterApplicationRequest(appInputGQL)

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
	req := fixRequestClientCredentialsForApplication(id)
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

func createApplicationTemplateFromInput(t *testing.T, ctx context.Context, input graphql.ApplicationTemplateInput) graphql.ApplicationTemplate {
	appTemplate, err := tc.graphqlizer.ApplicationTemplateInputToGQL(input)
	require.NoError(t, err)

	createApplicationTemplateRequest := fixCreateApplicationTemplateRequest(appTemplate)
	output := graphql.ApplicationTemplate{}

	err = tc.RunOperation(ctx, createApplicationTemplateRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	return output
}

func createApplicationTemplate(t *testing.T, ctx context.Context, name string) graphql.ApplicationTemplate {
	return createApplicationTemplateFromInput(t, ctx, fixApplicationTemplate(name))
}

func deleteApplicationTemplate(t *testing.T, ctx context.Context, id string) {
	req := fixDeleteApplicationTemplate(id)
	err := tc.RunOperation(ctx, req, nil)
	require.NoError(t, err)
}

//Runtime
func registerRuntime(t *testing.T, ctx context.Context, placeholder string) *graphql.RuntimeExt {
	input := fixRuntimeInput(placeholder)
	return registerRuntimeFromInput(t, ctx, &input)
}

func registerRuntimeFromInput(t *testing.T, ctx context.Context, input *graphql.RuntimeInput) *graphql.RuntimeExt {
	return registerRuntimeFromInputWithinTenant(t, ctx, input, testTenants.GetDefaultTenantID())
}

func registerRuntimeFromInputWithinTenant(t *testing.T, ctx context.Context, input *graphql.RuntimeInput, tenant string) *graphql.RuntimeExt {
	inputGQL, err := tc.graphqlizer.RuntimeInputToGQL(*input)
	require.NoError(t, err)

	createRequest := fixRegisterRuntimeRequest(inputGQL)
	var runtime graphql.RuntimeExt

	err = tc.RunOperationWithCustomTenant(ctx, tenant, createRequest, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)
	return &runtime
}

func getRuntime(t *testing.T, ctx context.Context, runtimeID string) *graphql.RuntimeExt {
	return getRuntimeWithinTenant(t, ctx, runtimeID, testTenants.GetDefaultTenantID())
}

func getRuntimeWithinTenant(t *testing.T, ctx context.Context, runtimeID string, tenant string) *graphql.RuntimeExt {
	var runtime graphql.RuntimeExt
	runtimeQuery := fixRuntimeRequest(runtimeID)

	err := tc.RunOperationWithCustomTenant(ctx, tenant, runtimeQuery, &runtime)
	require.NoError(t, err)
	return &runtime
}

func setRuntimeLabel(t *testing.T, ctx context.Context, runtimeID string, labelKey string, labelValue interface{}) *graphql.Label {
	return setRuntimeLabelWithinTenant(t, ctx, testTenants.GetDefaultTenantID(), runtimeID, labelKey, labelValue)
}

func setRuntimeLabelWithinTenant(t *testing.T, ctx context.Context, tenantID, runtimeID string, labelKey string, labelValue interface{}) *graphql.Label {
	setLabelRequest := fixSetRuntimeLabelRequest(runtimeID, labelKey, labelValue)
	label := graphql.Label{}

	err := tc.RunOperationWithCustomTenant(ctx, tenantID, setLabelRequest, &label)
	require.NoError(t, err)

	return &label
}

func unregisterRuntime(t *testing.T, id string) {
	unregisterRuntimeWithinTenant(t, id, testTenants.GetDefaultTenantID())
}

func unregisterRuntimeWithinTenant(t *testing.T, id string, tenantID string) {
	delReq := fixUnregisterRuntimeRequest(id)

	err := tc.RunOperationWithCustomTenant(context.Background(), tenantID, delReq, nil)
	require.NoError(t, err)
}

func generateClientCredentialsForRuntime(t *testing.T, ctx context.Context, id string) graphql.SystemAuth {
	req := fixRequestClientCredentialsForRuntime(id)
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
		Schema: marshalJSONSchema(t, schema),
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

func createScenariosLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, tenantID string, scenarios []string) *graphql.LabelDefinition {
	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": scenarios,
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}

	return createLabelDefinitionWithinTenant(t, ctx, "scenarios", jsonSchema, tenantID)
}

func deleteLabelDefinition(t *testing.T, ctx context.Context, labelDefinitionKey string, deleteRelatedResources bool) {
	deleteLabelDefinitionWithinTenant(t, ctx, labelDefinitionKey, deleteRelatedResources, testTenants.GetDefaultTenantID())
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

func addEventDefinition(t *testing.T, ctx context.Context, appID string, apiInput graphql.EventDefinitionInput) graphql.EventDefinition {
	eventAPIGQL, err := tc.graphqlizer.EventDefinitionInputToGQL(apiInput)
	require.NoError(t, err)
	addAPIRequest := fixAddEventAPIRequest(appID, eventAPIGQL)
	//WHEN
	outEventAPI := graphql.EventDefinition{}
	err = tc.RunOperation(ctx, addAPIRequest, &outEventAPI)
	require.NoError(t, err)
	return outEventAPI
}

//OneTimeToken

func requestOneTimeTokenForApplication(t *testing.T, ctx context.Context, id string) graphql.OneTimeTokenForApplicationExt {
	tokenRequest := fixRequestOneTimeTokenForApp(id)
	token := graphql.OneTimeTokenForApplicationExt{}
	err := tc.RunOperation(ctx, tokenRequest, &token)
	require.NoError(t, err)
	return token
}

func requestOneTimeTokenForRuntime(t *testing.T, ctx context.Context, id string) graphql.OneTimeTokenForRuntimeExt {
	tokenRequest := fixRequestOneTimeTokenForRuntime(id)
	token := graphql.OneTimeTokenForRuntimeExt{}
	err := tc.RunOperation(ctx, tokenRequest, &token)
	require.NoError(t, err)
	return token
}

// Integration System
func getIntegrationSystem(t *testing.T, ctx context.Context, id string) *graphql.IntegrationSystemExt {
	intSysRequest := fixIntegrationSystemRequest(id)
	intSys := graphql.IntegrationSystemExt{}
	require.NoError(t, tc.RunOperationWithoutTenant(ctx, intSysRequest, &intSys))
	return &intSys
}

func registerIntegrationSystem(t *testing.T, ctx context.Context, name string) *graphql.IntegrationSystemExt {
	input := graphql.IntegrationSystemInput{Name: name}
	in, err := tc.graphqlizer.IntegrationSystemInputToGQL(input)
	if err != nil {
		return nil
	}
	req := fixRegisterIntegrationSystemRequest(in)
	out := &graphql.IntegrationSystemExt{}
	err = tc.RunOperationWithoutTenant(ctx, req, out)
	require.NotEmpty(t, out)
	require.NoError(t, err)
	return out
}

func unregisterIntegrationSystem(t *testing.T, ctx context.Context, id string) {
	req := fixunregisterIntegrationSystem(id)
	err := tc.RunOperationWithoutTenant(ctx, req, nil)
	require.NoError(t, err)
}

func requestClientCredentialsForIntegrationSystem(t *testing.T, ctx context.Context, id string) graphql.SystemAuth {
	req := fixRequestClientCredentialsForIntegrationSystem(id)
	out := graphql.SystemAuth{}
	err := tc.RunOperationWithoutTenant(ctx, req, &out)
	require.NoError(t, err)
	return out
}

func deleteSystemAuthForIntegrationSystem(t *testing.T, ctx context.Context, id string) {
	req := fixDeleteSystemAuthForIntegrationSystemRequest(id)
	err := tc.RunOperationWithoutTenant(ctx, req, nil)
	require.NoError(t, err)
}

func setDefaultEventingForApplication(t *testing.T, ctx context.Context, appID string, runtimeID string) {
	req := fixSetDefaultEventingForApplication(appID, runtimeID)
	err := tc.RunOperation(ctx, req, nil)
	require.NoError(t, err)
}

func createPackage(t *testing.T, ctx context.Context, appID, pkgName string) graphql.PackageExt {
	in, err := tc.graphqlizer.PackageCreateInputToGQL(fixPackageCreateInput(pkgName))
	require.NoError(t, err)

	req := fixAddPackageRequest(appID, in)
	var resp graphql.PackageExt

	err = tc.RunOperation(ctx, req, &resp)
	require.NoError(t, err)

	return resp
}

func getPackage(t *testing.T, ctx context.Context, appID, pkgID string) graphql.PackageExt {
	req := fixPackageRequest(appID, pkgID)
	pkg := graphql.ApplicationExt{}
	require.NoError(t, tc.RunOperation(ctx, req, &pkg))
	return pkg.Package
}

func deletePackage(t *testing.T, ctx context.Context, id string) {
	req := fixDeletePackageRequest(id)

	require.NoError(t, tc.RunOperation(ctx, req, nil))
}

func addAPIToPackageWithInput(t *testing.T, ctx context.Context, pkgID string, input graphql.APIDefinitionInput) graphql.APIDefinitionExt {
	inStr, err := tc.graphqlizer.APIDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualApi := graphql.APIDefinitionExt{}
	req := fixAddAPIToPackageRequest(pkgID, inStr)
	err = tc.RunOperation(ctx, req, &actualApi)
	require.NoError(t, err)
	return actualApi
}

func addAPIToPackage(t *testing.T, ctx context.Context, pkgID string) graphql.APIDefinitionExt {
	return addAPIToPackageWithInput(t, ctx, pkgID, fixAPIDefinitionInput())
}

func addEventToPackageWithInput(t *testing.T, ctx context.Context, pkgID string, input graphql.EventDefinitionInput) graphql.EventDefinition {
	inStr, err := tc.graphqlizer.EventDefinitionInputToGQL(input)
	require.NoError(t, err)

	event := graphql.EventDefinition{}
	req := fixAddEventAPIToPackageRequest(pkgID, inStr)
	err = tc.RunOperation(ctx, req, &event)
	require.NoError(t, err)
	return event
}

func addEventToPackage(t *testing.T, ctx context.Context, pkgID string) graphql.EventDefinition {
	return addEventToPackageWithInput(t, ctx, pkgID, fixEventAPIDefinitionInput())
}

func addDocumentToPackageWithInput(t *testing.T, ctx context.Context, pkgID string, input graphql.DocumentInput) graphql.DocumentExt {
	inStr, err := tc.graphqlizer.DocumentInputToGQL(&input)
	require.NoError(t, err)

	actualDoc := graphql.DocumentExt{}
	req := fixAddDocumentToPackageRequest(pkgID, inStr)
	err = tc.RunOperation(ctx, req, &actualDoc)
	require.NoError(t, err)
	return actualDoc
}

func addDocumentToPackage(t *testing.T, ctx context.Context, pkgID string) graphql.DocumentExt {
	return addDocumentToPackageWithInput(t, ctx, pkgID, fixDocumentInput())
}

func createPackageInstanceAuth(t *testing.T, ctx context.Context, pkgID string) graphql.PackageInstanceAuth {
	authCtx, inputParams := fixPackageInstanceAuthContextAndInputParams(t)
	in, err := tc.graphqlizer.PackageInstanceAuthRequestInputToGQL(fixPackageInstanceAuthRequestInput(authCtx, inputParams))
	require.NoError(t, err)

	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestPackageInstanceAuthCreation(packageID: "%s", in: %s) {
				id
			}}`, pkgID, in))

	var resp graphql.PackageInstanceAuth

	err = tc.RunOperation(ctx, req, &resp)
	require.NoError(t, err)

	return resp
}

func fixPackageInstanceAuthContextAndInputParams(t *testing.T) (*graphql.JSON, *graphql.JSON) {
	authCtxPayload := map[string]interface{}{
		"ContextData": "ContextValue",
	}
	var authCtxData interface{} = authCtxPayload

	inputParamsPayload := map[string]interface{}{
		"InKey": "InValue",
	}
	var inputParamsData interface{} = inputParamsPayload

	authCtx := marshalJSON(t, authCtxData)
	inputParams := marshalJSON(t, inputParamsData)

	return authCtx, inputParams
}

func createAutomaticScenarioAssignmentFromInputWithinTenant(t *testing.T, ctx context.Context, input graphql.AutomaticScenarioAssignmentSetInput, tenantID string) graphql.AutomaticScenarioAssignment {
	inStr, err := tc.graphqlizer.AutomaticScenarioAssignmentSetInputToGQL(input)
	require.NoError(t, err)

	assignment := graphql.AutomaticScenarioAssignment{}
	req := fixCreateAutomaticScenarioAssignmentRequest(inStr)
	err = tc.RunOperationWithCustomTenant(ctx, tenantID, req, &assignment)
	require.NoError(t, err)
	return assignment
}

func createAutomaticScenarioAssignmentInTenant(t *testing.T, ctx context.Context, in graphql.AutomaticScenarioAssignmentSetInput, tenantID string) *graphql.AutomaticScenarioAssignment {
	assignmentInput, err := tc.graphqlizer.AutomaticScenarioAssignmentSetInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixCreateAutomaticScenarioAssignmentRequest(assignmentInput)

	assignment := graphql.AutomaticScenarioAssignment{}

	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, tenantID, createRequest, &assignment))
	require.NotEmpty(t, assignment.ScenarioName)
	return &assignment
}

func listAutomaticScenarioAssignmentsWithinTenant(t *testing.T, ctx context.Context, tenantID string) graphql.AutomaticScenarioAssignmentPage {
	assignmentsPage := graphql.AutomaticScenarioAssignmentPage{}
	req := fixAutomaticScenarioAssignmentsRequest()
	err := tc.RunOperationWithCustomTenant(ctx, tenantID, req, &assignmentsPage)
	require.NoError(t, err)
	return assignmentsPage
}

func deleteAutomaticScenarioAssignmentForScenarioWithinTenant(t *testing.T, ctx context.Context, tenantID, scenarioName string) graphql.AutomaticScenarioAssignment {
	assignment := graphql.AutomaticScenarioAssignment{}
	req := fixDeleteAutomaticScenarioAssignmentForScenarioRequest(scenarioName)
	err := tc.RunOperationWithCustomTenant(ctx, tenantID, req, &assignment)
	require.NoError(t, err)
	return assignment
}
