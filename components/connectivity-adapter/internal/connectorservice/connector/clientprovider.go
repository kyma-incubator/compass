package connector

import (
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
)

const (
	timeout = 30 * time.Second
)

//go:generate mockery -name=ClientProvider -output=automock -outpkg=automock -case=underscore
type ClientProvider interface {
	Client(r *http.Request) Client
}

type connectorClientProvider struct {
	connectorURL string
}

func NewClientProvider(connectorURL string) connectorClientProvider {
	return connectorClientProvider{
		connectorURL: connectorURL,
	}
}

func (dcp connectorClientProvider) Client(r *http.Request) Client {
	gqlClient := gqlcli.NewAuthorizedGraphQLClient(dcp.connectorURL, r)

	return NewClient(gqlClient, timeout)
}
