package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/stretchr/testify/require"
)

const (
	notExistingTenant     = "b1e46bd5-18ba-4a02-b96d-631a9e803504"
	emptyTenant           = ""
	tenantRequiredMessage = "Tenant is required"
	tenantNotFoundMessage = "Tenant not found"
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
	_ = pkg.RegisterApplicationFromInputWithinTenant(t, ctx, dexGraphQLClient, notExistingTenant, appInput)
	require.Contains(t, err.Error(), tenantNotFoundMessage)

	_ = pkg.RegisterApplicationFromInputWithinTenant(t, ctx, dexGraphQLClient, emptyTenant, appInput)
	require.Contains(t, err.Error(), tenantRequiredMessage)

	is := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, "test")
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, is.ID)

	req := pkg.FixRequestClientCredentialsForIntegrationSystem(is.ID)

	var credentials graphql.SystemAuth
	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, notExistingTenant, req, &credentials)
	require.Error(t, err)
	require.Contains(t, err.Error(), tenantNotFoundMessage)

	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, emptyTenant, req, &credentials)
	require.NoError(t, err)
	require.NotNil(t, credentials)
}
