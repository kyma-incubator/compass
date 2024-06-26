package asserters

import (
	"context"
	"testing"
	"time"

	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type FormationIsDeletedAsserter struct {
	certSecuredGraphQLClient *graphql.Client
	tenantID                 string
	formationName            string
	timeout                  time.Duration
	tick                     time.Duration
}

func NewFormationIsDeletedAsserter(certSecuredGraphQLClient *graphql.Client) *FormationIsDeletedAsserter {
	return &FormationIsDeletedAsserter{
		certSecuredGraphQLClient: certSecuredGraphQLClient,
		timeout:                  eventuallyTimeout,
		tick:                     eventuallyTick,
	}
}

func (a *FormationIsDeletedAsserter) WithFormationName(formationName string) *FormationIsDeletedAsserter {
	a.formationName = formationName
	return a
}

func (a *FormationIsDeletedAsserter) WithTenantID(tenantID string) *FormationIsDeletedAsserter {
	a.tenantID = tenantID
	return a
}

func (a *FormationIsDeletedAsserter) WithTimeout(timeout time.Duration) *FormationIsDeletedAsserter {
	a.timeout = timeout
	return a
}

func (a *FormationIsDeletedAsserter) WithTick(tick time.Duration) *FormationIsDeletedAsserter {
	a.tick = tick
	return a
}

func (a *FormationIsDeletedAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	t.Logf("Asserting formation assignments with eventually...")

	var formationName string
	if a.formationName != "" {
		formationName = a.formationName
	} else {
		formationName = ctx.Value(context_keys.FormationNameKey).(string)
	}

	tOnce := testingx.NewOnceLogger(t)
	require.Eventually(t, func() (isOkay bool) {
		// Get the formations for participant globally
		formationPage := fixtures.ListFormationsWithinTenant(t, ctx, a.tenantID, a.certSecuredGraphQLClient)
		foundFormation := false
		for _, formation := range formationPage.Data {
			if formation.Name == formationName {
				foundFormation = true
			}
		}

		if foundFormation {
			tOnce.Logf("Formation with name %s is not yet deleted", formationName)
			return false
		}
		tOnce.Logf("Successfully asserted formation with name %s is deleted", formationName)
		return true
	}, a.timeout, a.tick, "Timed out after %s while trying to assert formation with name %s is deleted.", a.timeout, formationName)
}
