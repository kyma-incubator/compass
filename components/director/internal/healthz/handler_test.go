package healthz

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/healthz/automock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"
)

func TestNewHTTPHandler(t *testing.T) {
	t.Run("should return 200 with ok inside response body", func(t *testing.T) {
		// GIVEN
		mockPinger := &automock.Pinger{}
		defer mockPinger.AssertExpectations(t)
		req, err := http.NewRequest("GET", "/healthz", nil)
		require.NoError(t, err)
		mockPinger.On("PingContext", req.Context()).Return(nil)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(NewLivenessHandler(mockPinger))
		// WHEN
		handler.ServeHTTP(rr, req)
		// THEN
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "ok", rr.Body.String())
	})

	t.Run("should return 500", func(t *testing.T) {
		// GIVEN
		mockPinger := &automock.Pinger{}
		defer mockPinger.AssertExpectations(t)
		req, err := http.NewRequest("GET", "/healthz", nil)
		require.NoError(t, err)
		mockPinger.On("PingContext", req.Context()).Return(errors.New("some error"))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(NewLivenessHandler(mockPinger))
		// WHEN
		handler.ServeHTTP(rr, req)
		// THEN
		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

}
