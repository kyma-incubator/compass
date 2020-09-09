package readiness

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/apimachinery/pkg/version"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type fakeApiServerClient struct {
	err error
}

func (f *fakeApiServerClient) ServerVersion() (*version.Info, error) {
	return nil, f.err
}

func TestHTTPHandler(t *testing.T) {
	t.Run("should return 200 with ok inside response body when api server responds", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/readiness", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		apiServerClient := &fakeApiServerClient{nil}
		handler := http.HandlerFunc(NewHTTPHandler(logrus.StandardLogger(), apiServerClient))

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "ok", rr.Body.String())
	})

	t.Run("should return 503 with Service Unavailable inside response body when api server does not respond", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/readiness", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		apiServerClient := &fakeApiServerClient{errors.New("error")}
		handler := http.HandlerFunc(NewHTTPHandler(logrus.StandardLogger(), apiServerClient))

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusServiceUnavailable, rr.Code)
		require.Equal(t, "Service Unavailable", rr.Body.String())
	})
}
