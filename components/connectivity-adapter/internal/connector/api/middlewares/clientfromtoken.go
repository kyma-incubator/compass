package middlewares

import (
	"context"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"net/http"
)

func NewClientFromTokenMiddleware() clientFromTokenMiddleware {
	return clientFromTokenMiddleware{}
}

type clientFromTokenMiddleware struct {
}

func (c clientFromTokenMiddleware) GetClientIdFromToken(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIdFromToken := r.Header.Get(oathkeeper.ClientIdFromTokenHeader)
		if clientIdFromToken == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), ClientIdFromTokenKey, clientIdFromToken))

		handler.ServeHTTP(w, r)
	})
}
