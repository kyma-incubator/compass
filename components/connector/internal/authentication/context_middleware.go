package authentication

import (
	"fmt"
	"net/http"
)

const (
	ConnectorTokenHeader string = "Connector-Token"
)

type authContextMiddleware struct {
}

func NewAuthenticationContextMiddleware() *authContextMiddleware {
	return &authContextMiddleware{}
}

func (acm *authContextMiddleware) PropagateAuthentication(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(ConnectorTokenHeader)
		r = r.WithContext(PutInContext(r.Context(), ConnectorTokenKey, token))

		// TODO: we should process the headers from Connector Hydrator here and put them into the Context

		fmt.Println(r.Header)

		handler.ServeHTTP(w, r)
	})
}
