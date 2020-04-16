package middlewares

import (
	"context"
	"net/http"
)

type Header string

const (
	Tenant       Header = "tenant"
	SubAccountID Header = "sub-account"
)

func ExtractTenant(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenant := r.Header.Get(string(Tenant))
		if tenant != "" {
			ctx = context.WithValue(ctx, Tenant, tenant)
		}

		subAccount := r.Header.Get(string(SubAccountID))
		if subAccount != "" {
			ctx = context.WithValue(ctx, SubAccountID, subAccount)
		}

		reqWithCtx := r.WithContext(ctx)

		handler.ServeHTTP(w, reqWithCtx)
	})
}
