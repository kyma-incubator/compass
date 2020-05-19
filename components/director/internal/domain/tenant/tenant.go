package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type key int

const (
	TenantContextKey         key = iota
	ExternalTenantContextKey key = iota
)

func LoadFromContext(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value(TenantContextKey).(string)

	if !ok {
		return "", apperrors.NewCannotReadTenantError()
	}

	if tenantID == "" {
		return "", apperrors.NewEmptyTenantError()
	}

	return tenantID, nil
}

func SaveInternalToContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, TenantContextKey, tenant)
}

func SaveExternalToContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, ExternalTenantContextKey, tenant)
}
