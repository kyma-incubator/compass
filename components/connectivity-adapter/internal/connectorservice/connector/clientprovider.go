package connector

import (
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
)

//go:generate mockery --name=ClientProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ClientProvider interface {
	Client(r *http.Request) Client
}

type connectorClientProvider struct {
	connectorURL string
	timeout      time.Duration
}

func NewClientProvider(connectorURL string, timeout time.Duration) connectorClientProvider {
	return connectorClientProvider{
		connectorURL: connectorURL,
		timeout:      timeout,
	}
}

func (dcp connectorClientProvider) Client(r *http.Request) Client {
	gqlClient := gqlcli.NewAuthorizedGraphQLClient(dcp.connectorURL, dcp.timeout, r)

	return NewClient(gqlClient)
}
