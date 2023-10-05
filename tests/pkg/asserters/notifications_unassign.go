package asserters

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg/operations"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
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
	externalServicesMockMtlsSecuredURL string
	client                             *http.Client
}

func NewUnassignNotificationsAsserter(op string, expectedNotificationsCountForOp int, targetObjectID string, sourceObjectID string, localTenantID string, appNamespace string, region string, tenant string, tenantParentCustomer string, externalServicesMockMtlsSecuredURL string, client *http.Client) *UnassignNotificationsAsserter {
	return &UnassignNotificationsAsserter{op: op, expectedNotificationsCountForOp: expectedNotificationsCountForOp, targetObjectID: targetObjectID, sourceObjectID: sourceObjectID, localTenantID: localTenantID, appNamespace: appNamespace, region: region, tenant: tenant, tenantParentCustomer: tenantParentCustomer, externalServicesMockMtlsSecuredURL: externalServicesMockMtlsSecuredURL, client: client}
}

func (a *UnassignNotificationsAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	formationID := ctx.Value(operations.FormationIDKey).(string)
	body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)

	notificationsForTarget := gjson.GetBytes(body, a.targetObjectID)
	assignNotificationAboutSource := notificationsForTarget.Array()[0]
	notificationsFoundCount := 0
	for _, notification := range notificationsForTarget.Array() {
		op := notification.Get("Operation").String()
		if op == a.op {
			notificationsFoundCount++
			assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t, assignNotificationAboutSource, assignOperation, formationID, a.sourceObjectID, a.localTenantID, a.appNamespace, a.region, a.tenant, a.tenantParentCustomer, nil)
		}
	}
	require.Equal(t, a.expectedNotificationsCountForOp, notificationsFoundCount, "two notifications for unassign app2 expected")
}
