//+build !ignore_external_dependencies

package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

//This test also test runtime/application auths custom resolver
//TODO: Currently we don't save OneTimeToken mutations in examples, because those tests are turn off in gen_examples.sh,
// because we need connector up and running, which requires k8s cluster running.
func TestTokenGeneration(t *testing.T) {
	t.Run("Generate one time token for Runtime", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()

		t.Log("Get Dex id_token")
		dexToken, err := idtokenprovider.GetDexToken()
		require.NoError(t, err)

		dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

		tenantId := tenant.TestTenants.GetDefaultTenantID()

		input := fixtures.FixRuntimeInput("test")
		runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
		defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)
		tokenRequestNumber := 3

		//WHEN
		for i := 0; i < tokenRequestNumber; i++ {
			token := fixtures.RequestOneTimeTokenForRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)
			assert.NotEmpty(t, token.Token)
			assert.NotEmpty(t, token.ConnectorURL)
		}
		//THEN
		runtimeExt := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)
		assert.Len(t, runtimeExt.Auths, tokenRequestNumber)
	})

	t.Run("Generate one time token for Application", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()

		t.Log("Get Dex id_token")
		dexToken, err := idtokenprovider.GetDexToken()
		require.NoError(t, err)

		dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

		tenantId := tenant.TestTenants.GetDefaultTenantID()

		app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "test", tenantId)
		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
		tokenRequestNumber := 3

		//WHEN
		for i := 0; i < tokenRequestNumber; i++ {
			token := fixtures.RequestOneTimeTokenForApplication(t, ctx, dexGraphQLClient, app.ID)
			assert.NotEmpty(t, token.Token)
			assert.NotEmpty(t, token.ConnectorURL)
			assert.NotEmpty(t, token.LegacyConnectorURL)
		}

		//THEN
		appExt := fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
		assert.Len(t, appExt.Auths, tokenRequestNumber)
	})
}
