package asserters

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type FormationAssignmentsAsyncCustomConfigMatcherAsserter struct {
	configMatcher func(t require.TestingT, expectedConfig, actualConfig *string) bool
	FormationAssignmentsAsyncAsserter
}

func NewFormationAssignmentsAsyncCustomConfigMatcherAsserter(configMatcher func(require.TestingT, *string, *string) bool, expectations map[string]map[string]fixtures.Assignment, expectedAssignmentsCount int, certSecuredGraphQLClient *graphql.Client, tenantID string) *FormationAssignmentsAsyncCustomConfigMatcherAsserter {
	f := FormationAssignmentsAsyncCustomConfigMatcherAsserter{
		configMatcher: configMatcher,
		FormationAssignmentsAsyncAsserter: FormationAssignmentsAsyncAsserter{
			FormationAssignmentsAsserter: FormationAssignmentsAsserter{
				expectations:             expectations,
				expectedAssignmentsCount: expectedAssignmentsCount,
				certSecuredGraphQLClient: certSecuredGraphQLClient,
				tenantID:                 tenantID,
			},
			timeout: eventuallyTimeout,
			tick:    eventuallyTick,
		},
	}
	return &f
}

func (a *FormationAssignmentsAsyncCustomConfigMatcherAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	formationID := ctx.Value(context_keys.FormationIDKey).(string)
	a.assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, a.certSecuredGraphQLClient, a.tenantID, formationID, a.expectedAssignmentsCount, a.expectations, a.configMatcher)
}

func (a *FormationAssignmentsAsyncCustomConfigMatcherAsserter) WithTimeout(timeout time.Duration) *FormationAssignmentsAsyncCustomConfigMatcherAsserter {
	a.timeout = timeout
	return a
}

func (a *FormationAssignmentsAsyncCustomConfigMatcherAsserter) WithTick(tick time.Duration) *FormationAssignmentsAsyncCustomConfigMatcherAsserter {
	a.tick = tick
	return a
}
