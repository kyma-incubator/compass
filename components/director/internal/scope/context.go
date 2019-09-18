package scope

import (
	"context"
	"errors"
)

type key int

const ScopesContextKey key = iota

var NoScopesError = errors.New("cannot read scopes from context")

func LoadFromContext(ctx context.Context) ([]string, error) {
	value := ctx.Value(ScopesContextKey)

	scopes, ok := value.([]string)

	if !ok {
		return nil, NoScopesError
	}

	return scopes, nil
}

func SaveToContext(ctx context.Context, scopes []string) context.Context {
	return context.WithValue(ctx, ScopesContextKey, scopes)
}
