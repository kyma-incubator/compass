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
	"github.com/stretchr/testify/require"
)

func TestCorrelationIDTransport_RoundTrip(t *testing.T) {
	const (
		testUrl       = "http://localhost:8080"
		correlationID = "123"
	)

	// GIVEN
	rt := &automock.HTTPRoundTripper{}
	rt.On("RoundTrip", mock.Anything).Return(nil, nil).Run(func(args mock.Arguments) {
		req, ok := args.Get(0).(*http.Request)
		require.True(t, ok)

		correlationIDFromRequest := correlation.CorrelationIDForRequest(req)
		require.Equal(t, correlationID, correlationIDFromRequest)
	})

	t.Run("sets correlation ID when it is present in a header", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodPost, testUrl, nil)
		assert.NoError(t, err)
		request.Header.Set("X-Correlation-ID", correlationID)

		correlationTransport := httputil.NewCorrelationIDTransport(rt)
		_, err = correlationTransport.RoundTrip(request)
		assert.NoError(t, err)
	})

	t.Run("sets correlation ID when it is present in the context", func(t *testing.T) {
		ctx := context.WithValue(context.TODO(), correlation.ContextField, correlationID)
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, testUrl, nil)
		assert.NoError(t, err)

		correlationTransport := httputil.NewCorrelationIDTransport(rt)
		_, err = correlationTransport.RoundTrip(request)
		assert.NoError(t, err)
	})
}
