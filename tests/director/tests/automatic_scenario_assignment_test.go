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

	formation1 := graphql.FormationInput{
		Name: testScenarioA,
	}

	formation2 := graphql.FormationInput{
		Name: testScenarioB,
	}

	testSelectorAGQL, err := testctx.Tc.Graphqlizer.LabelSelectorInputToGQL(testSelectorA)
	require.NoError(t, err)

	// setup available scenarios
	fixtures.UpsertScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, []string{"DEFAULT", testScenarioA, testScenarioB})
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, []string{"DEFAULT"})

	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation1, subaccount, tenantID)
	defer fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation1, subaccount, tenantID)
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation2, subaccount, tenantID)
	defer fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation2, subaccount, tenantID)

	// prepare queries
	getAssignmentForScenarioRequest := fixtures.FixAutomaticScenarioAssignmentForScenarioRequest(testScenarioA)
	listAssignmentsRequest := fixtures.FixAutomaticScenarioAssignmentsRequest()
	listAssignmentsForSelectorRequest := fixtures.FixAutomaticScenarioAssignmentsForSelectorRequest(testSelectorAGQL)

	actualAssignmentsPage := graphql.AutomaticScenarioAssignmentPage{}
	actualAssignmentForScenario := graphql.AutomaticScenarioAssignment{}
	actualAssignmentsForSelector := []*graphql.AutomaticScenarioAssignment{}

	// WHEN
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, listAssignmentsRequest, &actualAssignmentsPage)
	require.NoError(t, err)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, getAssignmentForScenarioRequest, &actualAssignmentForScenario)
	require.NoError(t, err)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, listAssignmentsForSelectorRequest, &actualAssignmentsForSelector)
	require.NoError(t, err)

	// THEN
	saveExample(t, listAssignmentsRequest.Query(), "query automatic scenario assignments")
	saveExample(t, getAssignmentForScenarioRequest.Query(), "query automatic scenario assignment for scenario")
	saveExample(t, listAssignmentsForSelectorRequest.Query(), "query automatic scenario assignments for selector")
	inputAssignment1 := graphql.AutomaticScenarioAssignmentSetInput{ScenarioName: testScenarioA, Selector: &testSelectorA}
	inputAssignment2 := graphql.AutomaticScenarioAssignmentSetInput{ScenarioName: testScenarioB, Selector: &testSelectorA}

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

	fixtures.UpsertScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, []string{prodScenario, manualScenario, devScenario, defaultScenario})
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, []string{"DEFAULT"})

	rtms := make([]*graphql.RuntimeExt, 3)
	for i := 0; i < 2; i++ {
		rmtInput := fixtures.FixRuntimeInput(fmt.Sprintf("runtime%d", i))

		rtm, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, subaccount, &rmtInput)
		rtms[i] = &rtm
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccount, &rtm)
		require.NoError(t, err)
		require.NotEmpty(t, rtm.ID)
	}

	rmtInput := fixtures.FixRuntimeInput(fmt.Sprintf("runtime%d", 2))

	rtm, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, &rmtInput)
	rtms[2] = &rtm
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &rtm)
	require.NoError(t, err)
	require.NotEmpty(t, rtm.ID)

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
		formationInput := graphql.FormationInput{Name: prodScenario}
		fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccount, tenantID)
		defer fixtures.CleanupUnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccount, tenantID)

		//THEN
		runtimes := fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, tenantID)
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
		formationInput := graphql.FormationInput{Name: prodScenario}
		fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccount, tenantID)
		runtimes := fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, tenantID)
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
		fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccount, tenantID)

		//THEN
		runtimes = fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, tenantID)
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
		formationInputProd := graphql.FormationInput{Name: prodScenario}
		formationInputDev := graphql.FormationInput{Name: devScenario}

		fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInputProd, subaccount, tenantID)
		fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInputDev, subaccount, tenantID)

		runtimes := fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, tenantID)
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
		fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantID, rtms[0].ID, "scenarios", []interface{}{prodScenario, devScenario, manualScenario})
		fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantID, rtms[1].ID, "scenarios", []interface{}{prodScenario, devScenario, manualScenario})
		fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInputProd, subaccount, tenantID)
		fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInputDev, subaccount, tenantID)

		//THEN
		runtimes = fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, tenantID)
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

	fixtures.UpsertScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, []string{defaultValue})

	assignment1 := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: scenario1,
		Selector:     selector,
	}
	assignment2 := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: scenario2,
		Selector:     selector,
	}

	formation1 := graphql.FormationInput{Name: scenario1}
	formation2 := graphql.FormationInput{Name: scenario2}

	var output graphql.Formation

	req := fixtures.FixAssignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), formation1.Name)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, req, nil)
	require.NoError(t, err)
	saveExample(t, req.Query(), "create automatic scenario assignment")

	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation2, subaccountID, tenantID)
	defer fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation2, subaccountID, tenantID)

	//WHEN
	req = fixtures.FixUnassignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), formation1.Name)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, req, &output)
	require.NoError(t, err)

	//THEN

	resultAssignment := graphql.AutomaticScenarioAssignment{ScenarioName: output.Name, Selector: &graphql.Label{
		Key:   selector.Key,
		Value: selector.Value,
	}}
	assertions.AssertAutomaticScenarioAssignment(t, assignment1, resultAssignment)

	allAssignments := fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID)
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

	fixtures.UpsertScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, []string{defaultValue})

	selector := &graphql.LabelSelectorInput{
		Value: subaccountID,
		Key:   "global_subaccount_id",
	}

	assignments := []graphql.AutomaticScenarioAssignmentSetInput{
		{ScenarioName: scenario1, Selector: selector},
		{ScenarioName: scenario2, Selector: selector},
	}
	formations := []graphql.FormationInput{
		{Name: scenario1},
		{Name: scenario2},
	}

	output := graphql.Formation{}
	var outputAssignments []*graphql.AutomaticScenarioAssignment
	labelForSelector := &graphql.Label{Key: selector.Key, Value: selector.Value}

	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formations[0], subaccountID, tenantID)
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formations[1], subaccountID, tenantID)

	//WHEN
	req := fixtures.FixUnassignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), formations[0].Name)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, req, &output)
	require.NoError(t, err)
	outputAssignments = append(outputAssignments, &graphql.AutomaticScenarioAssignment{ScenarioName: output.Name, Selector: labelForSelector})

	req = fixtures.FixUnassignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), formations[1].Name)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, req, &output)
	require.NoError(t, err)
	outputAssignments = append(outputAssignments, &graphql.AutomaticScenarioAssignment{ScenarioName: output.Name, Selector: labelForSelector})

	//THEN
	assertions.AssertAutomaticScenarioAssignments(t, assignments, outputAssignments)

	actualAssignments := fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID)
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

	fixtures.UpsertScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, []string{scenario, "DEFAULT"})
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, []string{"DEFAULT"})

	formation := graphql.FormationInput{Name: scenario}

	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation, subaccountID, tenantID)
	defer fixtures.CleanupUnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation, subaccountID, tenantID)

	rtmInput := graphql.RuntimeInput{
		Name: "test-name",
	}

	rtm, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, subaccountID, &rtmInput)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, &rtm)
	require.NoError(t, err)
	require.NotEmpty(t, rtm.ID)

	t.Run("Scenario is set when label matches selector", func(t *testing.T) {
		rtmWithScenarios := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, tenantID, rtm.ID)
		assertions.AssertScenarios(t, rtmWithScenarios.Labels, scenarios)
	})

	t.Run("Scenario is unset when automatic scenario assignment is deleted", func(t *testing.T) {
		fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: scenario}, subaccountID, tenantID)
		rtmWithoutScenarios := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, tenantID, rtm.ID)
		assertions.AssertScenarios(t, rtmWithoutScenarios.Labels, []interface{}{})
	})
}
