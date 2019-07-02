package tenant

import (
	"fmt"
	"net/http"
)

// TODO: Make one codebase for Director and Gateway

const TenantHeaderName = "tenant"

func RequireTenantHeader(excludedMethods ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantValue := r.Header.Get(TenantHeaderName)

			if !isExcludedMethod(r.Method, excludedMethods) && tenantValue == "" {
				errMessage := fmt.Sprintf("Header `%s` is required", TenantHeaderName)
				http.Error(w, errMessage, http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isExcludedMethod(method string, excludedMethods []string) bool {
	for _, excluded := range excludedMethods {
		if excluded == method {
			return true
		}
	}

	return false
}
