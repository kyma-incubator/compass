package director

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	gcli "github.com/machinebox/graphql"
)

func newAuthorizedGraphQLClient(bearerToken string) *gcli.Client {
	authorizedClient := newAuthorizedHTTPClient(bearerToken)
	return gcli.NewClient(getDirectorGraphqlURL(), gcli.WithHTTPClient(authorizedClient))
}

func getDirectorGraphqlURL() string {
	url := os.Getenv("DIRECTOR_URL")
	if url == "" {
		url = "http://127.0.0.1:3000/graphql"
	} else {
		url = url + "/graphql"
	}
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

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	return t.Transport.RoundTrip(req)
}
