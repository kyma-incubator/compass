package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	intSysInput := graphql.IntegrationSystemInput{Name: name}
	intSys, err := pkg.Tc.Graphqlizer.IntegrationSystemInputToGQL(intSysInput)
	require.NoError(t, err)

	registerIntegrationSystemRequest := pkg.FixRegisterIntegrationSystemRequest(intSys)
	output := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Register integration system")

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, registerIntegrationSystemRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, output.ID)

	//THEN
	require.NotEmpty(t, output.Name)
	saveExample(t, registerIntegrationSystemRequest.Query(), "register integration system")

	t.Log("Check if Integration System was registered")

	getIntegrationSystemRequest := pkg.FixGetIntegrationSystemRequest(output.ID)
	intSysOutput := graphql.IntegrationSystemExt{}

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, getIntegrationSystemRequest, &intSysOutput)

	require.NotEmpty(t, intSysOutput)
	assertIntegrationSystem(t, intSysInput, intSysOutput)
	saveExample(t, getIntegrationSystemRequest.Query(), "query integration system")
}

func TestUpdateIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "int-system"
	newName := "new-int-system"
	newDescription := "new description"
	t.Log("Register integration system")
	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, name)

	intSysInput := graphql.IntegrationSystemInput{Name: newName, Description: &newDescription}
	intSysGQL, err := pkg.Tc.Graphqlizer.IntegrationSystemInputToGQL(intSysInput)
	updateIntegrationSystemRequest := pkg.FixUpdateIntegrationSystemRequest(intSys.ID, intSysGQL)
	updateOutput := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Update integration system")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, updateIntegrationSystemRequest, &updateOutput)
	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, updateOutput.ID)

	//THEN
	t.Log("Check if Integration System was updated")
	assertIntegrationSystem(t, intSysInput, updateOutput)
	saveExample(t, updateIntegrationSystemRequest.Query(), "update integration system")
}

func TestUnregisterIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Register integration system")
	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, name)

	unregisterIntegrationSystemRequest := pkg.FixUnregisterIntegrationSystem(intSys.ID)
	deleteOutput := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Unregister integration system")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, unregisterIntegrationSystemRequest, &deleteOutput)
	require.NoError(t, err)

	//THEN
	t.Log("Check if Integration System was deleted")

	out := pkg.GetIntegrationSystem(t, ctx, dexGraphQLClient, intSys.ID)

	require.Empty(t, out)
	saveExample(t, unregisterIntegrationSystemRequest.Query(), "unregister integration system")
}

func TestQueryIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name := "int-system"

	t.Log("Register integration system")
	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, name)
	getIntegrationSystemRequest := pkg.FixGetIntegrationSystemRequest(intSys.ID)
	output := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Get integration system")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, getIntegrationSystemRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, output.ID)

	//THEN
	t.Log("Check if Integration System was received")
	assert.Equal(t, name, output.Name)
}

func TestQueryIntegrationSystems(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	name1 := "int-system-1"
	name2 := "int-system-2"

	t.Log("Register integration systems")
	intSys1 := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, name1)
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys1.ID)

	intSys2 := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, name2)
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys2.ID)

	first := 100
	after := ""

	getIntegrationSystemsRequest := pkg.FixGetIntegrationSystemsRequestWithPagination(first, after)
	output := graphql.IntegrationSystemPageExt{}

	// WHEN
	t.Log("List integration systems")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, getIntegrationSystemsRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if Integration Systems were received")
	assertIntegrationSystemNames(t, []string{name1, name2}, output)
	saveExample(t, getIntegrationSystemsRequest.Query(), "query integration systems")
}
