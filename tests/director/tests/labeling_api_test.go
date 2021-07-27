package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/json"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func TestCreateLabelWithoutLabelDefinition(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "label-without-label-def"
	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, name, tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	t.Log("Set label on application")
	labelKey := "test"
	labelValue := "val"

	setLabelRequest := fixtures.FixSetApplicationLabelRequest(application.ID, labelKey, labelValue)
	label := graphql.Label{}
	defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, labelKey, false, tenantId)
	defer fixtures.DeleteApplicationLabel(t, ctx, dexGraphQLClient, application.ID, labelKey)

	// WHEN
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, setLabelRequest, &label)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, label.Key)
	require.NotEmpty(t, label.Value)
	saveExample(t, setLabelRequest.Query(), "set application label")

	t.Log("Check if LabelDefinition was created internally")

	getLabelDefinitionRequest := fixtures.FixLabelDefinitionRequest(labelKey)
	labelDefinition := graphql.LabelDefinition{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, getLabelDefinitionRequest, &labelDefinition)

	require.NoError(t, err)
	require.NotEmpty(t, labelDefinition)
	assert.Equal(t, label.Key, labelDefinition.Key)
	saveExample(t, getLabelDefinitionRequest.Query(), "query label definition")
}

func TestCreateLabelWithExistingLabelDefinition(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	applicationName := "label-with-existing-label-def"

	t.Log("Create LabelDefinition")
	labelKey := "foo"

	jsonSchema := map[string]interface{}{
		"title": "foobarbaz",
		"type":  "object",
		"properties": map[string]interface{}{
			labelKey: map[string]interface{}{
				"type":        "string",
				"description": "foo",
			},
		},
		"required": []string{labelKey},
	}

	var schema interface{} = jsonSchema
	labelDefinitionInput := graphql.LabelDefinitionInput{
		Key:    labelKey,
		Schema: json.MarshalJSONSchema(t, schema),
	}

	labelDefinitionInputGQL, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
	require.NoError(t, err)

	t.Run("should fail - label value doesn't match json schema provided in label definition", func(t *testing.T) {
		createLabelDefinitionRequest := fixtures.FixCreateLabelDefinitionRequest(labelDefinitionInputGQL)
		labelDefinition := graphql.LabelDefinition{}

		t.Log("Create application")
		application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, applicationName, tenantId)
		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

		t.Log("Create label definition")
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createLabelDefinitionRequest, &labelDefinition)

		require.NoError(t, err)
		defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, labelKey, false, tenantId)
		assert.Equal(t, labelKey, labelDefinition.Key)

		invalidLabelValue := 123
		setLabelRequest := fixtures.FixSetApplicationLabelRequest(application.ID, labelKey, invalidLabelValue)

		// WHEN
		t.Log("Try to set label on application with invalid value against given json schema")
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, setLabelRequest, nil)

		//THEN
		require.Error(t, err)
		errMsg := fmt.Sprintf("reason=input value=%d, key=%s, is not valid against JSON Schema", invalidLabelValue, labelKey)
		assert.Contains(t, err.Error(), errMsg)
		saveExample(t, createLabelDefinitionRequest.Query(), "create label definition")

	})

	t.Run("should succeed - label value matches json schema in label definition", func(t *testing.T) {
		createLabelDefinitionRequest := fixtures.FixCreateLabelDefinitionRequest(labelDefinitionInputGQL)
		labelDefinition := graphql.LabelDefinition{}

		t.Log("Create application")
		application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, applicationName, tenantId)
		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

		t.Log("Create label definition")
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createLabelDefinitionRequest, &labelDefinition)

		t.Log("Set label on application with valid value")
		validLabelValue := map[string]interface{}{
			labelKey: "bar",
		}

		var appLabel interface{} = validLabelValue

		setLabelRequest := fixtures.FixSetApplicationLabelRequest(application.ID, labelKey, appLabel)
		label := graphql.Label{}

		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, setLabelRequest, &label)
		defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, labelKey, false, tenantId)
		defer fixtures.DeleteApplicationLabel(t, ctx, dexGraphQLClient, application.ID, labelKey)

		require.NoError(t, err)
		require.NotEmpty(t, label.Key)
		require.NotEmpty(t, label.Value)

		t.Log("Check if Label was set on application")
		queryAppReq := fixtures.FixGetApplicationRequest(application.ID)

		// WHEN
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, queryAppReq, &application)

		//THEN
		require.NoError(t, err)
		require.NotEmpty(t, application.Labels)
		assert.Equal(t, label.Value, application.Labels[labelKey])
	})
}

func TestEditLabelDefinition(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	labelKey := "foo"
	labelKeyBar := "bar"

	jsonSchema := map[string]interface{}{
		"title": "foobarbaz",
		"type":  "object",
		"properties": map[string]interface{}{
			labelKey: map[string]interface{}{
				"type":        "string",
				"description": "foo",
			},
		},
		"required": []string{labelKey},
	}

	invalidJsonSchema := map[string]interface{}{
		"title": "foobarbaz",
		"type":  "object",
		"properties": map[string]interface{}{
			labelKey: map[string]interface{}{
				"type":        "integer",
				"description": "integer value",
			},
		},
		"required": []string{labelKey},
	}

	newValidJsonSchema := map[string]interface{}{
		"title": "foobarbaz",
		"type":  "object",
		"properties": map[string]interface{}{
			labelKey: map[string]interface{}{
				"type":        "string",
				"description": "string value",
			},
			labelKeyBar: map[string]interface{}{
				"type":        "integer",
				"description": "integer value",
			},
		},
		"required": []string{labelKey},
	}

	var schema interface{} = jsonSchema
	labelDefinitionInput := graphql.LabelDefinitionInput{
		Key:    labelKey,
		Schema: json.MarshalJSONSchema(t, schema),
	}

	labelDefinitionInputGQL, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
	require.NoError(t, err)

	validLabelValue := map[string]interface{}{
		labelKey: labelKey,
	}
	var appLabel interface{} = validLabelValue

	t.Run("Try to edit LabelDefinition with incompatible data", func(t *testing.T) {
		createLabelDefinitionRequest := fixtures.FixCreateLabelDefinitionRequest(labelDefinitionInputGQL)
		labelDefinition := graphql.LabelDefinition{}

		t.Log("Create application")
		app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app", tenantId)
		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

		t.Log("Create label definition")
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createLabelDefinitionRequest, &labelDefinition)
		require.NoError(t, err)

		t.Log("Set label on application")
		setLabelRequest := fixtures.FixSetApplicationLabelRequest(app.ID, labelKey, appLabel)
		label := graphql.Label{}

		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, setLabelRequest, &label)
		defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, labelKey, false, tenantId)
		defer fixtures.DeleteApplicationLabel(t, ctx, dexGraphQLClient, app.ID, labelKey)

		var invalidSchema interface{} = invalidJsonSchema
		labelDefinitionInput = graphql.LabelDefinitionInput{
			Key:    labelKey,
			Schema: json.MarshalJSONSchema(t, invalidSchema),
		}

		ldInputGql, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
		require.NoError(t, err)

		updateLabelDefinitionReq := fixtures.FixUpdateLabelDefinitionRequest(ldInputGql)

		// WHEN
		t.Log("Try to edit LabelDefinition with incompatible data")
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateLabelDefinitionReq, nil)

		//THEN
		require.Error(t, err)
		errString := fmt.Sprintf(`reason=label with key="%s" is not valid against new schema for Application with ID="%s": %s: Invalid type. Expected: integer, given: string`, labelKey, app.ID, labelKey)
		assert.Contains(t, err.Error(), errString)
	})

	t.Run("Edit LabelDefinition with compatible data", func(t *testing.T) {
		createLabelDefinitionRequest := fixtures.FixCreateLabelDefinitionRequest(labelDefinitionInputGQL)
		labelDefinition := graphql.LabelDefinition{}

		t.Log("Create application")
		app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app", tenantId)
		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

		t.Log("Create label definition")
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createLabelDefinitionRequest, &labelDefinition)
		require.NoError(t, err)

		t.Log("Set label on application")
		setLabelRequest := fixtures.FixSetApplicationLabelRequest(app.ID, labelKey, appLabel)
		label := graphql.Label{}

		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, setLabelRequest, &label)
		defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, labelKey, false, tenantId)
		defer fixtures.DeleteApplicationLabel(t, ctx, dexGraphQLClient, app.ID, labelKey)

		var newSchema interface{} = newValidJsonSchema
		labelDefinitionInput = graphql.LabelDefinitionInput{
			Key:    labelKey,
			Schema: json.MarshalJSONSchema(t, newSchema),
		}

		ldInputGql, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
		require.NoError(t, err)

		updateLabelDefinitionReq := fixtures.FixUpdateLabelDefinitionRequest(ldInputGql)

		// WHEN
		t.Log("Edit LabelDefinition with compatible data")
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateLabelDefinitionReq, &labelDefinition)

		//THEN
		require.NoError(t, err)

		schemaVal, ok := (json.UnmarshalJSONSchema(t, labelDefinition.Schema)).(map[string]interface{})
		require.True(t, ok)
		actualProperties, ok := schemaVal["properties"].(map[string]interface{})
		require.True(t, ok)

		expectedProperties, ok := newValidJsonSchema["properties"].(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, expectedProperties, actualProperties)
		saveExample(t, updateLabelDefinitionReq.Query(), "update label definition")
	})
}

func TestCreateScenariosLabel(t *testing.T) {
	// GIVEN
	t.Log("Create application")
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

	t.Log("Check if scenarios LabelDefinition exists")
	labelKey := "scenarios"

	getLabelDefinition := fixtures.FixLabelDefinitionRequest(labelKey)
	ld := graphql.LabelDefinition{}

	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getLabelDefinition, &ld)
	require.NoError(t, err)

	t.Log("Check if app was labeled with scenarios=default")

	getApp := fixtures.FixGetApplicationRequest(app.ID)
	actualApp := graphql.Application{}
	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, getApp, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, app)
	if conf.DefaultScenarioEnabled {
		assert.Contains(t, app.Labels, labelKey)
		scenariosLabel, ok := app.Labels[labelKey].([]interface{})
		require.True(t, ok)

		var scenariosEnum []string
		for _, v := range scenariosLabel {
			scenariosEnum = append(scenariosEnum, v.(string))
		}

		assert.Contains(t, scenariosEnum, "DEFAULT")
	}
}

func TestUpdateScenariosLabelDefinitionValue(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application")
	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, tenantId, app.ID, conf.DefaultScenarioEnabled)
	labelKey := "scenarios"
	defaultValue := conf.DefaultScenario
	additionalValue := "ADDITIONAL"

	t.Logf("Update Label Definition scenarios enum with additional value %s", additionalValue)

	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": []string{defaultValue, additionalValue},
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}

	var schema interface{} = jsonSchema
	ldInput := graphql.LabelDefinitionInput{
		Key:    labelKey,
		Schema: json.MarshalJSONSchema(t, schema),
	}

	ldInputGQL, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(ldInput)
	require.NoError(t, err)

	updateLabelDefinitionRequest := fixtures.FixUpdateLabelDefinitionRequest(ldInputGQL)
	labelDefinition := graphql.LabelDefinition{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateLabelDefinitionRequest, &labelDefinition)

	require.NoError(t, err)

	scenarios := []string{additionalValue}
	if conf.DefaultScenarioEnabled {
		scenarios = []string{defaultValue, additionalValue}
	}
	var labelValue interface{} = scenarios

	t.Logf("Set scenario label value %s on application", additionalValue)
	fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, app.ID, labelKey, labelValue)

	t.Log("Check if new scenario label value was set correctly")
	appRequest := fixtures.FixGetApplicationRequest(app.ID)
	app = graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, appRequest, &app)
	require.NoError(t, err)

	scenariosLabel, ok := app.Labels[labelKey].([]interface{})
	require.True(t, ok)

	var actualScenariosEnum []string
	for _, v := range scenariosLabel {
		actualScenariosEnum = append(actualScenariosEnum, v.(string))
	}
	assert.Equal(t, scenarios, actualScenariosEnum)
}

func TestDeleteLabelDefinition(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	labelKey := "foo"

	jsonSchema := map[string]interface{}{
		"title": "foobarbaz",
		"type":  "object",
		"properties": map[string]interface{}{
			labelKey: map[string]interface{}{
				"type":        "string",
				"description": "foo",
			},
		},
		"required": []string{labelKey},
	}

	var schema interface{} = jsonSchema
	labelDefinitionInput := graphql.LabelDefinitionInput{
		Key:    labelKey,
		Schema: json.MarshalJSONSchema(t, schema),
	}

	ldInputGql, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
	require.NoError(t, err)

	t.Run("Try to delete Label Definition while it's being used by some labels with deleteRelatedLabels parameter set to false - should fail", func(t *testing.T) {

		t.Log("Create application")
		app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app", tenantId)
		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, tenantId, app.ID, conf.DefaultScenarioEnabled)

		t.Log("Create LabelDefinition")
		createLabelDefinitionRequest := fixtures.FixCreateLabelDefinitionRequest(ldInputGql)
		ld := graphql.LabelDefinition{}

		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createLabelDefinitionRequest, ld)
		require.NoError(t, err)

		t.Log("Set label on application")
		validLabelValue := map[string]interface{}{"foo": "test"}

		setLabelRequest := fixtures.FixSetApplicationLabelRequest(app.ID, labelKey, validLabelValue)
		label := graphql.Label{}

		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, setLabelRequest, &label)
		require.NoError(t, err)
		defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, labelKey, false, tenantId)
		defer fixtures.DeleteApplicationLabel(t, ctx, dexGraphQLClient, app.ID, labelKey)

		t.Log("Try to delete Label Definition while it's being used by some labels")

		deleteLabelDefinitionRequest := fixtures.FixDeleteLabelDefinitionRequest(labelKey, false)
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteLabelDefinitionRequest, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reason=could not delete label definition, it is already used by at least one label")
		saveExample(t, deleteLabelDefinitionRequest.Query(), "delete label definition")
	})

	t.Run("Delete Label Definition while it's being used by some labels with deleteRelatedLabels parameter set to true - should succeed", func(t *testing.T) {

		t.Log("Create LabelDefinition")
		createLabelDefinitionRequest := fixtures.FixCreateLabelDefinitionRequest(ldInputGql)
		ld := graphql.LabelDefinition{}

		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createLabelDefinitionRequest, ld)
		require.NoError(t, err)

		t.Log("Create application")
		app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app", tenantId)
		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, tenantId, app.ID, conf.DefaultScenarioEnabled)

		t.Log("Create runtime")
		input := fixtures.FixRuntimeInput("rtm")
		rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
		defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)

		t.Log("Set labels on application and runtime")
		fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, app.ID, labelKey, map[string]interface{}{labelKey: "app"})
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, rtm.ID, labelKey, map[string]interface{}{labelKey: "rtm"})

		t.Log("Delete Label Definition while it's being used by some labels")
		deleteLabelDefinitionRequest := fixtures.FixDeleteLabelDefinitionRequest(labelKey, true)
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteLabelDefinitionRequest, nil)
		require.NoError(t, err)

		t.Log("Assert labels were deleted from Application and Runtime")
		app = fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
		runtime := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantId, rtm.ID)

		assert.Empty(t, app.Labels[labelKey])
		assert.Empty(t, runtime.Labels[labelKey])

		t.Log("Assert Label definition was deleted")
		ldRequest := fixtures.FixLabelDefinitionRequest(labelKey)
		errMsg := fmt.Sprintf("graphql: label definition with key '%s' does not exist", labelKey)
		require.Nil(t, testctx.Tc.RunOperation(ctx, dexGraphQLClient, ldRequest, nil), errMsg)
	})

	t.Run("Delete Label from application, then delete the Label Definition - should succeed", func(t *testing.T) {

		t.Log("Create application")
		app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app", tenantId)
		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, tenantId, app.ID, conf.DefaultScenarioEnabled)

		t.Log("Create LabelDefinition")
		createLabelDefinitionRequest := fixtures.FixCreateLabelDefinitionRequest(ldInputGql)
		ld := graphql.LabelDefinition{}

		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createLabelDefinitionRequest, ld)
		require.NoError(t, err)

		t.Log("Set label on application")
		validLabelValue := map[string]interface{}{labelKey: "test"}

		setLabelRequest := fixtures.FixSetApplicationLabelRequest(app.ID, labelKey, validLabelValue)
		label := graphql.Label{}

		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, setLabelRequest, &label)
		require.NoError(t, err)

		deleteApplicationLabelRequest := fixtures.FixDeleteApplicationLabelRequest(app.ID, labelKey)
		label = graphql.Label{}

		err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteApplicationLabelRequest, &label)
		require.NoError(t, err)
		assert.Equal(t, labelKey, label.Key)

		deleteLabelDefinitionRequest := fixtures.FixDeleteLabelDefinitionRequest(labelKey, false)
		labelDefinition := graphql.LabelDefinition{}

		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteLabelDefinitionRequest, &labelDefinition)
		require.NoError(t, err)
		assertions.AssertGraphQLJSONSchema(t, labelDefinitionInput.Schema, labelDefinition.Schema)
	})
}

func TestDeleteDefaultValueInScenariosLabelDefinition(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application")
	app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, tenantId, app.ID, conf.DefaultScenarioEnabled)

	labelKey := "scenarios"
	defaultValue := conf.DefaultScenario

	t.Log("Try to update Label Definition with scenarios enum without DEFAULT value")

	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": []string{"NOTDEFAULT"},
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}

	var schema interface{} = jsonSchema
	ldInput := graphql.LabelDefinitionInput{
		Key:    labelKey,
		Schema: json.MarshalJSONSchema(t, schema),
	}

	ldInputGQL, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(ldInput)
	require.NoError(t, err)

	updateLabelDefinitionRequest := fixtures.FixUpdateLabelDefinitionRequest(ldInputGQL)
	labelDefinition := graphql.LabelDefinition{}

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateLabelDefinitionRequest, &labelDefinition)
	errMsg := fmt.Sprintf(`rule.validSchema=while validating schema for key %s: items.enum: At least one of the items must match, items.enum.0: items.enum.0 does not match: "%s"`, labelKey, defaultValue)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), errMsg)
}

func TestSearchApplicationsByLabels(t *testing.T) {
	// GIVEN
	//Create first application
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	labelKeyFoo := "foo"
	labelKeyBar := "bar"
	defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, labelKeyFoo, false, tenantId)
	defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, labelKeyBar, false, tenantId)

	firstApp := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "first", tenantId)
	require.NotEmpty(t, firstApp.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, firstApp.ID)

	//Create second application
	secondApp := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "second", tenantId)
	require.NotEmpty(t, secondApp.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, secondApp.ID)

	//Set label "foo" on both applications
	labelValueFoo := "val"

	firstAppLabel := fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, firstApp.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, firstAppLabel.Key)
	require.NotEmpty(t, firstAppLabel.Value)

	secondAppLabel := fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, secondApp.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, secondAppLabel.Key)
	require.NotEmpty(t, secondAppLabel.Value)

	//Set label "bar" on first application
	labelValueBar := "barval"

	firstAppBarLabel := fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, firstApp.ID, labelKeyBar, labelValueBar)
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
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, applicationRequest, &applicationPage)
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
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, applicationRequest, &applicationPage)
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
	defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, labelKeyFoo, false, tenantId)
	defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, labelKeyBar, false, tenantId)

	inputFirst := fixtures.FixRuntimeInput("first")
	firstRuntime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &inputFirst)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, firstRuntime.ID)

	//Create second runtime
	inputSecond := fixtures.FixRuntimeInput("second")
	secondRuntime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &inputSecond)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, secondRuntime.ID)

	//Set label "foo" on both runtimes
	labelValueFoo := "val"

	firstRuntimeLabel := fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, firstRuntime.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, firstRuntimeLabel.Key)
	require.NotEmpty(t, firstRuntimeLabel.Value)

	secondRuntimeLabel := fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, secondRuntime.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, secondRuntimeLabel.Key)
	require.NotEmpty(t, secondRuntimeLabel.Value)

	//Set label "bar" on first runtime
	labelValueBar := "barval"

	firstRuntimeBarLabel := fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, firstRuntime.ID, labelKeyBar, labelValueBar)
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
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, runtimesRequest, &runtimePage)
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
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, runtimesRequest, &runtimePage)
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
	ctx := context.TODO()

	firstSchema := map[string]interface{}{
		"test": "test",
	}
	firstLabelDefinition := fixtures.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, "first", firstSchema, tenantID)
	defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, firstLabelDefinition.Key, false, tenantID)

	secondSchema := map[string]interface{}{
		"test": "test",
	}
	secondLabelDefinition := fixtures.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, "second", secondSchema, tenantID)
	defer fixtures.DeleteLabelDefinition(t, ctx, dexGraphQLClient, secondLabelDefinition.Key, false, tenantID)

	//WHEN
	labelDefinitions, err := fixtures.ListLabelDefinitionsWithinTenant(t, ctx, dexGraphQLClient, tenantID)

	//THEN
	require.NoError(t, err)
	require.Len(t, labelDefinitions, 2)
	assert.Contains(t, labelDefinitions, firstLabelDefinition)
	assert.Contains(t, labelDefinitions, secondLabelDefinition)
}

func TestDeleteLastScenarioForApplication(t *testing.T) {
	//GIVEN
	ctx := context.TODO()

	tenantID := tenant.TestTenants.GetIDByName(t, tenant.DeleteLastScenarioForApplicationTenantName)
	name := "deleting-last-scenario-for-app-fail"
	scenarios := []string{conf.DefaultScenario, "Christmas", "New Year"}

	scenarioSchema := map[string]interface{}{
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
		"items": map[string]interface{}{
			"type": "string",
			"enum": scenarios,
		},
	}
	var schema interface{} = scenarioSchema

	fixtures.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, ScenariosLabel, schema, tenantID)

	appInput := graphql.ApplicationRegisterInput{
		Name: name,
		Labels: graphql.Labels{
			ScenariosLabel: []string{"Christmas", "New Year"},
		},
	}

	application, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, tenantID, appInput)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantID, application.ID)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, tenantID, application.ID, conf.DefaultScenarioEnabled)

	//WHEN
	appLabelRequest := fixtures.FixSetApplicationLabelRequest(application.ID, ScenariosLabel, []string{"Christmas"})
	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, appLabelRequest, nil))

	//remove last label
	appLabelRequest = fixtures.FixSetApplicationLabelRequest(application.ID, ScenariosLabel, []string{""})
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, appLabelRequest, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), `must be one of the following: "DEFAULT", "Christmas", "New Year"`)
}

func TestGetScenariosLabelDefinitionCreatesOneIfNotExists(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	tenantID := tenant.TestTenants.GetIDByName(t, "TestGetScenariosLabelDefinitionCreatesOneIfNotExists")
	getLabelDefinitionRequest := fixtures.FixLabelDefinitionRequest(ScenariosLabel)
	labelDefinition := graphql.LabelDefinition{}

	// WHEN
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, getLabelDefinitionRequest, &labelDefinition)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, labelDefinition)
	assert.Equal(t, ScenariosLabel, labelDefinition.Key)
	assert.NotEmpty(t, labelDefinition.Schema)
}
