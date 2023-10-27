package tests

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/claims"
	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
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
	providerAccountID := conf.TestProviderAccountID
	providerSubaccountID := conf.TestProviderSubaccountID

	// The accountToken is JWT token containing claim with provider account ID for tenant. In local setup that's 'testDefaultTenant'
	accountToken := token.GetUserToken(t, ctx, conf.UsernameAuthCfg.Account.TokenURL+conf.UsernameAuthCfg.Account.OAuthTokenPath, conf.UsernameAuthCfg.Account.ClientID, conf.UsernameAuthCfg.Account.ClientSecret, conf.BasicUsername, conf.BasicPassword, claims.AccountAuthenticatorClaimKey)
	// The subaccountToken is JWT token containing claim with provider subaccount ID for tenant. In local setup that's 'TestProviderSubaccount'
	subaccountToken := token.GetUserToken(t, ctx, conf.UsernameAuthCfg.Subaccount.TokenURL+conf.UsernameAuthCfg.Subaccount.OAuthTokenPath, conf.UsernameAuthCfg.Subaccount.ClientID, conf.UsernameAuthCfg.Subaccount.ClientSecret, conf.BasicUsername, conf.BasicPassword, claims.SubaccountAuthenticatorClaimKey)

	accountGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accountToken, conf.DirectorUserNameAuthenticatorURL)
	subaccountGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(subaccountToken, conf.DirectorUserNameAuthenticatorURL)

	testCases := []struct {
		name              string
		graphqlClient     *gcli.Client
		isSubaacountLevel bool
	}{
		{
			name:          "With account token",
			graphqlClient: accountGraphQLClient,
		},
		{
			name:              "With subaccount token",
			graphqlClient:     subaccountGraphQLClient,
			isSubaacountLevel: true,
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			accountApp, err := fixtures.RegisterApplication(t, ctx, ts.graphqlClient, "e2e-provider-account-app", providerAccountID)
			defer fixtures.CleanupApplication(t, ctx, ts.graphqlClient, providerAccountID, &accountApp)
			require.NoError(t, err)
			require.NotEmpty(t, accountApp.ID)

			subaccountApp, err := fixtures.RegisterApplication(t, ctx, ts.graphqlClient, "e2e-provider-subaccount-app", providerSubaccountID)
			defer fixtures.CleanupApplication(t, ctx, ts.graphqlClient, providerSubaccountID, &subaccountApp)
			require.NoError(t, err)
			require.NotEmpty(t, subaccountApp.ID)

			req := fixtures.FixApplicationsPageableRequest(5, "")
			var resp graphql.ApplicationPageExt
			err = testctx.Tc.RunOperationWithoutTenant(ctx, ts.graphqlClient, req, &resp)
			require.NoError(t, err)

			spew.Dump(resp)
			time.Sleep(30*time.Second)

			require.Equal(t, 1, resp.TotalCount)
			require.Len(t, resp.Data, 1)
			if ts.isSubaacountLevel {
				require.Equal(t, subaccountApp.ID, resp.Data[0].ID)
				require.Equal(t, subaccountApp.Name, resp.Data[0].Name)
			} else {
				require.Equal(t, accountApp.ID, resp.Data[0].ID)
				require.Equal(t, accountApp.Name, resp.Data[0].Name)
			}
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

	actualApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "e2e-tenant-access", tenant.TestTenants.GetDefaultTenantID())
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	customTenant := tenant.TestTenants.GetIDByName(t, tenant.TenantSeparationTenantName)
	anotherTenantsApps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, customTenant)
	assert.Empty(t, anotherTenantsApps.Data)

	in := graphql.TenantAccessInput{
		TenantID:     customTenant,
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

	anotherTenantsApps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, customTenant)
	require.Len(t, anotherTenantsApps.Data, 1)
	require.Equal(t, anotherTenantsApps.Data[0].ID, actualApp.ID)

	removeTenantAccessRequest := fixtures.FixRemoveTenantAccessRequest(customTenant, actualApp.ID, graphql.TenantAccessObjectTypeApplication)
	example.SaveExample(t, removeTenantAccessRequest.Query(), "remove tenant access")

	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, removeTenantAccessRequest, tenantAccess)
	require.NoError(t, err)

	anotherTenantsApps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, customTenant)
	assert.Empty(t, anotherTenantsApps.Data)
}

func TestSubstituteCaller(t *testing.T) {
	ctx := context.Background()

	substitutionTenant := tenant.TestTenants.GetDefaultSubaccountTenantID()

	actualApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "e2e-tenant-access", substitutionTenant) // todo::: change name + tenant var names
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	anotherTenantsApps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetIDByName(t, tenant.TestTenantSubstitutionAccount))
	assert.Empty(t, anotherTenantsApps.Data)

	anotherTenantsApps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetIDByName(t, tenant.TestTenantSubstitutionSubaccount))
	require.Len(t, anotherTenantsApps.Data, 1)
	require.Equal(t, anotherTenantsApps.Data[0].ID, actualApp.ID)
}
