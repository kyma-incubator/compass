package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type key int

const (
	// TenantContextKey is the key under which the TenantCtx is saved in a given context.Context.
	TenantContextKey key = iota

	// IsolationTypeKey is the key under which the IsolationType is saved in a given context.Context.
	IsolationTypeKey
)

type IsolationType string

const (
	SimpleIsolationType    IsolationType = "simple"
	RecursiveIsolationType IsolationType = "recursive"
)

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

// SaveToContext returns a child context of the provided context, including the provided tenant information.
// The internal tenant ID can be later retrieved from the context by calling LoadFromContext.
func SaveToContext(ctx context.Context, internalID, externalID string) context.Context {
	tenantCtx := TenantCtx{InternalID: internalID, ExternalID: externalID}
	return context.WithValue(ctx, TenantContextKey, tenantCtx)
}

// LoadIsolationTypeFromContext loads the tenant isolation type from context.
// If no valid isolation type is set, consider the recursive type as the default one.
func LoadIsolationTypeFromContext(ctx context.Context) IsolationType {
	if isolationType, ok := ctx.Value(IsolationTypeKey).(IsolationType); ok && isolationType.IsValid() {
		return isolationType
	}
	return RecursiveIsolationType
}

// SaveIsolationTypeToContext saves the isolation type into the provided context.
func SaveIsolationTypeToContext(ctx context.Context, isolationTypeString string) context.Context {
	return context.WithValue(ctx, IsolationTypeKey, IsolationType(isolationTypeString))
}

func (it IsolationType) IsValid() bool {
	return it == SimpleIsolationType ||
		it == RecursiveIsolationType
}
