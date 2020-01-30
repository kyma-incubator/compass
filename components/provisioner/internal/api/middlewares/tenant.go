package middlewares

import (
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"net/http"
)

type Header string

const Tenant Header = "tenant"

func ExtractTenant(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant := r.Header.Get(string(Tenant))

		if tenant == "" {
			util.RespondWithError(w, 400, errors.New("tenant header is empty"))
			return
		}

		reqWithCtx := r.WithContext(context.WithValue(r.Context(), Tenant, tenant))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
