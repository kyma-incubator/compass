package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type key int

const (
	TenantContextKey key = iota
)

type TenantCtx struct {
	InternalID string
	ExternalID string
}

func LoadFromContext(ctx context.Context) (string, error) {
	tenant, ok := ctx.Value(TenantContextKey).(TenantCtx)

	if !ok {
		return "", apperrors.NewCannotReadTenantError()
	}

	if tenant.InternalID == "" {
		return "", apperrors.NewTenantRequiredError()
	}

	return tenant.InternalID, nil
}

func SaveToContext(ctx context.Context, internalID, externalID string) context.Context {
	tenantCtx := TenantCtx{InternalID: internalID, ExternalID: externalID}
	return context.WithValue(ctx, TenantContextKey, tenantCtx)
}
