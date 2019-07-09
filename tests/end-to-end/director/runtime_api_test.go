package director

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeCreateUpdateAndDelete(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	givenInput := graphql.RuntimeInput{
		Name:        "runtime-1",
		Description: ptrString("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []string{"hhh"}},
		Annotations: &graphql.Annotations{"kkk": "lll"},
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	actualRuntime := graphql.Runtime{}

	// WHEN
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createRuntime(in: %s) {
					%s
				}
			}`, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
	saveQueryInExamples(t, createReq.Query(), "create runtime")
	err = tc.RunQuery(ctx, createReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)
	assertRuntime(t, givenInput, actualRuntime)
	assert.NotNil(t, actualRuntime.AgentAuth)

	// update runtime
	givenInput.Description = ptrString("modified-runtime-1-description")
	runtimeInGQL, err = tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)

	// WHEN
	updateReq := gcli.NewRequest(
		fmt.Sprintf(`mutation{ 
			result: updateRuntime(id: "%s", in: %s) {
					%s
				}
			}`, actualRuntime.ID, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
	saveQueryInExamples(t, updateReq.Query(), "update runtime")
	err = tc.RunQuery(ctx, updateReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, *givenInput.Description, *actualRuntime.Description)

	// add Label
	actualLabel := graphql.Label{}

	// WHEN
	addLabelReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addRuntimeLabel(runtimeID: "%s", key: "%s", values: %s) {
					%s
				}
			}`, actualRuntime.ID, "new-label", "[\"bbb\"]", tc.gqlFieldsProvider.ForLabel()))
	err = tc.RunQuery(ctx, addLabelReq, &actualLabel)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, "new-label", actualLabel.Key)
	assert.Len(t, actualLabel.Values, 1)
	assert.Contains(t, actualLabel.Values, "bbb")

	// get runtime and validate runtimes and annotations
	getRuntimeReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}
			}`, actualRuntime.ID, tc.gqlFieldsProvider.ForRuntime()))
	err = tc.RunQuery(ctx, getRuntimeReq, &actualRuntime)
	require.NoError(t, err)
	assert.Len(t, actualRuntime.Labels, 2)

	// delete label

	// WHEN
	delLabelReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: deleteRuntimeLabel(runtimeID: "%s", key: "%s") {
						%s
					}
				}
		`, actualRuntime.ID, "new-label", tc.gqlFieldsProvider.ForLabel()))
	err = tc.RunQuery(ctx, delLabelReq, nil)

	//THEN
	require.NoError(t, err)

	// delete runtime

	// WHEN
	delReq := gcli.NewRequest(fmt.Sprintf(`mutation{result: deleteRuntime(id: "%s") {%s}}`, actualRuntime.ID, tc.gqlFieldsProvider.ForRuntime()))
	saveQueryInExamples(t, delReq.Query(), "delete runtime")
	err = tc.RunQuery(ctx, delReq, nil)

	//THEN
	require.NoError(t, err)
}

func TestSetAndDeleteAPIAuth(t *testing.T) {
	// GIVEN
	// create application
	ctx := context.Background()
	placeholder := "app"
	in := generateSampleApplicationInput(placeholder)

	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: createApplication(in: %s) {
    					%s
					}
				}`, appInputGQL, tc.gqlFieldsProvider.ForApplication()))
	actualApp := ApplicationExt{}
	err = tc.RunQuery(ctx, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)

	// create runtime
	runtimeInput := graphql.RuntimeInput{
		Name:        "runtime-1",
		Description: ptrString("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []string{"hhh"}},
		Annotations: &graphql.Annotations{"kkk": "lll"},
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(runtimeInput)
	require.NoError(t, err)
	actualRuntime := graphql.Runtime{}
	createRuntimeReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createRuntime(in: %s) {
						%s
					}
				}`, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
	err = tc.RunQuery(ctx, createRuntimeReq, &actualRuntime)
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)

	defer deleteRuntime(t, actualRuntime.ID)

	actualRuntimeAuth := graphql.RuntimeAuth{}

	// WHEN
	// set Auth
	authIn := graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: "x-men",
				Password: "secret",
			}}}

	authInStr, err := tc.graphqlizer.AuthInputToGQL(&authIn)
	require.NoError(t, err)
	setAuthReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setAPIAuth(apiID: "%s", runtimeID: "%s", in: %s) {
					%s
				}
			}`, actualApp.Apis.Data[0].ID, actualRuntime.ID, authInStr, tc.gqlFieldsProvider.ForRuntimeAuth()))
	err = tc.RunQuery(ctx, setAuthReq, &actualRuntimeAuth)

	//THEN
	require.NoError(t, err)
	require.NotNil(t, actualRuntimeAuth.Auth)
	assert.Equal(t, actualRuntime.ID, actualRuntimeAuth.RuntimeID)
	actualBasic, ok := actualRuntimeAuth.Auth.Credential.(*graphql.BasicCredentialData)
	require.True(t, ok)
	assert.Equal(t, "x-men", actualBasic.Username)
	assert.Equal(t, "secret", actualBasic.Password)
	// delete Auth
	delAuthReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteAPIAuth(apiID: "%s",runtimeID: "%s") {
					%s
				} 
			}`, actualApp.Apis.Data[0].ID, actualRuntime.ID, tc.gqlFieldsProvider.ForRuntimeAuth()))
	err = tc.RunQuery(ctx, delAuthReq, nil)
	require.NoError(t, err)

}

func TestQueryRuntimes(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	idsToRemove := make([]string, 3)
	defer func() {
		for _, id := range idsToRemove {
			if id != "" {
				deleteRuntime(t, id)
			}
		}
	}()

	for i := 0; i < 3; i++ {
		givenInput := graphql.RuntimeInput{
			Name: fmt.Sprintf("runtime-%d", i),
		}
		runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
		require.NoError(t, err)
		createReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: createRuntime(in: %s) {
						%s
					} 
				}`, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
		actualRuntime := graphql.Runtime{}
		err = tc.RunQuery(ctx, createReq, &actualRuntime)
		require.NoError(t, err)
		require.NotEmpty(t, actualRuntime.ID)
		idsToRemove[i] = actualRuntime.ID
	}
	actualPage := graphql.RuntimePage{}

	// WHEN
	queryReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtimes {
					%s
				}
			}`, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForRuntime())))
	err := tc.RunQuery(ctx, queryReq, &actualPage)
	saveQueryInExamples(t, queryReq.Query(), "query runtimes")

	//THEN
	require.NoError(t, err)
	assert.Len(t, actualPage.Data, 3)
	assert.Equal(t, 3, actualPage.TotalCount)

}

func TestQuerySpecificRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	givenInput := graphql.RuntimeInput{
		Name: "runtime-1",
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createRuntime(in: %s) {
					%s
				}
			}`, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
	createdRuntime := graphql.Runtime{}
	err = tc.RunQuery(ctx, createReq, &createdRuntime)
	require.NoError(t, err)
	require.NotEmpty(t, createdRuntime.ID)

	defer deleteRuntime(t, createdRuntime.ID)
	queriedRuntime := graphql.Runtime{}

	// WHEN
	queryReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}
			}`, createdRuntime.ID, tc.gqlFieldsProvider.ForRuntime()))
	err = tc.RunQuery(ctx, queryReq, &queriedRuntime)
	saveQueryInExamples(t, queryReq.Query(), "query specific runtime")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdRuntime.ID, queriedRuntime.ID)
	assert.Equal(t, createdRuntime.Name, queriedRuntime.Name)
}

func deleteRuntime(t *testing.T, id string) {
	delReq := gcli.NewRequest(
		fmt.Sprintf(`mutation{deleteRuntime(id: "%s") {
				id
			}
		}`, id))
	err := tc.RunQuery(context.Background(), delReq, nil)
	require.NoError(t, err)
}
