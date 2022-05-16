package director

import (
	"crypto/tls"
	"net/http"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	gcli "github.com/machinebox/graphql"
)

//go:generate mockery --name=ClientProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ClientProvider interface {
	Client() Client
}

type DirectorClientProvider struct {
	directorURL       string
	timeout           time.Duration
	skipSSLValidation bool
}

func NewClientProvider(directorURL string, timeout time.Duration, skipSSLValidation bool) DirectorClientProvider {
	return DirectorClientProvider{
		directorURL:       directorURL,
		timeout:           timeout,
		skipSSLValidation: skipSSLValidation,
	}
}

func (cp DirectorClientProvider) Client() Client {
	authorizedClient := newAuthorizedHTTPClient(cp.timeout, cp.skipSSLValidation)
	gqlClient := gcli.NewClient(cp.directorURL, gcli.WithHTTPClient(authorizedClient))

	return NewClient(gqlClient)
}

func newAuthorizedHTTPClient(timeout time.Duration, skipSSLValidation bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
	}

	return &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransportWithHeader(tr, "Authorization")),
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
