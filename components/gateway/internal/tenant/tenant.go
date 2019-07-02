package tenant

import (
	"fmt"
	"net/http"
)

// TODO: Make one codebase for Director and Gateway

const TenantHeaderName = "tenant"

func RequireTenantHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantValue := r.Header.Get(TenantHeaderName)

		if r.Method != http.MethodGet {
			if tenantValue == "" {
				errMessage := fmt.Sprintf("Header `%s` is required", TenantHeaderName)
				http.Error(w, errMessage, http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}