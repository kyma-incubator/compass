package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
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
	inputString, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := pkg.FixRegisterRuntimeRequest(inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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

	input := pkg.FixRuntimeInput("validation-test-rtm")
	rtm := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)

	invalidInput := graphql.RuntimeInput{
		Name: "0invalid",
	}
	inputString, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := pkg.FixUpdateRuntimeRequest(rtm.ID, inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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
	inputString, err := pkg.Tc.Graphqlizer.LabelDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := pkg.FixCreateLabelDefinitionRequest(inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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
	ld := pkg.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, key, map[string]string{"type": "string"}, tenant)
	defer pkg.DeleteLabelDefinition(t, ctx, dexGraphQLClient, ld.Key, true, tenant)
	invalidInput := graphql.LabelDefinitionInput{
		Key: "",
	}
	inputString, err := pkg.Tc.Graphqlizer.LabelDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := pkg.FixUpdateLabelDefinitionRequest(inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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

	app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "validation-test-app", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)

	request := pkg.FixSetApplicationLabelRequest(app.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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

	input := pkg.FixRuntimeInput("validation-test-rtm")
	rtm := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, rtm.ID)

	request := pkg.FixSetRuntimeLabelRequest(rtm.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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

	app := pkg.FixSampleApplicationRegisterInputWithNameAndWebhooks("placeholder", "name")
	longDesc := strings.Repeat("a", 2001)
	app.Description = &longDesc

	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(app)
	require.NoError(t, err)
	createRequest := pkg.FixRegisterApplicationRequest(appInputGQL)

	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, createRequest, nil)

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

	app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-name", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)

	longDesc := strings.Repeat("a", 2001)
	appUpdate := graphql.ApplicationUpdateInput{ProviderName: str.Ptr("compass"), Description: &longDesc}
	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationUpdateInputToGQL(appUpdate)
	require.NoError(t, err)
	updateRequest := pkg.FixUpdateApplicationRequest(app.ID, appInputGQL)

	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, updateRequest, nil)

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

	app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-name", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, app.ID, "bndl")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	doc := pkg.FixDocumentInput(t)
	doc.DisplayName = strings.Repeat("a", 129)
	docInputGQL, err := pkg.Tc.Graphqlizer.DocumentInputToGQL(&doc)
	require.NoError(t, err)
	createRequest := pkg.FixAddDocumentToBundleRequest(bndl.ID, docInputGQL)

	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, createRequest, nil)

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

	isInputGQL, err := pkg.Tc.Graphqlizer.IntegrationSystemInputToGQL(intSys)
	require.NoError(t, err)
	createRequest := pkg.FixRegisterIntegrationSystemRequest(isInputGQL)

	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, createRequest, nil)

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

	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, "integration-system")
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)
	longDesc := strings.Repeat("a", 2001)
	intSysUpdate := graphql.IntegrationSystemInput{Name: "name", Description: &longDesc}
	isUpdateGQL, err := pkg.Tc.Graphqlizer.IntegrationSystemInputToGQL(intSysUpdate)
	require.NoError(t, err)
	update := pkg.FixUpdateIntegrationSystemRequest(intSys.ID, isUpdateGQL)

	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, update, nil)

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

	app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, app.ID, "bndl")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	api := graphql.APIDefinitionInput{Name: "name", TargetURL: "https://kyma project.io"}
	apiGQL, err := pkg.Tc.Graphqlizer.APIDefinitionInputToGQL(api)
	require.NoError(t, err)
	addAPIRequest := pkg.FixAddAPIToBundleRequest(bndl.ID, apiGQL)

	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, addAPIRequest, nil)

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

	app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, app.ID, "bndl")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	api := graphql.APIDefinitionInput{Name: "name", TargetURL: "https://kyma-project.io"}
	pkg.AddAPIToBundleWithInput(t, ctx, dexGraphQLClient, tenant, bndl.ID, api)

	api.TargetURL = "invalid URL"
	apiGQL, err := pkg.Tc.Graphqlizer.APIDefinitionInputToGQL(api)
	require.NoError(t, err)
	updateAPIRequest := pkg.FixUpdateAPIRequest(app.ID, apiGQL)

	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, updateAPIRequest, nil)

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

	app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, app.ID, "bndl")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	eventAPI := pkg.FixEventAPIDefinitionInput()
	longDesc := strings.Repeat("a", 2001)
	eventAPI.Description = &longDesc
	evenApiGQL, err := pkg.Tc.Graphqlizer.EventDefinitionInputToGQL(eventAPI)
	require.NoError(t, err)
	addEventAPIRequest := pkg.FixAddEventAPIToBundleRequest(bndl.ID, evenApiGQL)

	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, addEventAPIRequest, nil)

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

	app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "name", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, app.ID, "bndl")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	eventAPIUpdate := pkg.FixEventAPIDefinitionInput()
	eventAPI := pkg.AddEventToBundleWithInput(t, ctx, dexGraphQLClient, bndl.ID, eventAPIUpdate)

	longDesc := strings.Repeat("a", 2001)
	eventAPIUpdate.Description = &longDesc
	evenApiGQL, err := pkg.Tc.Graphqlizer.EventDefinitionInputToGQL(eventAPIUpdate)
	require.NoError(t, err)
	updateEventAPI := pkg.FixUpdateEventAPIRequest(eventAPI.ID, evenApiGQL)

	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, updateEventAPI, nil)

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

	appCreateInput := pkg.FixSampleApplicationRegisterInputWithWebhooks("placeholder")
	invalidInput := graphql.ApplicationTemplateInput{
		Name:             "",
		Placeholders:     []*graphql.PlaceholderDefinitionInput{},
		ApplicationInput: &appCreateInput,
		AccessLevel:      graphql.ApplicationTemplateAccessLevelGlobal,
	}
	inputString, err := pkg.Tc.Graphqlizer.ApplicationTemplateInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.ApplicationTemplate
	request := pkg.FixCreateApplicationTemplateRequest(inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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

	input := pkg.FixApplicationTemplate("validation-test-app-tpl")
	appTpl := pkg.CreateApplicationTemplate(t, ctx, dexGraphQLClient, tenant, input)
	defer pkg.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenant, appTpl.ID)

	appCreateInput := pkg.FixSampleApplicationRegisterInputWithWebhooks("placeholder")
	invalidInput := graphql.ApplicationTemplateInput{
		Name:             "",
		Placeholders:     []*graphql.PlaceholderDefinitionInput{},
		ApplicationInput: &appCreateInput,
		AccessLevel:      graphql.ApplicationTemplateAccessLevelGlobal,
	}
	inputString, err := pkg.Tc.Graphqlizer.ApplicationTemplateInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.ApplicationTemplate
	request := pkg.FixUpdateApplicationTemplateRequest(appTpl.ID, inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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

	input := pkg.FixApplicationTemplate("validation-app")
	tmpl := pkg.CreateApplicationTemplate(t, ctx, dexGraphQLClient, tenant, input)
	defer pkg.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenant, tmpl.ID)

	appFromTmpl := graphql.ApplicationFromTemplateInput{}
	appFromTmplGQL, err := pkg.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	registerAppFromTmpl := pkg.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, registerAppFromTmpl, nil)

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
	inputString, err := pkg.Tc.Graphqlizer.BundleCreateInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.BundleExt
	request := pkg.FixAddBundleRequest("", inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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
	inputString, err := pkg.Tc.Graphqlizer.BundleUpdateInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.BundleExt
	request := pkg.FixUpdateBundleRequest("", inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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
	inputString, err := pkg.Tc.Graphqlizer.BundleInstanceAuthSetInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.BundleInstanceAuth
	request := pkg.FixSetBundleInstanceAuthRequest("", inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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
	inputString, err := pkg.Tc.Graphqlizer.APIDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.APIDefinitionExt
	request := pkg.FixAddAPIToBundleRequest("", inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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
	inputString, err := pkg.Tc.Graphqlizer.EventDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.EventAPIDefinitionExt
	request := pkg.FixAddEventAPIToBundleRequest("", inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

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
	inputString, err := pkg.Tc.Graphqlizer.DocumentInputToGQL(&invalidInput)
	require.NoError(t, err)
	var result graphql.DocumentExt
	request := pkg.FixAddDocumentToBundleRequest("", inputString)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "description=cannot be blank")
}
