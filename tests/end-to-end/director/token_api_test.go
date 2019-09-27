//+build !no_token_test

package director

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

//This test also test runtime/application auths custom resolver
//TODO: Currently we don't save OneTimeToken mutations in examples, because those tests are turn off in gen_examples.sh,
// because we need connector up and running, whic requires k8s cluster running.
func TestTokenGeneration(t *testing.T) {
	t.Run("Generate one time token for Runtime", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		runtime := createRuntime(t, ctx, "test")
		defer deleteRuntime(t, runtime.ID)
		tokenRequestNumber := 3

		//WHEN
		for i := 0; i < tokenRequestNumber; i++ {
			token := generateOneTimeTokenForRuntime(t, ctx, runtime.ID)
			assert.NotEmpty(t, token.Token)
			assert.NotEmpty(t, token.ConnectorURL)
		}
		//THEN
		runtimeExt := getRuntime(t, ctx, runtime.ID)
		assert.Len(t, runtimeExt.Auths, tokenRequestNumber)
	})

	t.Run("Generate one time token for Application", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		app := createApplication(t, ctx, "test")
		defer deleteApplication(t, app.ID)
		tokenRequestNumber := 3

		//WHEN
		for i := 0; i < tokenRequestNumber; i++ {
			token := generateOneTimeTokenForApplication(t, ctx, app.ID)
			assert.NotEmpty(t, token.Token)
			assert.NotEmpty(t, token.ConnectorURL)
		}

		//THEN
		appExt := getApplication(t, ctx, app.ID)
		assert.Len(t, appExt.Auths, tokenRequestNumber)
	})
}
