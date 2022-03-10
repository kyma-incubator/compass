package tests

import (
	"context"
	"testing"

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
