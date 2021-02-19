package tests

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"testing"

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	givenInput := graphql.RuntimeInput{
		Name:        "runtime-create-update-delete",
		Description: ptr.String("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []interface{}{"hhh"}},
	}
	runtimeInGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	actualRuntime := graphql.RuntimeExt{}

	// WHEN
	registerReq := pkg.FixRegisterRuntimeRequest(runtimeInGQL)
	saveExampleInCustomDir(t, registerReq.Query(), RegisterRuntimeCategory, "register runtime")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)
	assertRuntime(t, givenInput, actualRuntime)

	// add Label
	actualLabel := graphql.Label{}

	// WHEN
	addLabelReq := pkg.FixSetRuntimeLabelRequest(actualRuntime.ID, "new_label", []string{"bbb"})
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, addLabelReq, &actualLabel)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, "new_label", actualLabel.Key)
	assert.Len(t, actualLabel.Value, 1)
	assert.Contains(t, actualLabel.Value, "bbb")

	// get runtime and validate runtimes
	getRuntimeReq := pkg.FixGetRuntimeRequest(actualRuntime.ID)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, getRuntimeReq, &actualRuntime)
	require.NoError(t, err)
	assert.Len(t, actualRuntime.Labels, 4)

	// add agent auth
	// GIVEN
	in := pkg.FixSampleApplicationRegisterInputWithWebhooks("app")

	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	createAppReq := pkg.FixRegisterApplicationRequest(appInputGQL)

	//WHEN
	actualApp := graphql.ApplicationExt{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, createAppReq, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, actualApp.ID)

	// update runtime, check if only simple values are updated
	//GIVEN
	givenInput.Name = "updated-name"
	givenInput.Description = ptr.String("updated-description")
	givenInput.Labels = &graphql.Labels{
		"key": []interface{}{"values", "aabbcc"},
	}
	runtimeStatusCond := graphql.RuntimeStatusConditionConnected
	givenInput.StatusCondition = &runtimeStatusCond

	runtimeInGQL, err = pkg.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	updateRuntimeReq := pkg.FixUpdateRuntimeRequest(actualRuntime.ID, runtimeInGQL)
	saveExample(t, updateRuntimeReq.Query(), "update runtime")
	//WHEN
	actualRuntime = graphql.RuntimeExt{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, updateRuntimeReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, givenInput.Name, actualRuntime.Name)
	assert.Equal(t, *givenInput.Description, *actualRuntime.Description)
	assert.Equal(t, len(actualRuntime.Labels), 2)
	assert.Equal(t, runtimeStatusCond, actualRuntime.Status.Condition)

	// delete runtime

	// WHEN
	delReq := pkg.FixUnregisterRuntimeRequest(actualRuntime.ID)
	saveExample(t, delReq.Query(), "unregister runtime")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, delReq, nil)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	givenInput := graphql.RuntimeInput{
		Name:        "runtime-with-scenario-assignments",
		Description: ptr.String("runtime-1-description"),
		Labels:      &graphql.Labels{labelKey: []interface{}{labelValue}},
	}
	runtimeInGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	actualRuntime := graphql.RuntimeExt{}

	// WHEN
	registerReq := pkg.FixRegisterRuntimeRequest(runtimeInGQL)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &actualRuntime)

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
	marshalledSchema := pkg.MarshalJSONSchema(t, schema)

	givenLabelDef := graphql.LabelDefinitionInput{
		Key:    ScenariosLabel,
		Schema: marshalledSchema,
	}
	labelDefInGQL, err := pkg.Tc.Graphqlizer.LabelDefinitionInputToGQL(givenLabelDef)
	require.NoError(t, err)
	actualLabelDef := graphql.LabelDefinition{}

	// WHEN
	updateLabelDefReq := pkg.FixUpdateLabelDefinitionRequest(labelDefInGQL)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, updateLabelDefReq, &actualLabelDef)

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

	scenarioAssignmentInGQL, err := pkg.Tc.Graphqlizer.AutomaticScenarioAssignmentSetInputToGQL(givenScenarioAssignment)
	require.NoError(t, err)
	actualScenarioAssignment := graphql.AutomaticScenarioAssignment{}

	// WHEN
	createAutomaticScenarioAssignmentReq := pkg.FixCreateAutomaticScenarioAssignmentRequest(scenarioAssignmentInGQL)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, createAutomaticScenarioAssignmentReq, &actualScenarioAssignment)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, givenScenarioAssignment.ScenarioName, actualScenarioAssignment.ScenarioName)
	assert.Equal(t, givenScenarioAssignment.Selector.Key, actualScenarioAssignment.Selector.Key)
	assert.Equal(t, givenScenarioAssignment.Selector.Value, actualScenarioAssignment.Selector.Value)

	// get runtime - verify it is in scenario
	getRuntimeReq := pkg.FixGetRuntimeRequest(actualRuntime.ID)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, getRuntimeReq, &actualRuntime)

	require.NoError(t, err)
	scenarios, hasScenarios := actualRuntime.Labels["scenarios"]
	assert.True(t, hasScenarios)
	assert.Len(t, scenarios, 2)
	assert.Contains(t, scenarios, testScenario)

	// delete runtime

	// WHEN
	delReq := pkg.FixUnregisterRuntimeRequest(actualRuntime.ID)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, delReq, nil)

	//THEN
	require.NoError(t, err)

	// get automatic scenario assignment - see that it's deleted
	actualScenarioAssignments := graphql.AutomaticScenarioAssignmentPage{}
	getScenarioAssignmentsReq := pkg.FixAutomaticScenarioAssignmentsRequest()
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, getScenarioAssignmentsReq, &actualScenarioAssignments)
	require.NoError(t, err)
	assert.Equal(t, actualScenarioAssignments.TotalCount, 0)
}

func TestRuntimeCreateUpdateDuplicatedNames(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	firstRuntimeName := "unique-name-1"
	givenInput := graphql.RuntimeInput{
		Name:        firstRuntimeName,
		Description: ptr.String("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []interface{}{"hhh"}},
	}
	runtimeInGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	firstRuntime := graphql.RuntimeExt{}
	registerReq := pkg.FixRegisterRuntimeRequest(runtimeInGQL)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &firstRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, firstRuntime.ID)
	assertRuntime(t, givenInput, firstRuntime)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, firstRuntime.ID)

	// try to create second runtime with first runtime name
	//GIVEN
	givenInput = graphql.RuntimeInput{
		Name:        firstRuntimeName,
		Description: ptr.String("runtime-1-description"),
	}
	runtimeInGQL, err = pkg.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	registerReq = pkg.FixRegisterRuntimeRequest(runtimeInGQL)
	saveExampleInCustomDir(t, registerReq.Query(), RegisterRuntimeCategory, "register runtime")

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, nil)

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
	runtimeInGQL, err = pkg.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	secondRuntime := graphql.RuntimeExt{}
	registerReq = pkg.FixRegisterRuntimeRequest(runtimeInGQL)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, secondRuntime.ID)
	assertRuntime(t, givenInput, secondRuntime)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, secondRuntime.ID)

	//Update first runtime with second runtime name, failed

	//GIVEN
	givenInput = graphql.RuntimeInput{
		Name:        secondRuntimeName,
		Description: ptr.String("runtime-1-description"),
	}
	runtimeInGQL, err = pkg.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	registerReq = pkg.FixUpdateRuntimeRequest(firstRuntime.ID, runtimeInGQL)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &secondRuntime)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not unique")
}

func TestQueryRuntimes(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	idsToRemove := make([]string, 0)
	defer func() {
		for _, id := range idsToRemove {
			if id != "" {
				pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, id)
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
		runtimeInGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
		require.NoError(t, err)
		createReq := pkg.FixRegisterRuntimeRequest(runtimeInGQL)
		actualRuntime := graphql.Runtime{}
		err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, createReq, &actualRuntime)
		require.NoError(t, err)
		require.NotEmpty(t, actualRuntime.ID)
		rtm.ID = actualRuntime.ID
		idsToRemove = append(idsToRemove, actualRuntime.ID)
	}
	actualPage := graphql.RuntimePage{}

	// WHEN
	queryReq := pkg.FixGetRuntimesRequestWithPagination()
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, queryReq, &actualPage)
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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	givenInput := graphql.RuntimeInput{
		Name: "runtime-specific-runtime",
	}
	runtimeInGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	registerReq := pkg.FixRegisterRuntimeRequest(runtimeInGQL)
	createdRuntime := graphql.Runtime{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &createdRuntime)
	require.NoError(t, err)
	require.NotEmpty(t, createdRuntime.ID)

	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, createdRuntime.ID)
	queriedRuntime := graphql.Runtime{}

	// WHEN
	queryReq := pkg.FixGetRuntimeRequest(createdRuntime.ID)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, queryReq, &queriedRuntime)
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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	runtimes := make(map[string]*graphql.Runtime)
	runtimesAmount := 10
	for i := 0; i < runtimesAmount; i++ {
		runtimeInput := graphql.RuntimeInput{
			Name: fmt.Sprintf("runtime-%d", i),
		}
		runtimeInputGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(runtimeInput)
		require.NoError(t, err)

		registerReq := pkg.FixRegisterRuntimeRequest(runtimeInputGQL)

		runtime := graphql.Runtime{}
		err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &runtime)

		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)
		defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)
		runtimes[runtime.ID] = &runtime
	}

	after := 3
	cursor := ""
	queriesForFullPage := int(runtimesAmount / after)

	for i := 0; i < queriesForFullPage; i++ {
		runtimesRequest := pkg.FixRuntimeRequestWithPaginationRequest(after, cursor)

		//WHEN
		runtimePage := graphql.RuntimePage{}
		err := pkg.Tc.RunOperation(ctx, dexGraphQLClient, runtimesRequest, &runtimePage)
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
	runtimesRequest := pkg.FixRuntimeRequestWithPaginationRequest(after, cursor)
	lastRuntimePage := graphql.RuntimePage{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, runtimesRequest, &lastRuntimePage)
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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "test-create-runtime-without-labels"
	runtimeInput := graphql.RuntimeInput{Name: name}

	runtime := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &runtimeInput)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)

	//WHEN
	fetchedRuntime := pkg.GetRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertRuntime(t, runtimeInput, fetchedRuntime)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := graphql.RuntimeInput{
		Name:        name,
		Description: ptr.String("runtime-1-description"),
		Labels:      &graphql.Labels{ScenariosLabel: []interface{}{"DEFAULT"}},
	}
	runtimeInGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(secondInput)
	require.NoError(t, err)
	updateReq := pkg.FixUpdateRuntimeRequest(fetchedRuntime.ID, runtimeInGQL)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, updateReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	assertRuntime(t, secondInput, secondRuntime)
}

func TestRegisterUpdateRuntimeWithIsNormalizedLabel(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "test-create-runtime-without-labels"
	runtimeInput := graphql.RuntimeInput{
		Name:   name,
		Labels: &graphql.Labels{IsNormalizedLabel: "false"},
	}

	runtime := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &runtimeInput)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)

	//WHEN
	fetchedRuntime := pkg.GetRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertRuntime(t, runtimeInput, fetchedRuntime)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := graphql.RuntimeInput{
		Name:        name,
		Description: ptr.String("runtime-1-description"),
		Labels:      &graphql.Labels{IsNormalizedLabel: "true", ScenariosLabel: []interface{}{"DEFAULT"}},
	}
	runtimeInGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(secondInput)
	require.NoError(t, err)
	updateReq := pkg.FixUpdateRuntimeRequest(fetchedRuntime.ID, runtimeInGQL)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, updateReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	assertRuntime(t, secondInput, secondRuntime)
}
