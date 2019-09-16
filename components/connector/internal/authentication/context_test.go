package authentication_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextExtension(t *testing.T) {

	t.Run("should put and get value form context", func(t *testing.T) {
		// given
		ctx := context.Background()

		// when
		ctx = authentication.PutIntoContext(ctx, authentication.ConnectorTokenKey, "abcd")

		// then
		token, err := authentication.GetStringFromContext(ctx, authentication.ConnectorTokenKey)
		require.NoError(t, err)
		assert.Equal(t, "abcd", token)
	})

	t.Run("should return error when failed to extract value", func(t *testing.T) {
		// given
		ctx := context.Background()

		// when
		ctx = context.WithValue(ctx, authentication.ConnectorTokenKey, struct{}{})

		// then
		token, err := authentication.GetStringFromContext(ctx, authentication.ConnectorTokenKey)
		require.Error(t, err)
		assert.Empty(t, token)
	})

}
