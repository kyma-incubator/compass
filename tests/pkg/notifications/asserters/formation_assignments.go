package asserters

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	gql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type FormationAssignmentsAsserter struct {
	expectations             map[string]map[string]fixtures.Assignment
	expectedAssignmentsCount int
	certSecuredGraphQLClient *graphql.Client
	tenantID                 string
}

func NewFormationAssignmentAsserter(expectations map[string]map[string]fixtures.Assignment, expectedAssignmentsCount int, certSecuredGraphQLClient *graphql.Client, tenantID string) *FormationAssignmentsAsserter {
	return &FormationAssignmentsAsserter{
		expectations:             expectations,
		expectedAssignmentsCount: expectedAssignmentsCount,
		certSecuredGraphQLClient: certSecuredGraphQLClient,
		tenantID:                 tenantID,
	}
}

func (a *FormationAssignmentsAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	formationID := ctx.Value(context_keys.FormationIDKey).(string)
	a.assertFormationAssignments(t, ctx, a.certSecuredGraphQLClient, a.tenantID, formationID, a.expectedAssignmentsCount, a.expectations)
	t.Log("Formation assignments are successfully asserted")
}

func (a *FormationAssignmentsAsserter) assertFormationAssignments(t *testing.T, ctx context.Context, certSecuredGraphQLClient *graphql.Client, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.Assignment) {
	spew.Dump(expectedAssignments)
	listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
	assignments := assignmentsPage.Data
	require.Equal(t, expectedAssignmentsCount, assignmentsPage.TotalCount)
	for _, assignment := range assignments {
		sourceAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
		require.Truef(t, ok, "Could not find expectations for assignment %q with source %q", assignment.ID, assignment.Source)

		assignmentExpectation, ok := sourceAssignmentsExpectations[assignment.Target]
		require.Truef(t, ok, "Could not find expectations for assignment %q with source %q and target %q", assignment.ID, assignment.Source, assignment.Target)

		// asserting state
		require.Equal(t, assignmentExpectation.AssignmentStatus.State, assignment.State)
		expectedAssignmentConfigStr := str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Config)
		assignmentConfiguration := str.PtrStrToStr(assignment.Configuration)
		if expectedAssignmentConfigStr != "" && expectedAssignmentConfigStr != "\"\"" && assignmentConfiguration != "" && assignmentConfiguration != "\"\"" {
			require.JSONEq(t, expectedAssignmentConfigStr, assignmentConfiguration)
		} else {
			require.Equal(t, expectedAssignmentConfigStr, assignmentConfiguration)
		}
		if str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Value) != "" && str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Value) != "\"\"" && str.PtrStrToStr(assignment.Value) != "" && str.PtrStrToStr(assignment.Value) != "\"\"" {
			require.JSONEq(t, str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Value), str.PtrStrToStr(assignment.Value))
		} else {
			require.Equal(t, expectedAssignmentConfigStr, assignmentConfiguration)
		}
		require.Equal(t, str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Error), str.PtrStrToStr(assignment.Error))

		// asserting operations
		require.Equal(t, len(assignmentExpectation.Operations), len(assignment.AssignmentOperations.Data))
		for _, expectedOperation := range assignmentExpectation.Operations {
			require.Truef(t, ContainsMatchingOperation(expectedOperation, assignment.AssignmentOperations.Data), "Could not find expected operation %v in assignment with ID %q", expectedOperation, assignment.ID)
		}
	}
}

func ContainsMatchingOperation(expectedOperation *fixtures.Operation, actualOperations []*gql.AssignmentOperation) bool {
	for _, actualOperation := range actualOperations {
		actualOperationIsFinished := actualOperation.FinishedAtTimestamp != nil
		if expectedOperation.Type == actualOperation.OperationType && expectedOperation.TriggeredBy == actualOperation.TriggeredBy && expectedOperation.IsFinished == actualOperationIsFinished {
			return true
		}
	}

	return false
}
