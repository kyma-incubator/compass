package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"

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

	runtimeIn := fixRuntimeInput("012345Myaccount_Runtime")
	inputString, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeIn)
	require.NoError(t, err)
	var result graphql.RuntimeExt
	request := fixtures.FixRegisterRuntimeRequest(inputString)
	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &result)

	// THEN
	require.NoError(t, err)
}

func TestCreateRuntime_ValidationFailure(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	runtimeIn := fixRuntimeInput("my runtime")
	inputString, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeIn)
	require.NoError(t, err)
	var result graphql.RuntimeExt
	request := fixtures.FixRegisterRuntimeRequest(inputString)
	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &result)

	// THEN
	require.Error(t, err)
	assert.EqualError(t, err, "graphql: Invalid data RuntimeRegisterInput [name=must be in a valid format]")
}

func TestUpdateRuntime_ValidationSuccess(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixRuntimeInput("validation-test-rtm")

	rtm, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &input)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &rtm)
	require.NoError(t, err)
	require.NotEmpty(t, rtm.ID)

	runtimeIn := fixRuntimeUpdateInput("012345Myaccount_Runtime")
	inputString, err := testctx.Tc.Graphqlizer.RuntimeUpdateInputToGQL(runtimeIn)
	require.NoError(t, err)
	var result graphql.RuntimeExt
	request := fixtures.FixUpdateRuntimeRequest(rtm.ID, inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

	// THEN
	require.NoError(t, err)
}

func TestUpdateRuntime_ValidationFailure(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixRuntimeInput("validation-test-rtm")

	rtm, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &input)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &rtm)
	require.NoError(t, err)
	require.NotEmpty(t, rtm.ID)

	runtimeIn := fixRuntimeUpdateInput("my runtime")
	inputString, err := testctx.Tc.Graphqlizer.RuntimeUpdateInputToGQL(runtimeIn)
	require.NoError(t, err)
	var result graphql.RuntimeExt
	request := fixtures.FixUpdateRuntimeRequest(rtm.ID, inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.EqualError(t, err, "graphql: Invalid data RuntimeUpdateInput [name=must be in a valid format]")
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
	var result graphql.RuntimeExt
	request := fixtures.FixCreateLabelDefinitionRequest(inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key=cannot be blank")
}

func TestUpdateLabelDefinition_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantID := tenant.TestTenants.GetIDByName(t, tenant.ListLabelDefinitionsTenantName)

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)

	defer tenant.TestTenants.CleanupTenant(tenantID)
	invalidInput := graphql.LabelDefinitionInput{
		Key: "",
	}
	inputString, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.RuntimeExt
	request := fixtures.FixUpdateLabelDefinitionRequest(inputString)
	saveExample(t, request.Query(), "update-label-definition")

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key=cannot be blank")
}

// Label Validation

func TestSetApplicationLabel_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "validation-test-app", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	request := fixtures.FixSetApplicationLabelRequest(app.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key=the length must be no more than 256")
}

func TestSetRuntimeLabel_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixRuntimeInput("validation-test-rtm")

	rtm, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &input)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &rtm)
	require.NoError(t, err)
	require.NotEmpty(t, rtm.ID)

	request := fixtures.FixSetRuntimeLabelRequest(rtm.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

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
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createRequest, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), longDescErrMsg)
}

func TestUpdateApplication_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "app-name", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	longDesc := strings.Repeat("a", 2001)
	appUpdate := graphql.ApplicationUpdateInput{ProviderName: str.Ptr("compass"), Description: &longDesc}
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(appUpdate)
	require.NoError(t, err)
	updateRequest := fixtures.FixUpdateApplicationRequest(app.ID, appInputGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateRequest, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), longDescErrMsg)
}

func TestAddDocument_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "app-name", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	doc := fixtures.FixDocumentInput(t)
	doc.DisplayName = strings.Repeat("a", 129)
	docInputGQL, err := testctx.Tc.Graphqlizer.DocumentInputToGQL(&doc)
	require.NoError(t, err)
	createRequest := fixtures.FixAddDocumentToBundleRequest(bndl.ID, docInputGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createRequest, nil)

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
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createRequest, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), longDescErrMsg)
}

func TestUpdateIntegrationSystem_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	longDesc := strings.Repeat("a", 2001)
	intSysUpdate := graphql.IntegrationSystemInput{Name: "name", Description: &longDesc}
	isUpdateGQL, err := testctx.Tc.Graphqlizer.IntegrationSystemInputToGQL(intSysUpdate)
	require.NoError(t, err)
	update := fixtures.FixUpdateIntegrationSystemRequest(intSys.ID, isUpdateGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, update, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), longDescErrMsg)
}

func TestAddAPI_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "name", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	api := graphql.APIDefinitionInput{Name: "name", TargetURL: "https://kyma project.io"}
	apiGQL, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(api)
	require.NoError(t, err)
	addAPIRequest := fixtures.FixAddAPIToBundleRequest(bndl.ID, apiGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addAPIRequest, nil)

	//THEN
	require.Error(t, err)
	require.Contains(t, err.Error(), "targetURL=must be a valid URL")
}

func TestUpdateAPI_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "name", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	api := graphql.APIDefinitionInput{Name: "name", TargetURL: "https://kyma-project.io"}
	fixtures.AddAPIToBundleWithInput(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID, api)

	api.TargetURL = "invalid URL"
	apiGQL, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(api)
	require.NoError(t, err)
	updateAPIRequest := fixtures.FixUpdateAPIRequest(app.ID, apiGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateAPIRequest, nil)

	//THEN
	require.Error(t, err)
	assert.EqualError(t, err, "graphql: Invalid data APIDefinitionInput [targetURL=must be a valid URL]")
}

func TestAddEventAPI_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "name", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	eventAPI := fixtures.FixEventAPIDefinitionInput()
	longDesc := strings.Repeat("a", 2001)
	eventAPI.Description = &longDesc
	evenApiGQL, err := testctx.Tc.Graphqlizer.EventDefinitionInputToGQL(eventAPI)
	require.NoError(t, err)
	addEventAPIRequest := fixtures.FixAddEventAPIToBundleRequest(bndl.ID, evenApiGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addEventAPIRequest, nil)

	//THEN
	require.Error(t, err)
	require.Contains(t, err.Error(), longDescErrMsg)
}

func TestUpdateEventAPI_Validation(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "name", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	eventAPIUpdate := fixtures.FixEventAPIDefinitionInput()
	eventAPI := fixtures.AddEventToBundleWithInput(t, ctx, certSecuredGraphQLClient, bndl.ID, eventAPIUpdate)

	longDesc := strings.Repeat("a", 2001)
	eventAPIUpdate.Description = &longDesc
	evenApiGQL, err := testctx.Tc.Graphqlizer.EventDefinitionInputToGQL(eventAPIUpdate)
	require.NoError(t, err)
	updateEventAPI := fixtures.FixUpdateEventAPIRequest(eventAPI.ID, evenApiGQL)

	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateEventAPI, nil)

	//THEN
	require.Error(t, err)
	require.Contains(t, err.Error(), longDescErrMsg)
}

// Application Template

func TestCreateApplicationTemplate_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appCreateInput := fixtures.FixSampleApplicationRegisterInputWithWebhooks("placeholder")
	invalidInput := fixAppTemplateInputWithDefaultDistinguishLabel("")
	invalidInput.Placeholders = []*graphql.PlaceholderDefinitionInput{}
	invalidInput.ApplicationInput = &appCreateInput
	invalidInput.AccessLevel = graphql.ApplicationTemplateAccessLevelGlobal
	inputString, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.ApplicationTemplate
	request := fixtures.FixCreateApplicationTemplateRequest(inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name=cannot be blank")
}

func TestUpdateApplicationTemplate_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appTemplateName := createAppTemplateName("validation-test-app-tpl")
	input := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)

	appTpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, input)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTpl)
	require.NoError(t, err)
	require.NotEmpty(t, appTpl.ID)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTpl.Labels[tenantfetcher.RegionKey])

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
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name=cannot be blank")
}

func TestRegisterApplicationFromTemplate_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appTemplateName := createAppTemplateName("validation-app")
	input := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)

	tmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, input)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, tmpl)
	require.NoError(t, err)
	require.NotEmpty(t, tmpl.ID)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, tmpl.Labels[tenantfetcher.RegionKey])

	appFromTmpl := graphql.ApplicationFromTemplateInput{}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	registerAppFromTmpl := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, registerAppFromTmpl, nil)

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
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

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
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

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
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

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
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

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
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

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
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "description=cannot be blank")
}
