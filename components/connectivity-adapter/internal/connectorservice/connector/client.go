package connector

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/retry"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/machinebox/graphql"
)

type client struct {
	queryProvider queryProvider
	gqlAPIClient  *graphql.Client
	timeout       time.Duration
}

func NewClient(compassConnectorAPIURL string, timeout time.Duration) (Client, error) {
	gqlExternalAPIClient := graphql.NewClient(compassConnectorAPIURL, graphql.WithHTTPClient(&http.Client{}))

	client := &client{
		gqlAPIClient: gqlExternalAPIClient,
		timeout:      timeout,
	}

	return client, nil
}

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	Configuration(headers map[string]string) (schema.Configuration, apperrors.AppError)
	SignCSR(csr string, headers map[string]string) (schema.CertificationResult, apperrors.AppError)
	Revoke(headers map[string]string) apperrors.AppError
}

func (c client) Configuration(headers map[string]string) (schema.Configuration, apperrors.AppError) {
	query := c.queryProvider.configuration()

	var response ConfigurationResponse

	err := c.executeExternal(headers, query, &response)
	if err != nil {
		return schema.Configuration{}, toAppError(err)
	}

	return response.Result, nil
}

func (c client) SignCSR(csr string, headers map[string]string) (schema.CertificationResult, apperrors.AppError) {
	query := c.queryProvider.signCSR(csr)

	var response CertificateResponse

	err := c.executeExternal(headers, query, &response)
	if err != nil {
		return schema.CertificationResult{}, toAppError(err)
	}

	return response.Result, nil
}

func (c client) Revoke(headers map[string]string) apperrors.AppError {
	query := c.queryProvider.revoke()

	var response RevokeResponse

	err := c.executeExternal(headers, query, response)

	return toAppError(err)
}

func (c *client) executeExternal(headers map[string]string, query string, res interface{}) error {
	return c.execute(c.gqlAPIClient, headers, query, res)
}

func (c *client) execute(client *graphql.Client, headers map[string]string, query string, res interface{}) error {
	req := graphql.NewRequest(query)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return retry.GQLRun(client.Run, ctx, req, res)
}

type ConfigurationResponse struct {
	Result schema.Configuration `json:"result"`
}

type CertificateResponse struct {
	Result schema.CertificationResult `json:"result"`
}

type RevokeResponse struct {
	Result bool `json:"result"`
}

type TokenResponse struct {
	Result schema.Token `json:"result"`
}
