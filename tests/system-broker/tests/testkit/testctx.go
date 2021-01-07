package testkit

import "net/http"

type TestContext struct {
	SystemBrokerURL string

	HttpClient *http.Client
}

func NewTestContext(cfg Config) *TestContext {
	return &TestContext{
		SystemBrokerURL: cfg.URL,

		HttpClient: http.DefaultClient,
	}
}
