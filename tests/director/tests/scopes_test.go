package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopesAuthorization(t *testing.T) {
	// given
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	id := uuid.New().String()

	testCases := []struct {
		Name                 string
		UseDefaultScopes     bool
		Scopes               []string
		ExpectedErrorMessage string
	}{
		{Name: "Different Scopes", Scopes: []string{"foo", "bar"}, ExpectedErrorMessage: "insufficient scopes provided"},
		{Name: "No scopes", Scopes: []string{}, ExpectedErrorMessage: "insufficient scopes provided"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			request := pkg.FixApplicationForRuntimeRequest(id)
			response := graphql.ApplicationPage{}

			// when
			err := pkg.Tc.RunOperationWithCustomScopes(ctx, dexGraphQLClient, tenant, testCase.Scopes, request, &response)

			// then
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
		})
	}
}
