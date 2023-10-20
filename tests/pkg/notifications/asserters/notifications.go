package asserters

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	unassignOperation = "unassign"
	assignOperation   = "assign"
)

type NotificationsAsserter struct {
	expectedNotificationsCount         int
	op                                 string
	targetObjectID                     string
	sourceObjectID                     string
	localTenantID                      string
	appNamespace                       string
	region                             string
	tenant                             string
	tenantParentCustomer               string
	externalServicesMockMtlsSecuredURL string
	client                             *http.Client
}

func NewNotificationsAsserter(expectedNotificationsCount int, op string, targetObjectID, sourceObjectID string, localTenantID string, appNamespace string, region string, tenant string, tenantParentCustomer string, externalServicesMockMtlsSecuredURL string, client *http.Client) *NotificationsAsserter {
	return &NotificationsAsserter{
		expectedNotificationsCount:         expectedNotificationsCount,
		op:                                 op,
		targetObjectID:                     targetObjectID,
		sourceObjectID:                     sourceObjectID,
		localTenantID:                      localTenantID,
		appNamespace:                       appNamespace,
		region:                             region,
		tenant:                             tenant,
		tenantParentCustomer:               tenantParentCustomer,
		externalServicesMockMtlsSecuredURL: externalServicesMockMtlsSecuredURL,
		client:                             client,
	}
}

func (a *NotificationsAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	formationID := ctx.Value(context_keys.FormationIDKey).(string)

	body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)
	assertNotificationsCount(t, body, a.targetObjectID, a.expectedNotificationsCount)

	notificationsForTarget := gjson.GetBytes(body, a.targetObjectID)
	assignNotificationAboutSource := notificationsForTarget.Array()[0]
	assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t, assignNotificationAboutSource, assignOperation, formationID, a.sourceObjectID, a.localTenantID, a.appNamespace, a.region, a.tenant, a.tenantParentCustomer, nil)
}

func getNotificationsFromExternalSvcMock(t *testing.T, client *http.Client, ExternalServicesMockMtlsSecuredURL string) []byte {
	t.Logf("Getting formation notifications recieved in external services mock")
	resp, err := client.Get(ExternalServicesMockMtlsSecuredURL + "/formation-callback")
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusOK, string(body)))
	return body
}

func assertNotificationsCount(t *testing.T, body []byte, objectID string, count int) {
	notifications := gjson.GetBytes(body, objectID)
	if count > 0 {
		require.True(t, notifications.Exists())
		require.Len(t, notifications.Array(), count)
	} else {
		require.False(t, notifications.Exists())
	}
}

func assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t *testing.T, notification gjson.Result, op, formationID, expectedAppID, expectedLocalTenantID, expectedAppNamespace, expectedAppRegion, expectedTenant, expectedCustomerID string, expectedConfig *string) {
	require.Equal(t, op, notification.Get("Operation").String())
	if op == unassignOperation {
		require.Equal(t, expectedAppID, notification.Get("ApplicationID").String())
	}
	require.Equal(t, formationID, notification.Get("RequestBody.ucl-formation-id").String())
	require.Equal(t, expectedTenant, notification.Get("RequestBody.globalAccountId").String())
	require.Equal(t, expectedCustomerID, notification.Get("RequestBody.crmId").String())

	notificationItems := notification.Get("RequestBody.items")
	require.True(t, notificationItems.Exists())
	require.Len(t, notificationItems.Array(), 1)

	app1FromNotification := notificationItems.Array()[0]
	require.Equal(t, expectedAppID, app1FromNotification.Get("ucl-system-tenant-id").String())
	require.Equal(t, expectedLocalTenantID, app1FromNotification.Get("tenant-id").String())
	require.Equal(t, expectedAppNamespace, app1FromNotification.Get("application-namespace").String())
	require.Equal(t, expectedAppRegion, app1FromNotification.Get("region").String())
	if expectedConfig != nil {
		require.Equal(t, *expectedConfig, notification.Get("RequestBody.config").String())
	}
}
