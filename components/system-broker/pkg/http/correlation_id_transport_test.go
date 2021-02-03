package http_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/http/httpfakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrelationIDTransport_RoundTripSetsCorrelationIDIfSuchIsPresent(t *testing.T) {
	const (
		rawUrl             = "http://localhost:8080"
		correlationID      = "test"
		requestIDHeaderKey = "x-request-id"
	)

	transport := &httpfakes.FakeHTTPRoundTripper{}
	transport.RoundTripStub = func(req *http.Request) (*http.Response, error) {
		headers := correlation.HeadersForRequest(req)
		require.Equal(t, correlationID, headers[requestIDHeaderKey])

		ctx := req.Context()
		require.Equal(t, correlation.Headers{"x-request-id": "test"}, ctx.Value("CorrelationHeaders"))

		return nil, nil
	}

	t.Run("sets correlation ID when it is present in the header", func(t *testing.T) {
		testUrl, err := url.Parse(rawUrl)
		require.NoError(t, err)
		request := &http.Request{
			Method: http.MethodPost,
			URL:    testUrl,
			Header: map[string][]string{},
		}
		request.Header.Set(requestIDHeaderKey, correlationID)

		correlationTransport := httputil.NewCorrelationIDTransport(transport)
		_, err = correlationTransport.RoundTrip(request)
		require.NoError(t, err)
	})

	t.Run("sets correlation ID when it is present in the context", func(t *testing.T) {
		testUrl, err := url.Parse(rawUrl)
		require.NoError(t, err)
		request := &http.Request{
			Method: http.MethodPost,
			URL:    testUrl,
			Header: map[string][]string{},
		}

		requestHeaders := correlation.Headers{requestIDHeaderKey: correlationID}
		ctx := context.WithValue(context.TODO(), correlation.HeadersContextKey, requestHeaders)
		request = request.WithContext(ctx)

		correlationTransport := httputil.NewCorrelationIDTransport(transport)
		_, err = correlationTransport.RoundTrip(request)
		assert.NoError(t, err)
	})
}
