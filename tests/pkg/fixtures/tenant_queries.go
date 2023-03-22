package fixtures

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
)

const (
	defaultCtxTimeout = 10 * time.Second
)

type TenantsResponse struct {
	Result *graphql.TenantPage `json:"result"`
}

type TenantResponse struct {
	Result *graphql.Tenant `json:"result"`
}

func GetTenants(gqlClient *gcli.Client) (*graphql.TenantPage, error) {
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

func FixAddTenantAccessRequest(tenantAccessInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addTenantAccess(in: %s) {
					%s
				}
			}`,
			tenantAccessInput, testctx.Tc.GQLFieldsProvider.ForTenantAccess()))
}

func FixRemoveTenantAccessRequest(tenantID, resourceID string, resourceType graphql.TenantAccessObjectType) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: removeTenantAccess(
 				tenantID: "%s"
      			resourceID: "%s"
      			resourceType: %s
			){
					%s
				}
			}`,
			tenantID, resourceID, resourceType, testctx.Tc.GQLFieldsProvider.ForTenantAccess()))
}
