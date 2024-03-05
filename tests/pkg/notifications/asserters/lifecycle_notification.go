package asserters

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
	"time"
)

type LifecycleNotificationsAsserter struct {
	operation                          string
	formationName                      string
	tenantID                           string
	parentTenantID                     string
	externalServicesMockMtlsSecuredURL string
	state                              string
	certSecuredGraphQLClient           *graphql.Client
	client                             *http.Client
	timeout                            time.Duration
	tick                               time.Duration
}

func NewLifecycleNotificationsAsserter(externalServicesMockMtlsSecuredURL string, gqlClient *graphql.Client, client *http.Client) *LifecycleNotificationsAsserter {
	return &LifecycleNotificationsAsserter{
		externalServicesMockMtlsSecuredURL: externalServicesMockMtlsSecuredURL,
		certSecuredGraphQLClient:           gqlClient,
		client:                             client,
		timeout:                            eventuallyTimeout,
		tick:                               eventuallyTick,
	}
}

func (a *LifecycleNotificationsAsserter) WithOperation(operation string) *LifecycleNotificationsAsserter {
	a.operation = operation
	return a
}

func (a *LifecycleNotificationsAsserter) WithFormationName(formationName string) *LifecycleNotificationsAsserter {
	a.formationName = formationName
	return a
}

func (a *LifecycleNotificationsAsserter) WithTenantID(tenantID string) *LifecycleNotificationsAsserter {
	a.tenantID = tenantID
	return a
}

func (a *LifecycleNotificationsAsserter) WithParentTenantID(parentTenantID string) *LifecycleNotificationsAsserter {
	a.parentTenantID = parentTenantID
	return a
}

func (a *LifecycleNotificationsAsserter) WithState(state string) *LifecycleNotificationsAsserter {
	a.state = state
	return a
}

func (a *LifecycleNotificationsAsserter) WithTimeout(timeout time.Duration) *LifecycleNotificationsAsserter {
	a.timeout = timeout
	return a
}

func (a *LifecycleNotificationsAsserter) WithTick(tick time.Duration) *LifecycleNotificationsAsserter {
	a.tick = tick
	return a
}

func (a *LifecycleNotificationsAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	formation := fixtures.GetFormationByName(t, ctx, a.certSecuredGraphQLClient, a.formationName, a.tenantID)
	body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)
	a.assertAsyncFormationNotificationFromCreationOrDeletionWithEventually(t, ctx, body, formation.ID, a.formationName, a.state, a.operation, a.tenantID, a.parentTenantID, a.timeout, a.tick)
}

func (a *LifecycleNotificationsAsserter) assertAsyncFormationNotificationFromCreationOrDeletionWithEventually(t *testing.T, ctx context.Context, body []byte, formationID, formationName, formationState, formationOperation, tenantID, parentTenantID string, timeout, tick time.Duration) {
	var shouldExpectDeleted bool
	if formationOperation == createFormationOperation || formationState == "DELETE_ERROR" {
		shouldExpectDeleted = false
	} else {
		shouldExpectDeleted = true
	}
	a.assertAsyncFormationNotificationFromCreationOrDeletionExpectDeletedWithEventually(t, ctx, body, formationID, formationName, formationState, formationOperation, tenantID, parentTenantID, shouldExpectDeleted, timeout, tick)
}
func (a *LifecycleNotificationsAsserter) assertAsyncFormationNotificationFromCreationOrDeletionExpectDeletedWithEventually(t *testing.T, ctx context.Context, body []byte, formationID, formationName, formationState, formationOperation, tenantID, parentTenantID string, shouldExpectDeleted bool, timeout, tick time.Duration) {
	t.Logf("Assert asynchronous formation lifecycle notifications are sent for %q operation...", formationOperation)
	notificationsForFormation := gjson.GetBytes(body, formationID)
	require.True(t, notificationsForFormation.Exists())
	require.Len(t, notificationsForFormation.Array(), 1)

	notificationForFormation := notificationsForFormation.Array()[0]
	require.Equal(t, formationOperation, notificationForFormation.Get("Operation").String())
	require.Equal(t, tenantID, notificationForFormation.Get("RequestBody.globalAccountId").String())
	require.Equal(t, parentTenantID, notificationForFormation.Get("RequestBody.crmId").String())

	notificationForFormationDetails := notificationForFormation.Get("RequestBody.details")
	require.True(t, notificationForFormationDetails.Exists())
	require.Equal(t, formationID, notificationForFormationDetails.Get("id").String())
	require.Equal(t, formationName, notificationForFormationDetails.Get("name").String())

	t.Logf("Asserting formation with eventually...")
	tOnce := testingx.NewOnceLogger(t)
	require.Eventually(t, func() (isOkay bool) {
		tOnce.Log("Assert formation lifecycle notifications are successfully processed...")
		formationPage := fixtures.ListFormationsWithinTenant(t, ctx, tenantID, a.certSecuredGraphQLClient)
		if shouldExpectDeleted {
			if formationPage.TotalCount != 0 {
				tOnce.Logf("Formation lifecycle notification is expected to have deleted formation with ID %q, but it is still there", formationID)
				return
			}
			if formationPage.Data != nil && len(formationPage.Data) > 0 {
				tOnce.Logf("Formation lifecycle notification is expected to have deleted formation with ID %q, but it is still there", formationID)
				return
			}
		} else {
			if formationPage.TotalCount != 1 {
				tOnce.Log("Formation count does not match")
				return
			}
			if formationPage.Data[0].State != formationState {
				tOnce.Logf("Formation state for formation with ID %q is %q, expected: %q", formationID, formationPage.Data[0].State, formationState)
				return
			}
			if formationPage.Data[0].ID != formationID {
				tOnce.Logf("Formation ID is %q, expected: %q", formationPage.Data[0].ID, formationID)
				return
			}
			if formationPage.Data[0].Name != formationName {
				tOnce.Logf("Formation name is %q, expected: %q", formationPage.Data[0].Name, formationName)
				return
			}
		}

		tOnce.Logf("Asynchronous formation lifecycle notifications are successfully validated for %q operation.", formationOperation)
		return true
	}, timeout, tick)
}
