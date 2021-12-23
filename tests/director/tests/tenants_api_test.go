package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/require"
)

var (
	trueVal  = true
	falseVal = false
)

func TestQueryTenants(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	getTenantsRequest := fixtures.FixTenantsRequest()
	actualTenantPage := graphql.TenantPage{}
	expectedTenants := expectedTenants()

	// WHEN
	t.Log("List tenants")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getTenantsRequest, &actualTenantPage)
	require.NoError(t, err)

	//THEN
	assertions.AssertTenants(t, expectedTenants, actualTenantPage.Data)
}

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

	tenantSearchTerm := "default"
	getTenantsRequest := fixtures.FixTenantsSearchRequest(tenantSearchTerm)
	actualTenantPage := graphql.TenantPage{}

	// WHEN
	t.Log("List tenants with search term")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getTenantsRequest, &actualTenantPage)
	require.NoError(t, err)

	//THEN
	assert.Len(t, actualTenantPage.Data, 3)
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

func expectedTenants() []*graphql.Tenant {
	testTnts := tenant.TestTenants.List()
	var expectedTenants []*graphql.Tenant

	for _, tnt := range testTnts {
		name := tnt.Name
		expectedTenants = append(expectedTenants, &graphql.Tenant{
			ID:   tnt.ExternalTenant,
			Name: &name,
		})
	}

	return expectedTenants
}

func expectedInitializedFieldForTenant(name string) *bool {
	switch name {
	case tenant.TenantsQueryInitializedTenantName:
		return &trueVal
	case tenant.TenantsQueryNotInitializedTenantName:
		return &falseVal
	}

	return nil
}
