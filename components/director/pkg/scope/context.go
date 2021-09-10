package scope

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type key int

// ScopesContextKey missing godoc
const ScopesContextKey key = iota

// LoadFromContext missing godoc
func LoadFromContext(ctx context.Context) ([]string, error) {
	value := ctx.Value(ScopesContextKey)
	scopes, ok := value.([]string)
	if !ok {
		return nil, apperrors.NewNoScopesInContextError()
	}
	return scopes, nil
}

// SaveToContext missing godoc
func SaveToContext(ctx context.Context, scopes []string) context.Context {
	return context.WithValue(ctx, ScopesContextKey, scopes)
}
