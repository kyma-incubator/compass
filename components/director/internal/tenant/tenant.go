package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

type key int

const TenantHeaderName = "tenant"

const TenantContextKey key = iota

var NoTenantError = errors.New("Cannot read tenant from context")

func LoadFromContext(ctx context.Context) (string, error) {
	value := ctx.Value(TenantContextKey)

	str, ok := value.(string)

	if !ok {
		return "", NoTenantError
	}

	return str, nil
}

func SaveToContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, TenantContextKey, tenant)
}

func RequireAndPassContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantValue := r.Header.Get(TenantHeaderName)

		if r.Method != http.MethodGet {
			if tenantValue == "" {
				errMessage := fmt.Sprintf("Header `%s` is required", TenantHeaderName)
				w.WriteHeader(http.StatusUnauthorized)
				err := writeJSONError(w, errMessage)
				if err != nil {
					log.Error(errors.Wrap(err, "while writing JSON error"))
					return
				}

				return
			}

			ctx := SaveToContext(r.Context(), tenantValue)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSONError(w http.ResponseWriter, errMessage string) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"errors": []string{errMessage},
	})
}
