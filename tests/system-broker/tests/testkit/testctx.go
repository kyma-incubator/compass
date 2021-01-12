package testkit

import (
	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"
)

type TestContext struct {
	Tenant string

	SystemBrokerURL             string
	DirectorURL                 string
	ConnectorTokenSecuredClient *connector.TokenSecuredClient
}

func NewTestContext(cfg Config) *TestContext {
	return &TestContext{
		Tenant:                      cfg.Tenant,
		SystemBrokerURL:             cfg.SystemBrokerURL,
		DirectorURL:                 cfg.DirectorURL,
		ConnectorTokenSecuredClient: connector.NewConnectorClient(cfg.ConnectorURL),
	}
}
