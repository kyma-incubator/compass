package director

import (
	"crypto/tls"
	"net/http"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	gcli "github.com/machinebox/graphql"
)

//go:generate mockery --name=ClientProvider --output=automock --outpkg=automock --case=underscore
type ClientProvider interface {
	Client() Client
}

type clientProvider struct {
	directorURL string
	timeout     time.Duration
}

func NewClientProvider(directorURL string, timeout time.Duration) clientProvider {
	return clientProvider{
		directorURL: directorURL,
		timeout:     timeout,
	}
}

func (cp clientProvider) Client() Client {
	authorizedClient := newAuthorizedHTTPClient(cp.timeout)
	gqlClient := gcli.NewClient(cp.directorURL, gcli.WithHTTPClient(authorizedClient))

	return NewClient(gqlClient)
}

type authenticatedTransport struct {
	http.Transport
}

func newAuthorizedHTTPClient(timeout time.Duration) *http.Client {
	transport := &authenticatedTransport{
		Transport: http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransport(transport)),
		Timeout:   timeout,
	}
}
