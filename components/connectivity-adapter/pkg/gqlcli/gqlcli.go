package gqlcli

import (
	"crypto/tls"
	"net/http"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	gcli "github.com/machinebox/graphql"
)

const AuthorizationHeaderKey = "Authorization"

func NewAuthorizedGraphQLClient(url string, timeout time.Duration, rq *http.Request) *gcli.Client {
	authorizationHeaderValue := rq.Header.Get(AuthorizationHeaderKey)
	authorizedClient := newAuthorizedHTTPClient(authorizationHeaderValue, timeout)
	return gcli.NewClient(url, gcli.WithHTTPClient(authorizedClient))
}

type authenticatedTransport struct {
	*http.Transport
	authorizationHeaderValue string
}

func newAuthorizedHTTPClient(authorizationHeaderValue string, timeout time.Duration) *http.Client {
	transport := &authenticatedTransport{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		authorizationHeaderValue: authorizationHeaderValue,
	}

	return &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransport(transport)),
		Timeout:   timeout,
	}
}

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", t.authorizationHeaderValue)
	return t.Transport.RoundTrip(req)
}

func (t *authenticatedTransport) Clone() httputil.HTTPRoundTripper {
	return &authenticatedTransport{
		Transport:                t.Transport.Clone(),
		authorizationHeaderValue: t.authorizationHeaderValue,
	}
}

func (t *authenticatedTransport) GetTransport() *http.Transport {
	return t.Transport
}
