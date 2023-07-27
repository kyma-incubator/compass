package accessstrategy

import (
	"context"
	"fmt"
	"net/http"
)

type openAccessStrategyExecutor struct{}

// NewOpenAccessStrategyExecutor creates a new Executor for the Open Access Strategy
func NewOpenAccessStrategyExecutor() *openAccessStrategyExecutor {
	return &openAccessStrategyExecutor{}
}

// Execute performs the access strategy's specific execution logic
func (*openAccessStrategyExecutor) Execute(_ context.Context, client *http.Client, documentURL, tnt string, additionalHeaders http.Header) (*http.Response, error) {
	req, err := http.NewRequest("GET", documentURL, nil)
	if err != nil {
		return nil, err
	}

	for header := range additionalHeaders {
		req.Header.Set(header, additionalHeaders.Get(header))
	}

	if len(tnt) > 0 {
		req.Header.Set(tenantHeader, tnt)
	}

	for header := range additionalHeaders {
		fmt.Println("ALEX Open Strategy header", header, additionalHeaders.Get(header))
	}

	return client.Do(req)
}
