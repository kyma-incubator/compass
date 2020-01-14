package tenant

import (
	"context"

	"github.com/pkg/errors"
)

type key int

const TenantContextKey key = iota

var NoTenantError = errors.New("cannot read tenant from context")

func LoadFromContext(ctx context.Context) (string, error) {
	value := ctx.Value(TenantContextKey)

	str, ok := value.(string)

	if !ok {
		return "", NoTenantError
	}

	return str, nil
}

func SaveToContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, TenantContextKey, tenant)
}
