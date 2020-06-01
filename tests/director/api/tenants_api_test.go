package api

import (
	"context"
	"testing"

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

	getTenantsRequest := fixTenantsRequest()
	var output []*graphql.Tenant
	expectedTenants := expectedTenants()

	t.Log("Initializing one of the tenants")
	initializedTenantID := testTenants.GetIDByName(t, tenantsQueryInitializedTenantName)
	unregisterApp := registerSimpleApp(t, initializedTenantID)
	defer unregisterApp()

	// WHEN
	t.Log("List tenant")
	err := tc.RunOperationWithoutTenant(ctx, getTenantsRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if tenants were received")

	assertTenants(t, expectedTenants, output)
	saveExample(t, getTenantsRequest.Query(), "query tenants")
}

func registerSimpleApp(t *testing.T, tenantID string) func() {
	ctx := context.Background()

	in := fixSampleApplicationRegisterInput("foo")
	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	var res graphql.Application
	req := fixRegisterApplicationRequest(appInputGQL)
	err = tc.RunOperationWithCustomTenant(ctx, tenantID, req, &res)
	require.NoError(t, err)

	return func() { unregisterApplicationInTenant(t, res.ID, tenantID) }
}

func expectedTenants() []*graphql.Tenant {
	testTnts := testTenants.List()
	var expectedTenants []*graphql.Tenant

	for _, tnt := range testTnts {
		name := tnt.Name
		expectedTenants = append(expectedTenants, &graphql.Tenant{
			ID:          tnt.ID,
			Name:        &name,
			Initialized: expectedInitializedFieldForTenant(name),
		})
	}

	return expectedTenants
}

func expectedInitializedFieldForTenant(name string) *bool {
	switch name {
	case tenantsQueryInitializedTenantName:
		return &trueVal
	case tenantsQueryNotInitializedTenantName:
		return &falseVal
	}

	return nil
}
