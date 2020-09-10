package readiness

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHTTPHandler(t *testing.T) {
	testFunc := func(statusCode int, body string, notificationCh chan struct{}, shouldNotify bool) {
		var isCacheLoaded = NewAtomicBool(false)
		req, err := http.NewRequest("GET", "/readiness", nil)
		require.NoError(t, err)

		handler := http.HandlerFunc(NewHTTPHandler(logrus.StandardLogger(), isCacheLoaded, notificationCh))

		if shouldNotify {
			notificationCh <- struct{}{}
			// wait to process notification
			for !isCacheLoaded.getValue() {

			}
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		require.Equal(t, statusCode, rr.Code)
		require.Equal(t, body, rr.Body.String())
	}

	t.Run("should return 200 with ok inside response body when cache is ready", func(t *testing.T) {
		notificationCh := make(chan struct{}, 1)
		testFunc(http.StatusOK, "ok", notificationCh, true)
	})

	t.Run("should return 503 with Service Unavailable inside response body when cache is not ready", func(t *testing.T) {
		notificationCh := make(chan struct{}, 1)
		testFunc(http.StatusServiceUnavailable, "Service Unavailable", notificationCh, false)
	})
}
