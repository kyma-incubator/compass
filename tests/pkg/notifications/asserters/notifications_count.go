package asserters

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type NotificationsCountAsserter struct {
	expectedNotificationsCount         int
	op                                 string
	targetObjectID                     string
	externalServicesMockMtlsSecuredURL string
	client                             *http.Client
}

func NewNotificationsCountAsserter(expectedNotificationsCount int, op string, targetObjectID, externalServicesMockMtlsSecuredURL string, client *http.Client) *NotificationsCountAsserter {
	return &NotificationsCountAsserter{
		expectedNotificationsCount:         expectedNotificationsCount,
		op:                                 op,
		targetObjectID:                     targetObjectID,
		externalServicesMockMtlsSecuredURL: externalServicesMockMtlsSecuredURL,
		client:                             client,
	}
}

func (nca *NotificationsCountAsserter) AssertExpectations(t *testing.T, _ context.Context) {
	body := getNotificationsFromExternalSvcMock(t, nca.client, nca.externalServicesMockMtlsSecuredURL)

	notificationsForTarget := gjson.GetBytes(body, nca.targetObjectID)
	notificationsAboutSource := notificationsForTarget.Array()
	nca.assertAtLeastNNotificationsOfTypeReceived(t, notificationsAboutSource)
	t.Logf("Successfully asserted assignment notification count for: %s", nca.targetObjectID)
}

func (nca *NotificationsCountAsserter) assertAtLeastNNotificationsOfTypeReceived(t *testing.T, notifications []gjson.Result) {
	notificationsForOperationCount := 0
	for _, notification := range notifications {
		if notification.Get("Operation").String() == nca.op {
			notificationsForOperationCount++
		}
	}
	require.LessOrEqual(t, nca.expectedNotificationsCount, notificationsForOperationCount, fmt.Sprintf("Mismatched assignment notification count for tenant: %s", nca.targetObjectID))
}
