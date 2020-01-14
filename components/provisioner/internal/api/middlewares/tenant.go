package middlewares

import (
	"context"
	"net/http"
)

const TenantHeader = "tenant"

func ExtractTenant(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant := r.Header.Get(TenantHeader)

		reqWithCtx := r.WithContext(context.WithValue(r.Context(), TenantHeader, tenant))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
