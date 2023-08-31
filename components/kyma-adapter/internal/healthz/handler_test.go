package healthz_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/healthz"

	"github.com/stretchr/testify/require"
)

func TestNewHTTPHandler(t *testing.T) {
	t.Run("should return 200 with ok inside response body", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/healthz", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(healthz.NewHTTPHandler())

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "ok", rr.Body.String())
	})
}
