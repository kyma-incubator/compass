package directorclient

import (
	"crypto/tls"
	"net/http"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	gcli "github.com/machinebox/graphql"
)

//go:generate mockery --name=ClientProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ClientProvider interface {
	Client() *gcli.Client
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

func (cp DirectorClientProvider) Client() *gcli.Client {
	authorizedClient := newAuthorizedHTTPClient(cp.timeout, cp.skipSSLValidation)
	return gcli.NewClient(cp.directorURL, gcli.WithHTTPClient(authorizedClient))
}

func newAuthorizedHTTPClient(timeout time.Duration, skipSSLValidation bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
	}

	return &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransportWithHeader(httputil.NewHTTPTransportWrapper(tr), "Authorization")),
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
