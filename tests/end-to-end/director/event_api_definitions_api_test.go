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

func TestQuerySpecificEventAPIDefinition(t *testing.T) {
	// GIVEN
	in := graphql.EventAPIDefinitionInput{
		Name: "test",
		Spec: &graphql.EventAPISpecInput{
			EventSpecType: "ASYNC_API",
			Format:        "YAML",
		},
	}
	EventAPIInputGQL, err := tc.graphqlizer.EventAPIDefinitionInputToGQL(in)
	require.NoError(t, err)
	applicationID := createApplication(t, context.Background(), "test").ID
	defer deleteApplication(t, applicationID)
	actualEventAPI := graphql.EventAPIDefinition{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addEventAPI(applicationID: "%s", in: %s) {
					%s
				}
			}`, applicationID, EventAPIInputGQL, tc.gqlFieldsProvider.ForEventAPI()))
	err = tc.RunOperation(context.TODO(), request, &actualEventAPI)
	require.NoError(t, err)
	require.NotEmpty(t, actualEventAPI.ID)
	createdID := actualEventAPI.ID
	defer deleteEventAPI(t, createdID)

	// WHEN
	queryAppReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: EventAPI(id: "%s") {
					%s
				}
			}`, actualEventAPI.ID, tc.gqlFieldsProvider.ForEventAPI()))
	err = tc.RunOperation(context.Background(), queryAppReq, &actualEventAPI)
	saveQueryInExamples(t, queryAppReq.Query(), "query event api")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdID, actualEventAPI.ID)
}

func deleteEventAPI(t *testing.T, id string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteEventAPI(id: "%s") {
			id
		}	
	}`, id))
	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}
