package asserters

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type NotificationsCountAsyncAsserter struct {
	NotificationsCountAsserter
	timeout time.Duration
	tick    time.Duration
}

func NewNotificationsCountAsyncAsserter(expectedNotificationsCount int, op string, targetObjectID, externalServicesMockMtlsSecuredURL string, client *http.Client) *NotificationsCountAsyncAsserter {
	return &NotificationsCountAsyncAsserter{
		NotificationsCountAsserter: NotificationsCountAsserter{
			expectedNotificationsCount:         expectedNotificationsCount,
			op:                                 op,
			targetObjectID:                     targetObjectID,
			externalServicesMockMtlsSecuredURL: externalServicesMockMtlsSecuredURL,
			client:                             client,
		},
		timeout: eventuallyTimeout,
		tick:    eventuallyTick,
	}
}

func (a *NotificationsCountAsyncAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	require.Eventually(t, func() (isOkay bool) {
		body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)
		notifications := gjson.GetBytes(body, a.targetObjectID)
		if a.expectedNotificationsCount > 0 {
			if !notifications.Exists() {
				return
			}

			if len(notifications.Array()) != a.expectedNotificationsCount {
				return
			}
			require.Len(t, notifications.Array(), a.expectedNotificationsCount)
		} else {
			if notifications.Exists() {
				return
			}
		}
		return true
	}, a.timeout, a.tick)
}
