package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type key int

const TenantContextKey key = iota

func LoadFromContext(ctx context.Context) (string, error) {
	value := ctx.Value(TenantContextKey)

	str, ok := value.(string)

	if !ok {
		return "", apperrors.NewCannotReadTenantError()
	}

	if str == "" {
		return "", apperrors.NewEmptyTenantError()
	}

	return str, nil
}

func SaveToContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, TenantContextKey, tenant)
}
