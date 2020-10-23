package request_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/request"
	"github.com/stretchr/testify/assert"
)

func TestRequest_NewHttpRequest(t *testing.T) {
	t.Run("when the correlation ID is present in the context the request has the header set", func(t *testing.T) {
		expectedHeader := "123"
		ctx := context.WithValue(context.Background(), correlation.ContextField, expectedHeader)

		req, err := request.NewHttpRequest(ctx, http.MethodGet, "/", nil)
		assert.NoError(t, err)

		actualHeader := req.Header.Get(correlation.RequestIDHeaderKey)
		assert.Equal(t, expectedHeader, actualHeader)
	})
	t.Run("when the correlation ID is not present in the context the request does not have the x-request-id header", func(t *testing.T) {
		req, err := request.NewHttpRequest(context.Background(), http.MethodGet, "/", nil)
		assert.NoError(t, err)

		header := req.Header.Get(correlation.RequestIDHeaderKey)
		assert.Empty(t, header)
	})
}

func TestRequest_NewGQLRequest(t *testing.T) {
	t.Run("when the correlation ID is present in the context the request has the x-request-id header set", func(t *testing.T) {
		expectedHeader := "123"
		ctx := context.WithValue(context.Background(), correlation.ContextField, expectedHeader)

		req := request.NewGQLRequest(ctx, "")

		actualHeader := req.Header.Get(correlation.RequestIDHeaderKey)
		assert.Equal(t, expectedHeader, actualHeader)
	})
	t.Run("when the correlation ID is not present in the context the request does not have the x-request-id header", func(t *testing.T) {
		req := request.NewGQLRequest(context.Background(), "")
		header := req.Header.Get(correlation.RequestIDHeaderKey)
		assert.Empty(t, header)
	})
}
