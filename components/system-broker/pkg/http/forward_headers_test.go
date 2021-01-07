package http_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"

	"github.com/stretchr/testify/require"
)

func TestHeaderForwarderForwardsGivenHeaderIfPresent(t *testing.T) {
	const (
		authHeaderKey  = "Authorization"
		authHeaderVal  = "42"
		unsetHeaderKey = "SomeHeader"
		unsetHeaderVal = "some-val"
	)

	var (
		forwardHeaders = map[string]string{authHeaderKey: authHeaderVal, unsetHeaderKey: unsetHeaderVal}
		response       = httptest.NewRecorder()
	)

	forwardHeadersKeys := make([]string, len(forwardHeaders))
	for k := range forwardHeaders {
		forwardHeadersKeys = append(forwardHeadersKeys, k)
	}

	testUrl, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodPost,
		URL:    testUrl,
		Header: map[string][]string{},
	}
	request.Header.Set(authHeaderKey, authHeaderVal)

	handler := httputil.HeaderForwarder(forwardHeadersKeys)
	handler(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		actualHeaders, err := httputil.LoadFromContext(ctx)
		require.NoError(t, err)
		require.Equal(t, len(actualHeaders), 1)

		authHeader, ok := actualHeaders[authHeaderKey]
		require.Equal(t, ok, true)
		require.Equal(t, authHeader, authHeaderVal)

		unsetHeader, ok := actualHeaders[unsetHeaderKey]
		require.Equal(t, ok, false)
		require.Equal(t, unsetHeader, "")

	})).ServeHTTP(response, request)
}
