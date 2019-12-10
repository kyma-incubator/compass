package clients

import (
	"crypto/tls"
	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ClientsProvider
type ClientsProvider interface {
	GetDirectorClient(certificate *tls.Certificate, url string, runtimeConfig string) (director.DirectorClient, error)
}

func NewClientsProvider(gqlClientConstr graphql.ClientConstructor, insecureConnectorCommunication, insecureConfigFetch, enableLogging bool) ClientsProvider {
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

func (cp *clientsProvider) GetDirectorClient(certificate *tls.Certificate, url string, runtimeConfig string) (director.DirectorClient, error) {
	gqlClient, err := cp.gqlClientConstructor(certificate, url, cp.enableLogging, cp.insecureConfigFetch)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return director.NewConfigurationClient(gqlClient, runtimeConfig), nil
}
