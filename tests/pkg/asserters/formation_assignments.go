package asserters

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/operations"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"testing"
)

type FormationAssignmentsAsserter struct {
	expectations             map[string]map[string]fixtures.AssignmentState
	expectedAssignmentsCount int
	certSecuredGraphQLClient *graphql.Client
	tenantID                 string
	delay                    int
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
	formationID := ctx.Value(operations.FormationIDKey).(string)
	a.assertFormationAssignments(t, ctx, a.certSecuredGraphQLClient, a.tenantID, formationID, a.expectedAssignmentsCount, a.expectations)
}

func (a *FormationAssignmentsAsserter) assertFormationAssignments(t *testing.T, ctx context.Context, certSecuredGraphQLClient *graphql.Client, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.AssignmentState) {
	listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
	assignments := assignmentsPage.Data
	require.Equal(t, expectedAssignmentsCount, assignmentsPage.TotalCount)
	for _, assignment := range assignments {
		targetAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
		require.Truef(t, ok, "Could not find expectations for assignment with source %q", assignment.Source)

		assignmentExpectation, ok := targetAssignmentsExpectations[assignment.Target]
		require.Truef(t, ok, "Could not find expectations for assignment with source %q and target %q", assignment.Source, assignment.Target)

		require.Equal(t, assignmentExpectation.State, assignment.State)
		require.Equal(t, str.PtrStrToStr(assignmentExpectation.Config), str.PtrStrToStr(assignment.Configuration))
		require.Equal(t, str.PtrStrToStr(assignmentExpectation.Value), str.PtrStrToStr(assignment.Value))
		require.Equal(t, str.PtrStrToStr(assignmentExpectation.Error), str.PtrStrToStr(assignment.Error))
	}
}
