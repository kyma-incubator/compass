package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/json"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func TestCreateLabel(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "label-without-label-def"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, name, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	t.Log("Set label on application")
	labelKey := "test"
	labelValue := "val"

	setLabelRequest := fixtures.FixSetApplicationLabelRequest(application.ID, labelKey, labelValue)
	label := graphql.Label{}

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, setLabelRequest, &label)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, label.Key)
	require.NotEmpty(t, label.Value)
	require.Equal(t, labelKey, label.Key)
	require.Equal(t, labelValue, label.Value)
	saveExample(t, setLabelRequest.Query(), "set application label")

	t.Log("Update label value on application")
	newLabelValue := "new-val"

	setLabelRequest = fixtures.FixSetApplicationLabelRequest(application.ID, labelKey, newLabelValue)
	updatedLabel := graphql.Label{}

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, setLabelRequest, &updatedLabel)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, updatedLabel.Key)
	require.NotEmpty(t, updatedLabel.Value)
	require.Equal(t, labelKey, updatedLabel.Key)
	require.Equal(t, newLabelValue, updatedLabel.Value)

	t.Log("Delete label for application")

	// WHEN
	deleteApplicationLabelRequest := fixtures.FixDeleteApplicationLabelRequest(application.ID, labelKey)
	delLabel := graphql.Label{}

	// THEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, deleteApplicationLabelRequest, &delLabel)
	require.NoError(t, err)
	assert.Equal(t, labelKey, delLabel.Key)
}

func TestCreateScenariosLabel(t *testing.T) {
	// GIVEN
	t.Log("Create application")
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "app", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)

	t.Log("Check if scenarios LabelDefinition exists")
	labelKey := "scenarios"

	getLabelDefinition := fixtures.FixLabelDefinitionRequest(labelKey)
	ld := graphql.LabelDefinition{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getLabelDefinition, &ld)
	require.NoError(t, err)

	getApp := fixtures.FixGetApplicationRequest(app.ID)
	actualApp := graphql.ApplicationExt{}
	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getApp, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp)
}

func TestUpdateScenariosLabelDefinitionValue(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application")
	app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "app", conf.ApplicationTypeLabelKey, createAppTemplateName("Cloud for Customer"), tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantId, app.ID)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	labelKey := "scenarios"
	additionalValue := "ADDITIONAL"

	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)

	t.Logf("Update Label Definition scenarios enum with additional value %s", additionalValue)

	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": []string{testScenario, additionalValue},
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}

	ldInput := graphql.LabelDefinitionInput{
		Key:    labelKey,
		Schema: json.MarshalJSONSchema(t, jsonSchema),
	}

	ldInputGQL, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(ldInput)
	require.NoError(t, err)

	updateLabelDefinitionRequest := fixtures.FixUpdateLabelDefinitionRequest(ldInputGQL)
	labelDefinition := graphql.LabelDefinition{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateLabelDefinitionRequest, &labelDefinition)

	require.NoError(t, err)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, additionalValue)

	scenarios := []string{additionalValue}

	t.Logf("Set scenario label value %s on application", additionalValue)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: additionalValue}, app.ID, tenantId)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: additionalValue}, app.ID, tenantId)

	t.Log("Check if new scenario label value was set correctly")
	appRequest := fixtures.FixGetApplicationRequest(app.ID)
	app = graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, appRequest, &app)
	require.NoError(t, err)

	scenariosLabel, ok := app.Labels[labelKey].([]interface{})
	require.True(t, ok)

	var actualScenariosEnum []string
	for _, v := range scenariosLabel {
		actualScenariosEnum = append(actualScenariosEnum, v.(string))
	}
	assert.Equal(t, scenarios, actualScenariosEnum)
}

func TestSearchApplicationsByLabels(t *testing.T) {
	// GIVEN
	//Create first application
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	labelKeyFoo := "foo"
	labelKeyBar := "bar"

	firstApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "first", tenantId)
	require.NotEmpty(t, firstApp.ID)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &firstApp)
	require.NoError(t, err)

	//Create second application
	secondApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "second", tenantId)
	require.NotEmpty(t, secondApp.ID)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &secondApp)
	require.NoError(t, err)

	//Set label "foo" on both applications
	labelValueFoo := "val"

	firstAppLabel := fixtures.SetApplicationLabel(t, ctx, certSecuredGraphQLClient, firstApp.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, firstAppLabel.Key)
	require.NotEmpty(t, firstAppLabel.Value)

	secondAppLabel := fixtures.SetApplicationLabel(t, ctx, certSecuredGraphQLClient, secondApp.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, secondAppLabel.Key)
	require.NotEmpty(t, secondAppLabel.Value)

	//Set label "bar" on first application
	labelValueBar := "barval"

	firstAppBarLabel := fixtures.SetApplicationLabel(t, ctx, certSecuredGraphQLClient, firstApp.ID, labelKeyBar, labelValueBar)
	require.NotEmpty(t, firstAppBarLabel.Key)
	require.NotEmpty(t, firstAppBarLabel.Value)

	// Query for application with LabelFilter "foo"
	labelFilter := graphql.LabelFilter{
		Key: labelKeyFoo,
	}

	//WHEN
	labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	applicationRequest := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 5, "")
	applicationPage := graphql.ApplicationPageExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, applicationRequest, &applicationPage)
	require.NoError(t, err)

	//THEN
	require.NotEmpty(t, applicationPage)
	assert.Equal(t, applicationPage.TotalCount, 2)
	assert.Contains(t, applicationPage.Data[0].Labels, labelKeyFoo)
	assert.Equal(t, applicationPage.Data[0].Labels[labelKeyFoo], labelValueFoo)
	assert.Contains(t, applicationPage.Data[1].Labels, labelKeyFoo)
	assert.Equal(t, applicationPage.Data[1].Labels[labelKeyFoo], labelValueFoo)

	// Query for application with LabelFilter "bar"
	labelFilter = graphql.LabelFilter{
		Key: labelKeyBar,
	}

	// WHEN
	labelFilterGQL, err = testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	applicationRequest = fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 5, "")
	applicationPage = graphql.ApplicationPageExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, applicationRequest, &applicationPage)
	require.NoError(t, err)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, applicationPage)
	assert.Equal(t, applicationPage.TotalCount, 1)
	assert.Contains(t, applicationPage.Data[0].Labels, labelKeyBar)
	assert.Equal(t, applicationPage.Data[0].Labels[labelKeyBar], labelValueBar)
	saveExampleInCustomDir(t, applicationRequest.Query(), queryApplicationsCategory, "query applications with label filter")
}

func TestSearchRuntimesByLabels(t *testing.T) {
	// GIVEN
	//Create first runtime
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	labelKeyFoo := "foo"
	labelKeyBar := "bar"

	inputFirst := fixRuntimeInput("first")
	firstRuntime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, inputFirst, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &firstRuntime)

	//Create second runtime
	inputSecond := fixRuntimeInput("second")
	secondRuntime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, inputSecond, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &secondRuntime)

	//Set label "foo" on both runtimes
	labelValueFoo := "val"

	firstRuntimeLabel := fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, firstRuntime.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, firstRuntimeLabel.Key)
	require.NotEmpty(t, firstRuntimeLabel.Value)

	secondRuntimeLabel := fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, secondRuntime.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, secondRuntimeLabel.Key)
	require.NotEmpty(t, secondRuntimeLabel.Value)

	//Set label "bar" on first runtime
	labelValueBar := "barval"

	firstRuntimeBarLabel := fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, firstRuntime.ID, labelKeyBar, labelValueBar)
	require.NotEmpty(t, firstRuntimeBarLabel.Key)
	require.NotEmpty(t, firstRuntimeBarLabel.Value)

	// Query for runtime with LabelFilter "foo"
	labelFilter := graphql.LabelFilter{
		Key: labelKeyFoo,
	}

	//WHEN
	labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	runtimesRequest := fixtures.FixRuntimesFilteredPageableRequest(labelFilterGQL, 5, "")
	runtimePage := graphql.RuntimePageExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, runtimesRequest, &runtimePage)
	require.NoError(t, err)

	//THEN
	require.NotEmpty(t, runtimePage)
	assert.Equal(t, runtimePage.TotalCount, 2)
	assert.Contains(t, runtimePage.Data[0].Labels, labelKeyFoo)
	assert.Equal(t, runtimePage.Data[0].Labels[labelKeyFoo], labelValueFoo)
	assert.Contains(t, runtimePage.Data[1].Labels, labelKeyFoo)
	assert.Equal(t, runtimePage.Data[1].Labels[labelKeyFoo], labelValueFoo)

	// Query for runtime with LabelFilter "bar"
	labelFilter = graphql.LabelFilter{
		Key: labelKeyBar,
	}

	// WHEN
	labelFilterGQL, err = testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	runtimesRequest = fixtures.FixRuntimesFilteredPageableRequest(labelFilterGQL, 5, "")
	runtimePage = graphql.RuntimePageExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, runtimesRequest, &runtimePage)
	require.NoError(t, err)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, runtimePage)
	assert.Equal(t, runtimePage.TotalCount, 1)
	assert.Contains(t, runtimePage.Data[0].Labels, labelKeyBar)
	assert.Equal(t, runtimePage.Data[0].Labels[labelKeyBar], labelValueBar)
	saveExampleInCustomDir(t, runtimesRequest.Query(), QueryRuntimesCategory, "query runtimes with label filter")
}

func TestListLabelDefinitions(t *testing.T) {
	//GIVEN
	tenantID := tenant.TestTenants.GetIDByName(t, tenant.ListLabelDefinitionsTenantName)
	defer tenant.TestTenants.CleanupTenant(tenantID)

	ctx := context.TODO()

	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": []string{testScenario},
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}

	input := graphql.LabelDefinitionInput{
		Key:    "scenarios",
		Schema: json.MarshalJSONSchema(t, jsonSchema),
	}

	in, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(input)
	require.NoError(t, err)

	createRequest := fixtures.FixCreateLabelDefinitionRequest(in)
	saveExample(t, createRequest.Query(), "create label definition")

	output := graphql.LabelDefinition{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createRequest, &output)
	require.NoError(t, err)

	firstLabelDefinition := &output

	//WHEN
	labelDefinitions, err := fixtures.ListLabelDefinitionsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID)
	saveExample(t, fixtures.FixLabelDefinitionsRequest().Query(), "query label definition")

	//THEN
	require.NoError(t, err)
	require.Len(t, labelDefinitions, 1)
	assert.Contains(t, labelDefinitions, firstLabelDefinition)
}

func TestDeleteLastScenarioForApplication(t *testing.T) {
	//GIVEN
	ctx := context.TODO()

	tenantID := tenant.TestTenants.GetIDByName(t, tenant.DeleteLastScenarioForApplicationTenantName)
	name := "deleting-last-scenario-for-app-fail"
	scenarios := []string{"Christmas", "New Year"}

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenarios[0])
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenarios[0])

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenarios[1])
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenarios[1])

	appInput := graphql.ApplicationRegisterInput{
		Name: name,
		Labels: graphql.Labels{
			ScenariosLabel:               scenarios,
			conf.ApplicationTypeLabelKey: createAppTemplateName("Cloud for Customer"),
		},
	}

	application, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, tenantID, appInput)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &application)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantID, application.ID)

	//WHEN
	appLabelRequest := fixtures.FixSetApplicationLabelRequest(application.ID, ScenariosLabel, []string{scenarios[0]})
	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, appLabelRequest, nil))

	//remove last label
	appLabelRequest = fixtures.FixSetApplicationLabelRequest(application.ID, ScenariosLabel, []string{""})
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, appLabelRequest, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), `must be one of the following: "Christmas", "New Year"`)
}
