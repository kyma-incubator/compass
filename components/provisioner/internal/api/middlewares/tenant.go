package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
)

type Header string

const (
	Tenant       Header = "tenant"
	SubAccountID Header = "sub-account"
)

func ExtractTenant(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant := r.Header.Get(string(Tenant))
		if tenant == "" {
			util.RespondWithError(w, 400, errors.New("tenant header is empty"))
			return
		}

		ctx := context.WithValue(r.Context(), Tenant, tenant)

		subAccount := r.Header.Get(string(SubAccountID))
		if subAccount != "" {
			ctx = context.WithValue(ctx, SubAccountID, subAccount)
		}

		reqWithCtx := r.WithContext(ctx)

		handler.ServeHTTP(w, reqWithCtx)
	})
}
