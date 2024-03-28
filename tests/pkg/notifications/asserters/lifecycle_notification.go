package asserters

import (
	"context"
	"net/http"
	"testing"
	"time"

	gql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type LifecycleNotificationsAsserter struct {
	operation                          string
	formationName                      string
	tenantID                           string
	parentTenantID                     string
	state                              string
	expectNotifications                bool
	externalServicesMockMtlsSecuredURL string
	certSecuredGraphQLClient           *graphql.Client
	client                             *http.Client
	timeout                            time.Duration
	tick                               time.Duration
}

func NewLifecycleNotificationsAsserter(externalServicesMockMtlsSecuredURL string, gqlClient *graphql.Client, client *http.Client) *LifecycleNotificationsAsserter {
	return &LifecycleNotificationsAsserter{
		expectNotifications:                true,
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

func (a *LifecycleNotificationsAsserter) WithExpectNotifications(expectNotifications bool) *LifecycleNotificationsAsserter {
	a.expectNotifications = expectNotifications
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
	var formationID string
	var formationName string
	if a.formationName != "" {
		formation := fixtures.GetFormationByName(t, ctx, a.certSecuredGraphQLClient, a.formationName, a.tenantID)
		formationID = formation.ID
		formationName = a.formationName
	} else {
		formationID = ctx.Value(context_keys.FormationIDKey).(string)
		formationName = ctx.Value(context_keys.FormationNameKey).(string)
	}

	body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)
	a.assertAsyncFormationNotificationFromCreationOrDeletionWithEventually(t, ctx, body, formationID, formationName, a.state, a.operation, a.tenantID, a.parentTenantID, a.expectNotifications, a.timeout, a.tick)
}

func (a *LifecycleNotificationsAsserter) assertAsyncFormationNotificationFromCreationOrDeletionWithEventually(t *testing.T, ctx context.Context, body []byte, formationID, formationName, formationState, formationOperation, tenantID, parentTenantID string, expectNotifications bool, timeout, tick time.Duration) {
	var shouldExpectDeleted bool
	if formationOperation == createFormationOperation || formationState == "DELETE_ERROR" {
		shouldExpectDeleted = false
	} else {
		shouldExpectDeleted = true
	}
	a.assertAsyncFormationNotificationFromCreationOrDeletionExpectDeletedWithEventually(t, ctx, body, formationID, formationName, formationState, formationOperation, tenantID, parentTenantID, expectNotifications, shouldExpectDeleted, timeout, tick)
}
func (a *LifecycleNotificationsAsserter) assertAsyncFormationNotificationFromCreationOrDeletionExpectDeletedWithEventually(t *testing.T, ctx context.Context, body []byte, formationID, formationName, formationState, formationOperation, tenantID, parentTenantID string, expectNotifications, shouldExpectDeleted bool, timeout, tick time.Duration) {
	t.Logf("Assert asynchronous formation lifecycle notifications are sent for %q operation...", formationOperation)
	notificationsForFormation := gjson.GetBytes(body, formationID)

	if !expectNotifications {
		require.False(t, notificationsForFormation.Exists())
		require.Len(t, notificationsForFormation.Array(), 0)
		return
	}

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

		var foundFormation *gql.Formation
		for _, formation := range formationPage.Data {
			if formation.Name == a.formationName {
				foundFormation = formation
			}
		}
		if shouldExpectDeleted {
			if foundFormation != nil {
				tOnce.Logf("Formation lifecycle notification is expected to have deleted formation with ID %q, but it is still there", formationID)
				return
			}
		} else {
			if foundFormation == nil {
				tOnce.Logf("Formation with ID %s was not found", formationID)
				return
			}
			if foundFormation.State != formationState {
				tOnce.Logf("Formation state for formation with ID %q is %q, expected: %q", formationID, foundFormation.State, formationState)
				return
			}
			if foundFormation.ID != formationID {
				tOnce.Logf("Formation ID is %q, expected: %q", foundFormation.ID, formationID)
				return
			}
			if foundFormation.Name != formationName {
				tOnce.Logf("Formation name is %q, expected: %q", foundFormation.Name, formationName)
				return
			}
		}

		tOnce.Logf("Asynchronous formation lifecycle notifications are successfully validated for %q operation.", formationOperation)
		return true
	}, timeout, tick)
}
