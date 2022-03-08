package director

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/retry"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/machinebox/graphql"
)

//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore
type Client interface {
	GetTenantByExternalID(ctx context.Context, tenantID string) (*schema.Tenant, error)
	GetTenantByInternalID(ctx context.Context, tenantID string) (*schema.Tenant, error)
	GetTenantByLowestOwnerForResource(ctx context.Context, resourceID, resourceType string) (string, error)
	GetSystemAuthByID(ctx context.Context, authID string) (schema.SystemAuth, error)
	GetSystemAuthByToken(ctx context.Context, token string) (schema.SystemAuth, error)
	UpdateSystemAuth(ctx context.Context, authID string, auth schema.Auth) (UpdateAuthResult, error)
	InvalidateSystemAuthOneTimeToken(ctx context.Context, authID string) (schema.SystemAuth, error)
	GetRuntimeByTokenIssuer(ctx context.Context, issuer string) (*schema.Runtime, error)
}

type Config struct {
	URL               string        `envconfig:"default=http://127.0.0.1:3000/graphql"`
	ClientTimeout     time.Duration `envconfig:"default=115s"`
	SkipSSLValidation bool          `envconfig:"default=false"`
}

func NewClient(gqlClient *graphql.Client) Client {
	return &client{
		gqlClient: gqlClient,
		timeout:   30 * time.Second,
	}
}

type client struct {
	gqlClient *graphql.Client
	timeout   time.Duration
}

type TenantResponse struct {
	Result *schema.Tenant `json:"result"`
}

type TenantByLowestOwnerForResourceResponse struct {
	Result string `json:"result"`
}

type SystemAuthResponse struct {
	Result schema.SystemAuth `json:"result"`
}

type RuntimeResponse struct {
	Result *schema.Runtime `json:"result"`
}

type UpdateAuthResult struct {
	ID string `json:"id"`
}
type UpdateSystemAuthResponse struct {
	Result UpdateAuthResult `json:"result"`
}

func (c *client) GetTenantByExternalID(ctx context.Context, tenantID string) (*schema.Tenant, error) {
	query := TenantByExternalIDQuery(tenantID)
	var response TenantResponse

	err := c.execute(ctx, c.gqlClient, query, &response)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

func (c *client) GetTenantByInternalID(ctx context.Context, tenantID string) (*schema.Tenant, error) {
	query := TenantByInternalIDQuery(tenantID)
	var response TenantResponse

	err := c.execute(ctx, c.gqlClient, query, &response)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

func (c *client) GetTenantByLowestOwnerForResource(ctx context.Context, resourceID, resourceType string) (string, error) {
	query := TenantByLowestOwnerForResourceQuery(resourceID, resourceType)
	var response TenantByLowestOwnerForResourceResponse

	err := c.execute(ctx, c.gqlClient, query, &response)
	if err != nil {
		return "", err
	}

	return response.Result, nil
}

func (c *client) GetSystemAuthByID(ctx context.Context, authID string) (schema.SystemAuth, error) {
	query := SystemAuthQuery(authID)

	var response SystemAuthResponse

	err := c.execute(ctx, c.gqlClient, query, &response)
	if err != nil {
		return schema.AppSystemAuth{}, err
	}

	return response.Result, nil
}

func (c *client) GetSystemAuthByToken(ctx context.Context, token string) (schema.SystemAuth, error) {
	query := SystemAuthByTokenQuery(token)

	var response SystemAuthResponse

	err := c.execute(ctx, c.gqlClient, query, &response)
	if err != nil {
		return schema.AppSystemAuth{}, err
	}

	return response.Result, nil
}

func (c *client) UpdateSystemAuth(ctx context.Context, authID string, auth schema.Auth) (UpdateAuthResult, error) {
	query, err := UpdateSystemAuthQuery(authID, auth)
	if err != nil {
		return UpdateAuthResult{}, err
	}

	var response UpdateSystemAuthResponse

	err = c.execute(ctx, c.gqlClient, query, &response)
	if err != nil {
		return UpdateAuthResult{}, err
	}

	return response.Result, nil
}

func (c *client) InvalidateSystemAuthOneTimeToken(ctx context.Context, authID string) (schema.SystemAuth, error) {
	query := InvalidateSystemAuthOneTimeTokenQuery(authID)

	var response SystemAuthResponse

	err := c.execute(ctx, c.gqlClient, query, &response)
	if err != nil {
		return schema.AppSystemAuth{}, err
	}

	return response.Result, nil
}

func (c *client) GetRuntimeByTokenIssuer(ctx context.Context, issuer string) (*schema.Runtime, error) {
	query := RuntimeByTokenIssuerQuery(issuer)
	var response RuntimeResponse

	err := c.execute(ctx, c.gqlClient, query, &response)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

func (c *client) execute(ctx context.Context, client *graphql.Client, query string, res interface{}) error {
	req := graphql.NewRequest(query)

	newCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return retry.GQLRun(client.Run, newCtx, req, res)
}
