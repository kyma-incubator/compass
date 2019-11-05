package connector

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type InternalClient struct {
	graphQlClient *gcli.Client
	queryProvider queryProvider
}

func NewInternalClient(endpoint string) *InternalClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))

	return &InternalClient{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c *InternalClient) GenerateApplicationToken(appID string) (externalschema.Token, error) {
	query := c.queryProvider.generateApplicationToken(appID)
	return c.generateToken(query)
}

func (c *InternalClient) GenerateRuntimeToken(runtimeID string) (externalschema.Token, error) {
	query := c.queryProvider.generateRuntimeToken(runtimeID)
	return c.generateToken(query)
}

func (c *InternalClient) generateToken(query string) (externalschema.Token, error) {
	req := gcli.NewRequest(query)

	var response TokenResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return externalschema.Token{}, errors.Wrap(err, "Failed to generate token")
	}
	return response.Result, nil
}

type TokenResponse struct {
	Result externalschema.Token `json:"result"`
}
