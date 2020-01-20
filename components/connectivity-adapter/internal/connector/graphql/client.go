package graphql

import (
	"context"
	"crypto/tls"
	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type client struct {
	queryProvider queryProvider
	gqlClient     *graphql.Client
	timeout       time.Duration
}

func NewClient(graphqlEndpoint string, insecureConfigFetch bool, timeout time.Duration) (Client, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureConfigFetch,
			},
		},
	}

	gqlClient := graphql.NewClient(graphqlEndpoint, graphql.WithHTTPClient(httpClient))

	client := &client{
		gqlClient: gqlClient,
		timeout:   timeout,
	}

	return client, nil
}

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	Configuration(clientID string) (schema.Configuration, error)
	SignCSR(csr string, clientID string) (schema.CertificationResult, error)
}

func (c client) Configuration(clientID string) (schema.Configuration, error) {
	query := c.queryProvider.configuration()

	var response ConfigurationResponse

	err := c.execute(clientID, query, &response)
	if err != nil {
		return schema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}

	return response.Result, nil
}

func (c client) SignCSR(csr string, clientID string) (schema.CertificationResult, error) {
	query := c.queryProvider.signCSR(csr)

	var response CertificateResponse

	err := c.execute(clientID, query, &response)
	if err != nil {
		return schema.CertificationResult{}, errors.Wrap(err, "Failed to sign csr")
	}

	return response.Result, nil
}

func (c *client) execute(clientID string, query string, res interface{}) error {

	req := graphql.NewRequest(query)
	req.Header.Set(oathkeeper.ClientIdFromTokenHeader, clientID)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	err := c.gqlClient.Run(ctx, req, res)

	return err
}

type ConfigurationResponse struct {
	Result schema.Configuration `json:"result"`
}

type CertificateResponse struct {
	Result schema.CertificationResult `json:"result"`
}
