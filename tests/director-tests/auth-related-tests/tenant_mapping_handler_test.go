package auth_related_tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director-tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director-tests/pkg/idtokenprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDifferentTenantAccessDenied(t *testing.T) {
	ctx := context.Background()
	notExistingTenant := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Try to create Application in not existing tenant")
	appInput := graphql.ApplicationRegisterInput{
		Name: "app-tmh-test",
	}
	_, err = registerApplicationWithinTenant(t, ctx, dexGraphQLClient, notExistingTenant, appInput)
	assert.Error(t, err)
	assert.Equal(t, "graphql: forbidden: invalid tenant", err.Error())
}
