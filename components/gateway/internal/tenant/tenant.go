package tenant

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

				w.WriteHeader(http.StatusUnauthorized)
				err := writeJSONError(w, errMessage)
				if err != nil {
					log.Error(errors.Wrap(err, "while writing JSON error"))
					return
				}

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeJSONError(w http.ResponseWriter, errMessage string) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"errors": []string{errMessage},
	})
}

func isExcludedMethod(method string, excludedMethods []string) bool {
	for _, excluded := range excludedMethods {
		if excluded == method {
			return true
		}
	}

	return false
}
