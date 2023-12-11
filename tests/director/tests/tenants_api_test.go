package tests

import (
	"context"
	"k8s.io/utils/strings/slices"
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
	var actualRootTenantsForSubaccount1 []graphql.Tenant
	getRootTenant := fixtures.FixRootTenantsRequest(subaccountExternalTenants[0])
	t.Logf("Query root tenants for external tenant: %q", subaccountExternalTenants[0])

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantsForSubaccount1)
	require.NoError(t, err)
	require.Equal(t, customerExternalTenant, actualRootTenantsForSubaccount1[0].ID)
	require.Equal(t, customerName, *actualRootTenantsForSubaccount1[0].Name)
	require.Equal(t, string(tenant.Customer), actualRootTenantsForSubaccount1[0].Type)
	example.SaveExample(t, getRootTenant.Query(), "get root tenant")

	// assert the top parent for subaccount 2
	var actualRootTenantsForSubaccount2 []graphql.Tenant
	getRootTenant = fixtures.FixRootTenantsRequest(subaccountExternalTenants[1])
	t.Logf("Query root tenant for external tenant: %q", subaccountExternalTenants[1])

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantsForSubaccount2)
	require.NoError(t, err)

	require.Equal(t, customerExternalTenant, actualRootTenantsForSubaccount2[0].ID)
	require.Equal(t, customerName, *actualRootTenantsForSubaccount2[0].Name)
	require.Equal(t, string(tenant.Customer), actualRootTenantsForSubaccount2[0].Type)

	// assert the top parent for account
	var actualRootTenantsForAccount []graphql.Tenant
	getRootTenant = fixtures.FixRootTenantsRequest(accountExternalTenant)
	t.Logf("Query root tenant for external tenant: %q", accountExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantsForAccount)
	require.NoError(t, err)
	require.Equal(t, customerExternalTenant, actualRootTenantsForAccount[0].ID)
	require.Equal(t, customerName, *actualRootTenantsForAccount[0].Name)
	require.Equal(t, string(tenant.Customer), actualRootTenantsForAccount[0].Type)

	// assert the top parent for customer
	var actualRootTenantsForCustomer []graphql.Tenant
	getRootTenant = fixtures.FixRootTenantsRequest(customerExternalTenant)
	t.Logf("Query root tenant for external tenant: %q", customerExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantsForCustomer)
	require.NoError(t, err)
	require.Empty(t, actualRootTenantsForCustomer)
}

func TestWriteTenants(t *testing.T) {
	testProvider := "e2e-test-provider"
	testLicenseType := "LICENSETYPE"

	// customerExternalTenantGUID to have leading zeros, so we can check if they are not accidentally trimmed
	customerExternalTenantGUID := "0022b8c3-bfda-47b4-8d1b-4c717b9940a3"
	customerExternalTenant := "0000customerID"
	customerTrimmedExternalTenant := "customerID"
	customerName := "customer-name"
	customerSubdomain := "customer-subdomain"

	accountExternalTenant := "account-external-tenant-1"
	accountName := "account-name"
	accountSubdomain := "account-subdomain"

	orgExternalTenant := "org-external-tenant-1"
	orgName := "org-name"
	orgSubdomain := "org-subdomain"

	region := "local"

	tenants := []graphql.BusinessTenantMappingInput{
		{
			Name:           customerName,
			ExternalTenant: customerExternalTenant,
			Parents:        nil,
			Subdomain:      &customerSubdomain,
			Region:         &region,
			Type:           string(tenant.Customer),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           customerName,
			ExternalTenant: customerExternalTenantGUID,
			Parents:        nil,
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
			Name:           orgName,
			ExternalTenant: orgExternalTenant,
			Parents:        []*string{&customerExternalTenant},
			Subdomain:      &orgSubdomain,
			Region:         &region,
			Type:           string(tenant.Organization),
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

	var actualCustomer1Tenant graphql.Tenant
	getTenant := fixtures.FixTenantRequest(customerTrimmedExternalTenant)
	t.Logf("Query tenant for external tenant: %q", customerTrimmedExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getTenant, &actualCustomer1Tenant)
	require.NoError(t, err)
	require.Equal(t, customerTrimmedExternalTenant, actualCustomer1Tenant.ID)
	require.Equal(t, customerName, *actualCustomer1Tenant.Name)
	require.Equal(t, string(tenant.Customer), actualCustomer1Tenant.Type)

	var actualCustomer2Tenant graphql.Tenant
	getTenant = fixtures.FixTenantRequest(customerExternalTenantGUID)
	t.Logf("Query tenant for external tenant: %q", customerExternalTenantGUID)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getTenant, &actualCustomer2Tenant)
	require.NoError(t, err)
	require.Equal(t, customerExternalTenantGUID, actualCustomer2Tenant.ID)
	require.Equal(t, customerName, *actualCustomer2Tenant.Name)
	require.Equal(t, string(tenant.Customer), actualCustomer2Tenant.Type)

	var actualAccountTenant graphql.Tenant
	getTenant = fixtures.FixTenantRequest(accountExternalTenant)
	t.Logf("Query tenant for external tenant: %q", accountExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getTenant, &actualAccountTenant)
	require.NoError(t, err)
	require.Equal(t, accountExternalTenant, actualAccountTenant.ID)
	require.Equal(t, accountName, *actualAccountTenant.Name)
	require.Equal(t, string(tenant.Account), actualAccountTenant.Type)
	require.True(t, slices.Contains(actualAccountTenant.Parents, actualCustomer1Tenant.InternalID))

	var actualOrgenant graphql.Tenant
	getTenant = fixtures.FixTenantRequest(orgExternalTenant)
	t.Logf("Query tenant for external tenant: %q", orgExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getTenant, &actualOrgenant)
	require.NoError(t, err)
	require.Equal(t, orgExternalTenant, actualOrgenant.ID)
	require.Equal(t, orgName, *actualOrgenant.Name)
	require.Equal(t, string(tenant.Organization), actualOrgenant.Type)
	require.True(t, slices.Contains(actualOrgenant.Parents, actualCustomer1Tenant.InternalID))
}
