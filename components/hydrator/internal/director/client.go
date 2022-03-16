package director

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/systemauth"
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
	GetSystemAuthByID(ctx context.Context, authID string) (*systemauth.SystemAuth, error)
	GetSystemAuthByToken(ctx context.Context, token string) (*systemauth.SystemAuth, error)
	UpdateSystemAuth(ctx context.Context, sysAuth *systemauth.SystemAuth) (UpdateAuthResult, error)
	InvalidateSystemAuthOneTimeToken(ctx context.Context, authID string) (*systemauth.SystemAuth, error)
	GetRuntimeByTokenIssuer(ctx context.Context, issuer string) (*schema.Runtime, error)
}

type SystemAuhConverter interface {
	GraphQLToModel(in *schema.AppSystemAuth) (*systemauth.SystemAuth, error)
}

type Config struct {
	URL               string        `envconfig:"default=http://127.0.0.1:3000/graphql"`
	ClientTimeout     time.Duration `envconfig:"default=115s"`
	SkipSSLValidation bool          `envconfig:"default=false"`
}

func NewClient(gqlClient *graphql.Client, sysAuthConv SystemAuhConverter) Client {
	return &client{
		gqlClient:   gqlClient,
		timeout:     30 * time.Second,
		sysAuthConv: sysAuthConv,
	}
}

type client struct {
	gqlClient   *graphql.Client
	timeout     time.Duration
	sysAuthConv SystemAuhConverter
}

type TenantResponse struct {
	Result *schema.Tenant `json:"result"`
}

type TenantByLowestOwnerForResourceResponse struct {
	Result string `json:"result"`
}

type SystemAuthResponse1 struct {
	Result *schema.AppSystemAuth `json:"result"`
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

func (c *client) GetSystemAuthByID(ctx context.Context, authID string) (*systemauth.SystemAuth, error) {
	query := SystemAuthQuery(authID)

	var response SystemAuthResponse1

	err := c.execute(ctx, c.gqlClient, query, &response)
	if err != nil {
		return nil, err
	}

	sysAuth, err := c.sysAuthConv.GraphQLToModel(response.Result)
	if err != nil {
		return nil, err
	}
	//if response.Result == nil {
	//	return nil, nil
	//}

	return sysAuth, nil

}

func (c *client) GetSystemAuthByToken(ctx context.Context, token string) (*systemauth.SystemAuth, error) {
	query := SystemAuthByTokenQuery(token)

	var response SystemAuthResponse1

	err := c.execute(ctx, c.gqlClient, query, &response)
	if err != nil {
		return nil, err
	}

	sysAuth, err := c.sysAuthConv.GraphQLToModel(response.Result)
	if err != nil {
		return nil, err
	}

	return sysAuth, nil
}

func (c *client) UpdateSystemAuth(ctx context.Context, sysAuth *systemauth.SystemAuth) (UpdateAuthResult, error) {
	query, err := UpdateSystemAuthQuery(sysAuth)
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

func (c *client) InvalidateSystemAuthOneTimeToken(ctx context.Context, authID string) (*systemauth.SystemAuth, error) {
	query := InvalidateSystemAuthOneTimeTokenQuery(authID)

	var response SystemAuthResponse1

	err := c.execute(ctx, c.gqlClient, query, &response)
	if err != nil {
		return nil, err
	}

	sysAuth, err := c.sysAuthConv.GraphQLToModel(response.Result)
	if err != nil {
		return nil, err
	}

	return sysAuth, nil
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
