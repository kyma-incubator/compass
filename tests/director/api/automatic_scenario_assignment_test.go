package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func TestAutomaticScenarioAssignmentQueries(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	tenantID := testTenants.GetIDByName(t, "ASA1")

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
	testSelectorAGQL, err := tc.graphqlizer.LabelSelectorInputToGQL(testSelectorA)
	require.NoError(t, err)

	// setup available scenarios
	createScenariosLabelDefinitionWithinTenant(t, ctx, tenantID, []string{"DEFAULT", testScenarioA, testScenarioB, testScenarioC})

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
	setAutomaticScenarioAssignmentFromInputWithinTenant(t, ctx, inputAssignment1, tenantID)
	defer deleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, tenantID, testScenarioA)
	setAutomaticScenarioAssignmentFromInputWithinTenant(t, ctx, inputAssignment2, tenantID)
	defer deleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, tenantID, testScenarioB)
	setAutomaticScenarioAssignmentFromInputWithinTenant(t, ctx, inputAssignment3, tenantID)
	defer deleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, tenantID, testScenarioC)

	// prepare queries
	getAssignmentForScenarioRequest := fixAutomaticScenarioAssignmentForScenarioRequest(testScenarioA)
	listAssignmentsRequest := fixAutomaticScenarioAssignmentsRequest()
	listAssignmentsForSelectorRequest := fixAutomaticScenarioAssignmentForSelectorRequest(testSelectorAGQL)

	actualAssignmentsPage := graphql.AutomaticScenarioAssignmentPage{}
	actualAssignmentForScenario := graphql.AutomaticScenarioAssignment{}
	actualAssignmentsForSelector := []*graphql.AutomaticScenarioAssignment{}

	// WHEN
	err = tc.RunOperationWithCustomTenant(ctx, tenantID, listAssignmentsRequest, &actualAssignmentsPage)
	require.NoError(t, err)
	err = tc.RunOperationWithCustomTenant(ctx, tenantID, getAssignmentForScenarioRequest, &actualAssignmentForScenario)
	require.NoError(t, err)
	err = tc.RunOperationWithCustomTenant(ctx, tenantID, listAssignmentsForSelectorRequest, &actualAssignmentsForSelector)
	require.NoError(t, err)

	// THEN
	saveExample(t, listAssignmentsRequest.Query(), "query automatic scenario assignments")
	saveExample(t, getAssignmentForScenarioRequest.Query(), "query automatic scenario assignment for scenario")
	saveExample(t, listAssignmentsForSelectorRequest.Query(), "query automatic scenario assignment for selector")

	assertAutomaticScenarioAssignments(t,
		[]graphql.AutomaticScenarioAssignmentSetInput{inputAssignment1, inputAssignment2, inputAssignment3},
		actualAssignmentsPage.Data)
	assertAutomaticScenarioAssignment(t, inputAssignment1, actualAssignmentForScenario)
	assertAutomaticScenarioAssignments(t,
		[]graphql.AutomaticScenarioAssignmentSetInput{inputAssignment1, inputAssignment2},
		actualAssignmentsForSelector)
}
