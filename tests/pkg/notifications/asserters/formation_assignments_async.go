package asserters

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type FormationAssignmentsAsyncAsserter struct {
	FormationAssignmentsAsserter
	timeout time.Duration
	tick time.Duration
	delay int64
}

func NewFormationAssignmentAsyncAsserter(expectations map[string]map[string]fixtures.AssignmentState, expectedAssignmentsCount int, certSecuredGraphQLClient *graphql.Client, tenantID string) *FormationAssignmentsAsyncAsserter {
	f := FormationAssignmentsAsyncAsserter{
		FormationAssignmentsAsserter: FormationAssignmentsAsserter{
			expectations:             expectations,
			expectedAssignmentsCount: expectedAssignmentsCount,
			certSecuredGraphQLClient: certSecuredGraphQLClient,
			tenantID:                 tenantID,
		},
		timeout: time.Second*5,
		tick: time.Millisecond*50,
	}
	return &f
}

func (a *FormationAssignmentsAsyncAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	formationID := ctx.Value(context_keys.FormationIDKey).(string)
	a.assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, a.certSecuredGraphQLClient, a.tenantID, formationID, a.expectedAssignmentsCount, a.expectations)
}

func (a *FormationAssignmentsAsyncAsserter) assertFormationAssignmentsAsynchronously(t *testing.T, ctx context.Context, certSecuredGraphQLClient *graphql.Client, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.AssignmentState) {
	t.Logf("Sleeping for %d milliseconds while the async formation assignment status is proccessed...", a.delay)
	time.Sleep(time.Millisecond * time.Duration(a.delay))
	listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
	assignments := assignmentsPage.Data
	require.Equal(t, expectedAssignmentsCount, assignmentsPage.TotalCount)
	for _, assignment := range assignments {
		sourceAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
		require.Truef(t, ok, "Could not find expectations for assignment with ID: %q and source %q", assignment.ID, assignment.Source)

		assignmentExpectation, ok := sourceAssignmentsExpectations[assignment.Target]
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

func (a *FormationAssignmentsAsyncAsserter) assertFormationAssignmentsAsynchronouslyWithEventually(t *testing.T, ctx context.Context, certSecuredGraphQLClient *graphql.Client, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.AssignmentState) {
	t.Logf("Asserting formation assignments with eventually...")
	require.Eventually(t, func() (isOkay bool) {
		t.Logf("Getting formation assignments...")
		listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
		assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
		if expectedAssignmentsCount != assignmentsPage.TotalCount {
			t.Logf("The expected assignments count: %d didn't match the actual: %d", expectedAssignmentsCount, assignmentsPage.TotalCount)
			return
		}
		t.Logf("There is/are: %d assignment(s), assert them with the expected ones...", assignmentsPage.TotalCount)

		assignments := assignmentsPage.Data
		for _, assignment := range assignments {
			sourceAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
			if !ok {
				t.Logf("Could not find expectations for assignment with ID: %q and source ID: %q", assignment.ID, assignment.Source)
				return
			}
			assignmentExpectation, ok := sourceAssignmentsExpectations[assignment.Target]
			if !ok {
				t.Logf("Could not find expectations for assignment with ID: %q, source ID: %q and target ID: %q", assignment.ID, assignment.Source, assignment.Target)
				return
			}
			if assignmentExpectation.State != assignment.State {
				t.Logf("The expected assignment state: %s doesn't match the actual: %s for assignment ID: %s", assignmentExpectation.State, assignment.State, assignment.ID)
				return
			}
			if isEqual := assertJSONStringEquality(t, assignmentExpectation.Error, assignment.Error); !isEqual {
				t.Logf("The expected assignment state: %s doesn't match the actual: %s for assignment ID: %s", str.PtrStrToStr(assignmentExpectation.Error), str.PtrStrToStr(assignment.Error), assignment.ID)
				return
			}
			if isEqual := assertJSONStringEquality(t, assignmentExpectation.Config, assignment.Configuration); !isEqual {
				t.Logf("The expected assignment config: %s doesn't match the actual: %s for assignment ID: %s", str.PtrStrToStr(assignmentExpectation.Config), str.PtrStrToStr(assignment.Configuration), assignment.ID)
				return
			}
		}

		t.Logf("Successfully asserted formation asssignments asynchronously")
		return true
	}, a.timeout, a.tick)
}

func assertJSONStringEquality(t *testing.T, expectedValue, actualValue *string) bool {
	expectedValueStr := str.PtrStrToStr(expectedValue)
	actualValueStr := str.PtrStrToStr(actualValue)
	if !isJSONStringEmpty(expectedValueStr) && !isJSONStringEmpty(actualValueStr) {
		return assert.JSONEq(t, expectedValueStr, actualValueStr)
	} else {
		return assert.Equal(t, expectedValueStr, actualValueStr)
	}
}

func isJSONStringEmpty(json string) bool {
	if json != "" && json != "\"\"" {
		return false
	}
	return true
}