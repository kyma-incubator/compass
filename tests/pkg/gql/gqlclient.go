package gql

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	gcli "github.com/machinebox/graphql"
)

func NewAuthorizedGraphQLClient(bearerToken string) *gcli.Client {
	return NewAuthorizedGraphQLClientWithCustomURL(bearerToken, GetDirectorGraphQLURL())
}

func NewAuthorizedGraphQLClientWithCustomURL(bearerToken, url string) *gcli.Client {
	authorizedClient := NewAuthorizedHTTPClient(bearerToken)
	return gcli.NewClient(url, gcli.WithHTTPClient(authorizedClient))
}

func GetDirectorGraphQLURL() string {
	return GetDirectorURL() + "/graphql"
}

func GetDirectorURL() string {
	url := os.Getenv("DIRECTOR_URL")
	if url == "" {
		url = "http://127.0.0.1:3000"
	}
	return url
}

type authenticatedTransport struct {
	http.Transport
	token string
}

func NewAuthorizedHTTPClient(bearerToken string) *http.Client {
	transport := &authenticatedTransport{
		Transport: http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		token: bearerToken,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30,
	}
}

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	return t.Transport.RoundTrip(req)
}
