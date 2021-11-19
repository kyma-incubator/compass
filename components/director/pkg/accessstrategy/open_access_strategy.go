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
func (*openAccessStrategyExecutor) Execute(_ context.Context, client *http.Client, documentURL string) (*http.Response, error) {
	return client.Get(documentURL)
}
