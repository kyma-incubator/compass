package healthz_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/healthz"

	"github.com/stretchr/testify/require"
)

func TestNewLivenessHandler(t *testing.T) {
	t.Run("should return 200 with ok inside response body", func(t *testing.T) {
		// GIVEN
		req, err := http.NewRequest("GET", "/healthz", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(healthz.NewLivenessHandler())

		// WHEN
		handler.ServeHTTP(rr, req)
		// THEN
		require.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestNewReadinessHandler(t *testing.T) {
	t.Run("should return 200 with ok inside response body", func(t *testing.T) {
		// GIVEN
		req, err := http.NewRequest("GET", "/readyz", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(healthz.NewReadinessHandler())

		// WHEN
		handler.ServeHTTP(rr, req)
		// THEN
		require.Equal(t, http.StatusOK, rr.Code)
	})
}
