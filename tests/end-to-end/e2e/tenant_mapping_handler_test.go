package e2e

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/gql"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/idtokenprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDifferentTenantAccessDenied(t *testing.T) {
	ctx := context.Background()
	notExistingTenant := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	t.Log("Get Dex id_token")
	config, err := idtokenprovider.LoadConfig()
	require.NoError(t, err)

	dexToken, err := idtokenprovider.Authenticate(config.IdProviderConfig)
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Try to create Application in not existing tenant")
	appInput := graphql.ApplicationRegisterInput{
		Name:                "app-tmh-test",
		ProviderDisplayName: "compass",
	}
	_, err = createApplicationWithinTenant(t, ctx, dexGraphQLClient, notExistingTenant, appInput)
	assert.Error(t, err)
	assert.Equal(t, "graphql: server returned a non-200 status code: 403", err.Error())
}
