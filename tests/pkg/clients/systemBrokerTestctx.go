package clients

import (
	"context"
	"crypto/rsa"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	//"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type SystemBrokerTestContext struct {
	Tenant  string
	Context context.Context

	SystemBrokerURL string
	DirectorURL     string

	ClientKey *rsa.PrivateKey

	ConnectorTokenSecuredClient *TokenSecuredClient
	DexGraphqlClient            *gcli.Client
}

func NewSystemBrokerTestContext(cfg config.SystemBrokerTestConfig) (*SystemBrokerTestContext, error) {
	token, err := idtokenprovider.GetDexToken()
	if err != nil {
		return nil, err
	}

	clientKey, err := certs.GenerateKey()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate private key")
	}

	return &SystemBrokerTestContext{
		Tenant:                      cfg.Tenant,
		Context:                     context.Background(),
		SystemBrokerURL:             cfg.SystemBrokerURL,
		DirectorURL:                 cfg.DirectorURL,
		ClientKey:                   clientKey,
		ConnectorTokenSecuredClient: NewTokenSecuredClient(cfg.ConnectorURL),
		DexGraphqlClient:            gql.NewAuthorizedGraphQLClientWithCustomURL(token, cfg.DirectorURL),
	}, nil
}
