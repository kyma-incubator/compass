package connector

import (
	"context"
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

func NewClient(gqlClient *graphql.Client) Client {
	return client{
		gqlAPIClient: gqlClient,
		timeout:      30 * time.Second,
	}
}

//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	Configuration(ctx context.Context, headers map[string]string) (schema.Configuration, apperrors.AppError)
	SignCSR(ctx context.Context, csr string, headers map[string]string) (schema.CertificationResult, apperrors.AppError)
	Revoke(ctx context.Context, headers map[string]string) apperrors.AppError
}

func (c client) Configuration(ctx context.Context, headers map[string]string) (schema.Configuration, apperrors.AppError) {
	query := c.queryProvider.configuration()

	var response ConfigurationResponse

	err := c.executeExternal(ctx, headers, query, &response)
	if err != nil {
		return schema.Configuration{}, toAppError(err)
	}

	return response.Result, nil
}

func (c client) SignCSR(ctx context.Context, csr string, headers map[string]string) (schema.CertificationResult, apperrors.AppError) {
	query := c.queryProvider.signCSR(csr)

	var response CertificateResponse

	err := c.executeExternal(ctx, headers, query, &response)
	if err != nil {
		return schema.CertificationResult{}, toAppError(err)
	}

	return response.Result, nil
}

func (c client) Revoke(ctx context.Context, headers map[string]string) apperrors.AppError {
	query := c.queryProvider.revoke()

	var response RevokeResponse

	err := c.executeExternal(ctx, headers, query, response)

	return toAppError(err)
}

func (c *client) executeExternal(ctx context.Context, headers map[string]string, query string, res interface{}) error {
	return c.execute(ctx, c.gqlAPIClient, headers, query, res)
}

func (c *client) execute(ctx context.Context, client *graphql.Client, headers map[string]string, query string, res interface{}) error {
	req := graphql.NewRequest(query)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	newCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return retry.GQLRun(client.Run, newCtx, req, res)
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
