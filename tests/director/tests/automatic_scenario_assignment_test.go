package tests

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func Test_AutomaticScenarioAssignmentQueries(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	tenantID := pkg.TestTenants.GetIDByName(t, "ASA1")

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	testScenarioA := "ASA1"
	testScenarioB := "ASA2"
	testScenarioC := "ASA3"
	testSelectorA := graphql.LabelSelectorInput{
		Key:   "keyA",
		Value: "valueA",
	}
	testSelectorB := graphql.LabelSelectorInput{
		Key:   "keyB",
		Value: "valueB",
	}
	testSelectorAGQL, err := pkg.Tc.Graphqlizer.LabelSelectorInputToGQL(testSelectorA)
	require.NoError(t, err)

	// setup available scenarios
	pkg.CreateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{"DEFAULT", testScenarioA, testScenarioB, testScenarioC})

	// create automatic scenario assignments
	inputAssignment1 := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: testScenarioA,
		Selector:     &testSelectorA,
	}
	inputAssignment2 := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: testScenarioB,
		Selector:     &testSelectorA,
	}
	inputAssignment3 := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: testScenarioC,
		Selector:     &testSelectorB,
	}
	pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, inputAssignment1, tenantID)
	defer pkg.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, testScenarioA)
	pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, inputAssignment2, tenantID)
	defer pkg.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, testScenarioB)
	pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, inputAssignment3, tenantID)
	defer pkg.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, testScenarioC)

	// prepare queries
	getAssignmentForScenarioRequest := pkg.FixAutomaticScenarioAssignmentForScenarioRequest(testScenarioA)
	listAssignmentsRequest := pkg.FixAutomaticScenarioAssignmentsRequest()
	listAssignmentsForSelectorRequest := pkg.FixAutomaticScenarioAssignmentsForSelectorRequest(testSelectorAGQL)

	actualAssignmentsPage := graphql.AutomaticScenarioAssignmentPage{}
	actualAssignmentForScenario := graphql.AutomaticScenarioAssignment{}
	actualAssignmentsForSelector := []*graphql.AutomaticScenarioAssignment{}

	// WHEN
	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, listAssignmentsRequest, &actualAssignmentsPage)
	require.NoError(t, err)
	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, getAssignmentForScenarioRequest, &actualAssignmentForScenario)
	require.NoError(t, err)
	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, listAssignmentsForSelectorRequest, &actualAssignmentsForSelector)
	require.NoError(t, err)

	// THEN
	saveExample(t, listAssignmentsRequest.Query(), "query automatic scenario assignments")
	saveExample(t, getAssignmentForScenarioRequest.Query(), "query automatic scenario assignment for scenario")
	saveExample(t, listAssignmentsForSelectorRequest.Query(), "query automatic scenario assignments for selector")

	assertAutomaticScenarioAssignments(t,
		[]graphql.AutomaticScenarioAssignmentSetInput{inputAssignment1, inputAssignment2, inputAssignment3},
		actualAssignmentsPage.Data)
	assertAutomaticScenarioAssignment(t, inputAssignment1, actualAssignmentForScenario)
	assertAutomaticScenarioAssignments(t,
		[]graphql.AutomaticScenarioAssignmentSetInput{inputAssignment1, inputAssignment2},
		actualAssignmentsForSelector)
}

func Test_AutomaticScenarioAssigmentForRuntime(t *testing.T) {
	//GIVEN
	ctx := context.TODO()

	tenantID := pkg.TestTenants.GetIDByName(t, "TestCreateAutomaticScenarioAssignment")

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	prodScenario := "PRODUCTION"
	devScenario := "DEVELOPMENT"
	manualScenario := "MANUAL"
	defaultScenario := "DEFAULT"
	pkg.CreateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{prodScenario, manualScenario, devScenario, defaultScenario})

	rtms := make([]*graphql.RuntimeExt, 3)
	for i := 0; i < 3; i++ {
		rmtInput := pkg.FixRuntimeInput(fmt.Sprintf("runtime%d", i))

		rtm := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantID, &rmtInput)
		rtms[i] = &rtm
		defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantID, rtm.ID)
	}

	selectorKey := "KEY"
	selectorValue := "VALUE"

	pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtms[0].ID, selectorKey, selectorValue)
	pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtms[1].ID, selectorKey, selectorValue)

	t.Run("Check automatic scenario assigment", func(t *testing.T) {
		//GIVEN
		expectedScenarios := map[string][]interface{}{
			rtms[0].ID: {defaultScenario, prodScenario},
			rtms[1].ID: {defaultScenario, prodScenario},
			rtms[2].ID: {defaultScenario},
		}

		//WHEN
		asaInput := pkg.FixAutomaticScenarioAssigmentInput(prodScenario, selectorKey, selectorValue)
		pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, tenantID)
		defer pkg.DeleteAutomaticScenarioAssigmentForSelector(t, ctx, dexGraphQLClient, tenantID, *asaInput.Selector)

		//THEN
		runtimes := pkg.ListRuntimes(t, ctx, dexGraphQLClient, tenantID)
		require.Len(t, runtimes.Data, 3)
		assertRuntimeScenarios(t, runtimes, expectedScenarios)
	})

	t.Run("Delete Automatic Scenario Assigment for scenario", func(t *testing.T) {
		//GIVEN
		scenarios := map[string][]interface{}{
			rtms[0].ID: {defaultScenario, prodScenario},
			rtms[1].ID: {defaultScenario, prodScenario},
			rtms[2].ID: {defaultScenario},
		}

		//WHEN
		asaInput := pkg.FixAutomaticScenarioAssigmentInput(prodScenario, selectorKey, selectorValue)
		pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, tenantID)
		runtimes := pkg.ListRuntimes(t, ctx, dexGraphQLClient, tenantID)
		assertRuntimeScenarios(t, runtimes, scenarios)

		expectedScenarios := map[string][]interface{}{
			rtms[0].ID: {defaultScenario},
			rtms[1].ID: {defaultScenario},
			rtms[2].ID: {defaultScenario},
		}

		//WHEN
		pkg.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, prodScenario)

		//THEN
		runtimes = pkg.ListRuntimes(t, ctx, dexGraphQLClient, tenantID)
		require.Len(t, runtimes.Data, 3)
		assertRuntimeScenarios(t, runtimes, expectedScenarios)
	})

	t.Run("Delete Automatic Scenario Assigment by selector, check also if manually added scenarios survived", func(t *testing.T) {
		//GIVEN
		scenarios := map[string][]interface{}{
			rtms[0].ID: {defaultScenario, prodScenario, devScenario},
			rtms[1].ID: {defaultScenario, prodScenario, devScenario},
			rtms[2].ID: {defaultScenario},
		}

		//WHEN
		asaInput := pkg.FixAutomaticScenarioAssigmentInput(prodScenario, selectorKey, selectorValue)
		pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, tenantID)
		asaInput.ScenarioName = devScenario
		pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, tenantID)
		runtimes := pkg.ListRuntimes(t, ctx, dexGraphQLClient, tenantID)
		assertRuntimeScenarios(t, runtimes, scenarios)

		expectedScenarios := map[string][]interface{}{
			rtms[0].ID: {manualScenario},
			rtms[1].ID: {manualScenario},
			rtms[2].ID: {defaultScenario},
		}

		//WHEN
		pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtms[0].ID, "scenarios", []interface{}{prodScenario, devScenario, manualScenario})
		pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtms[1].ID, "scenarios", []interface{}{prodScenario, devScenario, manualScenario})
		pkg.DeleteAutomaticScenarioAssigmentForSelector(t, ctx, dexGraphQLClient, tenantID, graphql.LabelSelectorInput{Key: selectorKey, Value: selectorValue})

		//THEN
		runtimes = pkg.ListRuntimes(t, ctx, dexGraphQLClient, tenantID)
		require.Len(t, runtimes.Data, 3)
		assertRuntimeScenarios(t, runtimes, expectedScenarios)
	})

}

func Test_DeleteAutomaticScenarioAssignmentForScenario(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	defaultValue := "DEFAULT"
	scenario1 := "test-scenario"
	scenario2 := "test-scenario-2"
	selector := &graphql.LabelSelectorInput{
		Value: "test-value",
		Key:   "test-key",
	}

	scenarios := []string{defaultValue, scenario1, scenario2}
	tenantID := pkg.TestTenants.GetIDByName(t, "TestDeleteAssignmentsForScenario")
	pkg.CreateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenarios)

	assignment1 := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: scenario1,
		Selector:     selector,
	}
	assignment2 := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: scenario2,
		Selector:     selector,
	}

	var output graphql.AutomaticScenarioAssignment

	assignment1Gql, err := pkg.Tc.Graphqlizer.AutomaticScenarioAssignmentSetInputToGQL(assignment1)
	require.NoError(t, err)

	req := pkg.FixCreateAutomaticScenarioAssignmentRequest(assignment1Gql)
	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, req, nil)
	require.NoError(t, err)
	saveExample(t, req.Query(), "create automatic scenario assignment")

	pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, assignment2, tenantID)
	defer pkg.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenario2)

	//WHEN
	req = pkg.FixDeleteAutomaticScenarioAssignmentForScenarioRequest(scenario1)
	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, req, &output)
	require.NoError(t, err)

	//THEN
	assertAutomaticScenarioAssignment(t, assignment1, output)

	allAssignments := pkg.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, dexGraphQLClient, tenantID)
	require.Len(t, allAssignments.Data, 1)
	require.Equal(t, 1, allAssignments.TotalCount)
	assertAutomaticScenarioAssignment(t, assignment2, *allAssignments.Data[0])

	saveExample(t, req.Query(), "delete automatic scenario assignment for scenario")
}

func Test_DeleteAutomaticScenarioAssignmentForSelector(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	defaultValue := "DEFAULT"
	scenario1 := "test-scenario"
	scenario2 := "test-scenario-2"
	scenario3 := "test-scenario-3"

	scenarios := []string{defaultValue, scenario1, scenario2, scenario3}

	tenantID := pkg.TestTenants.GetIDByName(t, "TestDeleteAssignmentsForSelector")
	pkg.CreateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenarios)

	selector := graphql.LabelSelectorInput{Key: "test-key", Value: "test-value"}
	selector2 := graphql.LabelSelectorInput{
		Key:   "test-key-2",
		Value: "test-value-2",
	}

	assignments := []graphql.AutomaticScenarioAssignmentSetInput{
		{ScenarioName: scenario1, Selector: &selector},
		{ScenarioName: scenario2, Selector: &selector},
	}
	anotherAssignment := graphql.AutomaticScenarioAssignmentSetInput{ScenarioName: scenario3, Selector: &selector2}

	var output []*graphql.AutomaticScenarioAssignment

	pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, assignments[0], tenantID)
	pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, assignments[1], tenantID)
	pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, anotherAssignment, tenantID)
	defer pkg.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenario3)

	selectorGql, err := pkg.Tc.Graphqlizer.LabelSelectorInputToGQL(selector)
	require.NoError(t, err)

	//WHEN
	req := pkg.FixDeleteAutomaticScenarioAssignmentsForSelectorRequest(selectorGql)
	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, req, &output)
	require.NoError(t, err)

	//THEN
	assertAutomaticScenarioAssignments(t, assignments, output)

	actualAssignments := pkg.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, dexGraphQLClient, tenantID)
	assert.Len(t, actualAssignments.Data, 1)
	require.Equal(t, 1, actualAssignments.TotalCount)
	assertAutomaticScenarioAssignment(t, anotherAssignment, *actualAssignments.Data[0])

	saveExample(t, req.Query(), "delete automatic scenario assignments for selector")

}

func TestAutomaticScenarioAssignmentsWholeScenario(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	defaultValue := "DEFAULT"
	scenario := "test"

	scenariosOnlyDefault := []interface{}{defaultValue}
	scenarios := []interface{}{scenario, defaultValue}
	tenantID := pkg.TestTenants.GetIDByName(t, "TestWholeScenario")
	pkg.CreateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{scenarios[0].(string), scenarios[1].(string)})

	selector := graphql.LabelSelectorInput{Key: "testkey", Value: "testvalue"}
	assignment := graphql.AutomaticScenarioAssignmentSetInput{ScenarioName: scenario, Selector: &selector}

	pkg.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, assignment, tenantID)

	rtmInput := graphql.RuntimeInput{
		Name:   "test-name",
		Labels: &graphql.Labels{selector.Key: selector.Value, "scenarios": []string{defaultValue}},
	}

	rtm := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantID, &rtmInput)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, rtm.ID, tenantID)

	t.Run("Scenario is set when label matches selector", func(t *testing.T) {
		rtmWithScenarios := pkg.GetRuntime(t, ctx, dexGraphQLClient, rtm.ID, tenantID)
		assertScenarios(t, rtmWithScenarios.Labels, scenarios)
	})

	selector2 := graphql.LabelSelectorInput{Key: "newtestkey", Value: "newtestvalue"}

	t.Run("Scenario is unset when label on runtime changes", func(t *testing.T) {
		pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtm.ID, selector.Key, selector2.Value)
		rtmWithScenarios := pkg.GetRuntime(t, ctx, dexGraphQLClient, rtm.ID, tenantID)
		assertScenarios(t, rtmWithScenarios.Labels, scenariosOnlyDefault)
	})

	t.Run("Scenario is set back when label on runtime matches selector", func(t *testing.T) {
		pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtm.ID, selector.Key, selector.Value)
		rtmWithScenarios := pkg.GetRuntime(t, ctx, dexGraphQLClient, rtm.ID, tenantID)
		assertScenarios(t, rtmWithScenarios.Labels, scenarios)
	})

	t.Run("Scenario is unset when automatic scenario assignment is deleted", func(t *testing.T) {
		pkg.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenario)
		rtmWithoutScenarios := pkg.GetRuntime(t, ctx, dexGraphQLClient, rtm.ID, tenantID)
		assertScenarios(t, rtmWithoutScenarios.Labels, scenariosOnlyDefault)
	})

}
