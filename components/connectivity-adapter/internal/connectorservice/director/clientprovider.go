package director

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
)

//go:generate mockery -name=ClientProvider -output=automock -outpkg=automock -case=underscore
type ClientProvider interface {
	Client(r *http.Request) Client
}

type directorClientProvider struct {
	directorURL string
}

func NewClientProvider(directorURL string) directorClientProvider {
	return directorClientProvider{
		directorURL: directorURL,
	}
}

func (dcp directorClientProvider) Client(r *http.Request) Client {
	gqlClient := gqlcli.NewAuthorizedGraphQLClient(dcp.directorURL, r)

	return NewClient(gqlClient)
}
