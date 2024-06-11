//go:build !ignore_external_dependencies
// +build !ignore_external_dependencies

package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/stretchr/testify/assert"
)

// This test also test runtime/application auths custom resolver
// TODO: Currently we don't save OneTimeToken mutations in examples, because those tests are turn off in gen_examples.sh,
// because we need connector up and running, which requires k8s cluster running.
func TestTokenGeneration(t *testing.T) {
	t.Run("Generate one time token for Runtime", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()

		tenantId := tenant.TestTenants.GetDefaultTenantID()

		input := fixtures.FixRuntimeRegisterInputWithoutLabels("test")

		var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
		runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &input)
		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)
		tokenRequestNumber := 3

		//WHEN
		for i := 0; i < tokenRequestNumber; i++ {
			token := fixtures.RequestOneTimeTokenForRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID)
			assert.NotEmpty(t, token.Token)
			assert.NotEmpty(t, token.ConnectorURL)
		}
		//THEN
		runtimeExt := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID)
		assert.Len(t, runtimeExt.Auths, tokenRequestNumber)
	})

	t.Run("Generate one time token for Application", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()

		tenantId := tenant.TestTenants.GetDefaultTenantID()

		app, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "test", tenantId)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
		require.NoError(t, err)
		require.NotEmpty(t, app.ID)
		tokenRequestNumber := 3

		//WHEN
		for i := 0; i < tokenRequestNumber; i++ {
			token := fixtures.RequestOneTimeTokenForApplication(t, ctx, certSecuredGraphQLClient, app.ID)
			assert.NotEmpty(t, token.Token)
			assert.NotEmpty(t, token.ConnectorURL)
			assert.NotEmpty(t, token.LegacyConnectorURL)
		}

		//THEN
		appExt := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, app.ID)
		assert.Len(t, appExt.Auths, tokenRequestNumber)
	})
}
