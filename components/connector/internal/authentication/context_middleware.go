package authentication

import (
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/connector/internal/oathkeeper"
)

const ()

type authContextMiddleware struct {
}

func NewAuthenticationContextMiddleware() *authContextMiddleware {
	return &authContextMiddleware{}
}

func (acm *authContextMiddleware) PropagateAuthentication(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is no longer used, we probably can delete it
		token := r.Header.Get(oathkeeper.ConnectorTokenHeader)
		r = r.WithContext(PutInContext(r.Context(), ConnectorTokenKey, token))

		// TODO - tests

		clientIdFromToken := r.Header.Get(oathkeeper.ClientIdFromTokenHeader)
		r = r.WithContext(PutInContext(r.Context(), ClientIdFromTokenKey, clientIdFromToken))

		tokenType := r.Header.Get(oathkeeper.TokenTypeHeader)
		r = r.WithContext(PutInContext(r.Context(), TokenTypeKey, tokenType))

		clientIdFromCertificate := r.Header.Get(oathkeeper.ClientIdFromCertificateHeader)
		r = r.WithContext(PutInContext(r.Context(), ClientIdFromCertificateKey, clientIdFromCertificate))

		clientCertificateHash := r.Header.Get(oathkeeper.ClientCertificateHashHeader)
		r = r.WithContext(PutInContext(r.Context(), ClientCertificateHash, clientCertificateHash))

		fmt.Println(r.Header)

		handler.ServeHTTP(w, r)
	})
}
