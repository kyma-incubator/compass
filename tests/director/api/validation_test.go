package api

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Runtime Validation

func TestCreateRuntime_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	invalidInput := graphql.RuntimeInput{
		Name: "0invalid",
	}
	inputString, err := tc.graphqlizer.RuntimeInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixRegisterRuntimeRequest(inputString)

	// WHEN
	err = tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type RuntimeInput")
}

func TestUpdateRuntime_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	rtm := registerRuntime(t, ctx, "validation-test-rtm")
	defer unregisterRuntime(t, rtm.ID)

	invalidInput := graphql.RuntimeInput{
		Name: "0invalid",
	}
	inputString, err := tc.graphqlizer.RuntimeInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixUpdateRuntimeRequest(rtm.ID, inputString)

	// WHEN
	err = tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type RuntimeInput")
}

// Label Definition Validation

func TestCreateLabelDefinition_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	invalidInput := graphql.LabelDefinitionInput{
		Key: "",
	}
	inputString, err := tc.graphqlizer.LabelDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixCreateLabelDefinitionRequest(inputString)

	// WHEN
	err = tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type LabelDefinitionInput")
}

func TestUpdateLabelDefinition_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	key := "test-validation-ld"
	ld := createLabelDefinitionWithinTenant(t, ctx, key, map[string]string{"type": "string"}, defaultTenant)
	defer deleteLabelDefinitionWithinTenant(t, ctx, ld.Key, true, defaultTenant)
	invalidInput := graphql.LabelDefinitionInput{
		Key: "",
	}
	inputString, err := tc.graphqlizer.LabelDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixUpdateLabelDefinitionRequest(inputString)

	// WHEN
	err = tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type LabelDefinitionInput")
}

// Label Validation

func TestSetApplicationLabel_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	app := registerApplication(t, ctx, "validation-test-app")
	defer unregisterApplication(t, app.ID)

	request := fixSetApplicationLabelRequest(app.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err := tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type LabelInput")
}

func TestSetRuntimeLabel_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	rtm := registerRuntime(t, ctx, "validation-test-rtm")
	defer unregisterRuntime(t, rtm.ID)

	request := fixSetRuntimeLabelRequest(rtm.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err := tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type LabelInput")
}

// Application Validation

const longDescErrorMsg = "graphql: validation error for type %s: description: the length must be no more than 2000."

func TestCreateApplication_Validation(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	app := fixSampleApplicationRegisterInputWithName("placeholder", "name")
	longDesc := strings.Repeat("a", 2001)
	app.Description = &longDesc

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(app)
	require.NoError(t, err)
	createRequest := fixRegisterApplicationRequest(appInputGQL)

	//WHEN
	err = tc.RunOperation(ctx, createRequest, nil)

	//THEN
	require.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf(longDescErrorMsg, "ApplicationRegisterInput"))
}

func TestUpdateApplication_Validation(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	app := registerApplication(t, ctx, "app-name")
	defer unregisterApplication(t, app.ID)

	longDesc := strings.Repeat("a", 2001)
	appUpdate := graphql.ApplicationUpdateInput{Name: "name", ProviderName: str.Ptr("compass"), Description: &longDesc}
	appInputGQL, err := tc.graphqlizer.ApplicationUpdateInputToGQL(appUpdate)
	require.NoError(t, err)
	updateRequest := fixUpdateApplicationRequest(app.ID, appInputGQL)

	//WHEN
	err = tc.RunOperation(ctx, updateRequest, nil)

	//THEN
	require.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf(longDescErrorMsg, "ApplicationUpdateInput"))
}

func TestAddDocument_Validation(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	app := registerApplication(t, ctx, "app-name")
	defer unregisterApplication(t, app.ID)

	doc := fixDocumentInput()
	doc.DisplayName = strings.Repeat("a", 129)
	docInputGQL, err := tc.graphqlizer.DocumentInputToGQL(&doc)
	require.NoError(t, err)
	createRequest := fixAddDocumentRequest(app.ID, docInputGQL)

	//WHEN
	err = tc.RunOperation(ctx, createRequest, nil)

	//THEN
	require.Error(t, err)
	assert.EqualError(t, err, "graphql: validation error for type DocumentInput: displayName: the length must be between 1 and 128.")
}

func TestCreateIntegrationSystem_Validation(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	intSys := graphql.IntegrationSystemInput{Name: "valid-name"}
	longDesc := strings.Repeat("a", 2001)
	intSys.Description = &longDesc

	isInputGQL, err := tc.graphqlizer.IntegrationSystemInputToGQL(intSys)
	require.NoError(t, err)
	createRequest := fixRegisterIntegrationSystemRequest(isInputGQL)

	//WHEN
	err = tc.RunOperation(ctx, createRequest, nil)

	//THEN
	require.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf(longDescErrorMsg, "IntegrationSystemInput"))
}

func TestUpdateIntegrationSystem_Validation(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	intSys := registerIntegrationSystem(t, ctx, "integration-system")
	defer unregisterIntegrationSystem(t, ctx, intSys.ID)
	longDesc := strings.Repeat("a", 2001)
	intSysUpdate := graphql.IntegrationSystemInput{Name: "name", Description: &longDesc}
	isUpdateGQL, err := tc.graphqlizer.IntegrationSystemInputToGQL(intSysUpdate)
	require.NoError(t, err)
	update := fixUpdateIntegrationSystemRequest(intSys.ID, isUpdateGQL)

	//WHEN
	err = tc.RunOperation(ctx, update, nil)

	//THEN
	require.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf(longDescErrorMsg, "IntegrationSystemInput"))
}

func TestAddAPI_Validation(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	app := registerApplication(t, ctx, "name")
	defer unregisterApplication(t, app.ID)

	api := graphql.APIDefinitionInput{Name: "name", TargetURL: "https://kyma project.io"}
	apiGQL, err := tc.graphqlizer.APIDefinitionInputToGQL(api)
	require.NoError(t, err)
	addAPIRequest := fixAddAPIRequest(app.ID, apiGQL)

	//WHEN
	err = tc.RunOperation(ctx, addAPIRequest, nil)

	//THEN
	require.Error(t, err)
	require.EqualError(t, err, "graphql: validation error for type APIDefinitionInput: targetURL: must be a valid URL.")
}

func TestUpdateAPI_Validation(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	app := registerApplication(t, ctx, "name")
	defer unregisterApplication(t, app.ID)

	api := graphql.APIDefinitionInput{Name: "name", TargetURL: "https://kyma-project.io"}
	addAPI(t, ctx, app.ID, api)

	api.TargetURL = "invalid URL"
	apiGQL, err := tc.graphqlizer.APIDefinitionInputToGQL(api)
	require.NoError(t, err)
	updateAPIRequest := fixUpdateAPIRequest(app.ID, apiGQL)

	//WHEN
	err = tc.RunOperation(ctx, updateAPIRequest, nil)

	//THEN
	require.Error(t, err)
	require.EqualError(t, err, "graphql: validation error for type APIDefinitionInput: targetURL: is not valid URL.")
}

func TestAddEventAPI_Validation(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	app := registerApplication(t, ctx, "name")
	defer unregisterApplication(t, app.ID)

	eventAPI := fixEventAPIDefinitionInput()
	longDesc := strings.Repeat("a", 2001)
	eventAPI.Description = &longDesc
	evenApiGQL, err := tc.graphqlizer.EventDefinitionInputToGQL(eventAPI)
	require.NoError(t, err)
	addEventAPIRequest := fixAddEventAPIRequest(app.ID, evenApiGQL)

	//WHEN
	err = tc.RunOperation(ctx, addEventAPIRequest, nil)

	//THEN
	require.Error(t, err)
	require.EqualError(t, err, fmt.Sprintf(longDescErrorMsg, "EventDefinitionInput"))
}

func TestUpdateEventAPI_Validation(t *testing.T) {
	ctx := context.TODO()
	app := registerApplication(t, ctx, "name")
	defer unregisterApplication(t, app.ID)

	eventAPIUpdate := fixEventAPIDefinitionInput()
	eventAPI := addEventDefinition(t, ctx, app.ID, eventAPIUpdate)

	longDesc := strings.Repeat("a", 2001)
	eventAPIUpdate.Description = &longDesc
	evenApiGQL, err := tc.graphqlizer.EventDefinitionInputToGQL(eventAPIUpdate)
	require.NoError(t, err)
	updateEventAPI := fixUpdateEventAPIRequest(eventAPI.ID, evenApiGQL)

	//WHEN
	err = tc.RunOperation(ctx, updateEventAPI, nil)

	//THEN
	require.Error(t, err)
	require.EqualError(t, err, fmt.Sprintf(longDescErrorMsg, "EventDefinitionInput"))
}

func fixEventAPIDefinitionInput() graphql.EventDefinitionInput {
	data := graphql.CLOB("data")
	return graphql.EventDefinitionInput{Name: "name",
		Spec: &graphql.EventSpecInput{
			Data:   &data,
			Type:   graphql.EventSpecTypeAsyncAPI,
			Format: graphql.SpecFormatJSON,
		}}
}

func fixAPIDefinitionInput() graphql.APIDefinitionInput {
	return graphql.APIDefinitionInput{
		Name:      "new-api-name",
		TargetURL: "https://target.url",
		Spec: &graphql.APISpecInput{
			Format: graphql.SpecFormatJSON,
			Type:   graphql.APISpecTypeOpenAPI,
			FetchRequest: &graphql.FetchRequestInput{
				URL: "https://foo.bar",
			},
		},
	}

}

// Application Template

func TestCreateApplicationTemplate_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appCreateInput := fixSampleApplicationRegisterInput("placeholder")
	invalidInput := graphql.ApplicationTemplateInput{
		Name:             "0invalid",
		Placeholders:     []*graphql.PlaceholderDefinitionInput{},
		ApplicationInput: &appCreateInput,
		AccessLevel:      graphql.ApplicationTemplateAccessLevelGlobal,
	}
	inputString, err := tc.graphqlizer.ApplicationTemplateInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.ApplicationTemplate
	request := fixCreateApplicationTemplateRequest(inputString)

	// WHEN
	err = tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type ApplicationTemplateInput")
}

func TestUpdateApplicationTemplate_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTpl := createApplicationTemplate(t, ctx, "validation-test-app-tpl")
	defer deleteApplicationTemplate(t, ctx, appTpl.ID)

	appCreateInput := fixSampleApplicationRegisterInput("placeholder")
	invalidInput := graphql.ApplicationTemplateInput{
		Name:             "0invalid",
		Placeholders:     []*graphql.PlaceholderDefinitionInput{},
		ApplicationInput: &appCreateInput,
		AccessLevel:      graphql.ApplicationTemplateAccessLevelGlobal,
	}
	inputString, err := tc.graphqlizer.ApplicationTemplateInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.ApplicationTemplate
	request := fixUpdateApplicationTemplateRequest(appTpl.ID, inputString)

	// WHEN
	err = tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type ApplicationTemplateInput")
}

func TestRegisterApplicationFromTemplate_Validation(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	tmpl := createApplicationTemplate(t, ctx, "validation-app")
	defer deleteApplicationTemplate(t, ctx, tmpl.ID)

	appFromTmpl := graphql.ApplicationFromTemplateInput{}
	appFromTmplGQL, err := tc.graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	registerAppFromTmpl := fixRegisterApplicationFromTemplate(appFromTmplGQL)
	//WHEN
	err = tc.RunOperation(ctx, registerAppFromTmpl, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TemplateName: cannot be blank.")
}

func fixDocumentInput() graphql.DocumentInput {
	return graphql.DocumentInput{
		Title:       "Readme",
		Description: "Detailed description of project",
		Format:      graphql.DocumentFormatMarkdown,
		DisplayName: "display-name",
		FetchRequest: &graphql.FetchRequestInput{
			URL:    "kyma-project.io",
			Mode:   ptr.FetchMode(graphql.FetchModePackage),
			Filter: ptr.String("/docs/README.md"),
			Auth:   fixBasicAuth(),
		},
	}
}
