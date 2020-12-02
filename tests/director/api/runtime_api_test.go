package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	scenariosLabel          = "scenarios"
	shouldNormalize         = "shouldNormalize"
	queryRuntimesCategory   = "query runtimes"
	registerRuntimeCategory = "register runtime"
)

func TestRuntimeRegisterUpdateAndUnregister(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	givenInput := graphql.RuntimeInput{
		Name:        "runtime-create-update-delete",
		Description: ptr.String("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []interface{}{"hhh"}},
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	actualRuntime := graphql.RuntimeExt{}

	// WHEN
	registerReq := fixRegisterRuntimeRequest(runtimeInGQL)
	saveExampleInCustomDir(t, registerReq.Query(), registerRuntimeCategory, "register runtime")
	err = tc.RunOperation(ctx, registerReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)
	assertRuntime(t, givenInput, actualRuntime)

	// add Label
	actualLabel := graphql.Label{}

	// WHEN
	addLabelReq := fixSetRuntimeLabelRequest(actualRuntime.ID, "new_label", []string{"bbb"})
	err = tc.RunOperation(ctx, addLabelReq, &actualLabel)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, "new_label", actualLabel.Key)
	assert.Len(t, actualLabel.Value, 1)
	assert.Contains(t, actualLabel.Value, "bbb")

	// get runtime and validate runtimes
	getRuntimeReq := fixRuntimeRequest(actualRuntime.ID)
	err = tc.RunOperation(ctx, getRuntimeReq, &actualRuntime)
	require.NoError(t, err)
	assert.Len(t, actualRuntime.Labels, 4)

	// add agent auth
	// GIVEN
	in := fixSampleApplicationRegisterInput("app")

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	createAppReq := fixRegisterApplicationRequest(appInputGQL)

	//WHEN
	actualApp := graphql.ApplicationExt{}
	err = tc.RunOperation(ctx, createAppReq, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)

	// update runtime, check if only simple values are updated
	//GIVEN
	givenInput.Name = "updated-name"
	givenInput.Description = ptr.String("updated-description")
	givenInput.Labels = &graphql.Labels{
		"key": []interface{}{"values", "aabbcc"},
	}
	runtimeStatusCond := graphql.RuntimeStatusConditionConnected
	givenInput.StatusCondition = &runtimeStatusCond

	runtimeInGQL, err = tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	updateRuntimeReq := fixUpdateRuntimeRequest(actualRuntime.ID, runtimeInGQL)
	saveExample(t, updateRuntimeReq.Query(), "update runtime")
	//WHEN
	actualRuntime = graphql.RuntimeExt{}
	err = tc.RunOperation(ctx, updateRuntimeReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, givenInput.Name, actualRuntime.Name)
	assert.Equal(t, *givenInput.Description, *actualRuntime.Description)
	assert.Equal(t, len(actualRuntime.Labels), 2)
	assert.Equal(t, runtimeStatusCond, actualRuntime.Status.Condition)

	// delete runtime

	// WHEN
	delReq := fixUnregisterRuntimeRequest(actualRuntime.ID)
	saveExample(t, delReq.Query(), "unregister runtime")
	err = tc.RunOperation(ctx, delReq, nil)

	//THEN
	require.NoError(t, err)
}

func TestRuntimeUnregisterDeletesScenarioAssignments(t *testing.T) {
	const (
		labelKey     = "labelKey"
		labelValue   = "labelValue"
		testScenario = "test-scenario"
	)
	// GIVEN
	ctx := context.Background()
	givenInput := graphql.RuntimeInput{
		Name:        "runtime-with-scenario-assignments",
		Description: ptr.String("runtime-1-description"),
		Labels:      &graphql.Labels{labelKey: []interface{}{labelValue}},
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	actualRuntime := graphql.RuntimeExt{}

	// WHEN
	registerReq := fixRegisterRuntimeRequest(runtimeInGQL)
	err = tc.RunOperation(ctx, registerReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)
	assertRuntime(t, givenInput, actualRuntime)

	// update label definition
	const defaultValue = "DEFAULT"
	enumValue := []string{defaultValue, testScenario}
	jsonSchema := map[string]interface{}{
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
		"items": map[string]interface{}{
			"type": "string",
			"enum": enumValue,
		},
	}
	var schema interface{} = jsonSchema
	marshalledSchema := marshalJSONSchema(t, schema)

	givenLabelDef := graphql.LabelDefinitionInput{
		Key:    scenariosLabel,
		Schema: marshalledSchema,
	}
	labelDefInGQL, err := tc.graphqlizer.LabelDefinitionInputToGQL(givenLabelDef)
	require.NoError(t, err)
	actualLabelDef := graphql.LabelDefinition{}

	// WHEN
	updateLabelDefReq := fixUpdateLabelDefinitionRequest(labelDefInGQL)
	err = tc.RunOperation(ctx, updateLabelDefReq, &actualLabelDef)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, givenLabelDef.Key, actualLabelDef.Key)
	assertGraphQLJSONSchema(t, givenLabelDef.Schema, actualLabelDef.Schema)

	// register automatic sccenario assignment
	givenScenarioAssignment := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: testScenario,
		Selector: &graphql.LabelSelectorInput{
			Key:   labelKey,
			Value: labelValue,
		},
	}

	scenarioAssignmentInGQL, err := tc.graphqlizer.AutomaticScenarioAssignmentSetInputToGQL(givenScenarioAssignment)
	require.NoError(t, err)
	actualScenarioAssignment := graphql.AutomaticScenarioAssignment{}

	// WHEN
	createAutomaticScenarioAssignmentReq := fixCreateAutomaticScenarioAssignmentRequest(scenarioAssignmentInGQL)
	err = tc.RunOperation(ctx, createAutomaticScenarioAssignmentReq, &actualScenarioAssignment)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, givenScenarioAssignment.ScenarioName, actualScenarioAssignment.ScenarioName)
	assert.Equal(t, givenScenarioAssignment.Selector.Key, actualScenarioAssignment.Selector.Key)
	assert.Equal(t, givenScenarioAssignment.Selector.Value, actualScenarioAssignment.Selector.Value)

	// get runtime - verify it is in scenario
	getRuntimeReq := fixRuntimeRequest(actualRuntime.ID)
	err = tc.RunOperation(ctx, getRuntimeReq, &actualRuntime)

	require.NoError(t, err)
	scenarios, hasScenarios := actualRuntime.Labels["scenarios"]
	assert.True(t, hasScenarios)
	assert.Len(t, scenarios, 2)
	assert.Contains(t, scenarios, testScenario)

	// delete runtime

	// WHEN
	delReq := fixUnregisterRuntimeRequest(actualRuntime.ID)
	err = tc.RunOperation(ctx, delReq, nil)

	//THEN
	require.NoError(t, err)

	// get automatic scenario assignment - see that it's deleted
	actualScenarioAssignments := graphql.AutomaticScenarioAssignmentPage{}
	getScenarioAssignmentsReq := fixAutomaticScenarioAssignmentsRequest()
	err = tc.RunOperation(ctx, getScenarioAssignmentsReq, &actualScenarioAssignments)
	require.NoError(t, err)
	assert.Equal(t, actualScenarioAssignments.TotalCount, 0)
}

func TestRuntimeCreateUpdateDuplicatedNames(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	firstRuntimeName := "unique-name-1"
	givenInput := graphql.RuntimeInput{
		Name:        firstRuntimeName,
		Description: ptr.String("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []interface{}{"hhh"}},
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	firstRuntime := graphql.RuntimeExt{}
	registerReq := fixRegisterRuntimeRequest(runtimeInGQL)

	// WHEN
	err = tc.RunOperation(ctx, registerReq, &firstRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, firstRuntime.ID)
	assertRuntime(t, givenInput, firstRuntime)
	defer unregisterRuntime(t, firstRuntime.ID)

	// try to create second runtime with first runtime name
	//GIVEN
	givenInput = graphql.RuntimeInput{
		Name:        firstRuntimeName,
		Description: ptr.String("runtime-1-description"),
	}
	runtimeInGQL, err = tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	registerReq = fixRegisterRuntimeRequest(runtimeInGQL)
	saveExampleInCustomDir(t, registerReq.Query(), registerRuntimeCategory, "register runtime")

	// WHEN
	err = tc.RunOperation(ctx, registerReq, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not unique")

	// create second runtime
	//GIVEN
	secondRuntimeName := "unique-name-2"
	givenInput = graphql.RuntimeInput{
		Name:        secondRuntimeName,
		Description: ptr.String("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []interface{}{"hhh"}},
	}
	runtimeInGQL, err = tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	secondRuntime := graphql.RuntimeExt{}
	registerReq = fixRegisterRuntimeRequest(runtimeInGQL)

	// WHEN
	err = tc.RunOperation(ctx, registerReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, secondRuntime.ID)
	assertRuntime(t, givenInput, secondRuntime)
	defer unregisterRuntime(t, secondRuntime.ID)

	//Update first runtime with second runtime name, failed

	//GIVEN
	givenInput = graphql.RuntimeInput{
		Name:        secondRuntimeName,
		Description: ptr.String("runtime-1-description"),
	}
	runtimeInGQL, err = tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	registerReq = fixUpdateRuntimeRequest(firstRuntime.ID, runtimeInGQL)

	// WHEN
	err = tc.RunOperation(ctx, registerReq, &secondRuntime)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not unique")
}

func TestQueryRuntimes(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	idsToRemove := make([]string, 0)
	defer func() {
		for _, id := range idsToRemove {
			if id != "" {
				unregisterRuntime(t, id)
			}
		}
	}()

	inputRuntimes := []*graphql.Runtime{
		{Name: "runtime-query-1", Description: ptr.String("test description")},
		{Name: "runtime-query-2", Description: ptr.String("another description")},
		{Name: "runtime-query-3"},
	}

	for _, rtm := range inputRuntimes {
		givenInput := graphql.RuntimeInput{
			Name:        rtm.Name,
			Description: rtm.Description,
		}
		runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
		require.NoError(t, err)
		createReq := fixRegisterRuntimeRequest(runtimeInGQL)
		actualRuntime := graphql.Runtime{}
		err = tc.RunOperation(ctx, createReq, &actualRuntime)
		require.NoError(t, err)
		require.NotEmpty(t, actualRuntime.ID)
		rtm.ID = actualRuntime.ID
		idsToRemove = append(idsToRemove, actualRuntime.ID)
	}
	actualPage := graphql.RuntimePage{}

	// WHEN
	queryReq := fixRuntimesRequest()
	err := tc.RunOperation(ctx, queryReq, &actualPage)
	saveExampleInCustomDir(t, queryReq.Query(), queryRuntimesCategory, "query runtimes")

	//THEN
	require.NoError(t, err)
	assert.Len(t, actualPage.Data, len(inputRuntimes))
	assert.Equal(t, len(inputRuntimes), actualPage.TotalCount)

	for _, inputRtm := range inputRuntimes {
		found := false
		for _, actualRtm := range actualPage.Data {
			if inputRtm.ID == actualRtm.ID {
				found = true
				assert.Equal(t, inputRtm.Name, actualRtm.Name)
				assert.Equal(t, inputRtm.Description, actualRtm.Description)
				break
			}
		}
		assert.True(t, found)
	}
}

func TestQuerySpecificRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	givenInput := graphql.RuntimeInput{
		Name: "runtime-specific-runtime",
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	registerReq := fixRegisterRuntimeRequest(runtimeInGQL)
	createdRuntime := graphql.Runtime{}
	err = tc.RunOperation(ctx, registerReq, &createdRuntime)
	require.NoError(t, err)
	require.NotEmpty(t, createdRuntime.ID)

	defer unregisterRuntime(t, createdRuntime.ID)
	queriedRuntime := graphql.Runtime{}

	// WHEN
	queryReq := fixRuntimeRequest(createdRuntime.ID)
	err = tc.RunOperation(ctx, queryReq, &queriedRuntime)
	saveExample(t, queryReq.Query(), "query runtime")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdRuntime.ID, queriedRuntime.ID)
	assert.Equal(t, createdRuntime.Name, queriedRuntime.Name)
	assert.Equal(t, createdRuntime.Description, queriedRuntime.Description)
}

func TestQueryRuntimesWithPagination(t *testing.T) {
	//GIVEN
	ctx := context.Background()
	runtimes := make(map[string]*graphql.Runtime)
	runtimesAmount := 10
	for i := 0; i < runtimesAmount; i++ {
		runtimeInput := graphql.RuntimeInput{
			Name: fmt.Sprintf("runtime-%d", i),
		}
		runtimeInputGQL, err := tc.graphqlizer.RuntimeInputToGQL(runtimeInput)
		require.NoError(t, err)

		registerReq := fixRegisterRuntimeRequest(runtimeInputGQL)

		runtime := graphql.Runtime{}
		err = tc.RunOperation(ctx, registerReq, &runtime)

		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)
		defer unregisterRuntime(t, runtime.ID)
		runtimes[runtime.ID] = &runtime
	}

	after := 3
	cursor := ""
	queriesForFullPage := int(runtimesAmount / after)

	for i := 0; i < queriesForFullPage; i++ {
		runtimesRequest := fixRuntimeRequestWithPaginationRequest(after, cursor)

		//WHEN
		runtimePage := graphql.RuntimePage{}
		err := tc.RunOperation(ctx, runtimesRequest, &runtimePage)
		require.NoError(t, err)

		//THEN
		assert.Equal(t, cursor, string(runtimePage.PageInfo.StartCursor))
		assert.True(t, runtimePage.PageInfo.HasNextPage)
		assert.Len(t, runtimePage.Data, after)
		assert.Equal(t, runtimesAmount, runtimePage.TotalCount)
		for _, runtime := range runtimePage.Data {
			assert.Equal(t, runtime, runtimes[runtime.ID])
			delete(runtimes, runtime.ID)
		}
		cursor = string(runtimePage.PageInfo.EndCursor)
	}

	//WHEN get last page with last runtime
	runtimesRequest := fixRuntimeRequestWithPaginationRequest(after, cursor)
	lastRuntimePage := graphql.RuntimePage{}
	err := tc.RunOperation(ctx, runtimesRequest, &lastRuntimePage)
	require.NoError(t, err)
	saveExampleInCustomDir(t, runtimesRequest.Query(), queryRuntimesCategory, "query runtimes with pagination")

	//THEN
	assert.False(t, lastRuntimePage.PageInfo.HasNextPage)
	assert.Empty(t, lastRuntimePage.PageInfo.EndCursor)
	require.Len(t, lastRuntimePage.Data, 1)
	assert.Equal(t, lastRuntimePage.Data[0], runtimes[lastRuntimePage.Data[0].ID])
	delete(runtimes, lastRuntimePage.Data[0].ID)
	assert.Len(t, runtimes, 0)
}

func TestRegisterUpdateRuntimeWithoutLabels(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	name := "test-create-runtime-without-labels"
	runtimeInput := graphql.RuntimeInput{Name: name}

	runtime := registerRuntimeFromInput(t, ctx, &runtimeInput)
	defer unregisterRuntime(t, runtime.ID)

	//WHEN
	fetchedRuntime := getRuntime(t, ctx, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertRuntime(t, runtimeInput, *fetchedRuntime)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := graphql.RuntimeInput{
		Name:        name,
		Description: ptr.String("runtime-1-description"),
		Labels:      &graphql.Labels{scenariosLabel: []interface{}{"DEFAULT"}},
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(secondInput)
	require.NoError(t, err)
	updateReq := fixUpdateRuntimeRequest(fetchedRuntime.ID, runtimeInGQL)

	// WHEN
	err = tc.RunOperation(ctx, updateReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	assertRuntime(t, secondInput, secondRuntime)
}

func TestRegisterUpdateRuntimeWithShouldNormalizeLabel(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	name := "test-create-runtime-without-labels"
	runtimeInput := graphql.RuntimeInput{
		Name:   name,
		Labels: &graphql.Labels{shouldNormalize: "false"},
	}

	runtime := registerRuntimeFromInput(t, ctx, &runtimeInput)
	defer unregisterRuntime(t, runtime.ID)

	//WHEN
	fetchedRuntime := getRuntime(t, ctx, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertRuntime(t, runtimeInput, *fetchedRuntime)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := graphql.RuntimeInput{
		Name:        name,
		Description: ptr.String("runtime-1-description"),
		Labels:      &graphql.Labels{shouldNormalize: "true", scenariosLabel: []interface{}{"DEFAULT"}},
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(secondInput)
	require.NoError(t, err)
	updateReq := fixUpdateRuntimeRequest(fetchedRuntime.ID, runtimeInGQL)

	// WHEN
	err = tc.RunOperation(ctx, updateReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	assertRuntime(t, secondInput, secondRuntime)
}
