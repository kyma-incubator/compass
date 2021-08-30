package fixtures

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
)

const (
	defaultCtxTimeout = 10 * time.Second
)

type TenantsResponse struct {
	Result []*graphql.Tenant `json:"result"`
}

type TenantResponse struct {
	Result *graphql.Tenant `json:"result"`
}

func GetTenants(gqlClient *gcli.Client) ([]*graphql.Tenant, error) {
	query := FixTenantsRequest().Query()
	req := gcli.NewRequest(query)

	var response TenantsResponse
	ctx, cancel := context.WithTimeout(context.Background(), defaultCtxTimeout)
	defer cancel()

	if err := gqlClient.Run(ctx, req, &response); err != nil {
		return nil, err
	}

	return response.Result, nil
}

func GetTenantByExternalID(gqlClient *gcli.Client, externalTenantID string) (*graphql.Tenant, error) {
	query := FixTenantRequest(externalTenantID).Query()
	req := gcli.NewRequest(query)

	var response TenantResponse
	ctx, cancel := context.WithTimeout(context.Background(), defaultCtxTimeout)
	defer cancel()

	if err := gqlClient.Run(ctx, req, &response); err != nil {
		return nil, err
	}

	return response.Result, nil
}
