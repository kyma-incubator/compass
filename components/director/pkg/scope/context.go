package scope

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type key int

// ScopesContextKey is the key under which the scopes are saved in a given context.Context.
const ScopesContextKey key = iota

// LoadFromContext retrieves the scopes from the provided context. It returns error if they cannot be found
func LoadFromContext(ctx context.Context) ([]string, error) {
	value := ctx.Value(ScopesContextKey)
	scopes, ok := value.([]string)
	if !ok {
		return nil, apperrors.NewNoScopesInContextError()
	}
	return scopes, nil
}

// SaveToContext returns a child context of the provided context, including the provided scopes information
func SaveToContext(ctx context.Context, scopes []string) context.Context {
	return context.WithValue(ctx, ScopesContextKey, scopes)
}
// IsGivenScopePresent returns whether an input scope is present in the provided scopes in the context. It returns error if scopes cannot be found
func IsGivenScopePresent(ctx context.Context, scope string) (bool, error) {
	value := ctx.Value(ScopesContextKey)
	scopes, ok := value.([]string)
	if !ok {
		return false, apperrors.NewNoScopesInContextError()
	}

	for _, s := range scopes {
		if s == scope {
			return true, nil
		}
	}
	return false, nil
}