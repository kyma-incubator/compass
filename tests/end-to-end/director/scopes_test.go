package director

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopesAuthorization(t *testing.T) {
	// given
	ctx := context.Background()
	id := uuid.New().String()
	requiredScopes, err := tc.scoper.GetScopesFor("queries.applicationsForRuntime")
	require.NoError(t, err)

	testCases := []struct {
		Name                 string
		UseDefaultScopes     bool
		Scopes               []string
		ExpectedErrorMessage string
	}{
		{Name: "Required Scopes", Scopes: requiredScopes, ExpectedErrorMessage: "runtime does not exist"},
		{Name: "Different Scopes", Scopes: []string{"foo", "bar"}, ExpectedErrorMessage: "insufficient scopes provided"},
		{Name: "No scopes", Scopes: []string{}, ExpectedErrorMessage: "insufficient scopes provided"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			request := fixApplicationForRuntimeRequest(id)
			response := graphql.ApplicationPage{}

			// when
			err := tc.RunOperationWithCustomScopes(ctx, testCase.Scopes, request, &response)

			// then
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
		})
	}
}
