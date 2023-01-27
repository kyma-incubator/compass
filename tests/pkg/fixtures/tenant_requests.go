package fixtures

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
)

func WriteTenants(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenants []graphql.BusinessTenantMappingInput) error {
	req := FixWriteTenantsRequest(t, tenants)
	return gqlClient.Run(ctx, req, nil)
}

func WriteTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant graphql.BusinessTenantMappingInput) error {
	req := FixWriteTenantRequest(t, tenant)
	return gqlClient.Run(ctx, req, nil)
}

func DeleteTenants(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenants []graphql.BusinessTenantMappingInput) error {
	req := FixDeleteTenantsRequest(t, tenants)
	return gqlClient.Run(ctx, req, nil)
}
