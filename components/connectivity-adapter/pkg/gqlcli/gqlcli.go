package gqlcli

import (
	"crypto/tls"
	gcli "github.com/machinebox/graphql"
	"net/http"
	"time"
)

const AuthorizationHeaderKey = "Authorization"

func NewAuthorizedGraphQLClient(url string, rq *http.Request) *gcli.Client {
	authorizationHeaderValue := rq.Header.Get(AuthorizationHeaderKey)
	authorizedClient := newAuthorizedHTTPClient(authorizationHeaderValue)
	return gcli.NewClient(url, gcli.WithHTTPClient(authorizedClient))
}

type authenticatedTransport struct {
	http.Transport
	authorizationHeaderValue string
}

func newAuthorizedHTTPClient(authorizationHeaderValue string) *http.Client {
	transport := &authenticatedTransport{
		Transport: http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		authorizationHeaderValue: authorizationHeaderValue,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30,
	}
}

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", t.authorizationHeaderValue)
	return t.Transport.RoundTrip(req)
}
