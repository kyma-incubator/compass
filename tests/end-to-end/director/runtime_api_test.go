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
		Name:        "runtime-create-update-delete",
		Description: ptrString("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []interface{}{"hhh"}},
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

	// add Label
	actualLabel := graphql.Label{}

	// WHEN
	addLabelReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setRuntimeLabel(runtimeID: "%s", key: "%s", value: %s) {
					%s
				}
			}`, actualRuntime.ID, "new-label", "[\"bbb\"]", tc.gqlFieldsProvider.ForLabel()))
	err = tc.RunQuery(ctx, addLabelReq, &actualLabel)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, "new-label", actualLabel.Key)
	assert.Len(t, actualLabel.Value, 1)
	assert.Contains(t, actualLabel.Value, "bbb")

	// get runtime and validate runtimes
	getRuntimeReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}
			}`, actualRuntime.ID, tc.gqlFieldsProvider.ForRuntime()))
	err = tc.RunQuery(ctx, getRuntimeReq, &actualRuntime)
	require.NoError(t, err)
	//assert.Len(t, actualRuntime.Labels, 2) // TODO: Make it work when labels are in place

	// add agent auth
	// GIVEN
	in := generateSampleApplicationInput("app")

	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	createAppReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: createApplication(in: %s) {
    					%s
					}
				}`, appInputGQL, tc.gqlFieldsProvider.ForApplication()))
	actualApp := ApplicationExt{}

	//WHEN
	err = tc.RunQuery(ctx, createAppReq, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)

	// set Auth
	// GIVEN
	authIn := graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: "x-men",
				Password: "secret",
			}}}
	actualRuntimeAuth := graphql.RuntimeAuth{}

	authInStr, err := tc.graphqlizer.AuthInputToGQL(&authIn)
	require.NoError(t, err)
	setAuthReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setAPIAuth(apiID: "%s", runtimeID: "%s", in: %s) {
					%s
				}
			}`, actualApp.Apis.Data[0].ID, actualRuntime.ID, authInStr, tc.gqlFieldsProvider.ForRuntimeAuth()))

	//WHEN
	err = tc.RunQuery(ctx, setAuthReq, &actualRuntimeAuth)

	//THEN
	require.NoError(t, err)
	require.NotNil(t, actualRuntimeAuth.Auth)
	assert.Equal(t, actualRuntime.ID, actualRuntimeAuth.RuntimeID)
	actualBasic, ok := actualRuntimeAuth.Auth.Credential.(*graphql.BasicCredentialData)
	require.True(t, ok)
	assert.Equal(t, "x-men", actualBasic.Username)
	assert.Equal(t, "secret", actualBasic.Password)

	// update runtime, check if only simple values are updated
	//GIVEN
	givenInput.Name = "updated-name"
	givenInput.Description = ptrString("updated-description")
	givenInput.Labels = &graphql.Labels{
		"key": []interface{}{"values", "aabbcc"},
	}
	runtimeInGQL, err = tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	actualRuntime = graphql.Runtime{ID: actualRuntime.ID}
	updateRuntimeReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: updateRuntime(id: "%s", in: %s) {
					%s
				}
		}
		`, actualRuntime.ID, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
	saveQueryInExamples(t, updateRuntimeReq.Query(), "update runtime")
	//WHEN
	err = tc.RunQuery(ctx, updateRuntimeReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, givenInput.Name, actualRuntime.Name)
	assert.Equal(t, *givenInput.Description, *actualRuntime.Description)
	assert.NotNil(t, actualRuntime.AgentAuth)

	// delete runtime

	// WHEN
	delReq := gcli.NewRequest(fmt.Sprintf(`mutation{result: deleteRuntime(id: "%s") {%s}}`, actualRuntime.ID, tc.gqlFieldsProvider.ForRuntime()))
	saveQueryInExamples(t, delReq.Query(), "delete runtime")
	err = tc.RunQuery(ctx, delReq, nil)

	//THEN
	require.NoError(t, err)
}

func TestRuntimeCreateUpdateDuplicatedNames(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	firstRuntimeName := "unique-name-1"
	givenInput := graphql.RuntimeInput{
		Name:        firstRuntimeName,
		Description: ptrString("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []interface{}{"hhh"}},
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	firstRuntime := graphql.Runtime{}
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createRuntime(in: %s) {
					%s
				}
			}`, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
	// WHEN
	err = tc.RunQuery(ctx, createReq, &firstRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, firstRuntime.ID)
	assertRuntime(t, givenInput, firstRuntime)
	assert.NotNil(t, firstRuntime.AgentAuth)
	defer deleteRuntime(t, firstRuntime.ID)

	// try to create second runtime with first runtime name
	//GIVEN
	givenInput = graphql.RuntimeInput{
		Name:        firstRuntimeName,
		Description: ptrString("runtime-1-description"),
	}
	runtimeInGQL, err = tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	createReq = gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createRuntime(in: %s) {
					%s
				}
			}`, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
	saveQueryInExamples(t, createReq.Query(), "create runtime")

	// WHEN
	err = tc.RunQuery(ctx, createReq, nil)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "runtime name is not unique within tenant")

	// create second runtime
	//GIVEN
	secondRuntimeName := "unique-name-2"
	givenInput = graphql.RuntimeInput{
		Name:        secondRuntimeName,
		Description: ptrString("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []interface{}{"hhh"}},
	}
	runtimeInGQL, err = tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	secondRuntime := graphql.Runtime{}
	createReq = gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createRuntime(in: %s) {
					%s
				}
			}`, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))

	// WHEN
	err = tc.RunQuery(ctx, createReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, secondRuntime.ID)
	assertRuntime(t, givenInput, secondRuntime)
	assert.NotNil(t, secondRuntime.AgentAuth)
	defer deleteRuntime(t, secondRuntime.ID)

	//Update first runtime with second runtime name, failed

	//GIVEN
	givenInput = graphql.RuntimeInput{
		Name:        secondRuntimeName,
		Description: ptrString("runtime-1-description"),
	}
	runtimeInGQL, err = tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	createReq = gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateRuntime(id: "%s", in :%s) {
					%s
				}
			}`, firstRuntime.ID, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))

	// WHEN
	err = tc.RunQuery(ctx, createReq, &secondRuntime)

	//THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "runtime name is not unique within tenant")
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
		Name:        "runtime-set-delete-api",
		Description: ptrString("runtime-1-description"),
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

	inputRuntimes := []*graphql.Runtime{
		{Name: "runtime-query-1", Description: ptrString("test description")},
		{Name: "runtime-query-2", Description: ptrString("another description")},
		{Name: "runtime-query-3"},
	}

	for i := 0; i < 3; i++ {
		givenInput := graphql.RuntimeInput{
			Name:        inputRuntimes[i].Name,
			Description: inputRuntimes[i].Description,
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
		inputRuntimes[i].ID = actualRuntime.ID
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
	runtimesContain(t, actualPage.Data, inputRuntimes)
}

func TestQuerySpecificRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	givenInput := graphql.RuntimeInput{
		Name: "runtime-specific-runtime",
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
	saveQueryInExamples(t, queryReq.Query(), "query runtime")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdRuntime.ID, queriedRuntime.ID)
	assert.Equal(t, createdRuntime.Name, queriedRuntime.Name)
	assert.Equal(t, createdRuntime.Description, queriedRuntime.Description)
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

// runtimesContain tests if every runtime from subset is present in set, comparing Name, ID and Description
func runtimesContain(t *testing.T, set, subset []*graphql.Runtime) {
	for _, subElement := range subset {
		assert.Condition(t, func() (success bool) {
			for _, superElement := range set {
				// check ID and Name
				if superElement.ID != subElement.ID || superElement.Name != subElement.Name {
					continue
				}
				// check Description
				if subElement.Description == nil || superElement.Description == nil {
					if subElement.Description == superElement.Description {
						return true
					} else {
						continue
					}
				}
				if *subElement.Description != *superElement.Description {
					continue
				}
				return true
			}
			return false
		})
	}
}
