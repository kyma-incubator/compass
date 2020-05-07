package scope

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type key int

const ScopesContextKey key = iota

func LoadFromContext(ctx context.Context) ([]string, error) {
	value := ctx.Value(ScopesContextKey)
	scopes, ok := value.([]string)
	if !ok {
		return nil, apperrors.NewNoScopesInContextError()
	}
	return scopes, nil
}

func SaveToContext(ctx context.Context, scopes []string) context.Context {
	return context.WithValue(ctx, ScopesContextKey, scopes)
}
