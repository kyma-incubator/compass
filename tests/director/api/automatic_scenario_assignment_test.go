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
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DeleteAutomaticScenarioAssignmentForScenario(t *testing.T) {
	defaultValue := "DEFAULT"
	scenario1 := "Test1"
	scenario2 := "Test2"
	scenarios := []string{defaultValue, scenario1, scenario2}
	ctx := context.Background()
	tenantID := testTenants.GetIDByName(t, "TestDeleteAssignmentsForScenario")
	createScenariosLabelDefinitionWithinTenant(t, ctx, tenantID, scenarios)

	assignment1 := graphql.AutomaticScenarioAssignmentSetInput{ScenarioName: scenario1,
		Selector: &graphql.LabelSelectorInput{
			Value: "test-value",
			Key:   "test-key",
		},
	}
	assignment2 := graphql.AutomaticScenarioAssignmentSetInput{ScenarioName: scenario2,
		Selector: &graphql.LabelSelectorInput{
			Value: "test-value",
			Key:   "test-key",
		},
	}

	var output graphql.AutomaticScenarioAssignment

	assignment1Gql, err := tc.graphqlizer.AutomaticScenarioAssignmentSetInputToGQL(assignment1)
	require.NoError(t, err)
	req := fixSetAutomaticScenarioAssignmentRequest(assignment1Gql)
	err = tc.RunOperationWithCustomTenant(ctx, tenantID, req, nil)
	require.NoError(t, err)
	saveExample(t, req.Query(), "set automatic scenario assignment")

	setAutomaticScenarioAssignmentInTenant(t, ctx, assignment2, tenantID)

	req = fixDeleteAutomaticScenarioAssignmentForScenarioRequest(scenario1)
	err = tc.RunOperationWithCustomTenant(ctx, tenantID, req, &output)
	require.NoError(t, err)

	assertAutomaticScenarioAssignment(t, assignment1, output)

	allAssignments := listAutomaticScenarioAssignmentsWithinTenant(t, ctx, tenantID)
	require.Equal(t, 1, allAssignments.TotalCount)
	assertAutomaticScenarioAssignment(t, assignment2, *allAssignments.Data[0])

	saveExample(t, req.Query(), "delete automatic scenario assignment for scenario")

}

func Test_DeleteAutomaticScenarioAssignmentForSelector(t *testing.T) {
	defaultValue := "DEFAULT"
	scenario1 := "Test1"
	scenario2 := "Test2"
	scenarios := []string{defaultValue, scenario1, scenario2}
	ctx := context.Background()
	tenantID := testTenants.GetIDByName(t, "TestDeleteAssignmentsForSelector")
	createScenariosLabelDefinitionWithinTenant(t, ctx, tenantID, scenarios)

	selector := graphql.LabelSelectorInput{Key: "test-key", Value: "test-value"}

	assignments := []graphql.AutomaticScenarioAssignmentSetInput{
		{
			ScenarioName: scenario1,
			Selector:     &selector,
		},
		{
			ScenarioName: scenario2,
			Selector:     &selector,
		},
	}

	var output []*graphql.AutomaticScenarioAssignment

	setAutomaticScenarioAssignmentInTenant(t, ctx, assignments[0], tenantID)
	setAutomaticScenarioAssignmentInTenant(t, ctx, assignments[1], tenantID)
	selectorGql, err := tc.graphqlizer.LabelSelectorInputToGQL(selector)
	require.NoError(t, err)

	req := fixDeleteAutomaticScenarioAssignmentsForSelectorRequest(selectorGql)
	err = tc.RunOperationWithCustomTenant(ctx, tenantID, req, &output)
	require.NoError(t, err)

	assertAutomaticScenarioAssignments(t, assignments, output)

	allAssignments := listAutomaticScenarioAssignmentForSelectorWithinTenant(t, ctx, tenantID, selector)
	assert.Equal(t, []*graphql.AutomaticScenarioAssignment{}, allAssignments)

	saveExample(t, req.Query(), "delete automatic scenario assignment for selector")

}

func createScenariosLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, tenantID string, scenarios []string) *graphql.LabelDefinition {
	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": scenarios,
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}
	return createLabelDefinitionWithinTenant(t, ctx, "scenarios", jsonSchema, tenantID)
}
