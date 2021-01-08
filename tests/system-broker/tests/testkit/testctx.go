package testkit

import (
	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"
)

type TestContext struct {
	SystemBrokerURL string

	InternalConnectorClient     *connector.InternalClient
	ConnectorTokenSecuredClient *connector.TokenSecuredClient
}

func NewTestContext(cfg Config) *TestContext {
	return &TestContext{
		SystemBrokerURL:             cfg.SystemBrokerURL,
		InternalConnectorClient:     connector.NewInternalClient(cfg.InternalConnectorURL),
		ConnectorTokenSecuredClient: connector.NewConnectorClient(cfg.ConnectorURL),
	}
}
