package connector

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/sirupsen/logrus"

	schema "github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	TokenHeader = "Connector-Token"
)

type ConnectorClient struct {
	graphQlClient *gcli.Client
	queryProvider queryProvider
}

func NewConnectorClient(endpoint string) *ConnectorClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))

	graphQlClient.Log = func(s string) {
		logrus.Info(s)
	}

	return &ConnectorClient{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c *ConnectorClient) Configuration(token string, headers ...http.Header) (schema.Configuration, error) {
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

func (c *ConnectorClient) GenerateCert(csr string, token string, headers ...http.Header) (schema.CertificationResult, error) {
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

type ConfigurationResponse struct {
	Result schema.Configuration `json:"result"`
}

type CertificationResponse struct {
	Result schema.CertificationResult `json:"result"`
}

type RevokeResult struct {
	Result bool `json:"result"`
}
