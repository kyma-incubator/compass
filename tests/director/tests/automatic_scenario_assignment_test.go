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
	tenantID := tenant.TestTenants.GetDefaultTenantID()
	subaccount := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	testScenarioA := "ASA1"
	testScenarioB := "ASA2"
	testSelectorA := graphql.LabelSelectorInput{
		Key:   "global_subaccount_id",
		Value: subaccount,
	}
	testSelectorAGQL, err := testctx.Tc.Graphqlizer.LabelSelectorInputToGQL(testSelectorA)
	require.NoError(t, err)

	// setup available scenarios
	fixtures.UpsertScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{"DEFAULT", testScenarioA, testScenarioB})
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{"DEFAULT"})

	// create automatic scenario assignments
	inputAssignment1 := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: testScenarioA,
		Selector:     &testSelectorA,
	}
	inputAssignment2 := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: testScenarioB,
		Selector:     &testSelectorA,
	}
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, inputAssignment1, tenantID)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, testScenarioA)
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, inputAssignment2, tenantID)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, testScenarioB)

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
		[]graphql.AutomaticScenarioAssignmentSetInput{inputAssignment1, inputAssignment2},
		actualAssignmentsPage.Data)
	assertions.AssertAutomaticScenarioAssignment(t, inputAssignment1, actualAssignmentForScenario)
	assertions.AssertAutomaticScenarioAssignments(t,
		[]graphql.AutomaticScenarioAssignmentSetInput{inputAssignment1, inputAssignment2},
		actualAssignmentsForSelector)
}

func Test_AutomaticScenarioAssigmentForRuntime(t *testing.T) {
	//GIVEN
	ctx := context.TODO()

	tenantID := tenant.TestTenants.GetDefaultTenantID()
	subaccount := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	prodScenario := "PRODUCTION"
	devScenario := "DEVELOPMENT"
	manualScenario := "MANUAL"
	defaultScenario := "DEFAULT"

	fixtures.UpsertScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{prodScenario, manualScenario, devScenario, defaultScenario})
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{"DEFAULT"})

	rtms := make([]*graphql.RuntimeExt, 3)
	for i := 0; i < 2; i++ {
		rmtInput := fixtures.FixRuntimeInput(fmt.Sprintf("runtime%d", i))

		rtm, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, subaccount, &rmtInput)
		rtms[i] = &rtm
		defer fixtures.CleanupRuntime(t, ctx, dexGraphQLClient, subaccount, &rtm)
		require.NoError(t, err)
		require.NotEmpty(t, rtm.ID)
	}

	rmtInput := fixtures.FixRuntimeInput(fmt.Sprintf("runtime%d", 2))

	rtm, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantID, &rmtInput)
	rtms[2] = &rtm
	defer fixtures.CleanupRuntime(t, ctx, dexGraphQLClient, tenantID, &rtm)
	require.NoError(t, err)
	require.NotEmpty(t, rtm.ID)

	selectorKey := "global_subaccount_id"
	selectorValue := subaccount

	t.Run("Check automatic scenario assigment", func(t *testing.T) {
		//GIVEN
		expectedScenarios := map[string][]interface{}{
			rtms[0].ID: {prodScenario},
			rtms[1].ID: {prodScenario},
			rtms[2].ID: {defaultScenario},
		}
		if !conf.DefaultScenarioEnabled {
			expectedScenarios[rtms[2].ID] = []interface{}{}
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
			rtms[2].ID: {defaultScenario},
		}
		if !conf.DefaultScenarioEnabled {
			scenarios[rtms[2].ID] = []interface{}{}
		}

		//WHEN
		asaInput := fixtures.FixAutomaticScenarioAssigmentInput(prodScenario, selectorKey, selectorValue)
		fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, tenantID)
		runtimes := fixtures.ListRuntimes(t, ctx, dexGraphQLClient, tenantID)
		assertions.AssertRuntimeScenarios(t, runtimes, scenarios)

		expectedScenarios := map[string][]interface{}{
			rtms[0].ID: {},
			rtms[1].ID: {},
			rtms[2].ID: {defaultScenario},
		}
		if !conf.DefaultScenarioEnabled {
			expectedScenarios[rtms[2].ID] = []interface{}{}
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
			rtms[2].ID: {defaultScenario},
		}
		if !conf.DefaultScenarioEnabled {
			scenarios[rtms[2].ID] = []interface{}{}
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
			rtms[2].ID: {defaultScenario},
		}
		if !conf.DefaultScenarioEnabled {
			expectedScenarios[rtms[2].ID] = []interface{}{}
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

	tenantID := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	defaultValue := "DEFAULT"
	scenario1 := "test-scenario"
	scenario2 := "test-scenario-2"
	selector := &graphql.LabelSelectorInput{
		Value: subaccountID,
		Key:   "global_subaccount_id",
	}

	scenarios := []string{defaultValue, scenario1, scenario2}

	fixtures.UpsertScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{"DEFAULT"})

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

	tenantID := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	defaultValue := "DEFAULT"
	scenario1 := "test-scenario"
	scenario2 := "test-scenario-2"

	scenarios := []string{defaultValue, scenario1, scenario2}

	fixtures.UpsertScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{"DEFAULT"})

	selector := &graphql.LabelSelectorInput{
		Value: subaccountID,
		Key:   "global_subaccount_id",
	}

	assignments := []graphql.AutomaticScenarioAssignmentSetInput{
		{ScenarioName: scenario1, Selector: selector},
		{ScenarioName: scenario2, Selector: selector},
	}

	var output []*graphql.AutomaticScenarioAssignment

	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, assignments[0], tenantID)
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, assignments[1], tenantID)

	selectorGql, err := testctx.Tc.Graphqlizer.LabelSelectorInputToGQL(*selector)
	require.NoError(t, err)

	//WHEN
	req := fixtures.FixDeleteAutomaticScenarioAssignmentsForSelectorRequest(selectorGql)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, req, &output)
	require.NoError(t, err)

	//THEN
	assertions.AssertAutomaticScenarioAssignments(t, assignments, output)

	actualAssignments := fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, dexGraphQLClient, tenantID)
	assert.Len(t, actualAssignments.Data, 0)
	require.Equal(t, 0, actualAssignments.TotalCount)

	saveExample(t, req.Query(), "delete automatic scenario assignments for selector")
}

func TestAutomaticScenarioAssignmentsWholeScenario(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	scenario := "test"

	scenarios := []interface{}{scenario}

	tenantID := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	fixtures.UpsertScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{scenario, "DEFAULT"})
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantID, []string{"DEFAULT"})

	selector := graphql.LabelSelectorInput{Key: "global_subaccount_id", Value: subaccountID}
	assignment := graphql.AutomaticScenarioAssignmentSetInput{ScenarioName: scenario, Selector: &selector}

	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, assignment, tenantID)
	defer fixtures.CleanUpAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenario)

	rtmInput := graphql.RuntimeInput{
		Name: "test-name",
	}

	rtm, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, subaccountID, &rtmInput)
	defer fixtures.CleanupRuntime(t, ctx, dexGraphQLClient, subaccountID, &rtm)
	require.NoError(t, err)
	require.NotEmpty(t, rtm.ID)

	t.Run("Scenario is set when label matches selector", func(t *testing.T) {
		rtmWithScenarios := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantID, rtm.ID)
		assertions.AssertScenarios(t, rtmWithScenarios.Labels, scenarios)
	})

	t.Run("Scenario is unset when automatic scenario assignment is deleted", func(t *testing.T) {
		fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantID, scenario)
		rtmWithoutScenarios := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantID, rtm.ID)
		assertions.AssertScenarios(t, rtmWithoutScenarios.Labels, []interface{}{})
	})
}
