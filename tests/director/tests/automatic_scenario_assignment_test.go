package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func Test_AutomaticScenarioAssignmentQueries(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetIDByName(t, tenant.AutomaticScenarioAssignmentQueriesTenantName)

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
	testSelectorAGQL, err := testctx.Tc.Graphqlizer.LabelSelectorInputToGQL(testSelectorA)
	require.NoError(t, err)

	// setup available scenarios
	fixtures.CreateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{"DEFAULT", testScenarioA, testScenarioB, testScenarioC})

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
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, inputAssignment1, tenantID)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, testScenarioA)
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, inputAssignment2, tenantID)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, testScenarioB)
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, inputAssignment3, tenantID)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, testScenarioC)

	// prepare queries
	getAssignmentForScenarioRequest := fixtures.FixAutomaticScenarioAssignmentForScenarioRequest(testScenarioA)
	listAssignmentsRequest := fixtures.FixAutomaticScenarioAssignmentsRequest()
	listAssignmentsForSelectorRequest := fixtures.FixAutomaticScenarioAssignmentsForSelectorRequest(testSelectorAGQL)

	actualAssignmentsPage := graphql.AutomaticScenarioAssignmentPage{}
	actualAssignmentForScenario := graphql.AutomaticScenarioAssignment{}
	actualAssignmentsForSelector := []*graphql.AutomaticScenarioAssignment{}

	// WHEN
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, listAssignmentsRequest, &actualAssignmentsPage)
	require.NoError(t, err)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, getAssignmentForScenarioRequest, &actualAssignmentForScenario)
	require.NoError(t, err)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, listAssignmentsForSelectorRequest, &actualAssignmentsForSelector)
	require.NoError(t, err)

	// THEN
	saveExample(t, listAssignmentsRequest.Query(), "query automatic scenario assignments")
	saveExample(t, getAssignmentForScenarioRequest.Query(), "query automatic scenario assignment for scenario")
	saveExample(t, listAssignmentsForSelectorRequest.Query(), "query automatic scenario assignments for selector")

	assertions.AssertAutomaticScenarioAssignments(t,
		[]graphql.AutomaticScenarioAssignmentSetInput{inputAssignment1, inputAssignment2, inputAssignment3},
		actualAssignmentsPage.Data)
	assertions.AssertAutomaticScenarioAssignment(t, inputAssignment1, actualAssignmentForScenario)
	assertions.AssertAutomaticScenarioAssignments(t,
		[]graphql.AutomaticScenarioAssignmentSetInput{inputAssignment1, inputAssignment2},
		actualAssignmentsForSelector)
}

func Test_AutomaticScenarioAssigmentForRuntime(t *testing.T) {
	//GIVEN
	ctx := context.TODO()

	tenantID := tenant.TestTenants.GetIDByName(t, tenant.AutomaticScenarioAssigmentForRuntimeTenantName)

	prodScenario := "PRODUCTION"
	devScenario := "DEVELOPMENT"
	manualScenario := "MANUAL"
	defaultScenario := "DEFAULT"
	fixtures.CreateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{prodScenario, manualScenario, devScenario, defaultScenario})

	rtms := make([]*graphql.RuntimeExt, 3)
	for i := 0; i < 3; i++ {
		rmtInput := fixtures.FixRuntimeInput(fmt.Sprintf("runtime%d", i))

		rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantID, &rmtInput)
		rtms[i] = &rtm
		defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantID, rtm.ID)
	}

	selectorKey := "KEY"
	selectorValue := "VALUE"

	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtms[0].ID, selectorKey, selectorValue)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtms[1].ID, selectorKey, selectorValue)

	t.Run("Check automatic scenario assigment", func(t *testing.T) {
		//GIVEN
		expectedScenarios := map[string][]interface{}{
			rtms[0].ID: {prodScenario},
			rtms[1].ID: {prodScenario},
			rtms[2].ID: {},
		}
		if conf.DefaultScenarioEnabled {
			expectedScenarios = map[string][]interface{}{
				rtms[0].ID: {defaultScenario, prodScenario},
				rtms[1].ID: {defaultScenario, prodScenario},
				rtms[2].ID: {defaultScenario},
			}
		}

		//WHEN
		asaInput := fixtures.FixAutomaticScenarioAssigmentInput(prodScenario, selectorKey, selectorValue)
		fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, tenantID)
		defer fixtures.DeleteAutomaticScenarioAssigmentForSelector(t, ctx, dexGraphQLClient, tenantID, *asaInput.Selector)

		//THEN
		runtimes := fixtures.ListRuntimes(t, ctx, dexGraphQLClient, tenantID)
		require.Len(t, runtimes.Data, 3)
		assertions.AssertRuntimeScenarios(t, runtimes, expectedScenarios)
	})

	t.Run("Delete Automatic Scenario Assigment for scenario", func(t *testing.T) {
		//GIVEN
		scenarios := map[string][]interface{}{
			rtms[0].ID: {prodScenario},
			rtms[1].ID: {prodScenario},
			rtms[2].ID: {},
		}
		if conf.DefaultScenarioEnabled {
			scenarios = map[string][]interface{}{
				rtms[0].ID: {defaultScenario, prodScenario},
				rtms[1].ID: {defaultScenario, prodScenario},
				rtms[2].ID: {defaultScenario},
			}
		}

		//WHEN
		asaInput := fixtures.FixAutomaticScenarioAssigmentInput(prodScenario, selectorKey, selectorValue)
		fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, tenantID)
		runtimes := fixtures.ListRuntimes(t, ctx, dexGraphQLClient, tenantID)
		assertions.AssertRuntimeScenarios(t, runtimes, scenarios)

		expectedScenarios := map[string][]interface{}{
			rtms[0].ID: {},
			rtms[1].ID: {},
			rtms[2].ID: {},
		}
		if conf.DefaultScenarioEnabled {
			expectedScenarios = map[string][]interface{}{
				rtms[0].ID: {defaultScenario},
				rtms[1].ID: {defaultScenario},
				rtms[2].ID: {defaultScenario},
			}
		}

		//WHEN
		fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, prodScenario)

		//THEN
		runtimes = fixtures.ListRuntimes(t, ctx, dexGraphQLClient, tenantID)
		require.Len(t, runtimes.Data, 3)
		assertions.AssertRuntimeScenarios(t, runtimes, expectedScenarios)
	})

	t.Run("Delete Automatic Scenario Assigment by selector, check also if manually added scenarios survived", func(t *testing.T) {
		//GIVEN
		scenarios := map[string][]interface{}{
			rtms[0].ID: {prodScenario, devScenario},
			rtms[1].ID: {prodScenario, devScenario},
			rtms[2].ID: {},
		}
		if conf.DefaultScenarioEnabled {
			scenarios = map[string][]interface{}{
				rtms[0].ID: {conf.DefaultScenario, prodScenario, devScenario},
				rtms[1].ID: {conf.DefaultScenario, prodScenario, devScenario},
				rtms[2].ID: {conf.DefaultScenario},
			}
		}

		//WHEN
		asaInput := fixtures.FixAutomaticScenarioAssigmentInput(prodScenario, selectorKey, selectorValue)
		fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, tenantID)
		asaInput.ScenarioName = devScenario
		fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, tenantID)
		runtimes := fixtures.ListRuntimes(t, ctx, dexGraphQLClient, tenantID)
		assertions.AssertRuntimeScenarios(t, runtimes, scenarios)

		expectedScenarios := map[string][]interface{}{
			rtms[0].ID: {manualScenario},
			rtms[1].ID: {manualScenario},
			rtms[2].ID: {},
		}
		if conf.DefaultScenarioEnabled {
			expectedScenarios = map[string][]interface{}{
				rtms[0].ID: {manualScenario},
				rtms[1].ID: {manualScenario},
				rtms[2].ID: {defaultScenario},
			}
		}

		//WHEN
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtms[0].ID, "scenarios", []interface{}{prodScenario, devScenario, manualScenario})
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtms[1].ID, "scenarios", []interface{}{prodScenario, devScenario, manualScenario})
		fixtures.DeleteAutomaticScenarioAssigmentForSelector(t, ctx, dexGraphQLClient, tenantID, graphql.LabelSelectorInput{Key: selectorKey, Value: selectorValue})

		//THEN
		runtimes = fixtures.ListRuntimes(t, ctx, dexGraphQLClient, tenantID)
		require.Len(t, runtimes.Data, 3)
		assertions.AssertRuntimeScenarios(t, runtimes, expectedScenarios)
	})

}

func Test_DeleteAutomaticScenarioAssignmentForScenario(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	defaultValue := "DEFAULT"
	scenario1 := "test-scenario"
	scenario2 := "test-scenario-2"
	selector := &graphql.LabelSelectorInput{
		Value: "test-value",
		Key:   "test-key",
	}

	scenarios := []string{defaultValue, scenario1, scenario2}
	tenantID := tenant.TestTenants.GetIDByName(t, tenant.DeleteAutomaticScenarioAssignmentForScenarioTenantName)
	fixtures.CreateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenarios)

	assignment1 := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: scenario1,
		Selector:     selector,
	}
	assignment2 := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: scenario2,
		Selector:     selector,
	}

	var output graphql.AutomaticScenarioAssignment

	assignment1Gql, err := testctx.Tc.Graphqlizer.AutomaticScenarioAssignmentSetInputToGQL(assignment1)
	require.NoError(t, err)

	req := fixtures.FixCreateAutomaticScenarioAssignmentRequest(assignment1Gql)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, req, nil)
	require.NoError(t, err)
	saveExample(t, req.Query(), "create automatic scenario assignment")

	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, assignment2, tenantID)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenario2)

	//WHEN
	req = fixtures.FixDeleteAutomaticScenarioAssignmentForScenarioRequest(scenario1)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, req, &output)
	require.NoError(t, err)

	//THEN
	assertions.AssertAutomaticScenarioAssignment(t, assignment1, output)

	allAssignments := fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, dexGraphQLClient, tenantID)
	require.Len(t, allAssignments.Data, 1)
	require.Equal(t, 1, allAssignments.TotalCount)
	assertions.AssertAutomaticScenarioAssignment(t, assignment2, *allAssignments.Data[0])

	saveExample(t, req.Query(), "delete automatic scenario assignment for scenario")
}

func Test_DeleteAutomaticScenarioAssignmentForSelector(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	defaultValue := "DEFAULT"
	scenario1 := "test-scenario"
	scenario2 := "test-scenario-2"
	scenario3 := "test-scenario-3"

	scenarios := []string{defaultValue, scenario1, scenario2, scenario3}

	tenantID := tenant.TestTenants.GetIDByName(t, tenant.DeleteAutomaticScenarioAssignmentForSelectorTenantName)
	fixtures.CreateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenarios)

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

	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, assignments[0], tenantID)
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, assignments[1], tenantID)
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, anotherAssignment, tenantID)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenario3)

	selectorGql, err := testctx.Tc.Graphqlizer.LabelSelectorInputToGQL(selector)
	require.NoError(t, err)

	//WHEN
	req := fixtures.FixDeleteAutomaticScenarioAssignmentsForSelectorRequest(selectorGql)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, req, &output)
	require.NoError(t, err)

	//THEN
	assertions.AssertAutomaticScenarioAssignments(t, assignments, output)

	actualAssignments := fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, dexGraphQLClient, tenantID)
	assert.Len(t, actualAssignments.Data, 1)
	require.Equal(t, 1, actualAssignments.TotalCount)
	assertions.AssertAutomaticScenarioAssignment(t, anotherAssignment, *actualAssignments.Data[0])

	saveExample(t, req.Query(), "delete automatic scenario assignments for selector")

}

func TestAutomaticScenarioAssignmentsWholeScenario(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	defaultValue := "DEFAULT"
	scenario := "test"

	scenariosOnlyDefault := []interface{}{defaultValue}
	scenarios := []interface{}{scenario, defaultValue}
	tenantID := tenant.TestTenants.GetIDByName(t, tenant.AutomaticScenarioAssignmentsWholeScenarioTenantName)
	fixtures.CreateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{scenarios[0].(string), scenarios[1].(string)})

	selector := graphql.LabelSelectorInput{Key: "testkey", Value: "testvalue"}
	assignment := graphql.AutomaticScenarioAssignmentSetInput{ScenarioName: scenario, Selector: &selector}

	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, assignment, tenantID)

	rtmInput := graphql.RuntimeInput{
		Name:   "test-name",
		Labels: graphql.Labels{selector.Key: selector.Value, "scenarios": []string{defaultValue}},
	}

	rtm := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantID, &rtmInput)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantID, rtm.ID)

	t.Run("Scenario is set when label matches selector", func(t *testing.T) {
		rtmWithScenarios := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantID, rtm.ID)
		assertions.AssertScenarios(t, rtmWithScenarios.Labels, scenarios)
	})

	selector2 := graphql.LabelSelectorInput{Key: "newtestkey", Value: "newtestvalue"}

	t.Run("Scenario is unset when label on runtime changes", func(t *testing.T) {
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtm.ID, selector.Key, selector2.Value)
		rtmWithScenarios := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantID, rtm.ID)
		assertions.AssertScenarios(t, rtmWithScenarios.Labels, scenariosOnlyDefault)
	})

	t.Run("Scenario is set back when label on runtime matches selector", func(t *testing.T) {
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantID, rtm.ID, selector.Key, selector.Value)
		rtmWithScenarios := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantID, rtm.ID)
		assertions.AssertScenarios(t, rtmWithScenarios.Labels, scenarios)
	})

	t.Run("Scenario is unset when automatic scenario assignment is deleted", func(t *testing.T) {
		fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenario)
		rtmWithoutScenarios := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantID, rtm.ID)
		assertions.AssertScenarios(t, rtmWithoutScenarios.Labels, scenariosOnlyDefault)
	})

}
