package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/require"
)

func TestQueryTenantsPage(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantLen := 5
	getTenantsRequest := fixtures.FixTenantsPageRequest(tenantLen)
	actualTenantPage := graphql.TenantPage{}

	// WHEN
	t.Log("List tenants with page size")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getTenantsRequest, &actualTenantPage)
	require.NoError(t, err)

	//THEN
	assert.Len(t, actualTenantPage.Data, tenantLen)
	assert.True(t, actualTenantPage.PageInfo.HasNextPage)
	assert.NotEmpty(t, actualTenantPage.PageInfo.EndCursor)
}

func TestQueryTenantsSearch(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantSearchTerm := tenant.TestDefaultCustomerTenant
	getTenantsRequest := fixtures.FixTenantsSearchRequest(tenantSearchTerm)
	actualTenantPage := graphql.TenantPage{}

	// WHEN
	t.Log("List tenants with search term")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getTenantsRequest, &actualTenantPage)
	require.NoError(t, err)

	//THEN
	assert.Len(t, actualTenantPage.Data, 1)
}

func TestQueryTenantsPageSearch(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantSearchTerm := "test"
	tenantLen := 3
	getTenantsRequest := fixtures.FixTenantsPageSearchRequest(tenantSearchTerm, tenantLen)
	actualTenantPage := graphql.TenantPage{}

	// WHEN
	t.Log("List tenants with search term and page size")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getTenantsRequest, &actualTenantPage)
	require.NoError(t, err)

	//THEN
	assert.Len(t, actualTenantPage.Data, 3)
	saveExample(t, getTenantsRequest.Query(), "query tenants")
}
