package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

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
	input := fixtures.FixRuntimeInput("customerRuntime")
	input.Labels[conf.SelfRegDistinguishLabelKey] = []interface{}{conf.SelfRegDistinguishLabelValue}
	input.Labels[RegionLabel] = conf.SelfRegRegion

	customerRuntime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, customerTenant, &input)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, customerTenant, &customerRuntime)
	require.NoError(t, err)
	require.NotEmpty(t, customerRuntime.ID)

	// Assert customer's runtime is visible in the customer's tenant
	customerRuntimes := fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, customerTenant)
	require.Len(t, customerRuntimes.Data, 1)
	require.Equal(t, customerRuntimes.Data[0].ID, customerRuntime.ID)

	// Assert customer's runtime is not visible in his child account tenant
	accountRuntimes := fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, accountTenant)
	require.Len(t, accountRuntimes.Data, 0)

	// Register runtime in account's tenant
	accountRuntimeInput := fixtures.FixRuntimeInput("accountRuntime")
	accountRuntimeInput.Labels[conf.SelfRegDistinguishLabelKey] = []interface{}{conf.SelfRegDistinguishLabelValue}
	accountRuntimeInput.Labels[RegionLabel] = conf.SelfRegRegion
	accountRuntime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, accountTenant, &accountRuntimeInput)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, accountTenant, &accountRuntime)
	require.NoError(t, err)
	require.NotEmpty(t, accountRuntime.ID)

	// Assert account's runtime is visible in the account's tenant
	accountRuntimes = fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, accountTenant)
	require.Len(t, accountRuntimes.Data, 1)
	require.Equal(t, accountRuntimes.Data[0].ID, accountRuntime.ID)

	// Assert both account's runtime and customer's runtime are visible in the customer's tenant
	customerRuntimes = fixtures.ListRuntimes(t, ctx, certSecuredGraphQLClient, customerTenant)
	assertions.AssertRuntimePageContainOnlyIDs(t, customerRuntimes, customerRuntime.ID, accountRuntime.ID)

	// Assert customer can update his own runtime
	customerRuntimeUpdateInput := fixtures.FixRuntimeInput("customerRuntimeUpdated")
	customerRuntime, err = fixtures.UpdateRuntimeWithinTenant(t, ctx, certSecuredGraphQLClient, customerTenant, customerRuntime.ID, customerRuntimeUpdateInput)
	require.NoError(t, err)

	// Assert customer can update his child account's runtime
	accountRuntimeUpdateInput := fixtures.FixRuntimeInput("accountRuntimeUpdated")
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
