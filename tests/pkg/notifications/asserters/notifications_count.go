package asserters

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
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

func (a *NotificationsCountAsserter) AssertExpectations(t *testing.T, _ context.Context) {
	body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)

	notificationsForTarget := gjson.GetBytes(body, a.targetObjectID)
	notificationsAboutSource := notificationsForTarget.Array()
	assertAtLeastNNotificationsOfTypeReceived(t, notificationsAboutSource, a.op, a.expectedNotificationsCount)
}

func assertAtLeastNNotificationsOfTypeReceived(t *testing.T, notifications []gjson.Result, op string, minCount int) {
	notificationsForOperationCount := 0
	for _, notification := range notifications {
		if notification.Get("Operation").String() == op {
			notificationsForOperationCount++
		}
	}
	require.LessOrEqual(t, minCount, notificationsForOperationCount)
}
