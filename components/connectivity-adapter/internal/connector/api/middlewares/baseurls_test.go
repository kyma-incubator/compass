package middlewares

import (
	"errors"
	mocks "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/api/middlewares/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_BaseUrls(t *testing.T) {

	t.Run("Should extract base paths", func(t *testing.T) {
		// given
		connectivityAdapterBaseUrl := "www.connectivity-adapter.com"
		eventServiceBaseURL := "www.event-service.com"

		runtimeBaseURLProvider := &mocks.RuntimeBaseURLProvider{}
		runtimeBaseURLProvider.On("EventServiceBaseURL").Return(eventServiceBaseURL, nil)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value, err := GetBaseURLsFromContext(r.Context(), BaseURLsKey)
			require.NoError(t, err)

			assert.Equal(t, value.ConnectivityAdapterBaseURL, connectivityAdapterBaseUrl)
			assert.Equal(t, value.EventServiceBaseURL, eventServiceBaseURL)
		})

		r := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", strings.NewReader(""))

		middleware := NewBaseURLsMiddleware(connectivityAdapterBaseUrl, runtimeBaseURLProvider)
		handlerWithMiddleware := middleware.GetBaseUrls(handler)

		// when
		handlerWithMiddleware.ServeHTTP(r, req)

		// then
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("Should return Internal Error when failed to get Events Base URL", func(t *testing.T) {
		// given
		connectivityAdapterBaseUrl := "www.connectivity-adapter.com"

		runtimeBaseURLProvider := &mocks.RuntimeBaseURLProvider{}
		runtimeBaseURLProvider.On("EventServiceBaseURL").Return("", errors.New("failed to get Events Base URL"))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Fail(t, "Handler must not be called")
		})

		r := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", strings.NewReader(""))

		middleware := NewBaseURLsMiddleware(connectivityAdapterBaseUrl, runtimeBaseURLProvider)
		handlerWithMiddleware := middleware.GetBaseUrls(handler)

		// when
		handlerWithMiddleware.ServeHTTP(r, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, r.Code)
	})

}
