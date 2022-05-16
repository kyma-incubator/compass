package director

import (
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
)

//go:generate mockery --name=ClientProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ClientProvider interface {
	Client(r *http.Request) Client
}

type directorClientProvider struct {
	directorURL string
	timeout     time.Duration
}

func NewClientProvider(directorURL string, timeout time.Duration) directorClientProvider {
	return directorClientProvider{
		directorURL: directorURL,
		timeout:     timeout,
	}
}

func (dcp directorClientProvider) Client(r *http.Request) Client {
	gqlClient := gqlcli.NewAuthorizedGraphQLClient(dcp.directorURL, dcp.timeout, r)

	return NewClient(gqlClient)
}
