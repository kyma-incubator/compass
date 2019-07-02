package tenant

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type key int

const TenantHeaderName = "tenant"

const TenantContextKey key = iota

func LoadFromContext(ctx context.Context) (string, error) {
	value := ctx.Value(TenantContextKey)

	str, ok := value.(string)

	if !ok {
		return "", errors.New("Cannot read tenant from context")
	}

	return str, nil
}

func SaveToContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, TenantContextKey, tenant)
}

func RequireHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantValue := r.Header.Get(TenantHeaderName)

		if r.Method != http.MethodGet {
			if tenantValue == "" {
				errMessage := fmt.Sprintf("Header `%s` is required", TenantHeaderName)
				http.Error(w, errMessage, http.StatusBadRequest)
				return
			}

			ctx := SaveToContext(r.Context(), tenantValue)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
