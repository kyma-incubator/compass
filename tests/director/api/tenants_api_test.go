package api

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryTenants(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	getTenantsRequest := fixTenantsRequest()
	var output []*graphql.Tenant
	defaultTenants := fixDefaultTenants()

	// WHEN
	t.Log("List tenant")
	err := tc.RunOperationWithoutTenant(ctx, getTenantsRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if tenants were received")

	for _, tenant := range defaultTenants {
		assert.Contains(t, output, tenant)
	}
	saveExample(t, getTenantsRequest.Query(), "query tenants")
}

func fixTenant(id, name string) *graphql.Tenant {
	return &graphql.Tenant{
		ID:   id,
		Name: str.Ptr(name),
	}
}

func fixDefaultTenants() []*graphql.Tenant {
	return append([]*graphql.Tenant{
		fixTenant("3e64ebae-38b5-46a0-b1ed-9ccee153a0ae", "default"),
		fixTenant("1eba80dd-8ff6-54ee-be4d-77944d17b10b", "foo"),
		fixTenant("af9f84a9-1d3a-4d9f-ae0c-94f883b33b6e", "bar"),
	}, tenantsToGraphql(fixTestTenants())...)
}
