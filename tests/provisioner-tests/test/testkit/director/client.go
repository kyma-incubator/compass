package director

import (
	"crypto/tls"
	"net/http"

	gcli "github.com/machinebox/graphql"
)

type Client interface {
}

type client struct {
	graphQLClient *gcli.Client
	queryProvider queryProvider
	graphqlizer   graphqlizer
}

func NewDirectorClient(endpoint string) Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))
	return &client{
		graphQLClient: graphQlClient,
		queryProvider: queryProvider{},
		graphqlizer:   graphqlizer{},
	}
}
