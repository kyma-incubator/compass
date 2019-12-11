package clients

import (
	"crypto/tls"
	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ClientsProvider
type GQLClientsProvider interface {
	GetDirectorClient(certificate *tls.Certificate, url string, runtimeConfig string) (director.DirectorClient, error)
}

func NewGQLClientsProvider(gqlClientConstr graphql.ClientConstructor, insecureConnectorCommunication bool, insecureConfigFetch bool, enableLogging bool) GQLClientsProvider {
	return &clientsProvider{
		gqlClientConstructor:            gqlClientConstr,
		insecureConnectionCommunication: insecureConnectorCommunication,
		insecureConfigFetch:             insecureConfigFetch,
		enableLogging:                   enableLogging,
	}
}

type clientsProvider struct {
	gqlClientConstructor            graphql.ClientConstructor
	insecureConnectionCommunication bool
	insecureConfigFetch             bool
	enableLogging                   bool
}

func (cp *clientsProvider) GetDirectorClient(certificate *tls.Certificate, url string, runtimeConfig string) (director.Service, error) {
	gqlClient, err := cp.gqlClientConstructor(certificate, url, cp.enableLogging, cp.insecureConfigFetch)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return director.NewDirectorClient(gqlClient), nil
}
