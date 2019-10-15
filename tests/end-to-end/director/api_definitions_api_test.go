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

func TestQuerySpecificAPIDefinition(t *testing.T) {
	// GIVEN
	in := graphql.APIDefinitionInput{
		Name:      "test",
		TargetURL: "test",
	}

	APIInputGQL, err := tc.graphqlizer.APIDefinitionInputToGQL(in)
	require.NoError(t, err)
	applicationID := createApplication(t, context.TODO(), "test").ID
	actualAPI := graphql.APIDefinition{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addApi(applicationID: %s, in: %s) {
					%s
				}
			}`, applicationID, APIInputGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
	err = tc.RunOperation(context.Background(), request, &actualAPI)
	require.NoError(t, err)
	require.NotEmpty(t, actualAPI.ID)
	createdID := actualAPI.ID
	defer deleteAPI(t, createdID)

	// WHEN
	queryAppReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: API(id: "%s") {
					%s
				}
			}`, actualAPI.ID, tc.gqlFieldsProvider.ForAPIDefinition()))
	err = tc.RunOperation(context.Background(), queryAppReq, &actualAPI)
	saveQueryInExamples(t, queryAppReq.Query(), "query api")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdID, actualAPI.ID)
}

func deleteAPI(t *testing.T, id string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteAPI(id: "%s") {
			id
		}	
	}`, id))
	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}
