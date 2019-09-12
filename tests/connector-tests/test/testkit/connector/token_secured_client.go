package connector

import (
	"context"
	"crypto/tls"
	"net/http"

	schema "github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	TokenHeader = "Connector-Token"
)

type TokenSecuredClient struct {
	graphQlClient *gcli.Client
	queryProvider queryProvider
}

func NewConnectorClient(endpoint string) *TokenSecuredClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))

	return &TokenSecuredClient{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c *TokenSecuredClient) Configuration(token string, headers ...http.Header) (schema.Configuration, error) {
	query := c.queryProvider.configuration()
	req := gcli.NewRequest(query)
	req.Header.Add(TokenHeader, token)

	var response ConfigurationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}
	return response.Result, nil
}

func (c *TokenSecuredClient) SignCSR(csr string, token string, headers ...http.Header) (schema.CertificationResult, error) {
	query := c.queryProvider.signCSR(csr)
	req := gcli.NewRequest(query)
	req.Header.Add(TokenHeader, token)

	var response CertificationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.CertificationResult{}, errors.Wrap(err, "Failed to generate certificate")
	}
	return response.Result, nil
}

type ConfigurationResponse struct {
	Result schema.Configuration `json:"result"`
}

type CertificationResponse struct {
	Result schema.CertificationResult `json:"result"`
}

type RevokeResult struct {
	Result bool `json:"result"`
}
