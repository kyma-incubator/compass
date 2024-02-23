package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/strings/slices"

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

	tenantSearchTerm := tenant.TestDeleteApplicationIfInScenario
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

func TestQueryRootTenants_CustomerRootTenant(t *testing.T) {
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

func TestQueryRootTenants_CostObjectRootTenant(t *testing.T) {
	ctx := context.TODO()

	testProvider := "e2e-test-provider"
	testLicenseType := "LICENSETYPE"

	costObjectExternalTenant := "cost-object-external-tenant"
	costObjectName := "cost-object-name"
	costObjectSubdomain := "cost-object-subdomain"

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
			Name:           costObjectName,
			ExternalTenant: costObjectExternalTenant,
			Parents:        []*string{},
			Subdomain:      &costObjectSubdomain,
			Region:         &region,
			Type:           string(tenant.CostObject),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           accountName,
			ExternalTenant: accountExternalTenant,
			Parents:        []*string{&costObjectExternalTenant},
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
	require.Equal(t, costObjectExternalTenant, actualRootTenantsForSubaccount1[0].ID)
	require.Equal(t, costObjectName, *actualRootTenantsForSubaccount1[0].Name)
	require.Equal(t, string(tenant.CostObject), actualRootTenantsForSubaccount1[0].Type)

	// assert the top parent for subaccount 2
	var actualRootTenantsForSubaccount2 []graphql.Tenant
	getRootTenant = fixtures.FixRootTenantsRequest(subaccountExternalTenants[1])
	t.Logf("Query root tenant for external tenant: %q", subaccountExternalTenants[1])

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantsForSubaccount2)
	require.NoError(t, err)

	require.Equal(t, costObjectExternalTenant, actualRootTenantsForSubaccount2[0].ID)
	require.Equal(t, costObjectName, *actualRootTenantsForSubaccount2[0].Name)
	require.Equal(t, string(tenant.CostObject), actualRootTenantsForSubaccount2[0].Type)

	// assert the top parent for account
	var actualRootTenantsForAccount []graphql.Tenant
	getRootTenant = fixtures.FixRootTenantsRequest(accountExternalTenant)
	t.Logf("Query root tenant for external tenant: %q", accountExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantsForAccount)
	require.NoError(t, err)
	require.Equal(t, costObjectExternalTenant, actualRootTenantsForAccount[0].ID)
	require.Equal(t, costObjectName, *actualRootTenantsForAccount[0].Name)
	require.Equal(t, string(tenant.CostObject), actualRootTenantsForAccount[0].Type)

	// assert the top parent for cost-object
	var actualRootTenantsForCostObject []graphql.Tenant
	getRootTenant = fixtures.FixRootTenantsRequest(costObjectExternalTenant)
	t.Logf("Query root tenant for external tenant: %q", costObjectExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantsForCostObject)
	require.NoError(t, err)
	require.Empty(t, actualRootTenantsForCostObject)
}

func TestQueryRootTenants_MultipleRootTenants(t *testing.T) {
	ctx := context.TODO()

	testProvider := "e2e-test-provider"
	testLicenseType := "LICENSETYPE"

	costObjectExternalTenant := "cost-object-external-tenant"
	costObjectName := "cost-object-name"
	costObjectSubdomain := "cost-object-subdomain"

	organizationExternalTenant := "organization-external-tenant"
	organizationName := "organization-name"
	organizationSubdomain := "organization-subdomain"

	folderExternalTenant := "folder-external-tenant"
	folderName := "folder-name"
	folderSubdomain := "folder-subdomain"

	resourceGroupExternalTenant := "resource-group-external-tenant"
	resourceGroupName := "resource-group-name"
	resourceGroupSubdomain := "resource-group-subdomain"

	region := "local"

	tenants := []graphql.BusinessTenantMappingInput{
		{
			Name:           costObjectName,
			ExternalTenant: costObjectExternalTenant,
			Parents:        []*string{},
			Subdomain:      &costObjectSubdomain,
			Region:         &region,
			Type:           string(tenant.CostObject),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           organizationName,
			ExternalTenant: organizationExternalTenant,
			Parents:        []*string{},
			Subdomain:      &organizationSubdomain,
			Region:         &region,
			Type:           string(tenant.Organization),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           folderName,
			ExternalTenant: folderExternalTenant,
			Parents:        []*string{&costObjectExternalTenant, &organizationExternalTenant},
			Subdomain:      &folderSubdomain,
			Region:         &region,
			Type:           string(tenant.Folder),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           resourceGroupName,
			ExternalTenant: resourceGroupExternalTenant,
			Parents:        []*string{&folderExternalTenant},
			Subdomain:      &resourceGroupSubdomain,
			Region:         &region,
			Type:           string(tenant.ResourceGroup),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
	}

	expectedRootTenants := []*graphql.Tenant{
		{
			ID:   costObjectExternalTenant,
			Name: &costObjectName,
			Type: "cost-object",
		},
		{
			ID:   organizationExternalTenant,
			Name: &organizationName,
			Type: "organization",
		},
	}

	err := fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
	assert.NoError(t, err)
	defer func() { // cleanup tenants
		err := fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, tenants)
		assert.NoError(t, err)
		log.D().Info("Successfully cleanup tenants")
	}()

	// assert the top parents for resource-group
	var actualRootTenantsForResourceGroup []*graphql.Tenant
	getRootTenant := fixtures.FixRootTenantsRequest(resourceGroupExternalTenant)
	t.Logf("Query root tenants for external tenant: %q", resourceGroupExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantsForResourceGroup)
	require.NoError(t, err)
	require.Equal(t, 2, len(actualRootTenantsForResourceGroup))
	assertions.AssertTenants(t, expectedRootTenants, actualRootTenantsForResourceGroup)

	// assert the top parent for folder
	var actualRootTenantsForFolder []*graphql.Tenant
	getRootTenant = fixtures.FixRootTenantsRequest(folderExternalTenant)
	t.Logf("Query root tenant for external tenant: %q", folderExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantsForFolder)
	require.NoError(t, err)
	require.Equal(t, 2, len(actualRootTenantsForFolder))
	assertions.AssertTenants(t, expectedRootTenants, actualRootTenantsForFolder)

	// assert the top parent for cost-object
	var actualRootTenantsForCostObject []graphql.Tenant
	getRootTenant = fixtures.FixRootTenantsRequest(costObjectExternalTenant)
	t.Logf("Query root tenant for external tenant: %q", costObjectExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantsForCostObject)
	require.NoError(t, err)
	require.Empty(t, actualRootTenantsForCostObject)

	// assert the top parent for organization
	var actualRootTenantsForOrganization []graphql.Tenant
	getRootTenant = fixtures.FixRootTenantsRequest(organizationExternalTenant)
	t.Logf("Query root tenant for external tenant: %q", organizationExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRootTenant, &actualRootTenantsForOrganization)
	require.NoError(t, err)
	require.Empty(t, actualRootTenantsForOrganization)
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

func TestWriteTenantsTenantAccess(t *testing.T) {
	testProvider := "e2e-test-provider"
	testLicenseType := "LICENSETYPE"

	customerExternalTenant := "customerID"
	customerName := "customer-name"

	accountExternalTenant := "account-external-tenant-1"
	accountName := "account-name"
	accountSubdomain := "account-subdomain"
	customerSubdomain := "customer-subdomain"

	region := "local"

	tenants := []graphql.BusinessTenantMappingInput{
		{
			Name:           accountName,
			ExternalTenant: accountExternalTenant,
			Subdomain:      &accountSubdomain,
			Region:         &region,
			Type:           string(tenant.Account),
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

	var actualAccountTenant graphql.Tenant
	getTenant := fixtures.FixTenantRequest(accountExternalTenant)
	t.Logf("Query tenant for external tenant: %q", accountExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getTenant, &actualAccountTenant)
	require.NoError(t, err)
	require.Equal(t, accountExternalTenant, actualAccountTenant.ID)
	require.Equal(t, accountName, *actualAccountTenant.Name)
	require.Equal(t, string(tenant.Account), actualAccountTenant.Type)

	in := graphql.ApplicationRegisterInput{
		Name:           "e2e-test-app",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("application used for e2e tests"),
		HealthCheckURL: ptr.String("http://mye2etest.com/health"),
		Labels: graphql.Labels{
			"group": []interface{}{"e2e-test"},
		},
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	// WHEN
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)

	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, accountExternalTenant, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, accountExternalTenant, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp)
	require.NotEmpty(t, actualApp.ID)
	assertions.AssertApplication(t, in, actualApp)

	tenants = []graphql.BusinessTenantMappingInput{
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
			Name:           accountName,
			ExternalTenant: accountExternalTenant,
			Parents:        []*string{&customerExternalTenant},
			Subdomain:      &accountSubdomain,
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
	}

	err = fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
	assert.NoError(t, err)
	defer func() { // cleanup tenants
		err := fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, tenants)
		assert.NoError(t, err)
		log.D().Info("Successfully cleanup tenants")
	}()

	// Check that the customer has access to the application
	app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, customerExternalTenant, actualApp.ID)
	require.NotEmpty(t, app)
	require.NotEmpty(t, app.ID)
	assertions.AssertApplication(t, in, app)
}

func TestTenantParentUpdateWithCostObject(t *testing.T) {
	// GIVEN
	isolatedAccount := tenant.TestTenants.GetIDByName(t, tenant.TestIsolatedAccountName)
	in := graphql.ApplicationRegisterInput{
		Name:           "e2e-test-app",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("application used for e2e tests"),
		HealthCheckURL: ptr.String("http://mye2etest.com/health"),
		Labels: graphql.Labels{
			"group": []interface{}{"e2e-test"},
		},
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	// WHEN
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)

	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, isolatedAccount, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, isolatedAccount, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp)
	require.NotEmpty(t, actualApp.ID)
	assertions.AssertApplication(t, in, actualApp)

	// Create cost-object tenant
	costObjectExternalTenantID := "cost-obj"
	costObjectTenant := graphql.BusinessTenantMappingInput{
		Name:           "cost-object-name",
		ExternalTenant: costObjectExternalTenantID,
		Parents:        []*string{},
		Type:           string(tenant.CostObject),
	}

	err = fixtures.WriteTenant(t, ctx, directorInternalGQLClient, costObjectTenant)
	defer func() {
		err := fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, []graphql.BusinessTenantMappingInput{costObjectTenant})
		assert.NoError(t, err)
		log.D().Info("Successfully cleanup tenants")
	}()
	assert.NoError(t, err)

	// Update parents of the isolated account to include the cost-object.
	ga, err := fixtures.GetTenantByExternalID(directorInternalGQLClient, isolatedAccount)
	assert.NoError(t, err)
	assert.NotEmpty(t, ga.ID)

	updateAccountInput := graphql.BusinessTenantMappingInput{
		Name:           *ga.Name,
		ExternalTenant: ga.ID,
		Parents:        []*string{&costObjectExternalTenantID},
		Type:           ga.Type,
		Provider:       ga.Provider,
	}
	_, err = fixtures.UpdateTenant(t, ctx, directorInternalGQLClient, ga.InternalID, updateAccountInput)
	assert.NoError(t, err)

	// Check that the cost-object has access to the application
	app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, costObjectExternalTenantID, actualApp.ID)
	require.NotEmpty(t, app)
	require.NotEmpty(t, app.ID)
	assertions.AssertApplication(t, in, app)

	// Update parents of the isolated account - remove the cost-object.
	updateAccountInput = graphql.BusinessTenantMappingInput{
		Name:           *ga.Name,
		ExternalTenant: ga.ID,
		Parents:        []*string{},
		Type:           ga.Type,
		Provider:       ga.Provider,
	}
	_, err = fixtures.UpdateTenant(t, ctx, directorInternalGQLClient, ga.InternalID, updateAccountInput)
	assert.NoError(t, err)

	// Check that the cost-object has no longer access to the application
	app = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, costObjectExternalTenantID, actualApp.ID)
	assert.Empty(t, app)

	// Check that the isolated account still has access to the application
	app = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, isolatedAccount, actualApp.ID)
	require.NotEmpty(t, app)
	require.NotEmpty(t, app.ID)
	assertions.AssertApplication(t, in, app)
}

func TestTenantParentDeleteCostObjectParent(t *testing.T) {
	// GIVEN
	isolatedAccount := tenant.TestTenants.GetIDByName(t, tenant.TestIsolatedAccountName)
	in := graphql.ApplicationRegisterInput{
		Name:           "e2e-test-app",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("application used for e2e tests"),
		HealthCheckURL: ptr.String("http://mye2etest.com/health"),
		Labels: graphql.Labels{
			"group": []interface{}{"e2e-test"},
		},
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	// WHEN
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)

	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, isolatedAccount, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, isolatedAccount, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp)
	require.NotEmpty(t, actualApp.ID)
	assertions.AssertApplication(t, in, actualApp)

	// Create cost-object tenant
	costObjectExternalTenantID := "cost-obj"
	costObjectTenant := graphql.BusinessTenantMappingInput{
		Name:           "cost-object-name",
		ExternalTenant: costObjectExternalTenantID,
		Parents:        []*string{},
		Type:           string(tenant.CostObject),
	}

	err = fixtures.WriteTenant(t, ctx, directorInternalGQLClient, costObjectTenant)
	defer func() {
		err := fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, []graphql.BusinessTenantMappingInput{costObjectTenant})
		assert.NoError(t, err)
		log.D().Info("Successfully cleanup tenants")
	}()
	assert.NoError(t, err)

	// Update parents of the isolated account to include the cost-object.
	ga, err := fixtures.GetTenantByExternalID(directorInternalGQLClient, isolatedAccount)
	assert.NoError(t, err)
	assert.NotEmpty(t, ga.ID)

	updateAccountInput := graphql.BusinessTenantMappingInput{
		Name:           *ga.Name,
		ExternalTenant: ga.ID,
		Parents:        []*string{&costObjectExternalTenantID},
		Type:           ga.Type,
		Provider:       ga.Provider,
	}
	_, err = fixtures.UpdateTenant(t, ctx, directorInternalGQLClient, ga.InternalID, updateAccountInput)
	assert.NoError(t, err)

	// Check that the cost-object has access to the application
	app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, costObjectExternalTenantID, actualApp.ID)
	require.NotEmpty(t, app)
	require.NotEmpty(t, app.ID)
	assertions.AssertApplication(t, in, app)

	// Delete cost-object tenant
	err = fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, []graphql.BusinessTenantMappingInput{costObjectTenant})
	require.NoError(t, err)

	// Get isolated account and check that it has no parents
	ga, err = fixtures.GetTenantByExternalID(directorInternalGQLClient, isolatedAccount)
	assert.NoError(t, err)
	assert.NotEmpty(t, ga.ID)
	assert.Empty(t, ga.Parents)

	// Check that the isolated account still has access to the application
	app = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, isolatedAccount, actualApp.ID)
	require.NotEmpty(t, app)
	require.NotEmpty(t, app.ID)
	assertions.AssertApplication(t, in, app)
}

func TestSetTenantLabel(t *testing.T) {
	testProvider := "e2e-test-provider"
	testLicenseType := "LICENSETYPE"
	region := "local"

	accountExternalTenant := "account-external-tenant"
	accountName := "account-name"
	accountSubdomain := "account-subdomain"

	labelKey := "label_key"
	labelValue := "label_value"

	account := graphql.BusinessTenantMappingInput{
		Name:           accountName,
		ExternalTenant: accountExternalTenant,
		Parents:        []*string{},
		Subdomain:      &accountSubdomain,
		Region:         &region,
		Type:           string(tenant.Account),
		Provider:       testProvider,
		LicenseType:    &testLicenseType,
	}
	err := fixtures.WriteTenant(t, ctx, directorInternalGQLClient, account)
	assert.NoError(t, err)
	defer func() {
		err := fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, []graphql.BusinessTenantMappingInput{account})
		assert.NoError(t, err)
		log.D().Info("Successfully cleanup tenants")
	}()

	var actualAccountTenant graphql.Tenant
	getTenant := fixtures.FixTenantRequest(accountExternalTenant)
	t.Logf("Query tenant for external tenant: %q", accountExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getTenant, &actualAccountTenant)
	require.NoError(t, err)

	// Set tenant Label
	var res map[string]interface{}
	setLabel := fixtures.FixSetTenantLabelRequest(actualAccountTenant.InternalID, labelKey, labelValue)
	example.SaveExample(t, setLabel.Query(), "set tenant label")
	t.Logf("Set tenant label on tenant with external id: %q", accountExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, setLabel, &res)
	require.NoError(t, err)

	// Assert that the label is present
	getTenant = fixtures.FixTenantRequest(accountExternalTenant)
	t.Logf("Query tenant for external tenant: %q and validate that the label is present", accountExternalTenant)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getTenant, &actualAccountTenant)
	require.NoError(t, err)

	val, ok := actualAccountTenant.Labels[labelKey]
	require.True(t, ok)

	value, ok := val.(string)
	require.True(t, ok)
	require.Equal(t, labelValue, value)
}
