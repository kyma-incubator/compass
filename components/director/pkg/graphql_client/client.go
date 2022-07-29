package graphqlclient

import (
	"crypto/tls"
	"net/http"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	gcli "github.com/machinebox/graphql"
)

// NewGraphQLClient missing godoc
func NewGraphQLClient(url string, timeout time.Duration) *gcli.Client {
	return gcli.NewClient(url, gcli.WithHTTPClient(newAuthorizedHTTPClient(timeout)))
}

func newAuthorizedHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(transport)),
		Timeout:   timeout,
	}
}
