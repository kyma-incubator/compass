package scope_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopesContext(t *testing.T) {
	t.Run("load returns scopes previously saved in context", func(t *testing.T) {
		// GIVEN
		givenScopes := []string{"aaa", "bbb"}
		ctx := scope.SaveToContext(context.Background(), givenScopes)
		// WHEN
		actual, err := scope.LoadFromContext(ctx)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenScopes, actual)
	})
	t.Run("load returns error if scopes not found in ctx", func(t *testing.T) {
		// WHEN
		_, err := scope.LoadFromContext(context.TODO())
		// THEN
		assert.Equal(t, scope.NoScopesInContextError, err)
	})

	t.Run("cannot override scopes accidentally", func(t *testing.T) {
		// GIVEN
		givenScopes := []string{"aaa", "bbb"}
		ctx := scope.SaveToContext(context.Background(), givenScopes)
		ctx = context.WithValue(ctx, 0, "some random value")
		// WHEN
		actual, err := scope.LoadFromContext(ctx)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenScopes, actual)
	})

}
