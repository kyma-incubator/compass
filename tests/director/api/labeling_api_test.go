package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func TestCreateLabelWithoutLabelDefinition(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "label-without-label-def"
	application := registerApplication(t, ctx, name)
	defer unregisterApplication(t, application.ID)

	t.Log("Set label on application")
	labelKey := "test"
	labelValue := "val"

	setLabelRequest := fixSetApplicationLabelRequest(application.ID, labelKey, labelValue)
	label := graphql.Label{}
	defer deleteLabelDefinition(t, ctx, labelKey, false)
	defer deleteApplicationLabel(t, ctx, application.ID, labelKey)

	// WHEN
	err := tc.RunOperation(ctx, setLabelRequest, &label)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, label.Key)
	require.NotEmpty(t, label.Value)
	saveExample(t, setLabelRequest.Query(), "set application label")

	t.Log("Check if LabelDefinition was created internally")

	getLabelDefinitionRequest := fixLabelDefinitionRequest(labelKey)
	labelDefinition := graphql.LabelDefinition{}

	err = tc.RunOperation(ctx, getLabelDefinitionRequest, &labelDefinition)

	require.NoError(t, err)
	require.NotEmpty(t, labelDefinition)
	assert.Equal(t, label.Key, labelDefinition.Key)
	saveExample(t, getLabelDefinitionRequest.Query(), "query label definition")
}

func TestCreateLabelWithExistingLabelDefinition(t *testing.T) {
	// GIVEN
	ctx := context.Background()

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
		Schema: marshalJSONSchema(t, schema),
	}

	labelDefinitionInputGQL, err := tc.graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
	require.NoError(t, err)

	t.Run("should fail - label value doesn't match json schema provided in label definition", func(t *testing.T) {
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(labelDefinitionInputGQL)
		labelDefinition := graphql.LabelDefinition{}

		t.Log("Create application")
		application := registerApplication(t, ctx, applicationName)
		defer unregisterApplication(t, application.ID)

		t.Log("Create label definition")
		err = tc.RunOperation(ctx, createLabelDefinitionRequest, &labelDefinition)

		require.NoError(t, err)
		defer deleteLabelDefinition(t, ctx, labelKey, false)
		assert.Equal(t, labelKey, labelDefinition.Key)

		invalidLabelValue := 123
		setLabelRequest := fixSetApplicationLabelRequest(application.ID, labelKey, invalidLabelValue)

		// WHEN
		t.Log("Try to set label on application with invalid value against given json schema")
		err = tc.RunOperation(ctx, setLabelRequest, nil)

		//THEN
		require.Error(t, err)
		errMsg := fmt.Sprintf("graphql: while creating label for Application: while validating Label value for '%s': while validating value %d against JSON Schema", labelKey, invalidLabelValue)
		assert.Contains(t, err.Error(), errMsg)
		saveExample(t, createLabelDefinitionRequest.Query(), "create label definition")

	})

	t.Run("should succeed - label value matches json schema in label definition", func(t *testing.T) {
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(labelDefinitionInputGQL)
		labelDefinition := graphql.LabelDefinition{}

		t.Log("Create application")
		application := registerApplication(t, ctx, applicationName)
		defer unregisterApplication(t, application.ID)

		t.Log("Create label definition")
		err = tc.RunOperation(ctx, createLabelDefinitionRequest, &labelDefinition)

		t.Log("Set label on application with valid value")
		validLabelValue := map[string]interface{}{
			labelKey: "bar",
		}

		var appLabel interface{} = validLabelValue

		setLabelRequest := fixSetApplicationLabelRequest(application.ID, labelKey, appLabel)
		label := graphql.Label{}

		err = tc.RunOperation(ctx, setLabelRequest, &label)
		defer deleteLabelDefinition(t, ctx, labelKey, false)
		defer deleteApplicationLabel(t, ctx, application.ID, labelKey)

		require.NoError(t, err)
		require.NotEmpty(t, label.Key)
		require.NotEmpty(t, label.Value)

		t.Log("Check if Label was set on application")
		queryAppReq := fixApplicationRequest(application.ID)

		// WHEN
		err = tc.RunOperation(context.Background(), queryAppReq, &application)

		//THEN
		require.NoError(t, err)
		require.NotEmpty(t, application.Labels)
		assert.Equal(t, label.Value, application.Labels[labelKey])
	})
}

func TestEditLabelDefinition(t *testing.T) {
	// GIVEN
	ctx := context.Background()

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
		Schema: marshalJSONSchema(t, schema),
	}

	labelDefinitionInputGQL, err := tc.graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
	require.NoError(t, err)

	validLabelValue := map[string]interface{}{
		labelKey: labelKey,
	}
	var appLabel interface{} = validLabelValue

	t.Run("Try to edit LabelDefinition with incompatible data", func(t *testing.T) {
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(labelDefinitionInputGQL)
		labelDefinition := graphql.LabelDefinition{}

		t.Log("Create application")
		app := registerApplication(t, ctx, "app")
		defer unregisterApplication(t, app.ID)

		t.Log("Create label definition")
		err = tc.RunOperation(ctx, createLabelDefinitionRequest, &labelDefinition)
		require.NoError(t, err)

		t.Log("Set label on application")
		setLabelRequest := fixSetApplicationLabelRequest(app.ID, labelKey, appLabel)
		label := graphql.Label{}

		err = tc.RunOperation(ctx, setLabelRequest, &label)
		defer deleteLabelDefinition(t, ctx, labelKey, false)
		defer deleteApplicationLabel(t, ctx, app.ID, labelKey)

		var invalidSchema interface{} = invalidJsonSchema
		labelDefinitionInput = graphql.LabelDefinitionInput{
			Key:    labelKey,
			Schema: marshalJSONSchema(t, invalidSchema),
		}

		ldInputGql, err := tc.graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
		require.NoError(t, err)

		updateLabelDefinitionReq := fixUpdateLabelDefinitionRequest(ldInputGql)

		// WHEN
		t.Log("Try to edit LabelDefinition with incompatible data")
		err = tc.RunOperation(context.Background(), updateLabelDefinitionReq, nil)

		//THEN
		require.Error(t, err)
		errString := fmt.Sprintf("graphql: while updating label definition: label with key \"%s\" is not valid against new schema for Application with ID \"%s\": %s: Invalid type. Expected: integer, given: string", labelKey, app.ID, labelKey)
		assert.EqualError(t, err, errString)
	})

	t.Run("Edit LabelDefinition with compatible data", func(t *testing.T) {
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(labelDefinitionInputGQL)
		labelDefinition := graphql.LabelDefinition{}

		t.Log("Create application")
		app := registerApplication(t, ctx, "app")
		defer unregisterApplication(t, app.ID)

		t.Log("Create label definition")
		err = tc.RunOperation(ctx, createLabelDefinitionRequest, &labelDefinition)
		require.NoError(t, err)

		t.Log("Set label on application")
		setLabelRequest := fixSetApplicationLabelRequest(app.ID, labelKey, appLabel)
		label := graphql.Label{}

		err = tc.RunOperation(ctx, setLabelRequest, &label)
		defer deleteLabelDefinition(t, ctx, labelKey, false)
		defer deleteApplicationLabel(t, ctx, app.ID, labelKey)

		var newSchema interface{} = newValidJsonSchema
		labelDefinitionInput = graphql.LabelDefinitionInput{
			Key:    labelKey,
			Schema: marshalJSONSchema(t, newSchema),
		}

		ldInputGql, err := tc.graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
		require.NoError(t, err)

		updateLabelDefinitionReq := fixUpdateLabelDefinitionRequest(ldInputGql)

		// WHEN
		t.Log("Edit LabelDefinition with compatible data")
		err = tc.RunOperation(context.Background(), updateLabelDefinitionReq, &labelDefinition)

		//THEN
		require.NoError(t, err)

		schemaVal, ok := (unmarshalJSONSchema(t, labelDefinition.Schema)).(map[string]interface{})
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
	app := registerApplication(t, ctx, "app")
	defer unregisterApplication(t, app.ID)

	t.Log("Check if scenarios LabelDefinition exists")
	labelKey := "scenarios"

	getLabelDefinition := fixLabelDefinitionRequest(labelKey)
	ld := graphql.LabelDefinition{}

	err := tc.RunOperation(ctx, getLabelDefinition, &ld)
	require.NoError(t, err)

	t.Log("Check if app was labeled with scenarios=default")

	getApp := fixApplicationRequest(app.ID)
	actualApp := graphql.Application{}
	// WHEN
	err = tc.RunOperation(ctx, getApp, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, app)
	assert.Contains(t, app.Labels, labelKey)

	scenariosLabel, ok := app.Labels[labelKey].([]interface{})
	require.True(t, ok)

	var scenariosEnum []string
	for _, v := range scenariosLabel {
		scenariosEnum = append(scenariosEnum, v.(string))
	}

	assert.Contains(t, scenariosEnum, "DEFAULT")
}

func TestUpdateScenariosLabelDefinitionValue(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	t.Log("Create application")
	app := registerApplication(t, ctx, "app")
	defer unregisterApplication(t, app.ID)
	labelKey := "scenarios"
	defaultValue := "DEFAULT"
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
		Schema: marshalJSONSchema(t, schema),
	}

	ldInputGQL, err := tc.graphqlizer.LabelDefinitionInputToGQL(ldInput)
	require.NoError(t, err)

	updateLabelDefinitionRequest := fixUpdateLabelDefinitionRequest(ldInputGQL)
	labelDefinition := graphql.LabelDefinition{}

	err = tc.RunOperation(ctx, updateLabelDefinitionRequest, &labelDefinition)

	require.NoError(t, err)

	scenarios := []string{defaultValue, additionalValue}
	var labelValue interface{} = scenarios

	t.Logf("Set scenario label value %s on application", additionalValue)
	setApplicationLabel(t, ctx, app.ID, labelKey, labelValue)

	t.Log("Check if new scenario label value was set correctly")
	appRequest := fixApplicationRequest(app.ID)
	app = graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, appRequest, &app)
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
		Schema: marshalJSONSchema(t, schema),
	}

	ldInputGql, err := tc.graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
	require.NoError(t, err)

	t.Run("Try to delete Label Definition while it's being used by some labels with deleteRelatedLabels parameter set to false - should fail", func(t *testing.T) {

		t.Log("Create application")
		app := registerApplication(t, ctx, "app")
		defer unregisterApplication(t, app.ID)

		t.Log("Create LabelDefinition")
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(ldInputGql)
		ld := graphql.LabelDefinition{}

		err = tc.RunOperation(ctx, createLabelDefinitionRequest, ld)
		require.NoError(t, err)

		t.Log("Set label on application")
		validLabelValue := map[string]interface{}{"foo": "test"}

		setLabelRequest := fixSetApplicationLabelRequest(app.ID, labelKey, validLabelValue)
		label := graphql.Label{}

		err = tc.RunOperation(ctx, setLabelRequest, &label)
		require.NoError(t, err)
		defer deleteLabelDefinition(t, ctx, labelKey, false)
		defer deleteApplicationLabel(t, ctx, app.ID, labelKey)

		t.Log("Try to delete Label Definition while it's being used by some labels")

		deleteLabelDefinitionRequest := fixDeleteLabelDefinitionRequest(labelKey, false)
		err = tc.RunOperation(context.Background(), deleteLabelDefinitionRequest, nil)
		require.Error(t, err)
		assert.EqualError(t, err, "graphql: could not delete label definition, it is already used by at least one label")
		saveExample(t, deleteLabelDefinitionRequest.Query(), "delete label definition")
	})

	t.Run("Delete Label Definition while it's being used by some labels with deleteRelatedLabels parameter set to true - should succeed", func(t *testing.T) {

		t.Log("Create LabelDefinition")
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(ldInputGql)
		ld := graphql.LabelDefinition{}

		err = tc.RunOperation(ctx, createLabelDefinitionRequest, ld)
		require.NoError(t, err)

		t.Log("Create application")
		app := registerApplication(t, ctx, "app")
		defer unregisterApplication(t, app.ID)

		t.Log("Create runtime")
		rtm := registerRuntime(t, ctx, "rtm")
		defer unregisterRuntimeWithinTenant(t, rtm.ID, testTenants.GetDefaultTenantID())

		t.Log("Set labels on application and runtime")
		setApplicationLabel(t, ctx, app.ID, labelKey, map[string]interface{}{labelKey: "app"})
		setRuntimeLabel(t, ctx, rtm.ID, labelKey, map[string]interface{}{labelKey: "rtm"})

		t.Log("Delete Label Definition while it's being used by some labels")
		deleteLabelDefinitionRequest := fixDeleteLabelDefinitionRequest(labelKey, true)
		err = tc.RunOperation(context.Background(), deleteLabelDefinitionRequest, nil)
		require.NoError(t, err)

		t.Log("Assert labels were deleted from Application and Runtime")
		app = getApplication(t, ctx, app.ID)
		runtime := getRuntime(t, ctx, rtm.ID)

		assert.Empty(t, app.Labels[labelKey])
		assert.Empty(t, runtime.Labels[labelKey])

		t.Log("Assert Label definition was deleted")
		ldRequest := fixLabelDefinitionRequest(labelKey)
		errMsg := fmt.Sprintf("graphql: label definition with key '%s' does not exist", labelKey)
		require.Nil(t, tc.RunOperation(ctx, ldRequest, nil), errMsg)
	})

	t.Run("Delete Label from application, then delete the Label Definition - should succeed", func(t *testing.T) {

		t.Log("Create application")
		app := registerApplication(t, ctx, "app")
		defer unregisterApplication(t, app.ID)

		t.Log("Create LabelDefinition")
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(ldInputGql)
		ld := graphql.LabelDefinition{}

		err = tc.RunOperation(ctx, createLabelDefinitionRequest, ld)
		require.NoError(t, err)

		t.Log("Set label on application")
		validLabelValue := map[string]interface{}{labelKey: "test"}

		setLabelRequest := fixSetApplicationLabelRequest(app.ID, labelKey, validLabelValue)
		label := graphql.Label{}

		err = tc.RunOperation(ctx, setLabelRequest, &label)
		require.NoError(t, err)

		deleteApplicationLabelRequest := fixDeleteApplicationLabelRequest(app.ID, labelKey)
		label = graphql.Label{}

		err := tc.RunOperation(ctx, deleteApplicationLabelRequest, &label)
		require.NoError(t, err)
		assert.Equal(t, labelKey, label.Key)

		deleteLabelDefinitionRequest := fixDeleteLabelDefinitionRequest(labelKey, false)
		labelDefinition := graphql.LabelDefinition{}

		err = tc.RunOperation(context.Background(), deleteLabelDefinitionRequest, &labelDefinition)
		require.NoError(t, err)
		assertGraphQLJSONSchema(t, labelDefinitionInput.Schema, labelDefinition.Schema)
	})
}

func TestDeleteScenariosLabel(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	t.Log("Create application")
	app := registerApplication(t, ctx, "app")
	defer unregisterApplication(t, app.ID)

	t.Log("Try to delete scenarios label on application")
	labelKey := "scenarios"
	deleteApplicationLabelRequest := fixDeleteApplicationLabelRequest(app.ID, labelKey)

	// WHEN
	err := tc.RunOperation(ctx, deleteApplicationLabelRequest, nil)

	//THEN
	require.Error(t, err)
	assert.EqualError(t, err, "graphql: scenarios label can not be deleted from application")
}

func TestDeleteDefaultValueInScenariosLabelDefinition(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	t.Log("Create application")
	app := registerApplication(t, ctx, "app")
	defer unregisterApplication(t, app.ID)
	labelKey := "scenarios"
	defaultValue := "DEFAULT"

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
		Schema: marshalJSONSchema(t, schema),
	}

	ldInputGQL, err := tc.graphqlizer.LabelDefinitionInputToGQL(ldInput)
	require.NoError(t, err)

	updateLabelDefinitionRequest := fixUpdateLabelDefinitionRequest(ldInputGQL)
	labelDefinition := graphql.LabelDefinition{}

	// WHEN
	err = tc.RunOperation(ctx, updateLabelDefinitionRequest, &labelDefinition)
	errMsg := fmt.Sprintf(`graphql: validation error for type LabelDefinitionInput: Rule.ValidSchema: while validating schema for key %s: items.enum: At least one of the items must match, items.enum.0: items.enum.0 does not match: "%s".`, labelKey, defaultValue)

	// THEN
	require.Error(t, err)
	assert.EqualError(t, err, errMsg)
}

func TestSearchApplicationsByLabels(t *testing.T) {
	// GIVEN
	//Create first application
	ctx := context.Background()
	labelKeyFoo := "foo"
	labelKeyBar := "bar"
	defer deleteLabelDefinition(t, ctx, labelKeyFoo, false)
	defer deleteLabelDefinition(t, ctx, labelKeyBar, false)

	firstApp := registerApplication(t, ctx, "first")
	require.NotEmpty(t, firstApp.ID)
	defer unregisterApplication(t, firstApp.ID)

	//Create second application
	secondApp := registerApplication(t, ctx, "second")
	require.NotEmpty(t, secondApp.ID)
	defer unregisterApplication(t, secondApp.ID)

	//Set label "foo" on both applications
	labelValueFoo := "val"

	firstAppLabel := setApplicationLabel(t, ctx, firstApp.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, firstAppLabel.Key)
	require.NotEmpty(t, firstAppLabel.Value)

	secondAppLabel := setApplicationLabel(t, ctx, secondApp.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, secondAppLabel.Key)
	require.NotEmpty(t, secondAppLabel.Value)

	//Set label "bar" on first application
	labelValueBar := "barval"

	firstAppBarLabel := setApplicationLabel(t, ctx, firstApp.ID, labelKeyBar, labelValueBar)
	require.NotEmpty(t, firstAppBarLabel.Key)
	require.NotEmpty(t, firstAppBarLabel.Value)

	// Query for application with LabelFilter "foo"
	labelFilter := graphql.LabelFilter{
		Key: labelKeyFoo,
	}

	//WHEN
	labelFilterGQL, err := tc.graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	applicationRequest := fixApplicationsFilteredPageableRequest(labelFilterGQL, 5, "")
	applicationPage := graphql.ApplicationPageExt{}
	err = tc.RunOperation(ctx, applicationRequest, &applicationPage)
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
	labelFilterGQL, err = tc.graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	applicationRequest = fixApplicationsFilteredPageableRequest(labelFilterGQL, 5, "")
	applicationPage = graphql.ApplicationPageExt{}
	err = tc.RunOperation(ctx, applicationRequest, &applicationPage)
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
	labelKeyFoo := "foo"
	labelKeyBar := "bar"
	defer deleteLabelDefinition(t, ctx, labelKeyFoo, false)
	defer deleteLabelDefinition(t, ctx, labelKeyBar, false)

	firstRuntime := registerRuntime(t, ctx, "first")
	defer unregisterRuntime(t, firstRuntime.ID)

	//Create second runtime
	secondRuntime := registerRuntime(t, ctx, "second")
	defer unregisterRuntime(t, secondRuntime.ID)

	//Set label "foo" on both runtimes
	labelValueFoo := "val"

	firstRuntimeLabel := setRuntimeLabel(t, ctx, firstRuntime.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, firstRuntimeLabel.Key)
	require.NotEmpty(t, firstRuntimeLabel.Value)

	secondRuntimeLabel := setRuntimeLabel(t, ctx, secondRuntime.ID, labelKeyFoo, labelValueFoo)
	require.NotEmpty(t, secondRuntimeLabel.Key)
	require.NotEmpty(t, secondRuntimeLabel.Value)

	//Set label "bar" on first runtime
	labelValueBar := "barval"

	firstRuntimeBarLabel := setRuntimeLabel(t, ctx, firstRuntime.ID, labelKeyBar, labelValueBar)
	require.NotEmpty(t, firstRuntimeBarLabel.Key)
	require.NotEmpty(t, firstRuntimeBarLabel.Value)

	// Query for runtime with LabelFilter "foo"
	labelFilter := graphql.LabelFilter{
		Key: labelKeyFoo,
	}

	//WHEN
	labelFilterGQL, err := tc.graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	runtimesRequest := fixRuntimesRequest(labelFilterGQL, 5, "")
	runtimePage := graphql.RuntimePageExt{}
	err = tc.RunOperation(ctx, runtimesRequest, &runtimePage)
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
	labelFilterGQL, err = tc.graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	runtimesRequest = fixRuntimesRequest(labelFilterGQL, 5, "")
	runtimePage = graphql.RuntimePageExt{}
	err = tc.RunOperation(ctx, runtimesRequest, &runtimePage)
	require.NoError(t, err)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, runtimePage)
	assert.Equal(t, runtimePage.TotalCount, 1)
	assert.Contains(t, runtimePage.Data[0].Labels, labelKeyBar)
	assert.Equal(t, runtimePage.Data[0].Labels[labelKeyBar], labelValueBar)
	saveExampleInCustomDir(t, runtimesRequest.Query(), queryRuntimesCategory, "query runtimes with label filter")
}

func TestListLabelDefinitions(t *testing.T) {
	//GIVEN
	tenantID := testTenants.GetIDByName(t, "Test3")
	ctx := context.TODO()
	firstSchema := map[string]interface{}{
		"test": "test",
	}
	firstLabelDefinition := createLabelDefinitionWithinTenant(t, ctx, "first", firstSchema, tenantID)
	defer deleteLabelDefinitionWithinTenant(t, ctx, firstLabelDefinition.Key, false, tenantID)

	secondSchema := map[string]interface{}{
		"test": "test",
	}
	secondLabelDefinition := createLabelDefinitionWithinTenant(t, ctx, "second", secondSchema, tenantID)
	defer deleteLabelDefinitionWithinTenant(t, ctx, secondLabelDefinition.Key, false, tenantID)

	//WHEN
	labelDefinitions, err := listLabelDefinitionsWithinTenant(t, ctx, tenantID)

	//THEN
	require.NoError(t, err)
	require.Len(t, labelDefinitions, 2)
	assert.Contains(t, labelDefinitions, firstLabelDefinition)
	assert.Contains(t, labelDefinitions, secondLabelDefinition)
}

func TestDeleteLastScenarioForApplication(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	tenantID := testTenants.GetIDByName(t, "Test4")
	name := "deleting-last-scenario-for-app-fail"
	scenarios := []string{"DEFAULT", "Christmas", "New Year"}

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

	createLabelDefinitionWithinTenant(t, ctx, scenariosLabel, schema, tenantID)

	appInput := graphql.ApplicationRegisterInput{
		Name: name,
		Labels: &graphql.Labels{
			scenariosLabel: []string{"Christmas", "New Year"},
		},
	}

	application := registerApplicationFromInputWithinTenant(t, ctx, appInput, tenantID)
	require.NotEmpty(t, application.ID)
	defer unregisterApplicationInTenant(t, application.ID, tenantID)

	//WHEN
	appLabelRequest := fixSetApplicationLabelRequest(application.ID, scenariosLabel, []string{"Christmas"})
	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, tenantID, appLabelRequest, nil))

	//remove last label
	appLabelRequest = fixSetApplicationLabelRequest(application.ID, scenariosLabel, []string{""})
	err := tc.RunOperationWithCustomTenant(ctx, tenantID, appLabelRequest, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), `must be one of the following: "DEFAULT", "Christmas", "New Year"`)
}
