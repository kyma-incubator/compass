package scope_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/scope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopesContext(t *testing.T) {
	t.Run("load returns scopes previously saved in context", func(t *testing.T) {
		givenScopes := []string{"aaa", "bbb"}
		ctx := scope.SaveToContext(context.Background(), givenScopes)
		actual, err := scope.LoadFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, givenScopes, actual)
	})
	t.Run("load returns error if scopes not found in ctx", func(t *testing.T) {
		_, err := scope.LoadFromContext(context.TODO())
		assert.Equal(t, scope.NoScopesError, err)
	})

	t.Run("cannot override scopes accidentally", func(t *testing.T) {
		givenScopes := []string{"aaa", "bbb"}
		ctx := scope.SaveToContext(context.Background(), givenScopes)
		ctx = context.WithValue(ctx, 0, "some random value")
		actual, err := scope.LoadFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, givenScopes, actual)
	})

}
