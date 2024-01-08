package asserters

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type FormationAssignmentsAsserter struct {
	expectations             map[string]map[string]fixtures.AssignmentState
	expectedAssignmentsCount int
	certSecuredGraphQLClient *graphql.Client
	tenantID                 string
}

func NewFormationAssignmentAsserter(expectations map[string]map[string]fixtures.AssignmentState, expectedAssignmentsCount int, certSecuredGraphQLClient *graphql.Client, tenantID string) *FormationAssignmentsAsserter {
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

func (a *FormationAssignmentsAsserter) assertFormationAssignments(t *testing.T, ctx context.Context, certSecuredGraphQLClient *graphql.Client, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.AssignmentState) {
	listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
	assignments := assignmentsPage.Data
	require.Equal(t, expectedAssignmentsCount, assignmentsPage.TotalCount)
	for _, assignment := range assignments {
		sourceAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
		require.Truef(t, ok, "Could not find expectations for assignment with source %q", assignment.Source)

		assignmentExpectation, ok := sourceAssignmentsExpectations[assignment.Target]
		require.Truef(t, ok, "Could not find expectations for assignment with source %q and target %q", assignment.Source, assignment.Target)

		require.Equal(t, assignmentExpectation.State, assignment.State)
		expectedAssignmentConfigStr := str.PtrStrToStr(assignmentExpectation.Config)
		assignmentConfiguration := str.PtrStrToStr(assignment.Configuration)
		if expectedAssignmentConfigStr != "" && expectedAssignmentConfigStr != "\"\"" && assignmentConfiguration != "" && assignmentConfiguration != "\"\"" {
			require.JSONEq(t, expectedAssignmentConfigStr, assignmentConfiguration)
		} else {
			require.Equal(t, expectedAssignmentConfigStr, assignmentConfiguration)
		}
		if str.PtrStrToStr(assignmentExpectation.Value) != "" && str.PtrStrToStr(assignmentExpectation.Value) != "\"\"" && str.PtrStrToStr(assignment.Value) != "" && str.PtrStrToStr(assignment.Value) != "\"\"" {
			require.JSONEq(t, str.PtrStrToStr(assignmentExpectation.Value), str.PtrStrToStr(assignment.Value))
		} else {
			require.Equal(t, expectedAssignmentConfigStr, assignmentConfiguration)
		}
		require.Equal(t, str.PtrStrToStr(assignmentExpectation.Error), str.PtrStrToStr(assignment.Error))
	}
}
