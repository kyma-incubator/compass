package graphql_client

import (
	"crypto/tls"
	"net/http"
	"time"

	gcli "github.com/machinebox/graphql"
)

func NewGraphQLClient(URL string, timeout time.Duration) *gcli.Client {
	return gcli.NewClient(URL, gcli.WithHTTPClient(newAuthorizedHTTPClient(timeout)))
}

func newAuthorizedHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
