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
	saveQueryInExamples(t, createIntegrationSystemRequest.Query(), "create integration system")

	t.Log("Check if Integration System was created")

	getIntegrationSystemRequest := fixIntegrationSystemRequest(output.ID)
	intSysOutput := graphql.IntegrationSystemExt{}

	err = tc.RunOperation(ctx, getIntegrationSystemRequest, &intSysOutput)

	require.NotEmpty(t, intSysOutput)
	assert.Equal(t, name, intSysOutput.Name)
	saveQueryInExamples(t, getIntegrationSystemRequest.Query(), "query integration system")
}

func TestUpdateIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"
	newName := "new-int-system"

	t.Log("Create integration system")
	intSys := createIntegrationSystem(t, ctx, name)
	require.NotEmpty(t, intSys)

	intSysInput := graphql.IntegrationSystemInput{Name: newName}
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
	assert.Equal(t, newName, updateOutput.Name)
	saveQueryInExamples(t, updateIntegrationSystemRequest.Query(), "update integration system")
}

func TestDeleteIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	t.Log("Create integration system")
	intSys := createIntegrationSystem(t, ctx, name)
	require.NotEmpty(t, intSys)

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
	saveQueryInExamples(t, deleteIntegrationSystemRequest.Query(), "delete integration system")
}

func TestQueryIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "int-system"

	t.Log("Create integration system")
	intSys := createIntegrationSystem(t, ctx, name)
	require.NotEmpty(t, intSys)

	getIntegrationSystemRequest := fixIntegrationSystemRequest(intSys.ID)
	output := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Get integration system")
	err := tc.RunOperation(ctx, getIntegrationSystemRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	defer deleteIntegrationSystem(t, ctx, output.ID)

	//THEN
	t.Log("Check if Integration System was received")
	assert.Equal(t, name, output.Name)
	saveQueryInExamples(t, getIntegrationSystemRequest.Query(), "get integration system")
}

func TestQueryIntegrationSystems(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name1 := "int-system-1"
	name2 := "int-system-2"

	t.Log("Create integration systems")
	intSys1 := createIntegrationSystem(t, ctx, name1)
	require.NotEmpty(t, intSys1)
	defer deleteIntegrationSystem(t, ctx, intSys1.ID)

	intSys2 := createIntegrationSystem(t, ctx, name2)
	require.NotEmpty(t, intSys2)
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
	saveQueryInExamples(t, getIntegrationSystemsRequest.Query(), "get many integration systems")
}
