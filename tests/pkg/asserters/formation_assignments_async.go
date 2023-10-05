package asserters

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type FormationAssignmentsAsyncAsserter struct {
	FormationAssignmentsAsserter
	delay int64
}

func NewFormationAssignmentAsyncAsserter(expectations map[string]map[string]fixtures.AssignmentState, expectedAssignmentsCount int, certSecuredGraphQLClient *graphql.Client, tenantID string, formationID string, delay int64) *FormationAssignmentsAsyncAsserter {
	f := FormationAssignmentsAsyncAsserter{}
	f.expectations = expectations
	f.expectedAssignmentsCount = expectedAssignmentsCount
	f.certSecuredGraphQLClient = certSecuredGraphQLClient
	f.tenantID = tenantID
	f.formationID = formationID
	f.delay = delay
	return &f
}

func (a *FormationAssignmentsAsyncAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	a.assertFormationAssignmentsAsynchronously(t, ctx, a.certSecuredGraphQLClient, a.tenantID, a.formationID, a.expectedAssignmentsCount, a.expectations)
}

func (a *FormationAssignmentsAsyncAsserter) assertFormationAssignmentsAsynchronously(t *testing.T, ctx context.Context, certSecuredGraphQLClient *graphql.Client, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.AssignmentState) {
	t.Logf("Sleeping for %d seconds while the async formation assignment status is proccessed...", a.delay)
	time.Sleep(time.Second * time.Duration(a.delay))
	listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
	require.Equal(t, expectedAssignmentsCount, assignmentsPage.TotalCount)

	assignments := assignmentsPage.Data

	for _, assignment := range assignments {
		targetAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
		require.Truef(t, ok, "Could not find expectations for assignment with ID: %q and source %q", assignment.ID, assignment.Source)

		assignmentExpectation, ok := targetAssignmentsExpectations[assignment.Target]
		require.Truef(t, ok, "Could not find expectations for assignment with ID: %q, source %q and target %q", assignment.ID, assignment.Source, assignment.Target)
		require.Equal(t, assignmentExpectation.State, assignment.State, "Assignment with ID: %q has different state than expected", assignment.ID)

		require.Equal(t, str.PtrStrToStr(assignmentExpectation.Error), str.PtrStrToStr(assignment.Error))

		expectedAssignmentConfigStr := str.PtrStrToStr(assignmentExpectation.Config)
		actualAssignmentConfigStr := str.PtrStrToStr(assignment.Configuration)
		if expectedAssignmentConfigStr != "" && expectedAssignmentConfigStr != "\"\"" && actualAssignmentConfigStr != "" && actualAssignmentConfigStr != "\"\"" {
			require.JSONEq(t, expectedAssignmentConfigStr, actualAssignmentConfigStr)
			require.JSONEq(t, str.PtrStrToStr(assignmentExpectation.Config), actualAssignmentConfigStr)
		} else {
			require.Equal(t, expectedAssignmentConfigStr, actualAssignmentConfigStr)
		}
	}
}
