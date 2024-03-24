package asserters

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/json"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

const (
	eventuallyTimeout = 8 * time.Second
	eventuallyTick    = 50 * time.Millisecond
)

var exactJSONConfigMatcher = func(t require.TestingT, expectedConfig, actualConfig *string) bool {
	return json.AssertJSONStringEquality(t, expectedConfig, actualConfig)
}

type FormationAssignmentsAsyncAsserter struct {
	FormationAssignmentsAsserter
	formationName string
	timeout       time.Duration
	tick          time.Duration
}

func NewFormationAssignmentAsyncAsserter(expectations map[string]map[string]fixtures.AssignmentState, expectedAssignmentsCount int, certSecuredGraphQLClient *graphql.Client, tenantID string) *FormationAssignmentsAsyncAsserter {
	f := FormationAssignmentsAsyncAsserter{
		FormationAssignmentsAsserter: FormationAssignmentsAsserter{
			expectations:             expectations,
			expectedAssignmentsCount: expectedAssignmentsCount,
			certSecuredGraphQLClient: certSecuredGraphQLClient,
			tenantID:                 tenantID,
		},
		timeout: eventuallyTimeout,
		tick:    eventuallyTick,
	}
	return &f
}

func (a *FormationAssignmentsAsyncAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	var formationID string
	if a.formationName != "" {
		formation := fixtures.GetFormationByName(t, ctx, a.certSecuredGraphQLClient, a.formationName, a.tenantID)
		formationID = formation.ID
	} else {
		formationID = ctx.Value(context_keys.FormationIDKey).(string)
	}
	a.assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, a.certSecuredGraphQLClient, a.tenantID, formationID, a.expectedAssignmentsCount, a.expectations, exactJSONConfigMatcher)
}

func (a *FormationAssignmentsAsyncAsserter) WithFormationName(formationName string) *FormationAssignmentsAsyncAsserter {
	a.formationName = formationName
	return a
}

func (a *FormationAssignmentsAsyncAsserter) WithTimeout(timeout time.Duration) *FormationAssignmentsAsyncAsserter {
	a.timeout = timeout
	return a
}

func (a *FormationAssignmentsAsyncAsserter) WithTick(tick time.Duration) *FormationAssignmentsAsyncAsserter {
	a.tick = tick
	return a
}

func (a *FormationAssignmentsAsyncAsserter) assertFormationAssignmentsAsynchronouslyWithEventually(t *testing.T, ctx context.Context, certSecuredGraphQLClient *graphql.Client, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.AssignmentState, configMatcher func(t require.TestingT, expectedConfig, actualConfig *string) bool) {
	t.Logf("Asserting formation assignments with eventually...")
	tOnce := testingx.NewOnceLogger(t)
	require.Eventually(t, func() (isOkay bool) {
		tOnce.Logf("Getting formation assignments...")
		listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
		assignmentsPage := fixtures.ListFormationAssignments(tOnce, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
		if expectedAssignmentsCount != assignmentsPage.TotalCount {
			tOnce.Logf("The expected assignments count: %d didn't match the actual: %d", expectedAssignmentsCount, assignmentsPage.TotalCount)
			return
		}
		tOnce.Logf("There is/are: %d assignment(s), assert them with the expected ones...", assignmentsPage.TotalCount)

		assignments := assignmentsPage.Data
		for _, assignment := range assignments {
			sourceAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
			if !ok {
				tOnce.Logf("Could not find expectations for assignment with ID: %q and source ID: %q", assignment.ID, assignment.Source)
				return
			}
			assignmentExpectation, ok := sourceAssignmentsExpectations[assignment.Target]
			if !ok {
				tOnce.Logf("Could not find expectations for assignment with ID: %q, source ID: %q and target ID: %q", assignment.ID, assignment.Source, assignment.Target)
				return
			}
			if assignmentExpectation.State != assignment.State {
				tOnce.Logf("The expected assignment state: %s doesn't match the actual: %s for assignment ID: %s", assignmentExpectation.State, assignment.State, assignment.ID)
				return
			}
			if isEqual := json.AssertJSONStringEquality(tOnce, assignmentExpectation.Error, assignment.Error); !isEqual {
				tOnce.Logf("The expected assignment state: %s doesn't match the actual: %s for assignment ID: %s", str.PtrStrToStr(assignmentExpectation.Error), str.PtrStrToStr(assignment.Error), assignment.ID)
				return
			}
			if isEqual := configMatcher(t, assignmentExpectation.Config, assignment.Configuration); !isEqual {
				tOnce.Logf("The expected assignment config: %s doesn't match the actual: %s for assignment ID: %s", str.PtrStrToStr(assignmentExpectation.Config), str.PtrStrToStr(assignment.Configuration), assignment.ID)
				return
			}
		}

		tOnce.Logf("Successfully asserted formation assignments asynchronously")
		return true
	}, a.timeout, a.tick, "Timed out after %s while trying to assert formation assignments.", a.timeout)
}
