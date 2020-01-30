package connector

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"

	externalSchema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type client struct {
	queryProvider        queryProvider
	gqlExternalAPIClient *graphql.Client
	gqlInternalAPIClient *graphql.Client
	timeout              time.Duration
}

func NewClient(compassConnectorAPIURL string, compassConnectorInternalAPIURL string, timeout time.Duration) (Client, error) {
	gqlExternalAPIClient := graphql.NewClient(compassConnectorAPIURL, graphql.WithHTTPClient(&http.Client{}))
	gqlInternalAPIClient := graphql.NewClient(compassConnectorInternalAPIURL, graphql.WithHTTPClient(&http.Client{}))

	client := &client{
		gqlExternalAPIClient: gqlExternalAPIClient,
		gqlInternalAPIClient: gqlInternalAPIClient,
		timeout:              timeout,
	}

	return client, nil
}

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	Configuration(headers map[string]string) (externalSchema.Configuration, apperrors.AppError)
	SignCSR(csr string, headers map[string]string) (externalSchema.CertificationResult, apperrors.AppError)
	Revoke(headers map[string]string) apperrors.AppError
	Token(application string) (string, apperrors.AppError)
}

func (c client) Configuration(headers map[string]string) (externalSchema.Configuration, apperrors.AppError) {
	query := c.queryProvider.configuration()

	var response ConfigurationResponse

	err := c.executeExternal(headers, query, &response)
	if err != nil {
		return externalSchema.Configuration{}, toAppError(errors.Wrap(err, "Failed to get configuration"))
	}

	return response.Result, nil
}

func (c client) SignCSR(csr string, headers map[string]string) (externalSchema.CertificationResult, apperrors.AppError) {
	query := c.queryProvider.signCSR(csr)

	var response CertificateResponse

	err := c.executeExternal(headers, query, &response)
	if err != nil {
		return externalSchema.CertificationResult{}, toAppError(errors.Wrap(err, "Failed to sign csr"))
	}

	return response.Result, nil
}

func (c client) Revoke(headers map[string]string) apperrors.AppError {
	query := c.queryProvider.revoke()

	var response RevokeResponse

	err := c.executeExternal(headers, query, response)

	return toAppError(err)
}

func (c client) Token(application string) (string, apperrors.AppError) {
	query := c.queryProvider.token(application)

	var response TokenResponse
	err := c.executeInternal(query, &response)
	if err != nil {
		return "", toAppError(errors.Wrap(err, "Failed to get token"))
	}

	return response.Result.Token, nil
}

func (c *client) executeExternal(headers map[string]string, query string, res interface{}) error {
	return c.execute(c.gqlExternalAPIClient, headers, query, res)
}

func (c *client) executeInternal(query string, res interface{}) error {
	return c.execute(c.gqlInternalAPIClient, map[string]string{}, query, res)
}

func (c *client) execute(client *graphql.Client, headers map[string]string, query string, res interface{}) error {

	req := graphql.NewRequest(query)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	err := client.Run(ctx, req, res)

	return err
}

type ConfigurationResponse struct {
	Result externalSchema.Configuration `json:"result"`
}

type CertificateResponse struct {
	Result externalSchema.CertificationResult `json:"result"`
}

type RevokeResponse struct {
	Result bool `json:"result"`
}

type TokenResponse struct {
	Result externalSchema.Token `json:"result"`
}
