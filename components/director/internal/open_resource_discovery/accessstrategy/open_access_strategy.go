package accessstrategy

import (
	"context"
	"net/http"
)

type openAccessStrategyExecutor struct{}

// Execute performs the access strategy's specific execution logic
func (*openAccessStrategyExecutor) Execute(_ context.Context, client *http.Client, documentURL string) (*http.Response, error) {
	return client.Get(documentURL)
}
