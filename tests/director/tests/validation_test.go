package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Runtime Validation

func TestCreateRuntime_ValidationSuccess(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	runtimeIn := graphql.RuntimeInput{
		Name: "012345Myaccount_Runtime",
	}
	inputString, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeIn)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixtures.FixRegisterRuntimeRequest(inputString)
	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, result.ID)

	// THEN
	require.NoError(t, err)
}

func TestCreateRuntime_ValidationFailure(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	runtimeIn := graphql.RuntimeInput{
		Name: "012345Myaccount_Runtime_aaaaaaaaaaaaаа",
	}
	inputString, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeIn)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixtures.FixRegisterRuntimeRequest(inputString)
	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)
	if err == nil {
		defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, result.ID)
	}

	// THEN
	require.Error(t, err)
	assert.EqualError(t, err, "graphql: Invalid data RuntimeInput [name=the length must be between 1 and 36]")
}

func TestUpdateRuntime_ValidationSuccess(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixtures.FixRuntimeInput("validation-test-rtm")
	rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)

	runtimeIn := graphql.RuntimeInput{
		Name: "012345Myaccount_Runtime",
	}
	inputString, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeIn)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixtures.FixUpdateRuntimeRequest(rtm.ID, inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.NoError(t, err)
}

func TestUpdateRuntime_ValidationFailure(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixtures.FixRuntimeInput("validation-test-rtm")
	rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)

	runtimeIn := graphql.RuntimeInput{
		Name: "012345Myaccount_Runtime_aaaaaaaaaaaaаа",
	}
	inputString, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeIn)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixtures.FixUpdateRuntimeRequest(rtm.ID, inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.EqualError(t, err, "graphql: Invalid data RuntimeInput [name=the length must be between 1 and 36]")
}

// Label Definition Validation

func TestCreateLabelDefinition_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	invalidInput := graphql.LabelDefinitionInput{
		Key: "",
	}
	inputString, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixtures.FixCreateLabelDefinitionRequest(inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key=cannot be blank")
}

func TestUpdateLabelDefinition_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	key := "test_validation_ld"
	ld := fixtures.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, key, map[string]string{"type": "string"}, tenantId)
	defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, ld.Key, true, tenantId)
	invalidInput := graphql.LabelDefinitionInput{
		Key: "",
	}
	inputString, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixtures.FixUpdateLabelDefinitionRequest(inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key=cannot be blank")
}

// Label Validation

func TestSetApplicationLabel_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "validation-test-app", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

	request := fixtures.FixSetApplicationLabelRequest(app.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key=the length must be no more than 256")
}

func TestSetRuntimeLabel_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixtures.FixRuntimeInput("validation-test-rtm")
	rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)

	request := fixtures.FixSetRuntimeLabelRequest(rtm.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key=the length must be no more than 256")
}

// Application Validation

const longDescErrMsg = "description=the length must be no more than 2000"

func TestCreateApplication_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	app := fixtures.FixSampleApplicationRegisterInputWithNameAndWebhooks("placeholder", "name")
	longDesc := strings.Repeat("a", 2001)
	app.Description = &longDesc

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(app)
	require.NoError(t, err)
	createRequest := fixtures.FixRegisterApplicationRequest(appInputGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createRequest, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), longDescErrMsg)
}

func TestUpdateApplication_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-name", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

	longDesc := strings.Repeat("a", 2001)
	appUpdate := graphql.ApplicationUpdateInput{ProviderName: str.Ptr("compass"), Description: &longDesc}
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(appUpdate)
	require.NoError(t, err)
	updateRequest := fixtures.FixUpdateApplicationRequest(app.ID, appInputGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateRequest, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), longDescErrMsg)
}

func TestAddDocument_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-name", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	doc := fixtures.FixDocumentInput(t)
	doc.DisplayName = strings.Repeat("a", 129)
	docInputGQL, err := testctx.Tc.Graphqlizer.DocumentInputToGQL(&doc)
	require.NoError(t, err)
	createRequest := fixtures.FixAddDocumentToBundleRequest(bndl.ID, docInputGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createRequest, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "displayName=the length must be between 1 and 128")
}

func TestCreateIntegrationSystem_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	intSys := graphql.IntegrationSystemInput{Name: "valid-name"}
	longDesc := strings.Repeat("a", 2001)
	intSys.Description = &longDesc

	isInputGQL, err := testctx.Tc.Graphqlizer.IntegrationSystemInputToGQL(intSys)
	require.NoError(t, err)
	createRequest := fixtures.FixRegisterIntegrationSystemRequest(isInputGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createRequest, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), longDescErrMsg)
}

func TestUpdateIntegrationSystem_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, "integration-system")
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys.ID)
	longDesc := strings.Repeat("a", 2001)
	intSysUpdate := graphql.IntegrationSystemInput{Name: "name", Description: &longDesc}
	isUpdateGQL, err := testctx.Tc.Graphqlizer.IntegrationSystemInputToGQL(intSysUpdate)
	require.NoError(t, err)
	update := fixtures.FixUpdateIntegrationSystemRequest(intSys.ID, isUpdateGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, update, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), longDescErrMsg)
}

func TestAddAPI_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	api := graphql.APIDefinitionInput{Name: "name", TargetURL: "https://kyma project.io"}
	apiGQL, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(api)
	require.NoError(t, err)
	addAPIRequest := fixtures.FixAddAPIToBundleRequest(bndl.ID, apiGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, addAPIRequest, nil)

	//THEN
	require.Error(t, err)
	require.Contains(t, err.Error(), "targetURL=must be a valid URL")
}

func TestUpdateAPI_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	api := graphql.APIDefinitionInput{Name: "name", TargetURL: "https://kyma-project.io"}
	fixtures.AddAPIToBundleWithInput(t, ctx, dexGraphQLClient, tenantId, bndl.ID, api)

	api.TargetURL = "invalid URL"
	apiGQL, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(api)
	require.NoError(t, err)
	updateAPIRequest := fixtures.FixUpdateAPIRequest(app.ID, apiGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateAPIRequest, nil)

	//THEN
	require.Error(t, err)
	assert.EqualError(t, err, "graphql: Invalid data APIDefinitionInput [targetURL=must be a valid URL]")
}

func TestAddEventAPI_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	eventAPI := fixtures.FixEventAPIDefinitionInput()
	longDesc := strings.Repeat("a", 2001)
	eventAPI.Description = &longDesc
	evenApiGQL, err := testctx.Tc.Graphqlizer.EventDefinitionInputToGQL(eventAPI)
	require.NoError(t, err)
	addEventAPIRequest := fixtures.FixAddEventAPIToBundleRequest(bndl.ID, evenApiGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, addEventAPIRequest, nil)

	//THEN
	require.Error(t, err)
	require.Contains(t, err.Error(), longDescErrMsg)
}

func TestUpdateEventAPI_Validation(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	eventAPIUpdate := fixtures.FixEventAPIDefinitionInput()
	eventAPI := fixtures.AddEventToBundleWithInput(t, ctx, dexGraphQLClient, bndl.ID, eventAPIUpdate)

	longDesc := strings.Repeat("a", 2001)
	eventAPIUpdate.Description = &longDesc
	evenApiGQL, err := testctx.Tc.Graphqlizer.EventDefinitionInputToGQL(eventAPIUpdate)
	require.NoError(t, err)
	updateEventAPI := fixtures.FixUpdateEventAPIRequest(eventAPI.ID, evenApiGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateEventAPI, nil)

	//THEN
	require.Error(t, err)
	require.Contains(t, err.Error(), longDescErrMsg)
}

// Application Template

func TestCreateApplicationTemplate_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appCreateInput := fixtures.FixSampleApplicationRegisterInputWithWebhooks("placeholder")
	invalidInput := graphql.ApplicationTemplateInput{
		Name:             "",
		Placeholders:     []*graphql.PlaceholderDefinitionInput{},
		ApplicationInput: &appCreateInput,
		AccessLevel:      graphql.ApplicationTemplateAccessLevelGlobal,
	}
	inputString, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.ApplicationTemplate
	request := fixtures.FixCreateApplicationTemplateRequest(inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name=cannot be blank")
}

func TestUpdateApplicationTemplate_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixtures.FixApplicationTemplate("validation-test-app-tpl")
	appTpl := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenantId, input)
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, appTpl.ID)

	appCreateInput := fixtures.FixSampleApplicationRegisterInputWithWebhooks("placeholder")
	invalidInput := graphql.ApplicationTemplateInput{
		Name:             "",
		Placeholders:     []*graphql.PlaceholderDefinitionInput{},
		ApplicationInput: &appCreateInput,
		AccessLevel:      graphql.ApplicationTemplateAccessLevelGlobal,
	}
	inputString, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.ApplicationTemplate
	request := fixtures.FixUpdateApplicationTemplateRequest(appTpl.ID, inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name=cannot be blank")
}

func TestRegisterApplicationFromTemplate_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixtures.FixApplicationTemplate("validation-app")
	tmpl := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenantId, input)
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, tmpl.ID)

	appFromTmpl := graphql.ApplicationFromTemplateInput{}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	registerAppFromTmpl := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerAppFromTmpl, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "templateName=cannot be blank")
}

// BUNDLE API

func TestAddBundle_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	invalidInput := graphql.BundleCreateInput{}
	inputString, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.BundleExt
	request := fixtures.FixAddBundleRequest("", inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name=cannot be blank")
}

func TestUpdateBundle_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	invalidInput := graphql.BundleUpdateInput{}
	inputString, err := testctx.Tc.Graphqlizer.BundleUpdateInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.BundleExt
	request := fixtures.FixUpdateBundleRequest("", inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name=cannot be blank")
}

func TestSetBundleInstanceAuth_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	invalidInput := graphql.BundleInstanceAuthSetInput{}
	inputString, err := testctx.Tc.Graphqlizer.BundleInstanceAuthSetInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.BundleInstanceAuth
	request := fixtures.FixSetBundleInstanceAuthRequest("", inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reason=at least one field (Auth or Status) has to be provided")
}

func TestAddAPIDefinitionToBundle_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	invalidInput := graphql.APIDefinitionInput{}
	inputString, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.APIDefinitionExt
	request := fixtures.FixAddAPIToBundleRequest("", inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name=cannot be blank")
}

func TestAddEventDefinitionToBundle_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	invalidInput := graphql.EventDefinitionInput{}
	inputString, err := testctx.Tc.Graphqlizer.EventDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.EventAPIDefinitionExt
	request := fixtures.FixAddEventAPIToBundleRequest("", inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name=cannot be blank")
}

func TestAddDocumentToBundle_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	invalidInput := graphql.DocumentInput{
		Format: graphql.DocumentFormatMarkdown,
	}
	inputString, err := testctx.Tc.Graphqlizer.DocumentInputToGQL(&invalidInput)
	require.NoError(t, err)
	var result graphql.DocumentExt
	request := fixtures.FixAddDocumentToBundleRequest("", inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "description=cannot be blank")
}
