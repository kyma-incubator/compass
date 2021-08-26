package fixtures

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gqlTools "github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/jwtbuilder"
	gcli "github.com/machinebox/graphql"
)

const (
	tenantScope       = "tenant:read"
	defaultCtxTimeout = 10 * time.Second
)

type TenantsResponse struct {
	Result []*graphql.Tenant `json:"result"`
}

type TenantResponse struct {
	Result *graphql.Tenant `json:"result"`
}

func GetTenants(directorURL string, externalTenantID string) ([]*graphql.Tenant, error) {
	query := FixTenantsRequest().Query()
	req := gcli.NewRequest(query)

	token, err := jwtbuilder.Build(externalTenantID, []string{tenantScope}, &jwtbuilder.Consumer{})
	if err != nil {
		return nil, err
	}

	client := gqlTools.NewAuthorizedGraphQLClientWithCustomURL(token, directorURL)

	var response TenantsResponse
	ctx, cancel := context.WithTimeout(context.Background(), defaultCtxTimeout)
	defer cancel()

	if err = client.Run(ctx, req, &response); err != nil {
		return nil, err
	}

	return response.Result, nil
}

func GetTenantByExternalID(directorURL string, requestTenant, externalTenantID string) (*graphql.Tenant, error) {
	query := FixTenantRequest(externalTenantID).Query()
	req := gcli.NewRequest(query)

	token, err := jwtbuilder.Build(requestTenant, []string{tenantScope}, &jwtbuilder.Consumer{})
	if err != nil {
		return nil, err
	}

	client := gqlTools.NewAuthorizedGraphQLClientWithCustomURL(token, directorURL)

	var response TenantResponse
	ctx, cancel := context.WithTimeout(context.Background(), defaultCtxTimeout)
	defer cancel()

	if err = client.Run(ctx, req, &response); err != nil {
		return nil, err
	}

	return response.Result, nil
}
