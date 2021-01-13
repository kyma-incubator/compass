package testkit

import (
	"context"
	"crypto/rsa"

	connectorTestkit "github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type TestContext struct {
	Tenant  string
	Context context.Context

	SystemBrokerURL string
	DirectorURL     string

	ClientKey *rsa.PrivateKey

	ConnectorTokenSecuredClient *connector.TokenSecuredClient
	DexGraphqlClient            *gcli.Client
}

func NewTestContext(cfg Config) (*TestContext, error) {
	token, err := idtokenprovider.GetDexToken()
	if err != nil {
		return nil, err
	}

	clientKey, err := connectorTestkit.GenerateKey()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate private key")
	}

	return &TestContext{
		Tenant:                      cfg.Tenant,
		Context:                     context.Background(),
		SystemBrokerURL:             cfg.SystemBrokerURL,
		DirectorURL:                 cfg.DirectorURL,
		ClientKey:                   clientKey,
		ConnectorTokenSecuredClient: connector.NewConnectorClient(cfg.ConnectorURL),
		DexGraphqlClient:            gql.NewAuthorizedGraphQLClientWithCustomURL(token, cfg.DirectorURL),
	}, nil
}
