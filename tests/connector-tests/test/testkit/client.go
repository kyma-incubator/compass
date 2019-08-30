package testkit

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

type ConnectorClient interface {
	GenerateToken(appID string) (schema.Token, error)
	Configuration(token string) (schema.Configuration, error)
	GenerateCert(csr string, token string) (schema.CertificationResult, error)
}

type client struct {
	graphQlClient *gcli.Client
	queryProvider queryProvider
}

func NewConnectorClient(endpoint string) ConnectorClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))

	return &client{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c client) GenerateToken(appID string) (schema.Token, error) {
	query := c.queryProvider.generateToken(appID)
	req := gcli.NewRequest(query)

	var response TokenResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.Token{}, errors.Wrap(err, "Failed to generate token")
	}
	return response.Result, nil
}

func (c client) Configuration(token string) (schema.Configuration, error) {
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

func (c client) GenerateCert(csr string, token string) (schema.CertificationResult, error) {
	query := c.queryProvider.generateCert(csr)
	req := gcli.NewRequest(query)
	req.Header.Add(TokenHeader, token)

	var response CertificationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.CertificationResult{}, errors.Wrap(err, "Failed to generate certificate")
	}
	return response.Result, nil
}

type TokenResponse struct {
	Result schema.Token `json:"result"`
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
