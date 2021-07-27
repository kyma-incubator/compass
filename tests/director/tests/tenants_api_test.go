package tests

import (
	"context"
	"testing"

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
	var output []*graphql.Tenant
	expectedTenants := expectedTenants()

	t.Log("Initializing one of the tenants")
	initializedTenantID := tenant.TestTenants.GetIDByName(t, tenant.TenantsQueryInitializedTenantName)
	unregisterApp := fixtures.RegisterSimpleApp(t, ctx, dexGraphQLClient, initializedTenantID)
	defer unregisterApp()

	// WHEN
	t.Log("List tenants")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getTenantsRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if tenants were received")

	assertions.AssertTenants(t, expectedTenants, output)
	saveExample(t, getTenantsRequest.Query(), "query tenants")
}

func expectedTenants() []*graphql.Tenant {
	testTnts := tenant.TestTenants.List()
	var expectedTenants []*graphql.Tenant

	for _, tnt := range testTnts {
		name := tnt.Name
		expectedTenants = append(expectedTenants, &graphql.Tenant{
			ID:          tnt.ExternalTenant,
			Name:        &name,
			Initialized: expectedInitializedFieldForTenant(name),
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
