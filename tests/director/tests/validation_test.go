package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Runtime Validation

func TestCreateRuntime_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	invalidInput := graphql.RuntimeInput{
		Name: "0invalid",
	}
	inputString, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixtures.FixRegisterRuntimeRequest(inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.EqualError(t, err, "graphql: Invalid data RuntimeInput [name=cannot start with digit]")
}

func TestUpdateRuntime_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	input := fixtures.FixRuntimeInput("validation-test-rtm")
	rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)

	invalidInput := graphql.RuntimeInput{
		Name: "0invalid",
	}
	inputString, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixtures.FixUpdateRuntimeRequest(rtm.ID, inputString)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.EqualError(t, err, "graphql: Invalid data RuntimeInput [name=cannot start with digit]")
}

// Label Definition Validation

func TestCreateLabelDefinition_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	key := "test_validation_ld"
	ld := fixtures.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, key, map[string]string{"type": "string"}, tenant)
	defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, ld.Key, true, tenant)
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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "validation-test-app", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)

	request := fixtures.FixSetApplicationLabelRequest(app.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key=the length must be no more than 256")
}

func TestSetRuntimeLabel_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	input := fixtures.FixRuntimeInput("validation-test-rtm")
	rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)

	request := fixtures.FixSetRuntimeLabelRequest(rtm.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key=the length must be no more than 256")
}

// Application Validation

const longDescErrMsg = "description=the length must be no more than 2000"

func TestCreateApplication_Validation(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-name", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-name", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenant, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, "integration-system")
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)
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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenant, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenant, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	api := graphql.APIDefinitionInput{Name: "name", TargetURL: "https://kyma-project.io"}
	fixtures.AddAPIToBundleWithInput(t, ctx, dexGraphQLClient, tenant, bndl.ID, api)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenant, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenant, app.ID, "bndl")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	input := fixtures.FixApplicationTemplate("validation-test-app-tpl")
	appTpl := fixtures.CreateApplicationTemplate(t, ctx, dexGraphQLClient, tenant, input)
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenant, appTpl.ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	input := fixtures.FixApplicationTemplate("validation-app")
	tmpl := fixtures.CreateApplicationTemplate(t, ctx, dexGraphQLClient, tenant, input)
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenant, tmpl.ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

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
