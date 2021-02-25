package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	getTenantsRequest := fixtures.FixTenantsRequest()
	var output []*graphql.Tenant
	expectedTenants := expectedTenants()

	t.Log("Initializing one of the tenants")
	initializedTenantID := pkg.TestTenants.GetIDByName(t, pkg.TenantsQueryInitializedTenantName)
	unregisterApp := fixtures.RegisterSimpleApp(t,ctx,dexGraphQLClient, initializedTenantID)
	defer unregisterApp()

	// WHEN
	t.Log("List tenant")
	err = testctx.Tc.RunOperation(ctx,dexGraphQLClient, getTenantsRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if tenants were received")

	assertions.AssertTenants(t, expectedTenants, output)
	saveExample(t, getTenantsRequest.Query(), "query tenants")
}



func expectedTenants() []*graphql.Tenant {
	testTnts := pkg.TestTenants.List()
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
	case pkg.TenantsQueryInitializedTenantName:
		return &trueVal
	case pkg.TenantsQueryNotInitializedTenantName:
		return &falseVal
	}

	return nil
}
