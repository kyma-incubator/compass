package info_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/info"

	"github.com/stretchr/testify/require"
)

func TestNewInfoHandler(t *testing.T) {
	t.Run("should return 500 when cert cache is empty", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		certCache := certloader.NewCertificateCache()

		req, err := http.NewRequest("GET", "/v1/info", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(info.NewInfoHandler(ctx, info.Config{}, certCache))

		// WHEN
		handler.ServeHTTP(rr, req)
		// THEN
		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
