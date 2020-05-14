package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type key int

const (
	TenantContextKey = iota
	ExternalTenantContextKey
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

func SaveToContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, TenantContextKey, tenant)
}
