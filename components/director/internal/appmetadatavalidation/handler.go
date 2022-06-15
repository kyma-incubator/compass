package appmetadatavalidation

import (
	"context"
	"net/http"
)

type contextKey string

const TenantHeader contextKey = "tenantHeaderKey"

// Handler is an object with dependencies
type handler struct {
}

// NewHandler is a constructor for Handler object
func NewHandler() *handler {
	return &handler{}
}

// Handler adds the tenant header to the context
func (u *handler) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), TenantHeader, r.Header.Get("Tenant"))
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		})
	}
}
