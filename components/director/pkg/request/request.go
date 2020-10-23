package request

import (
	"context"
	"io"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	gcli "github.com/machinebox/graphql"
)

func NewHttpRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if id := correlation.IDFromContext(ctx); id != "" {
		req.Header.Set(correlation.RequestIDHeaderKey, id)
	}

	return req, nil
}

func NewGQLRequest(ctx context.Context, query string) *gcli.Request {
	req := gcli.NewRequest(query)
	if id := correlation.IDFromContext(ctx); id != "" {
		req.Header.Set(correlation.RequestIDHeaderKey, id)
	}

	return req
}
