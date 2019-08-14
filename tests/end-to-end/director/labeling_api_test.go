package director

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func TestCreateLabelWithoutLabelDefinition(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "create-label-without-label-definition"
	application := createApplication(t, ctx, name)
	defer deleteApplication(t, application.ID)

	t.Log("Set label on application")
	labelKey := "test"
	labelValue := "val"

	setLabelRequest := fixSetApplicationLabelRequest(application.ID, labelKey, labelValue)
	label := graphql.Label{}
	defer deleteLabelDefinitionWithinDefaultTenant(t, ctx, labelKey, false)
	defer deleteApplicationLabel(t, ctx, application.ID, labelKey)

	// WHEN
	err := tc.RunQuery(ctx, setLabelRequest, &label)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, label.Key)
	require.NotEmpty(t, label.Value)
	saveQueryInExamples(t, setLabelRequest.Query(), "set application label")

	t.Log("Check if LabelDefinition was created internally")

	getLabelDefinitionRequest := fixLabelDefinitionRequest(labelKey)
	labelDefinition := graphql.LabelDefinition{}

	err = tc.RunQuery(ctx, getLabelDefinitionRequest, &labelDefinition)

	require.NoError(t, err)
	require.NotEmpty(t, labelDefinition)
	assert.Equal(t, label.Key, labelDefinition.Key)
	saveQueryInExamples(t, getLabelDefinitionRequest.Query(), "query label definition")
}

func TestCreateLabelWithExistingLabelDefinition_ShouldFail(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	applicationName := "create-label-with-existing-label-definition"

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
		Schema: &schema,
	}

	labelDefinitionInputGQL, err := tc.graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
	require.NoError(t, err)

	t.Run("should fail - label value doesn't match json schema provided in label definition", func(t *testing.T) {
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(labelDefinitionInputGQL)
		labelDefinition := graphql.LabelDefinition{}

		t.Log("Create application")
		application := createApplication(t, ctx, applicationName)
		defer deleteApplication(t, application.ID)

		t.Log("Create label definition")
		err = tc.RunQuery(ctx, createLabelDefinitionRequest, &labelDefinition)

		require.NoError(t, err)
		defer deleteLabelDefinitionWithinDefaultTenant(t, ctx, labelKey, false)
		assert.Equal(t, labelKey, labelDefinition.Key)

		invalidLabelValue := 123
		setLabelRequest := fixSetApplicationLabelRequest(application.ID, labelKey, invalidLabelValue)

		// WHEN
		t.Log("Try to set label on application with invalid value against given json schema")
		err = tc.RunQuery(ctx, setLabelRequest, nil)

		//THEN
		require.Error(t, err)
		errMsg := fmt.Sprintf("graphql: while creating label for Application: while validating Label value for '%s': while validating value %d against JSON Schema: map[properties:map[foo:map[description:foo type:string]] required:[foo] title:foobarbaz type:object]: (root): Invalid type. Expected: object, given: integer", labelKey, invalidLabelValue)
		assert.EqualError(t, err, errMsg)
		saveQueryInExamples(t, createLabelDefinitionRequest.Query(), "create label definition")

	})

	t.Run("should succeed - label value matches json schema in label definition", func(t *testing.T) {
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(labelDefinitionInputGQL)
		labelDefinition := graphql.LabelDefinition{}

		t.Log("Create application")
		application := createApplication(t, ctx, applicationName)
		defer deleteApplication(t, application.ID)

		t.Log("Create label definition")
		err = tc.RunQuery(ctx, createLabelDefinitionRequest, &labelDefinition)

		t.Log("Set label on application with valid value")
		validLabelValue := map[string]interface{}{
			labelKey: "bar",
		}

		var appLabel interface{} = validLabelValue

		setLabelRequest := fixSetApplicationLabelRequest(application.ID, labelKey, appLabel)
		label := graphql.Label{}

		err = tc.RunQuery(ctx, setLabelRequest, &label)
		defer deleteLabelDefinitionWithinDefaultTenant(t, ctx, labelKey, false)
		defer deleteApplicationLabel(t, ctx, application.ID, labelKey)

		require.NoError(t, err)
		require.NotEmpty(t, label.Key)
		require.NotEmpty(t, label.Value)

		t.Log("Check if Label was set on application")
		queryAppReq := fixApplicationRequest(application.ID)

		// WHEN
		err = tc.RunQuery(context.Background(), queryAppReq, &application)

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
		Schema: &schema,
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
		app := createApplication(t, ctx, "app")
		defer deleteApplication(t, app.ID)

		t.Log("Create label definition")
		err = tc.RunQuery(ctx, createLabelDefinitionRequest, &labelDefinition)
		require.NoError(t, err)

		t.Log("Set label on application")
		setLabelRequest := fixSetApplicationLabelRequest(app.ID, labelKey, appLabel)
		label := graphql.Label{}

		err = tc.RunQuery(ctx, setLabelRequest, &label)
		defer deleteLabelDefinitionWithinDefaultTenant(t, ctx, labelKey, false)
		defer deleteApplicationLabel(t, ctx, app.ID, labelKey)

		var invalidSchema interface{} = invalidJsonSchema
		labelDefinitionInput = graphql.LabelDefinitionInput{
			Key:    labelKey,
			Schema: &invalidSchema,
		}

		ldInputGql, err := tc.graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
		require.NoError(t, err)

		updateLabelDefinitionReq := fixUpdateLabelDefinitionRequest(ldInputGql)

		// WHEN
		t.Log("Try to edit LabelDefinition with incompatible data")
		err = tc.RunQuery(context.Background(), updateLabelDefinitionReq, nil)

		//THEN
		require.Error(t, err)
		errString := fmt.Sprintf("graphql: while updating label definition: label with key \"%s\" is not valid against new schema for Application with ID \"%s\": %s: Invalid type. Expected: integer, given: string", labelKey, app.ID, labelKey)
		assert.EqualError(t, err, errString)
	})

	t.Run("Edit LabelDefinition with compatible data", func(t *testing.T) {
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(labelDefinitionInputGQL)
		labelDefinition := graphql.LabelDefinition{}

		t.Log("Create application")
		app := createApplication(t, ctx, "app")
		defer deleteApplication(t, app.ID)

		t.Log("Create label definition")
		err = tc.RunQuery(ctx, createLabelDefinitionRequest, &labelDefinition)
		require.NoError(t, err)

		t.Log("Set label on application")
		setLabelRequest := fixSetApplicationLabelRequest(app.ID, labelKey, appLabel)
		label := graphql.Label{}

		err = tc.RunQuery(ctx, setLabelRequest, &label)
		defer deleteLabelDefinitionWithinDefaultTenant(t, ctx, labelKey, false)
		defer deleteApplicationLabel(t, ctx, app.ID, labelKey)

		var newSchema interface{} = newValidJsonSchema
		labelDefinitionInput = graphql.LabelDefinitionInput{
			Key:    labelKey,
			Schema: &newSchema,
		}

		ldInputGql, err := tc.graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
		require.NoError(t, err)

		updateLabelDefinitionReq := fixUpdateLabelDefinitionRequest(ldInputGql)

		// WHEN
		t.Log("Edit LabelDefinition with compatible data")
		err = tc.RunQuery(context.Background(), updateLabelDefinitionReq, &labelDefinition)

		//THEN
		require.NoError(t, err)

		schemaVal, ok := (*labelDefinition.Schema).(map[string]interface{})
		require.True(t, ok)
		actualProperties, ok := schemaVal["properties"].(map[string]interface{})
		require.True(t, ok)

		expectedProperties, ok := newValidJsonSchema["properties"].(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, expectedProperties, actualProperties)
		saveQueryInExamples(t, updateLabelDefinitionReq.Query(), "update label definition")
	})
}

func TestCreateScenariosLabel(t *testing.T) {
	// GIVEN
	t.Log("Create application")
	ctx := context.Background()
	app := createApplication(t, ctx, "app")
	defer deleteApplication(t, app.ID)

	t.Log("Check if scenarios LabelDefinition exists")
	labelKey := "scenarios"

	getLabelDefinition := fixLabelDefinitionRequest(labelKey)
	ld := graphql.LabelDefinition{}

	err := tc.RunQuery(ctx, getLabelDefinition, &ld)
	require.NoError(t, err)

	t.Log("Check if app was labeled with scenarios=default")

	getApp := fixApplicationRequest(app.ID)
	actualApp := graphql.Application{}
	// WHEN
	err = tc.RunQuery(ctx, getApp, &actualApp)

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
	app := createApplication(t, ctx, "app")
	defer deleteApplication(t, app.ID)
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
		Schema: &schema,
	}

	ldInputGQL, err := tc.graphqlizer.LabelDefinitionInputToGQL(ldInput)
	require.NoError(t, err)

	updateLabelDefinitionRequest := fixUpdateLabelDefinitionRequest(ldInputGQL)
	labelDefinition := graphql.LabelDefinition{}

	err = tc.RunQuery(ctx, updateLabelDefinitionRequest, &labelDefinition)

	require.NoError(t, err)

	scenarios := []string{defaultValue, additionalValue}
	var labelValue interface{} = scenarios

	t.Logf("Set scenario label value %s on application", additionalValue)
	setApplicationLabel(t, ctx, app.ID, labelKey, labelValue)

	t.Log("Check if new scenario label value was set correctly")
	appRequest := fixApplicationRequest(app.ID)
	app = ApplicationExt{}

	err = tc.RunQuery(ctx, appRequest, &app)
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
		Schema: &schema,
	}

	ldInputGql, err := tc.graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
	require.NoError(t, err)

	t.Run("Try to delete Label Definition while it's being used by some labels - should fail", func(t *testing.T) {

		t.Log("Create application")
		app := createApplication(t, ctx, "app")
		defer deleteApplication(t, app.ID)

		t.Log("Create LabelDefinition")
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(ldInputGql)
		ld := graphql.LabelDefinition{}

		err = tc.RunQuery(ctx, createLabelDefinitionRequest, ld)
		require.NoError(t, err)

		t.Log("Set label on application")
		validLabelValue := map[string]interface{}{"foo": "test"}

		setLabelRequest := fixSetApplicationLabelRequest(app.ID, labelKey, validLabelValue)
		label := graphql.Label{}

		err = tc.RunQuery(ctx, setLabelRequest, &label)
		require.NoError(t, err)
		defer deleteLabelDefinitionWithinDefaultTenant(t, ctx, labelKey, false)
		defer deleteApplicationLabel(t, ctx, app.ID, labelKey)

		t.Log("Try to delete Label Definition while it's being used by some labels")

		deleteLabelDefinitionRequest := fixDeleteLabelDefinition(labelKey, false)
		err = tc.RunQuery(context.Background(), deleteLabelDefinitionRequest, nil)
		require.Error(t, err)
		assert.EqualError(t, err, "graphql: could not delete label definition, it is already used by at least one label")
		saveQueryInExamples(t, deleteLabelDefinitionRequest.Query(), "delete label definition")
	})

	t.Run("Delete Label from application, then delete the Label Definition - should succeed", func(t *testing.T) {

		t.Log("Create application")
		app := createApplication(t, ctx, "app")
		defer deleteApplication(t, app.ID)

		t.Log("Create LabelDefinition")
		createLabelDefinitionRequest := fixCreateLabelDefinitionRequest(ldInputGql)
		ld := graphql.LabelDefinition{}

		err = tc.RunQuery(ctx, createLabelDefinitionRequest, ld)
		require.NoError(t, err)

		t.Log("Set label on application")
		validLabelValue := map[string]interface{}{labelKey: "test"}

		setLabelRequest := fixSetApplicationLabelRequest(app.ID, labelKey, validLabelValue)
		label := graphql.Label{}

		err = tc.RunQuery(ctx, setLabelRequest, &label)
		require.NoError(t, err)

		deleteApplicationLabelRequest := fixDeleteApplicationLabel(app.ID, labelKey)
		label = graphql.Label{}

		err := tc.RunQuery(ctx, deleteApplicationLabelRequest, &label)
		require.NoError(t, err)
		assert.Equal(t, labelKey, label.Key)

		deleteLabelDefinitionRequest := fixDeleteLabelDefinition(labelKey, false)
		labelDefinition := graphql.LabelDefinition{}

		err = tc.RunQuery(context.Background(), deleteLabelDefinitionRequest, &labelDefinition)
		require.NoError(t, err)
		assert.ObjectsAreEqualValues(labelDefinitionInput.Schema, labelDefinition.Schema)
	})
}

func TestDeleteScenariosLabel(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	t.Log("Create application")
	app := createApplication(t, ctx, "app")
	defer deleteApplication(t, app.ID)

	t.Log("Try to delete scenarios label on application")
	labelKey := "scenarios"
	deleteApplicationLabelRequest := fixDeleteApplicationLabel(app.ID, labelKey)

	// WHEN
	err := tc.RunQuery(ctx, deleteApplicationLabelRequest, nil)

	//THEN
	require.Error(t, err)
	assert.EqualError(t, err, "graphql: scenarios label can not be deleted from application")
}

func TestDeleteDefaultValueInScenariosLabelDefinition(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	t.Log("Create application")
	app := createApplication(t, ctx, "app")
	defer deleteApplication(t, app.ID)
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
		Schema: &schema,
	}

	ldInputGQL, err := tc.graphqlizer.LabelDefinitionInputToGQL(ldInput)
	require.NoError(t, err)

	updateLabelDefinitionRequest := fixUpdateLabelDefinitionRequest(ldInputGQL)
	labelDefinition := graphql.LabelDefinition{}

	// WHEN
	err = tc.RunQuery(ctx, updateLabelDefinitionRequest, &labelDefinition)
	errMsg := fmt.Sprintf(`graphql: while updating label definition: while validating Label Definition: while validating schema for key %s: items.enum: At least one of the items must match, items.enum.0: items.enum.0 does not match: "%s"`, labelKey, defaultValue)

	// THEN
	require.Error(t, err)
	assert.EqualError(t, err, errMsg)
}

func TestSearchByLabels(t *testing.T) {
	//TODO: return this test after implementing filtering on applications
	t.SkipNow()
	//Create first application
	// GIVEN
	ctx := context.Background()

	first := 5
	after := ""
	firstApp := createApplication(t, ctx, "first")
	require.NotEmpty(t, firstApp.ID)
	defer deleteApplication(t, firstApp.ID)

	//Create second application
	secondApp := createApplication(t, ctx, "second")
	require.NotEmpty(t, firstApp.ID)
	defer deleteApplication(t, firstApp.ID)

	//Set label "foo" on both applications
	labelKeyFoo := "foo"
	labelValueFoo := "val"

	setLabelRequest := fixSetApplicationLabelRequest(firstApp.ID, labelKeyFoo, labelValueFoo)
	firstAppLabel := graphql.Label{}
	err := tc.RunQuery(ctx, setLabelRequest, &firstAppLabel)
	require.NoError(t, err)
	require.NotEmpty(t, firstAppLabel.Key)
	require.NotEmpty(t, firstAppLabel.Value)

	setLabelRequest = fixSetApplicationLabelRequest(secondApp.ID, labelKeyFoo, labelValueFoo)
	secondAppLabel := graphql.Label{}
	err = tc.RunQuery(ctx, setLabelRequest, &secondAppLabel)
	require.NoError(t, err)
	require.NotEmpty(t, secondAppLabel.Key)
	require.NotEmpty(t, secondAppLabel.Value)

	//Set label "bar" on first application
	labelKeyBar := "bar"
	labelValueBar := "barval"

	setLabelRequest = fixSetApplicationLabelRequest(firstApp.ID, labelKeyBar, labelValueBar)
	firstAppBarLabel := graphql.Label{}
	err = tc.RunQuery(ctx, setLabelRequest, &firstAppBarLabel)
	require.NoError(t, err)
	require.NotEmpty(t, firstAppLabel.Key)
	require.NotEmpty(t, firstAppLabel.Value)

	// Query for application with LabelFilter "foo"
	labelFilter := graphql.LabelFilter{
		Key:   labelKeyFoo,
		Query: nil,
	}

	//WHEN
	labelFilterGQL, err := tc.graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	applicationRequest := fixApplications(labelFilterGQL, first, after)
	applicationPage := ApplicationPageExt{}
	err = tc.RunQuery(ctx, applicationRequest, &applicationPage)
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
		Key:   labelKeyBar,
		Query: nil,
	}

	// WHEN
	labelFilterGQL, err = tc.graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	applicationRequest = fixApplications(labelFilterGQL, first, after)
	applicationPage = ApplicationPageExt{}
	err = tc.RunQuery(ctx, applicationRequest, &applicationPage)
	require.NoError(t, err)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, applicationPage)
	assert.Equal(t, applicationPage.TotalCount, 1)
	assert.Contains(t, applicationPage.Data[0].Labels, labelKeyBar)
	assert.Equal(t, applicationPage.Data[0].Labels[labelKeyBar], labelValueBar)
	saveQueryInExamples(t, applicationRequest.Query(), "query application with filter")

}

func TestListLabelDefinitions(t *testing.T) {
	//GIVEN
	tenantID := uuid.New().String()
	ctx := context.TODO()
	firstSchema := map[string]interface{}{
		"test": "test",
	}

	var schema interface{} = firstSchema

	firstLabelDefinitionKey := "first"
	firstLabelDefinitionInput := graphql.LabelDefinitionInput{
		Key:    firstLabelDefinitionKey,
		Schema: &schema,
	}
	firstLabelDefinition := createLabelDefinitionWithinTenant(t, ctx, firstLabelDefinitionInput, tenantID)
	defer deleteLabelDefinitionWithinTenant(t, ctx, firstLabelDefinitionKey, false, tenantID)

	secondSchema := map[string]interface{}{
		"test": "test",
	}
	schema = secondSchema

	secondLabelDefinitionKey := "second"
	secondLabelDefinitionInput := graphql.LabelDefinitionInput{
		Key:    secondLabelDefinitionKey,
		Schema: &schema,
	}
	secondLabelDefinition := createLabelDefinitionWithinTenant(t, ctx, secondLabelDefinitionInput, tenantID)
	defer deleteLabelDefinitionWithinTenant(t, ctx, secondLabelDefinitionKey, false, tenantID)

	//WHEN
	labelDefinitions, err := listLabelDefinitionsWithinTenant(t, ctx, tenantID)

	//THEN
	require.NoError(t, err)
	require.Len(t, labelDefinitions, 2)
	assert.Contains(t, labelDefinitions, firstLabelDefinition)
	assert.Contains(t, labelDefinitions, secondLabelDefinition)
}

func TestDeletingLastScenarioForApplication_ShouldFail(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	tenant := uuid.New().String()
	name := "test-deleting-last-scenario-for-application-should-fail"
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

	labelDefinitionInput := graphql.LabelDefinitionInput{
		Key:    scenariosLabel,
		Schema: &schema,
	}
	createLabelDefinitionWithinTenant(t, ctx, labelDefinitionInput, tenant)

	appInput := graphql.ApplicationInput{Name: name,
		Labels: &graphql.Labels{
			scenariosLabel: []string{"Christmas", "New Year"},
		}}

	application := createApplicationFromInputWithinTenant(t, ctx, appInput, tenant)
	require.NotEmpty(t, application.ID)
	defer deleteApplicationInTenant(t, application.ID, tenant)

	//WHEN
	appLabelRequest := fixSetApplicationLabelRequest(application.ID, scenariosLabel, []string{"Christmas"})
	appLabelRequest.Header["Tenant"] = []string{tenant}
	require.NoError(t, tc.RunQuery(ctx, appLabelRequest, nil))

	//remove last label
	appLabelRequest = fixSetApplicationLabelRequest(application.ID, scenariosLabel, []string{""})
	appLabelRequest.Header["Tenant"] = []string{tenant}
	err := tc.RunQuery(ctx, appLabelRequest, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), `0 must be one of the following: "DEFAULT", "Christmas", "New Year"`)
}

func TestCreateRuntimeWithoutLabels(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	name := "test-create-runtime-without-labels"
	runtimeInput := &graphql.RuntimeInput{Name: name}

	runtime := createRuntimeFromInput(t, ctx, runtimeInput)
	defer deleteRuntime(t, runtime.ID)

	//WHEN
	fetchedRuntime := getRuntime(t, ctx, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
}
