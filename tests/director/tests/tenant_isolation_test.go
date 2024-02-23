package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/claims"
	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantIsolation(t *testing.T) {
	ctx := context.Background()

	actualApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "tenantseparation", tenant.TestTenants.GetDefaultTenantID())
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	customTenant := tenant.TestTenants.GetIDByName(t, tenant.TenantSeparationTenantName)
	anotherTenantsApps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, customTenant)

	assert.Empty(t, anotherTenantsApps.Data)
}

func TestTenantIsolationWithMultipleUsernameAuthenticators(t *testing.T) {
	ctx := context.Background()

	accountTokenURL, err := token.ChangeSubdomain(conf.UsernameAuthCfg.Account.TokenURL, conf.UsernameAuthCfg.Account.Subdomain, conf.UsernameAuthCfg.Account.OAuthTokenPath)
	require.NoError(t, err)
	require.NotEmpty(t, accountTokenURL)

	subaccountTokenURL, err := token.ChangeSubdomain(conf.UsernameAuthCfg.Subaccount.TokenURL, conf.UsernameAuthCfg.Subaccount.Subdomain, conf.UsernameAuthCfg.Subaccount.OAuthTokenPath)
	require.NoError(t, err)
	require.NotEmpty(t, subaccountTokenURL)

	// The accountToken is JWT token containing claim with account ID for tenant. In local setup that's 'ApplicationsForRuntimeTenantName'
	accountToken := token.GetUserToken(t, ctx, accountTokenURL, conf.UsernameAuthCfg.Account.ClientID, conf.UsernameAuthCfg.Account.ClientSecret, conf.BasicUsername, conf.BasicPassword, claims.AccountAuthenticatorClaimKey)
	// The subaccountToken is JWT token containing claim with subaccount ID for tenant. In local setup that's 'TestTenantSubstitutionSubaccount2' test tenant, and it has 'customerId' label with value external tenant ID of 'ApplicationsForRuntimeTenantName'
	subaccountToken := token.GetUserToken(t, ctx, subaccountTokenURL, conf.UsernameAuthCfg.Subaccount.ClientID, conf.UsernameAuthCfg.Subaccount.ClientSecret, conf.BasicUsername, conf.BasicPassword, claims.SubaccountAuthenticatorClaimKey)

	accountGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accountToken, conf.DirectorUserNameAuthenticatorURL)
	subaccountGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(subaccountToken, conf.DirectorUserNameAuthenticatorURL)

	testCases := []struct {
		name               string
		graphqlClient      *gcli.Client
		isSubaccountTenant bool
	}{
		{
			name:          "with account token",
			graphqlClient: accountGraphQLClient,
		},
		{
			name:               "with subaccount token",
			graphqlClient:      subaccountGraphQLClient,
			isSubaccountTenant: true,
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			// the tenant will be derived from the token part of the graphql client
			app, err := fixtures.RegisterApplication(t, ctx, ts.graphqlClient, "e2e-user-auth-app", "")
			defer fixtures.CleanupApplication(t, ctx, ts.graphqlClient, "", &app)
			require.NoError(t, err)
			require.NotEmpty(t, app.ID)

			accountResp := fixtures.GetApplicationPageMinimal(t, ctx, accountGraphQLClient, "")
			subaccountResp := fixtures.GetApplicationPageMinimal(t, ctx, subaccountGraphQLClient, "")

			require.True(t, assertions.DoesAppExistsInAppPageData(app.ID, accountResp))
			require.True(t, assertions.DoesAppExistsInAppPageData(app.ID, subaccountResp))
		})
	}
}

func TestHierarchicalTenantIsolation(t *testing.T) {
	ctx := context.Background()

	customerTenant := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant)
	accountTenant := tenant.TestTenants.GetDefaultTenantID()

	// Register app in customer's tenant
	customerApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "customerApp", customerTenant)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, customerTenant, &customerApp)
	require.NoError(t, err)
	require.NotEmpty(t, customerApp.ID)

	// Assert customer's app is visible in the customer's tenant
	customerApps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, customerTenant)
	require.Len(t, customerApps.Data, 1)
	require.Equal(t, customerApps.Data[0].ID, customerApp.ID)

	// Assert customer's app is not visible in his child account tenant
	accountApps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, accountTenant)
	require.Len(t, accountApps.Data, 0)

	// Register app in account's tenant
	accountApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "accountApp", accountTenant)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, accountTenant, &accountApp)
	require.NoError(t, err)
	require.NotEmpty(t, accountApp.ID)

	// Assert account's app is visible in the account's tenant
	accountApps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, accountTenant)
	require.Len(t, accountApps.Data, 1)
	require.Equal(t, accountApps.Data[0].ID, accountApp.ID)

	// Assert both account's app and customer's app are visible in the customer's tenant
	customerApps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, customerTenant)
	assertions.AssertApplicationPageContainOnlyIDs(t, customerApps, customerApp.ID, accountApp.ID)

	// Assert customer can update his own application
	customerAppUpdateInput := fixtures.FixSampleApplicationUpdateInput("customerAppUpdated")
	customerApp, err = fixtures.UpdateApplicationWithinTenant(t, ctx, certSecuredGraphQLClient, customerTenant, customerApp.ID, customerAppUpdateInput)
	require.NoError(t, err)

	// Assert customer can update his child account's application
	accountAppUpdateInput := fixtures.FixSampleApplicationUpdateInput("accountAppUpdated")
	accountApp, err = fixtures.UpdateApplicationWithinTenant(t, ctx, certSecuredGraphQLClient, customerTenant, accountApp.ID, accountAppUpdateInput)
	require.NoError(t, err)

	// Assert customer can add bundle to his own application
	customerBundle := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, customerTenant, customerApp.ID, "newCustomerBundle")
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, customerTenant, customerBundle.ID)

	// Assert customer can add bundle to his child account's application
	accountBundle := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, customerTenant, accountApp.ID, "newCustomerBundleInAccountsApp")
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, customerTenant, accountBundle.ID)
}

func TestHierarchicalTenantIsolationRuntimeAndRuntimeContext(t *testing.T) {
	ctx := context.Background()

	customerTenant := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant)
	accountTenant := tenant.TestTenants.GetDefaultTenantID()

	// Register runtime in customer's tenant
	input := fixtures.FixRuntimeRegisterInputWithoutLabels("customerRuntime")

	var customerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, customerTenant, &customerRuntime)
	customerRuntime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, customerTenant, input, conf.GatewayOauth)

	// Assert customer's runtime is visible in the customer's tenant
	customerRuntimes := fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, customerTenant)
	require.Len(t, customerRuntimes.Data, 1)
	require.Equal(t, customerRuntimes.Data[0].ID, customerRuntime.ID)

	// Assert customer's runtime is not visible in his child account tenant
	accountRuntimes := fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, accountTenant)
	require.Len(t, accountRuntimes.Data, 0)

	// Register runtime in account's tenant
	accountRuntimeInput := fixtures.FixRuntimeRegisterInputWithoutLabels("accountRuntime")
	var accountRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, accountTenant, &accountRuntime)
	accountRuntime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, accountTenant, accountRuntimeInput, conf.GatewayOauth)

	// Assert account's runtime is visible in the account's tenant
	accountRuntimes = fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, accountTenant)
	require.Len(t, accountRuntimes.Data, 1)
	require.Equal(t, accountRuntimes.Data[0].ID, accountRuntime.ID)

	// Assert both account's runtime and customer's runtime are visible in the customer's tenant
	customerRuntimes = fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, customerTenant)
	assertions.AssertRuntimePageContainOnlyIDs(t, customerRuntimes, customerRuntime.ID, accountRuntime.ID)

	// Assert customer can update his own runtime
	customerRuntimeUpdateInput := fixtures.FixRuntimeUpdateInputWithoutLabels("customerRuntimeUpdated")
	customerRuntime, err := fixtures.UpdateRuntimeWithinTenant(t, ctx, certSecuredGraphQLClient, customerTenant, customerRuntime.ID, customerRuntimeUpdateInput)
	require.NoError(t, err)

	// Assert customer can update his child account's runtime
	accountRuntimeUpdateInput := fixtures.FixRuntimeUpdateInputWithoutLabels("accountRuntimeUpdated")
	accountRuntime, err = fixtures.UpdateRuntimeWithinTenant(t, ctx, certSecuredGraphQLClient, customerTenant, accountRuntime.ID, accountRuntimeUpdateInput)
	require.NoError(t, err)

	// Assert customer can add runtimeContext to his own runtime
	customerRuntimeContextForCustomerRuntime := fixtures.CreateRuntimeContext(t, ctx, certSecuredGraphQLClient, customerTenant, customerRuntime.ID, "key", "CreatedByCustomer")
	defer fixtures.DeleteRuntimeContext(t, ctx, certSecuredGraphQLClient, customerTenant, customerRuntimeContextForCustomerRuntime.ID)

	// Assert customer can add runtimeContext to his child account's runtime
	customerRuntimeContextForAccountRuntime := fixtures.CreateRuntimeContext(t, ctx, certSecuredGraphQLClient, customerTenant, accountRuntime.ID, "key", "CreatedByCustomer")
	defer fixtures.DeleteRuntimeContext(t, ctx, certSecuredGraphQLClient, customerTenant, customerRuntimeContextForAccountRuntime.ID)

	// Assert runtimeContext added to customer runtime is visible in the customer's tenant
	customerRuntime = fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, customerTenant, customerRuntime.ID)
	require.Equal(t, 1, len(customerRuntime.RuntimeContexts.Data))
	require.Equal(t, &customerRuntimeContextForCustomerRuntime, customerRuntime.RuntimeContexts.Data[0])

	// Assert runtimeContext added to account runtime is visible in the customer's tenant
	accountRuntime = fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, customerTenant, accountRuntime.ID)
	require.Equal(t, 1, len(accountRuntime.RuntimeContexts.Data))
	require.Equal(t, &customerRuntimeContextForAccountRuntime, accountRuntime.RuntimeContexts.Data[0])

	// Assert account can add runtimeContext to his runtime
	accountRuntimeContextForAccountRuntime := fixtures.CreateRuntimeContext(t, ctx, certSecuredGraphQLClient, accountTenant, accountRuntime.ID, "key", "CreatedByAccount")
	defer fixtures.DeleteRuntimeContext(t, ctx, certSecuredGraphQLClient, customerTenant, accountRuntimeContextForAccountRuntime.ID)

	// Assert only runtimeContext added to account's runtime by account is visible in the customer's tenant
	accountRuntime = fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, accountTenant, accountRuntime.ID)
	require.Equal(t, 1, len(accountRuntime.RuntimeContexts.Data))
	require.Equal(t, &accountRuntimeContextForAccountRuntime, accountRuntime.RuntimeContexts.Data[0])

	// Assert both runtimeContext added to account's runtime by account and customer are visible in the customer's tenant
	accountRuntime = fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, customerTenant, accountRuntime.ID)
	require.Equal(t, 2, len(accountRuntime.RuntimeContexts.Data))
	require.ElementsMatch(t, []*graphql.RuntimeContextExt{&customerRuntimeContextForAccountRuntime, &accountRuntimeContextForAccountRuntime}, accountRuntime.RuntimeContexts.Data)
}

func TestTenantAccess(t *testing.T) {
	ctx := context.Background()

	testProvider := "e2e-test-provider"
	testLicenseType := "LICENSETYPE"

	customerExternalTenant := "customer-external-tenant"
	customerName := "customer-name"
	customerSubdomain := "customer-subdomain"

	orgExternalTenant := "org-external-tenant"
	orgName := "org-name"
	orgSubdomain := "org-subdomain"

	folder1ExternalTenant := "folder-1-external-tenant"
	folder1Name := "folder-1-name"
	folder1Subdomain := "folder-1-subdomain"

	folder2ExternalTenant := "folder-2-external-tenant"
	folder2Name := "folder-2-name"
	folder2Subdomain := "folder-2-subdomain"

	folder3ExternalTenant := "folder-3-external-tenant"
	folder3Name := "folder-3-name"
	folder3Subdomain := "folder-3-subdomain"

	resourceGroup1ExternalTenant := "resource-group-1-external-tenant"
	resourceGroup1Name := "resource-group-1-name"
	resourceGroup1Subdomain := "resource-group-1-subdomain"

	resourceGroup2ExternalTenant := "resource-group-2-external-tenant"
	resourceGroup2Name := "resource-2-group-name"
	resourceGroup2Subdomain := "resource-2-group-subdomain"

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
			Name:           orgName,
			ExternalTenant: orgExternalTenant,
			Parents:        []*string{&customerExternalTenant},
			Subdomain:      &orgSubdomain,
			Region:         &region,
			Type:           string(tenant.Folder),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           folder1Name,
			ExternalTenant: folder1ExternalTenant,
			Parents:        []*string{&orgExternalTenant},
			Subdomain:      &folder1Subdomain,
			Region:         &region,
			Type:           string(tenant.Folder),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           folder2Name,
			ExternalTenant: folder2ExternalTenant,
			Parents:        []*string{&folder1ExternalTenant},
			Subdomain:      &folder2Subdomain,
			Region:         &region,
			Type:           string(tenant.Folder),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           folder3Name,
			ExternalTenant: folder3ExternalTenant,
			Parents:        []*string{&folder1ExternalTenant},
			Subdomain:      &folder3Subdomain,
			Region:         &region,
			Type:           string(tenant.Folder),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           resourceGroup1Name,
			ExternalTenant: resourceGroup1ExternalTenant,
			Parents:        []*string{&folder2ExternalTenant},
			Subdomain:      &resourceGroup1Subdomain,
			Region:         &region,
			Type:           string(tenant.ResourceGroup),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           resourceGroup2Name,
			ExternalTenant: resourceGroup2ExternalTenant,
			Parents:        []*string{&folder3ExternalTenant},
			Subdomain:      &resourceGroup2Subdomain,
			Region:         &region,
			Type:           string(tenant.ResourceGroup),
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

	actualApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "e2e-tenant-access", resourceGroup1ExternalTenant)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, resourceGroup1ExternalTenant, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	folder3Apps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, folder3ExternalTenant)
	assert.Empty(t, folder3Apps.Data)

	resourceGroup2Apps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, resourceGroup2ExternalTenant)
	assert.Empty(t, resourceGroup2Apps.Data)

	resourceGroup1Apps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, resourceGroup1ExternalTenant)
	require.Len(t, resourceGroup1Apps.Data, 1)
	require.Equal(t, resourceGroup1Apps.Data[0].ID, actualApp.ID)

	folder2Apps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, folder2ExternalTenant)
	require.Len(t, folder2Apps.Data, 1)
	require.Equal(t, folder2Apps.Data[0].ID, actualApp.ID)

	folder1Apps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, folder1ExternalTenant)
	require.Len(t, folder1Apps.Data, 1)
	require.Equal(t, folder1Apps.Data[0].ID, actualApp.ID)

	orgApps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, orgExternalTenant)
	require.Len(t, orgApps.Data, 1)
	require.Equal(t, orgApps.Data[0].ID, actualApp.ID)

	customerApps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, customerExternalTenant)
	require.Len(t, customerApps.Data, 1)
	require.Equal(t, customerApps.Data[0].ID, actualApp.ID)

	in := graphql.TenantAccessInput{
		TenantID:     resourceGroup2ExternalTenant,
		ResourceType: graphql.TenantAccessObjectTypeApplication,
		ResourceID:   actualApp.ID,
		Owner:        true,
	}

	tenantAccessInputString, err := testctx.Tc.Graphqlizer.TenantAccessInputToGQL(in)
	require.NoError(t, err)

	addTenantAccessRequest := fixtures.FixAddTenantAccessRequest(tenantAccessInputString)
	example.SaveExample(t, addTenantAccessRequest.Query(), "add tenant access")

	tenantAccess := &graphql.TenantAccess{}
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, addTenantAccessRequest, tenantAccess)
	require.NoError(t, err)

	folder3Apps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, folder3ExternalTenant)
	require.Len(t, folder3Apps.Data, 1)
	require.Equal(t, folder3Apps.Data[0].ID, actualApp.ID)

	resourceGroup2Apps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, resourceGroup2ExternalTenant)
	require.Len(t, resourceGroup2Apps.Data, 1)
	require.Equal(t, resourceGroup2Apps.Data[0].ID, actualApp.ID)

	resourceGroup1Apps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, resourceGroup1ExternalTenant)
	require.Len(t, resourceGroup1Apps.Data, 1)
	require.Equal(t, resourceGroup1Apps.Data[0].ID, actualApp.ID)

	folder2Apps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, folder2ExternalTenant)
	require.Len(t, folder2Apps.Data, 1)
	require.Equal(t, folder2Apps.Data[0].ID, actualApp.ID)

	folder1Apps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, folder1ExternalTenant)
	require.Len(t, folder1Apps.Data, 1)
	require.Equal(t, folder1Apps.Data[0].ID, actualApp.ID)

	orgApps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, orgExternalTenant)
	require.Len(t, orgApps.Data, 1)
	require.Equal(t, orgApps.Data[0].ID, actualApp.ID)

	customerApps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, customerExternalTenant)
	require.Len(t, customerApps.Data, 1)
	require.Equal(t, customerApps.Data[0].ID, actualApp.ID)

	removeTenantAccessRequest := fixtures.FixRemoveTenantAccessRequest(resourceGroup2ExternalTenant, actualApp.ID, graphql.TenantAccessObjectTypeApplication)
	example.SaveExample(t, removeTenantAccessRequest.Query(), "remove tenant access")

	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, removeTenantAccessRequest, tenantAccess)
	require.NoError(t, err)

	folder3Apps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, folder3ExternalTenant)
	assert.Empty(t, folder3Apps.Data)

	resourceGroup2Apps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, resourceGroup2ExternalTenant)
	assert.Empty(t, resourceGroup2Apps.Data)

	resourceGroup1Apps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, resourceGroup1ExternalTenant)
	require.Len(t, resourceGroup1Apps.Data, 1)
	require.Equal(t, resourceGroup1Apps.Data[0].ID, actualApp.ID)

	folder2Apps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, folder2ExternalTenant)
	require.Len(t, folder2Apps.Data, 1)
	require.Equal(t, folder2Apps.Data[0].ID, actualApp.ID)

	folder1Apps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, folder1ExternalTenant)
	require.Len(t, folder1Apps.Data, 1)
	require.Equal(t, folder1Apps.Data[0].ID, actualApp.ID)

	orgApps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, orgExternalTenant)
	require.Len(t, orgApps.Data, 1)
	require.Equal(t, orgApps.Data[0].ID, actualApp.ID)

	customerApps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, customerExternalTenant)
	require.Len(t, customerApps.Data, 1)
	require.Equal(t, customerApps.Data[0].ID, actualApp.ID)
}

func TestSubstituteCaller(t *testing.T) {
	ctx := context.Background()
	substitutionTenant := tenant.TestTenants.GetDefaultTenantID()

	actualApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "e2e-test-substitution-app", substitutionTenant)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, substitutionTenant, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	tenantsApps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetIDByName(t, tenant.TestTenantSubstitutionAccount))
	assert.Empty(t, tenantsApps.Data)

	// The 'TestTenantSubstitutionSubaccount' tenant substitute 'testDefaultSubaccountTenant' that has 'testDefaultTenant' as parent.
	// That's why when we call with 'TestTenantSubstitutionSubaccount' we see the apps registered in the 'testDefaultTenant'
	tenantsApps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetIDByName(t, tenant.TestTenantSubstitutionSubaccount))
	require.Len(t, tenantsApps.Data, 1)
	require.Equal(t, tenantsApps.Data[0].ID, actualApp.ID)
}
