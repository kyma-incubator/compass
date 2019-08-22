package authentication

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextExtension(t *testing.T) {

	t.Run("should put and get value form context", func(t *testing.T) {
		// given
		ctx := context.Background()

		// when
		ctx = PutInContext(ctx, ConnectorTokenKey, "abcd")

		// then
		token, err := GetStringFromContext(ctx, ConnectorTokenKey)
		require.NoError(t, err)
		assert.Equal(t, "abcd", token)
	})

	t.Run("should return error when failed to extract value", func(t *testing.T) {
		// given
		ctx := context.Background()

		// when
		ctx = context.WithValue(ctx, ConnectorTokenKey, struct{}{})

		// then
		token, err := GetStringFromContext(ctx, ConnectorTokenKey)
		require.Error(t, err)
		assert.Empty(t, token)
	})

}
