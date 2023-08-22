package accessstrategy

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"net/http"
	"sync"
)

type openAccessStrategyExecutor struct{}

// NewOpenAccessStrategyExecutor creates a new Executor for the Open Access Strategy
func NewOpenAccessStrategyExecutor() *openAccessStrategyExecutor {
	return &openAccessStrategyExecutor{}
}

// Execute performs the access strategy's specific execution logic
func (*openAccessStrategyExecutor) Execute(_ context.Context, client *http.Client, documentURL, tnt string, additionalHeaders *sync.Map) (*http.Response, error) {
	req, err := http.NewRequest("GET", documentURL, nil)
	if err != nil {
		return nil, err
	}

	if additionalHeaders != nil {
		additionalHeaders.Range(func(key, value any) bool {
			req.Header.Set(str.CastOrEmpty(key), str.CastOrEmpty(value))
			return true
		})
	}

	if len(tnt) > 0 {
		req.Header.Set(tenantHeader, tnt)
	}

	return client.Do(req)
}
