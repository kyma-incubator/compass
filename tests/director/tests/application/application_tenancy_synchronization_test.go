package application

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateTenantAccessForNewApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	in := graphql.ApplicationRegisterInput{
		Name:           "test-atom-application",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			"group": []interface{}{"production", "experimental"},
		},
	}

	resourceGroupTnt := tenant.TestTenants.GetIDByName(t, tenant.TestAtomResourceGroup)

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	// WHEN
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)

	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, resourceGroupTnt, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, resourceGroupTnt, &actualApp)
	require.NoError(t, err)

	//THEN
	require.NotEmpty(t, actualApp)
	require.NotEmpty(t, actualApp.ID)
	assertions.AssertApplication(t, in, actualApp)
	assert.Equal(t, graphql.ApplicationStatusConditionInitial, actualApp.Status.Condition)

	app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)
	assert.Equal(t, actualApp.ID, app.ID)
}

func TestCreateTenantAccessForNewTenants(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	newExternalTenantID := "ga-tenant-multiple"

	in := graphql.ApplicationRegisterInput{
		Name:           "test-atom-application",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			"group": []interface{}{"production", "experimental"},
		},
	}

	parentExternalID := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant)
	tenants := []graphql.BusinessTenantMappingInput{
		{
			Name:           "test-new-tenant",
			ExternalTenant: newExternalTenantID,
			Parents:        []*string{&parentExternalID},
			Type:           string(tenant.Account),
			Provider:       "provide",
		},
	}

	resourceGroupTnt := tenant.TestTenants.GetIDByName(t, tenant.TestAtomResourceGroup)

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	request := fixtures.FixRegisterApplicationRequest(appInputGQL)

	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, resourceGroupTnt, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, resourceGroupTnt, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp)

	// WHEN
	err = fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
	require.NoError(t, err)

	defer func() {
		tenantsToDelete := []graphql.BusinessTenantMappingInput{
			{
				ExternalTenant: newExternalTenantID,
			},
		}
		err := fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, tenantsToDelete)
		assert.NoError(t, err)
	}()

	//THEN
	app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, newExternalTenantID, actualApp.ID)
	assert.Equal(t, actualApp.ID, app.ID)
}

func TestCreateTenantAccessForNewTenant(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	newExternalTenantID := "ga-tenant-single"

	parent, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant))
	assert.NoError(t, err)

	in := graphql.ApplicationRegisterInput{
		Name:           "test-atom-application",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			"group": []interface{}{"production", "experimental"},
		},
	}

	newTenant := graphql.BusinessTenantMappingInput{
		Name:           "test-new-tenant",
		ExternalTenant: newExternalTenantID,
		Parents:        []*string{&parent.InternalID},
		Type:           string(tenant.Account),
		Provider:       "provide",
	}

	resourceGroupTnt := tenant.TestTenants.GetIDByName(t, tenant.TestAtomResourceGroup)

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	request := fixtures.FixRegisterApplicationRequest(appInputGQL)

	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, resourceGroupTnt, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, resourceGroupTnt, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp)

	// WHEN
	err = fixtures.WriteTenant(t, ctx, directorInternalGQLClient, newTenant)
	require.NoError(t, err)

	defer func() {
		tenantsToDelete := []graphql.BusinessTenantMappingInput{
			{
				ExternalTenant: newExternalTenantID,
			},
		}
		err := fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, tenantsToDelete)
		assert.NoError(t, err)
	}()

	//THEN
	app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, newExternalTenantID, actualApp.ID)
	assert.Equal(t, actualApp.ID, app.ID)
}
