package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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

	t.Log("Register application as Static User")
	appInput := graphql.ApplicationRegisterInput{
		Name:         "app-static-user",
		ProviderName: ptr.String("compass"),
	}
	_, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, notExistingTenant, appInput)
	require.Error(t, err)
	require.Contains(t, err.Error(), tenantNotFoundMessage)

	_, err = fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, emptyTenant, appInput)
	require.Error(t, err)
	require.Contains(t, err.Error(), tenantRequiredMessage)

	is, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, "test")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, is)
	require.NoError(t, err)
	require.NotEmpty(t, is.ID)

	req := fixtures.FixRequestClientCredentialsForIntegrationSystem(is.ID)

	var credentials graphql.IntSysSystemAuth
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, notExistingTenant, req, &credentials)
	require.Error(t, err)
	require.Contains(t, err.Error(), tenantNotFoundMessage)

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, emptyTenant, req, &credentials)
	require.NoError(t, err)
	require.NotNil(t, credentials)
}
