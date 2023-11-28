package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/director/tests/example"
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
	t.Logf("List tenants with page size: %d", tenantLen)
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getTenantsRequest, &actualTenantPage)
	require.NoError(t, err)

	//THEN
	require.NotNil(t, actualTenantPage)
	require.NotNil(t, actualTenantPage.Data)
	require.NotNil(t, actualTenantPage.PageInfo)
	require.Len(t, actualTenantPage.Data, tenantLen)
	require.True(t, actualTenantPage.PageInfo.HasNextPage)
	require.NotEmpty(t, actualTenantPage.PageInfo.EndCursor)
}

func TestQueryTenantsSearch(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantSearchTerm := tenant.TestDefaultCustomerTenant
	getTenantsRequest := fixtures.FixTenantsSearchRequest(tenantSearchTerm)
	actualTenantPage := graphql.TenantPage{}

	// WHEN
	t.Log("List tenants with search term")
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getTenantsRequest, &actualTenantPage)
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
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getTenantsRequest, &actualTenantPage)
	require.NoError(t, err)

	//THEN
	assert.Len(t, actualTenantPage.Data, 3)
	example.SaveExample(t, getTenantsRequest.Query(), "query tenants")
}

func TestQueryRootTenants(t *testing.T) {
	ctx := context.TODO()

	testProvider := "e2e-test-provider"
	testLicenseType := "LICENSETYPE"

	customerExternalTenant := "customer-external-tenant"
	customerName := "customer-name"
	customerSubdomain := "customer-subdomain"

	accountExternalTenant := "account-external-tenant"
	accountName := "account-name"
	accountSubdomain := "account-subdomain"

	subaccountNames := []string{"subaccount-name", "subaccount-name-2"}
	subaccountExternalTenants := []string{"subaccount-external-tenant", "subaccount-external-tenant-2"}
	subaccountRegion := "test"
	subaccountSubdomain := "sub1"

	region := "local"

	tenants := []graphql.BusinessTenantMappingInput{
		{
			Name:           customerName,
			ExternalTenant: customerExternalTenant,
			Parents:        []*string{},
			Subdomain:      &customerSubdomain,
			Region:         &region,
			Type:           string(tenant.Customer),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           accountName,
			ExternalTenant: accountExternalTenant,
			Parents:        []*string{&customerExternalTenant},
			Subdomain:      &accountSubdomain,
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           subaccountNames[0],
			ExternalTenant: subaccountExternalTenants[0],
			Parents:        []*string{&accountExternalTenant},
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Subaccount),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           subaccountNames[1],
			ExternalTenant: subaccountExternalTenants[1],
			Parents:        []*string{&accountExternalTenant},
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Subaccount),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
	}
	err := fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
	assert.NoError(t, err)
	defer func() { // cleanup tenants
		err := fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, tenants)
		assert.NoError(t, err)
		log.D().Info("Successfully cleanup tenants")
	}()

	// assert the top parent for subaccount 1
	var actualRootTenantForSubaccount1 graphql.Tenant
	getRootTenant := fixtures.FixRootTenantRequest(subaccountExternalTenants[0]) // TODO fixme
	t.Logf("Query root tenant for external tenant: %q", subaccountExternalTenants[0])

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantForSubaccount1)
	require.NoError(t, err)
	require.Equal(t, customerExternalTenant, actualRootTenantForSubaccount1.ID)
	require.Equal(t, customerName, *actualRootTenantForSubaccount1.Name)
	require.Equal(t, string(tenant.Customer), actualRootTenantForSubaccount1.Type)
	example.SaveExample(t, getRootTenant.Query(), "get root tenant")

	// assert the top parent for subaccount 2
	var actualRootTenantForSubaccount2 graphql.Tenant
	getRootTenant = fixtures.FixRootTenantRequest(subaccountExternalTenants[1])
	t.Logf("Query root tenant for external tenant: %q", subaccountExternalTenants[1])

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantForSubaccount2)
	require.NoError(t, err)

	require.Equal(t, customerExternalTenant, actualRootTenantForSubaccount2.ID)
	require.Equal(t, customerName, *actualRootTenantForSubaccount2.Name)
	require.Equal(t, string(tenant.Customer), actualRootTenantForSubaccount2.Type)

	// assert the top parent for account
	var actualRootTenantForAccount graphql.Tenant
	getRootTenant = fixtures.FixRootTenantRequest(accountExternalTenant)
	t.Logf("Query root tenant for external tenant: %q", accountExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantForAccount)
	require.NoError(t, err)
	require.Equal(t, customerExternalTenant, actualRootTenantForAccount.ID)
	require.Equal(t, customerName, *actualRootTenantForAccount.Name)
	require.Equal(t, string(tenant.Customer), actualRootTenantForAccount.Type)

	// assert the top parent for customer
	var actualRootTenantForCustomer graphql.Tenant
	getRootTenant = fixtures.FixRootTenantRequest(customerExternalTenant)
	t.Logf("Query root tenant for external tenant: %q", customerExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantForCustomer)
	require.NoError(t, err)
	require.Equal(t, customerExternalTenant, actualRootTenantForCustomer.ID)
	require.Equal(t, customerName, *actualRootTenantForCustomer.Name)
	require.Equal(t, string(tenant.Customer), actualRootTenantForCustomer.Type)
}
