package tenant

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type tntCtxKey int
type localTntIDCtxKey int

// TenantContextKey is the key under which the TenantCtx is saved in a given context.Context.
const TenantContextKey tntCtxKey = iota

// LocalTenantIDContextKey is the key under which the local tenant id is saved in a given context.Context.
const LocalTenantIDContextKey localTntIDCtxKey = iota

// TenantCtx is the structure can be saved in a request context. It is used to determine the tenant context in which the request is being executed.
type TenantCtx struct {
	InternalID string
	ExternalID string
}

// LoadFromContext retrieves the internal tenant ID from the provided context. It returns error if such ID cannot be found.
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

// LoadTenantPairFromContext retrieves the whole tenant context from the provided request context. It returns error if such ID cannot be found.
func LoadTenantPairFromContext(ctx context.Context) (TenantCtx, error) {
	tenant, ok := ctx.Value(TenantContextKey).(TenantCtx)

	if !ok {
		return TenantCtx{}, apperrors.NewCannotReadTenantError()
	}

	if tenant.InternalID == "" {
		return TenantCtx{}, apperrors.NewTenantRequiredError()
	}

	return tenant, nil
}

// SaveToContext returns a child context of the provided context, including the provided tenant information.
// The internal tenant ID can be later retrieved from the context by calling LoadFromContext.
func SaveToContext(ctx context.Context, internalID, externalID string) context.Context {
	tenantCtx := TenantCtx{InternalID: internalID, ExternalID: externalID}
	return context.WithValue(ctx, TenantContextKey, tenantCtx)
}

// LoadLocalTenantIDFromContext retrieves the local tenant ID from the provided context. It returns error if such ID cannot be found.
func LoadLocalTenantIDFromContext(ctx context.Context) (string, error) {
	localTenantID, ok := ctx.Value(LocalTenantIDContextKey).(string)

	if !ok {
		return "", apperrors.NewInternalError("cannot read local tenant id from context")
	}

	localTenantID = strings.TrimSpace(localTenantID)
	if localTenantID == "" {
		return "", apperrors.NewInternalError("local tenant id is required")
	}

	return localTenantID, nil
}

// SaveLocalTenantIDToContext returns a child context of the provided context, including the provided local tenant id.
func SaveLocalTenantIDToContext(ctx context.Context, localTenantID string) context.Context {
	return context.WithValue(ctx, LocalTenantIDContextKey, localTenantID)
}
