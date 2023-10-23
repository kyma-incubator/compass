package asserters

import (
	"context"
	"net/http"
	"testing"

	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type UnassignNotificationsAsserter struct {
	op                                 string
	expectedNotificationsCountForOp    int
	targetObjectID                     string
	sourceObjectID                     string
	localTenantID                      string
	appNamespace                       string
	region                             string
	tenant                             string
	tenantParentCustomer               string
	config                             string
	externalServicesMockMtlsSecuredURL string
	client                             *http.Client
}

func NewUnassignNotificationsAsserter(expectedNotificationsCountForOp int, targetObjectID string, sourceObjectID string, localTenantID string, appNamespace string, region string, tenant string, tenantParentCustomer string, config string, externalServicesMockMtlsSecuredURL string, client *http.Client) *UnassignNotificationsAsserter {
	return &UnassignNotificationsAsserter{
		op:                                 unassignOperation,
		expectedNotificationsCountForOp:    expectedNotificationsCountForOp,
		targetObjectID:                     targetObjectID,
		sourceObjectID:                     sourceObjectID,
		localTenantID:                      localTenantID,
		appNamespace:                       appNamespace,
		region:                             region,
		tenant:                             tenant,
		tenantParentCustomer:               tenantParentCustomer,
		config:                             config,
		externalServicesMockMtlsSecuredURL: externalServicesMockMtlsSecuredURL,
		client:                             client,
	}
}

func (a *UnassignNotificationsAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	formationID := ctx.Value(context_keys.FormationIDKey).(string)
	body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)

	notificationsForTarget := gjson.GetBytes(body, a.targetObjectID)
	notificationsFoundCount := 0
	for _, notification := range notificationsForTarget.Array() {
		op := notification.Get("Operation").String()
		if op == a.op {
			notificationsFoundCount++
			err := verifyFormationAssignmentNotification(t, notification, unassignOperation, formationID, a.sourceObjectID, a.localTenantID, a.appNamespace, a.region, a.config, a.tenant, a.tenantParentCustomer, false)
			require.NoError(t, err)
		}
	}
	require.Equal(t, a.expectedNotificationsCountForOp, notificationsFoundCount, "expected %s notifications for target object %s", a.expectedNotificationsCountForOp, a.targetObjectID)
}
