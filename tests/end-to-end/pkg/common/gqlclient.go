package common

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	gcli "github.com/machinebox/graphql"
)

func NewAuthorizedGraphQLClient(bearerToken string) *gcli.Client {
	return NewAuthorizedGraphQLClientWithCustomURL(bearerToken, getDirectorGraphqlURL())
}

func NewAuthorizedGraphQLClientWithCustomURL(bearerToken, url string) *gcli.Client {
	authorizedClient := newAuthorizedHTTPClient(bearerToken)
	return gcli.NewClient(url, gcli.WithHTTPClient(authorizedClient))
}

func getDirectorGraphqlURL() string {
	url := os.Getenv("DIRECTOR_URL")
	if url == "" {
		url = "http://127.0.0.1:3000"
	}
	url = url + "/graphql"
	return url
}

type authenticatedTransport struct {
	http.Transport
	token string
}

func newAuthorizedHTTPClient(bearerToken string) *http.Client {
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

func GetDexGraphQLClient(bearerToken string) *gcli.Client {
	return NewAuthorizedGraphQLClient(bearerToken)
}

func GetOauthGraphQLClient(token string, url string) *gcli.Client {
	gqlClient := NewAuthorizedGraphQLClientWithCustomURL(token, url)
	return gqlClient
}

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	return t.Transport.RoundTrip(req)
}
