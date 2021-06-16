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

	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ScenariosLabel          = "scenarios"
	IsNormalizedLabel       = "isNormalized"
	QueryRuntimesCategory   = "query runtimes"
	RegisterRuntimeCategory = "register runtime"
)

func TestRuntimeRegisterUpdateAndUnregister(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	givenInput := graphql.RuntimeInput{
		Name:        "runtime-create-update-delete",
		Description: ptr.String("runtime-1-description"),
		Labels:      graphql.Labels{"ggg": []interface{}{"hhh"}},
	}
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	actualRuntime := graphql.RuntimeExt{}

	// WHEN
	registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
	saveExampleInCustomDir(t, registerReq.Query(), RegisterRuntimeCategory, "register runtime")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)
	assertions.AssertRuntime(t, givenInput, actualRuntime, conf.DefaultScenarioEnabled)

	// add Label
	actualLabel := graphql.Label{}

	// WHEN
	addLabelReq := fixtures.FixSetRuntimeLabelRequest(actualRuntime.ID, "new_label", []string{"bbb"})
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, addLabelReq, &actualLabel)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, "new_label", actualLabel.Key)
	assert.Len(t, actualLabel.Value, 1)
	assert.Contains(t, actualLabel.Value, "bbb")

	// get runtime and validate runtimes
	getRuntimeReq := fixtures.FixGetRuntimeRequest(actualRuntime.ID)
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, getRuntimeReq, &actualRuntime)
	require.NoError(t, err)
	if conf.DefaultScenarioEnabled {
		assert.Len(t, actualRuntime.Labels, 4)
	} else {
		assert.Len(t, actualRuntime.Labels, 3)
	}

	// add agent auth
	// GIVEN
	in := fixtures.FixSampleApplicationRegisterInputWithWebhooks("app")

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	createAppReq := fixtures.FixRegisterApplicationRequest(appInputGQL)

	//WHEN
	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createAppReq, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, actualApp.ID)

	// update runtime, check if only simple values are updated
	//GIVEN
	givenInput.Name = "updated-name"
	givenInput.Description = ptr.String("updated-description")
	givenInput.Labels = graphql.Labels{
		"key": []interface{}{"values", "aabbcc"},
	}
	runtimeStatusCond := graphql.RuntimeStatusConditionConnected
	givenInput.StatusCondition = &runtimeStatusCond

	runtimeInGQL, err = testctx.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	updateRuntimeReq := fixtures.FixUpdateRuntimeRequest(actualRuntime.ID, runtimeInGQL)
	saveExample(t, updateRuntimeReq.Query(), "update runtime")
	//WHEN
	actualRuntime = graphql.RuntimeExt{}
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateRuntimeReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, givenInput.Name, actualRuntime.Name)
	assert.Equal(t, *givenInput.Description, *actualRuntime.Description)
	assert.Equal(t, len(actualRuntime.Labels), 2)
	assert.Equal(t, runtimeStatusCond, actualRuntime.Status.Condition)

	// delete runtime

	// WHEN
	delReq := fixtures.FixUnregisterRuntimeRequest(actualRuntime.ID)
	saveExample(t, delReq.Query(), "unregister runtime")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, delReq, nil)

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
		Labels:      graphql.Labels{labelKey: []interface{}{labelValue}},
	}
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	actualRuntime := graphql.RuntimeExt{}

	// WHEN
	registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)
	assertions.AssertRuntime(t, givenInput, actualRuntime, conf.DefaultScenarioEnabled)

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
	marshalledSchema := json.MarshalJSONSchema(t, schema)

	givenLabelDef := graphql.LabelDefinitionInput{
		Key:    ScenariosLabel,
		Schema: marshalledSchema,
	}
	labelDefInGQL, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(givenLabelDef)
	require.NoError(t, err)
	actualLabelDef := graphql.LabelDefinition{}

	// WHEN
	updateLabelDefReq := fixtures.FixUpdateLabelDefinitionRequest(labelDefInGQL)
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateLabelDefReq, &actualLabelDef)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, givenLabelDef.Key, actualLabelDef.Key)
	assertions.AssertGraphQLJSONSchema(t, givenLabelDef.Schema, actualLabelDef.Schema)

	// register automatic sccenario assignment
	givenScenarioAssignment := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: testScenario,
		Selector: &graphql.LabelSelectorInput{
			Key:   labelKey,
			Value: labelValue,
		},
	}

	scenarioAssignmentInGQL, err := testctx.Tc.Graphqlizer.AutomaticScenarioAssignmentSetInputToGQL(givenScenarioAssignment)
	require.NoError(t, err)
	actualScenarioAssignment := graphql.AutomaticScenarioAssignment{}

	// WHEN
	createAutomaticScenarioAssignmentReq := fixtures.FixCreateAutomaticScenarioAssignmentRequest(scenarioAssignmentInGQL)
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createAutomaticScenarioAssignmentReq, &actualScenarioAssignment)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, givenScenarioAssignment.ScenarioName, actualScenarioAssignment.ScenarioName)
	assert.Equal(t, givenScenarioAssignment.Selector.Key, actualScenarioAssignment.Selector.Key)
	assert.Equal(t, givenScenarioAssignment.Selector.Value, actualScenarioAssignment.Selector.Value)

	// get runtime - verify it is in scenario
	getRuntimeReq := fixtures.FixGetRuntimeRequest(actualRuntime.ID)
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, getRuntimeReq, &actualRuntime)

	require.NoError(t, err)
	scenarios, hasScenarios := actualRuntime.Labels["scenarios"]
	assert.True(t, hasScenarios)
	if conf.DefaultScenarioEnabled {
		assert.Len(t, scenarios, 2)
	} else {
		assert.Len(t, scenarios, 1)
	}
	assert.Contains(t, scenarios, testScenario)

	// delete runtime

	// WHEN
	delReq := fixtures.FixUnregisterRuntimeRequest(actualRuntime.ID)
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, delReq, nil)

	//THEN
	require.NoError(t, err)

	// get automatic scenario assignment - see that it's deleted
	actualScenarioAssignments := graphql.AutomaticScenarioAssignmentPage{}
	getScenarioAssignmentsReq := fixtures.FixAutomaticScenarioAssignmentsRequest()
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, getScenarioAssignmentsReq, &actualScenarioAssignments)
	require.NoError(t, err)
	assert.Equal(t, actualScenarioAssignments.TotalCount, 0)
}

func TestRuntimeCreateUpdateDuplicatedNames(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	firstRuntimeName := "unique-name-1"
	givenInput := graphql.RuntimeInput{
		Name:        firstRuntimeName,
		Description: ptr.String("runtime-1-description"),
		Labels:      graphql.Labels{"ggg": []interface{}{"hhh"}},
	}
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	firstRuntime := graphql.RuntimeExt{}
	registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &firstRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, firstRuntime.ID)
	assertions.AssertRuntime(t, givenInput, firstRuntime, conf.DefaultScenarioEnabled)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, firstRuntime.ID)

	// try to create second runtime with first runtime name
	//GIVEN
	givenInput = graphql.RuntimeInput{
		Name:        firstRuntimeName,
		Description: ptr.String("runtime-1-description"),
	}
	runtimeInGQL, err = testctx.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	registerReq = fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
	saveExampleInCustomDir(t, registerReq.Query(), RegisterRuntimeCategory, "register runtime")

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not unique")

	// create second runtime
	//GIVEN
	secondRuntimeName := "unique-name-2"
	givenInput = graphql.RuntimeInput{
		Name:        secondRuntimeName,
		Description: ptr.String("runtime-1-description"),
		Labels:      graphql.Labels{"ggg": []interface{}{"hhh"}},
	}
	runtimeInGQL, err = testctx.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	secondRuntime := graphql.RuntimeExt{}
	registerReq = fixtures.FixRegisterRuntimeRequest(runtimeInGQL)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, secondRuntime.ID)
	assertions.AssertRuntime(t, givenInput, secondRuntime, conf.DefaultScenarioEnabled)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, secondRuntime.ID)

	//Update first runtime with second runtime name, failed

	//GIVEN
	givenInput = graphql.RuntimeInput{
		Name:        secondRuntimeName,
		Description: ptr.String("runtime-1-description"),
	}
	runtimeInGQL, err = testctx.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	registerReq = fixtures.FixUpdateRuntimeRequest(firstRuntime.ID, runtimeInGQL)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &secondRuntime)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not unique")
}

func TestQueryRuntimes(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	idsToRemove := make([]string, 0)
	defer func() {
		for _, id := range idsToRemove {
			if id != "" {
				fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, id)
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
		runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
		require.NoError(t, err)
		createReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
		actualRuntime := graphql.Runtime{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createReq, &actualRuntime)
		require.NoError(t, err)
		require.NotEmpty(t, actualRuntime.ID)
		rtm.ID = actualRuntime.ID
		idsToRemove = append(idsToRemove, actualRuntime.ID)
	}
	actualPage := graphql.RuntimePage{}

	// WHEN
	queryReq := fixtures.FixGetRuntimesRequestWithPagination()
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, queryReq, &actualPage)
	saveExampleInCustomDir(t, queryReq.Query(), QueryRuntimesCategory, "query runtimes")

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

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	givenInput := graphql.RuntimeInput{
		Name: "runtime-specific-runtime",
	}
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
	createdRuntime := graphql.Runtime{}
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &createdRuntime)
	require.NoError(t, err)
	require.NotEmpty(t, createdRuntime.ID)

	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, createdRuntime.ID)
	queriedRuntime := graphql.Runtime{}

	// WHEN
	queryReq := fixtures.FixGetRuntimeRequest(createdRuntime.ID)
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, queryReq, &queriedRuntime)
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

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	runtimes := make(map[string]*graphql.Runtime)
	runtimesAmount := 10
	for i := 0; i < runtimesAmount; i++ {
		runtimeInput := graphql.RuntimeInput{
			Name: fmt.Sprintf("runtime-%d", i),
		}
		runtimeInputGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeInput)
		require.NoError(t, err)

		registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInputGQL)

		runtime := graphql.Runtime{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &runtime)

		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)
		defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)
		runtimes[runtime.ID] = &runtime
	}

	after := 3
	cursor := ""
	queriesForFullPage := int(runtimesAmount / after)

	for i := 0; i < queriesForFullPage; i++ {
		runtimesRequest := fixtures.FixRuntimeRequestWithPaginationRequest(after, cursor)

		//WHEN
		runtimePage := graphql.RuntimePage{}
		err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, runtimesRequest, &runtimePage)
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
	runtimesRequest := fixtures.FixRuntimeRequestWithPaginationRequest(after, cursor)
	lastRuntimePage := graphql.RuntimePage{}
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, runtimesRequest, &lastRuntimePage)
	require.NoError(t, err)
	saveExampleInCustomDir(t, runtimesRequest.Query(), QueryRuntimesCategory, "query runtimes with pagination")

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
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "test-create-runtime-without-labels"
	runtimeInput := graphql.RuntimeInput{Name: name}

	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &runtimeInput)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	//WHEN
	fetchedRuntime := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertions.AssertRuntime(t, runtimeInput, fetchedRuntime, conf.DefaultScenarioEnabled)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := graphql.RuntimeInput{
		Name:        name,
		Description: ptr.String("runtime-1-description"),
		Labels:      graphql.Labels{ScenariosLabel: []interface{}{"DEFAULT"}},
	}
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(secondInput)
	require.NoError(t, err)
	updateReq := fixtures.FixUpdateRuntimeRequest(fetchedRuntime.ID, runtimeInGQL)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	assertions.AssertRuntime(t, secondInput, secondRuntime, conf.DefaultScenarioEnabled)
}

func TestRegisterUpdateRuntimeWithIsNormalizedLabel(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "test-create-runtime-without-labels"
	runtimeInput := graphql.RuntimeInput{
		Name:   name,
		Labels: graphql.Labels{IsNormalizedLabel: "false"},
	}

	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &runtimeInput)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	//WHEN
	fetchedRuntime := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertions.AssertRuntime(t, runtimeInput, fetchedRuntime, conf.DefaultScenarioEnabled)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := graphql.RuntimeInput{
		Name:        name,
		Description: ptr.String("runtime-1-description"),
		Labels:      graphql.Labels{IsNormalizedLabel: "true", ScenariosLabel: []interface{}{"DEFAULT"}},
	}
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(secondInput)
	require.NoError(t, err)
	updateReq := fixtures.FixUpdateRuntimeRequest(fetchedRuntime.ID, runtimeInGQL)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	assertions.AssertRuntime(t, secondInput, secondRuntime, conf.DefaultScenarioEnabled)
}
