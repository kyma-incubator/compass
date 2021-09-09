package http_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/http/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCorrelationIDTransport_RoundTrip(t *testing.T) {
	const (
		testURL            = "http://localhost:8080"
		requestIDHeaderKey = "x-request-id"
		requestID          = "123"
	)

	// GIVEN
	rt := &automock.HTTPRoundTripper{}
	rt.On("RoundTrip", mock.Anything).Return(nil, nil).Run(func(args mock.Arguments) {
		req, ok := args.Get(0).(*http.Request)
		assert.True(t, ok)

		correlationHeaders := correlation.HeadersForRequest(req)
		assert.Equal(t, requestID, correlationHeaders[requestIDHeaderKey])
	})

	t.Run("sets correlation ID when it is present in a header", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodPost, testURL, nil)
		assert.NoError(t, err)
		request.Header.Set(requestIDHeaderKey, requestID)

		correlationTransport := httputil.NewCorrelationIDTransport(rt)
		_, err = correlationTransport.RoundTrip(request)
		assert.NoError(t, err)
	})

	t.Run("sets correlation ID when it is present in the context", func(t *testing.T) {
		requestHeaders := correlation.Headers{requestIDHeaderKey: requestID}
		ctx := context.WithValue(context.TODO(), correlation.HeadersContextKey, requestHeaders)
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
		assert.NoError(t, err)

		correlationTransport := httputil.NewCorrelationIDTransport(rt)
		_, err = correlationTransport.RoundTrip(request)
		assert.NoError(t, err)
	})
}
