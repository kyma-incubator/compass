package graphql

import (
	"context"
	"crypto/tls"
	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
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
	Configuration(headers map[string]string) (schema.Configuration, error)
	SignCSR(csr string, headers map[string]string) (schema.CertificationResult, error)
}

func (c client) Configuration(headers map[string]string) (schema.Configuration, error) {
	query := c.queryProvider.configuration()

	var response ConfigurationResponse

	err := c.execute(headers, query, &response)
	if err != nil {
		return schema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}

	return response.Result, nil
}

func (c client) SignCSR(csr string, headers map[string]string) (schema.CertificationResult, error) {
	query := c.queryProvider.signCSR(csr)

	var response CertificateResponse

	err := c.execute(headers, query, &response)
	if err != nil {
		return schema.CertificationResult{}, errors.Wrap(err, "Failed to sign csr")
	}

	return response.Result, nil
}

func (c *client) execute(headers map[string]string, query string, res interface{}) error {

	req := graphql.NewRequest(query)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

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
