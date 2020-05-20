package gateway_integration

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"
	"github.com/stretchr/testify/require"
)

const (
	wrongTenant        = "b1e46bd5-18ba-4a02-b96d-631a9e803504"
	emptyTenant        = ""
	noTenantMessage    = "tenant not found"
	wrongTenantMessage = "No tenant found :"
)

func TestTenantErrors(t *testing.T) {

	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Register application as Static User")
	appInput := graphql.ApplicationRegisterInput{
		Name:         "app-static-user",
		ProviderName: ptr.String("compass"),
	}
	_, err = registerApplicationWithinTenant(t, ctx, dexGraphQLClient, wrongTenant, appInput)
	require.Error(t, err)
	require.Contains(t, err.Error(), wrongTenantMessage)

	_, err = registerApplicationWithinTenant(t, ctx, dexGraphQLClient, emptyTenant, appInput)
	require.Error(t, err)
	require.Contains(t, err.Error(), noTenantMessage)

	is := registerIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, "test")

	req := fixGenerateClientCredentialsForIntegrationSystem(is.ID)

	var credentials graphql.SystemAuth
	err = tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, wrongTenant, req, &credentials)
	require.Error(t, err)
	require.Contains(t, err.Error(), wrongTenantMessage)

	err = tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, emptyTenant, req, &credentials)
	require.NoError(t, err)
	require.NotNil(t, credentials)
}
