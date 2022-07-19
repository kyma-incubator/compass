package accessstrategy

import (
	"context"
	"net/http"
)

type openAccessStrategyExecutor struct{}

// NewOpenAccessStrategyExecutor creates a new Executor for the Open Access Strategy
func NewOpenAccessStrategyExecutor() *openAccessStrategyExecutor {
	return &openAccessStrategyExecutor{}
}

// Execute performs the access strategy's specific execution logic
func (*openAccessStrategyExecutor) Execute(_ context.Context, client *http.Client, documentURL, tnt string) (*http.Response, error) {
	req, err := http.NewRequest("GET", documentURL, nil)
	if err != nil {
		return nil, err
	}

	if len(tnt) > 0 {
		req.Header.Set(tenantHeader, tnt)
	}

	return client.Do(req)
}
