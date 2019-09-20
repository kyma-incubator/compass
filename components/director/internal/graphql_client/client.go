package graphql_client

import (
	"crypto/tls"
	"net/http"
	"time"

	gcli "github.com/machinebox/graphql"
)

func NewGraphQLClient(URL string) *gcli.Client {
	return gcli.NewClient(URL, gcli.WithHTTPClient(newAuthorizedHTTPClient()))
}

func newAuthorizedHTTPClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30,
	}
}
