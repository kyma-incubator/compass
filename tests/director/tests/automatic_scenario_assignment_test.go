package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestAutomaticScenarioAssignmentQueries(t *testing.T) {
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
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenarioA)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenarioA)

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenarioB)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenarioB)

	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation1, subaccount, tenantID)
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation1.Name, subaccount, tenantID)
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation2, subaccount, tenantID)
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation2.Name, subaccount, tenantID)

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

func TestAutomaticScenarioAssignmentForRuntime(t *testing.T) {
	//GIVEN
	ctx := context.TODO()

	tenantID := tenant.TestTenants.GetDefaultTenantID()
	subaccount := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	prodScenario := "PRODUCTION"
	devScenario := "DEVELOPMENT"
	manualScenario := "MANUAL"
	scenarios := []string{prodScenario, devScenario, manualScenario}

	for _, scenario := range scenarios {
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenario)
		fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenario)
	}

	rtm0 := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccount, fixRuntimeInput("runtime0"), conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccount, &rtm0)

	rtm1 := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccount, fixRuntimeInput("runtime1"), conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccount, &rtm1)

	rtm2 := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantID, fixRuntimeInput("runtime2"), conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &rtm2)

	t.Run("Check automatic scenario assigment", func(t *testing.T) {
		//GIVEN
		expectedScenarios := map[string][]interface{}{
			rtm0.ID: {prodScenario},
			rtm1.ID: {prodScenario},
			rtm2.ID: {},
		}

		//WHEN
		formationInput := graphql.FormationInput{Name: prodScenario}
		fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccount, tenantID)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput.Name, subaccount, tenantID)

		//THEN
		runtimes := fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, tenantID)
		require.Len(t, runtimes.Data, 3)
		assertions.AssertRuntimeScenarios(t, runtimes, expectedScenarios)
	})

	t.Run("Delete Automatic Scenario Assigment for scenario", func(t *testing.T) {
		//GIVEN
		scenarios := map[string][]interface{}{
			rtm0.ID: {prodScenario},
			rtm1.ID: {prodScenario},
			rtm2.ID: {},
		}

		//WHEN
		formationInput := graphql.FormationInput{Name: prodScenario}
		fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccount, tenantID)
		runtimes := fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, tenantID)
		assertions.AssertRuntimeScenarios(t, runtimes, scenarios)

		expectedScenarios := map[string][]interface{}{
			rtm0.ID: {},
			rtm1.ID: {},
			rtm2.ID: {},
		}

		//WHEN
		fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccount, tenantID)

		//THEN
		runtimes = fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, tenantID)
		require.Len(t, runtimes.Data, 3)
		assertions.AssertRuntimeScenarios(t, runtimes, expectedScenarios)
	})
}

func TestAutomaticScenarioAssignmentsWholeScenario(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	scenario := "test"

	scenarios := []interface{}{scenario}

	tenantID := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenario)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenario)

	formation := graphql.FormationInput{Name: scenario}

	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation, subaccountID, tenantID)
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formation.Name, subaccountID, tenantID)

	rtm := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, fixRuntimeInput("test-name"), conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, &rtm)

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
