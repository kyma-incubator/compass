package director

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	intSysInput := graphql.IntegrationSystemInput{Name: name}
	intSys, err := tc.graphqlizer.IntegrationSystemInputToGQL(intSysInput)
	require.NoError(t, err)

	createIntegrationSystemRequest := fixCreateIntegrationSystemRequest(intSys)
	output := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Create integration system")

	err = tc.RunOperation(ctx, createIntegrationSystemRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	defer deleteIntegrationSystem(t, ctx, output.ID)

	//THEN
	require.NotEmpty(t, output.Name)
	saveExample(t, createIntegrationSystemRequest.Query(), "create integration system")

	t.Log("Check if Integration System was created")

	getIntegrationSystemRequest := fixIntegrationSystemRequest(output.ID)
	intSysOutput := graphql.IntegrationSystemExt{}

	err = tc.RunOperation(ctx, getIntegrationSystemRequest, &intSysOutput)

	require.NotEmpty(t, intSysOutput)
	assertIntegrationSystem(t, intSysInput, intSysOutput)
	saveExample(t, getIntegrationSystemRequest.Query(), "query integration system")
}

func TestUpdateIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"
	newName := "new-int-system"
	newDescription := "new description"
	t.Log("Create integration system")
	intSys := createIntegrationSystem(t, ctx, name)

	intSysInput := graphql.IntegrationSystemInput{Name: newName, Description: &newDescription}
	intSysGQL, err := tc.graphqlizer.IntegrationSystemInputToGQL(intSysInput)
	updateIntegrationSystemRequest := fixUpdateIntegrationSystemRequest(intSys.ID, intSysGQL)
	updateOutput := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Update integration system")
	err = tc.RunOperation(ctx, updateIntegrationSystemRequest, &updateOutput)
	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)
	defer deleteIntegrationSystem(t, ctx, updateOutput.ID)

	//THEN
	t.Log("Check if Integration System was updated")
	assertIntegrationSystem(t, intSysInput, updateOutput)
	saveExample(t, updateIntegrationSystemRequest.Query(), "update integration system")
}

func TestDeleteIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	t.Log("Create integration system")
	intSys := createIntegrationSystem(t, ctx, name)

	deleteIntegrationSystemRequest := fixDeleteIntegrationSystem(intSys.ID)
	deleteOutput := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Delete integration system")
	err := tc.RunOperation(ctx, deleteIntegrationSystemRequest, &deleteOutput)
	require.NoError(t, err)

	//THEN
	t.Log("Check if Integration System was deleted")

	out := getIntegrationSystem(t, ctx, intSys.ID)

	require.Empty(t, out)
	saveExample(t, deleteIntegrationSystemRequest.Query(), "delete integration system")
}

func TestQueryIntegrationSystems(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name1 := "int-system-1"
	name2 := "int-system-2"

	t.Log("Create integration systems")
	intSys1 := createIntegrationSystem(t, ctx, name1)
	defer deleteIntegrationSystem(t, ctx, intSys1.ID)

	intSys2 := createIntegrationSystem(t, ctx, name2)
	defer deleteIntegrationSystem(t, ctx, intSys2.ID)

	first := 2
	after := ""

	getIntegrationSystemsRequest := fixIntegrationSystemsRequest(first, after)
	output := graphql.IntegrationSystemPageExt{}

	// WHEN
	t.Log("List integration systems")
	err := tc.RunOperation(ctx, getIntegrationSystemsRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if Integration Systems were received")
	assert.Equal(t, 2, output.TotalCount)
	saveExample(t, getIntegrationSystemsRequest.Query(), "query integration systems")
}
