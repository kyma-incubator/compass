//+build !no_token_test

package director

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//This test also test runtime/application auths custom resolver
func TestTokenGeneration(t *testing.T) {
	t.Run("Generate one time token for Runtime", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		runtime := createRuntime(t, ctx, "test")
		defer deleteRuntime(t, runtime.ID)
		tokenRequest := fixGeneratOneTimeRuntimeTokenForRuntime(runtime.ID)
		tokenRequestAmount := 3

		//WHEN
		for i := 0; i < tokenRequestAmount; i++ {
			token := graphql.OneTimeToken{}
			err := tc.RunQuery(ctx, tokenRequest, &token)
			require.NoError(t, err)
			assert.NotEmpty(t, token.Token)
			assert.NotEmpty(t, token.ConnectorURL)
		}
		//THEN
		runtimeExt := getRuntime(t, ctx, runtime.ID)
		assert.Len(t, runtimeExt.Auths, tokenRequestAmount)
	})

	t.Run("Generate one time token for Application", func(t *testing.T) {
		ctx := context.TODO()
		app := createApplication(t, ctx, "test")
		defer deleteApplication(t, app.ID)
		tokenRequest := fixGeneratOneTimeRuntimeTokenForApp(app.ID)
		tokenRequestAmount := 3

		//WHEN
		for i := 0; i < tokenRequestAmount; i++ {
			token := graphql.OneTimeToken{}
			err := tc.RunQuery(ctx, tokenRequest, &token)
			require.NoError(t, err)
			assert.NotEmpty(t, token.Token)
			assert.NotEmpty(t, token.ConnectorURL)
		}

		//THEN
		appExt := getApplication(t, ctx, app.ID)
		assert.Len(t, appExt.Auths, tokenRequestAmount)
	})
}
