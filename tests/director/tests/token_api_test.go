//+build !ignore_external_dependencies

package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/stretchr/testify/require"
	"testing"

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

		tenant := pkg.TestTenants.GetDefaultTenantID()

		input := pkg.FixRuntimeInput("test")
		runtime := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
		defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)
		tokenRequestNumber := 3

		//WHEN
		for i := 0; i < tokenRequestNumber; i++ {
			token := pkg.RequestOneTimeTokenForRuntime(t, ctx, dexGraphQLClient, runtime.ID)
			assert.NotEmpty(t, token.Token)
			assert.NotEmpty(t, token.ConnectorURL)
		}
		//THEN
		runtimeExt := pkg.GetRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)
		assert.Len(t, runtimeExt.Auths, tokenRequestNumber)
	})

	t.Run("Generate one time token for Application", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()

		t.Log("Get Dex id_token")
		dexToken, err := idtokenprovider.GetDexToken()
		require.NoError(t, err)

		dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

		tenant := pkg.TestTenants.GetDefaultTenantID()

		app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "test", tenant)
		defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
		tokenRequestNumber := 3

		//WHEN
		for i := 0; i < tokenRequestNumber; i++ {
			token := pkg.RequestOneTimeTokenForApplication(t, ctx, dexGraphQLClient, app.ID)
			assert.NotEmpty(t, token.Token)
			assert.NotEmpty(t, token.ConnectorURL)
			assert.NotEmpty(t, token.LegacyConnectorURL)
		}

		//THEN
		appExt := pkg.GetApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
		assert.Len(t, appExt.Auths, tokenRequestNumber)
	})
}
