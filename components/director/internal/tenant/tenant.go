package tenant

import (
	"context"
)

type key int

const TenantCtxKey key = iota

func LoadFromContext(ctx context.Context) (string, error) {
	value := ctx.Value(TenantCtxKey)

	str, ok := value.(string)

	if !ok {
		//return "", errors.New("Cannot read tenant from context")
		return "sample", nil //TODO: Remove sample tenant
	}

	return str, nil
}

func SaveToContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, TenantCtxKey, tenant)
}
