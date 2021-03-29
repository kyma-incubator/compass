package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	intSysInput := graphql.IntegrationSystemInput{Name: name}
	intSys, err := testctx.Tc.Graphqlizer.IntegrationSystemInputToGQL(intSysInput)
	require.NoError(t, err)

	registerIntegrationSystemRequest := fixtures.FixRegisterIntegrationSystemRequest(intSys)
	output := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Register integration system")

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerIntegrationSystemRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, output.ID)

	//THEN
	require.NotEmpty(t, output.Name)
	saveExample(t, registerIntegrationSystemRequest.Query(), "register integration system")

	t.Log("Check if Integration System was registered")

	getIntegrationSystemRequest := fixtures.FixGetIntegrationSystemRequest(output.ID)
	intSysOutput := graphql.IntegrationSystemExt{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, getIntegrationSystemRequest, &intSysOutput)

	require.NotEmpty(t, intSysOutput)
	assertions.AssertIntegrationSystem(t, intSysInput, intSysOutput)
	saveExample(t, getIntegrationSystemRequest.Query(), "query integration system")
}

func TestUpdateIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "int-system"
	newName := "new-int-system"
	newDescription := "new description"
	t.Log("Register integration system")
	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, name)

	intSysInput := graphql.IntegrationSystemInput{Name: newName, Description: &newDescription}
	intSysGQL, err := testctx.Tc.Graphqlizer.IntegrationSystemInputToGQL(intSysInput)
	updateIntegrationSystemRequest := fixtures.FixUpdateIntegrationSystemRequest(intSys.ID, intSysGQL)
	updateOutput := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Update integration system")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateIntegrationSystemRequest, &updateOutput)
	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, updateOutput.ID)

	//THEN
	t.Log("Check if Integration System was updated")
	assertions.AssertIntegrationSystem(t, intSysInput, updateOutput)
	saveExample(t, updateIntegrationSystemRequest.Query(), "update integration system")
}

func TestUnregisterIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Register integration system")
	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, name)

	unregisterIntegrationSystemRequest := fixtures.FixUnregisterIntegrationSystem(intSys.ID)
	deleteOutput := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Unregister integration system")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, unregisterIntegrationSystemRequest, &deleteOutput)
	require.NoError(t, err)

	//THEN
	t.Log("Check if Integration System was deleted")

	out := fixtures.GetIntegrationSystem(t, ctx, dexGraphQLClient, intSys.ID)

	require.Empty(t, out)
	saveExample(t, unregisterIntegrationSystemRequest.Query(), "unregister integration system")
}

func TestQueryIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Register integration system")
	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, name)
	getIntegrationSystemRequest := fixtures.FixGetIntegrationSystemRequest(intSys.ID)
	output := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Get integration system")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getIntegrationSystemRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, output.ID)

	//THEN
	t.Log("Check if Integration System was received")
	assert.Equal(t, name, output.Name)
}

func TestQueryIntegrationSystems(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name1 := "int-system-1"
	name2 := "int-system-2"

	t.Log("Register integration systems")
	intSys1 := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, name1)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys1.ID)

	intSys2 := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, name2)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys2.ID)

	first := 100
	after := ""

	getIntegrationSystemsRequest := fixtures.FixGetIntegrationSystemsRequestWithPagination(first, after)
	output := graphql.IntegrationSystemPageExt{}

	// WHEN
	t.Log("List integration systems")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getIntegrationSystemsRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if Integration Systems were received")
	assertions.AssertIntegrationSystemNames(t, []string{name1, name2}, output)
	saveExample(t, getIntegrationSystemsRequest.Query(), "query integration systems")
}
