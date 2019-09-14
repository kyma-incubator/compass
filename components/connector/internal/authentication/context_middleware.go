package authentication

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
)

type authContextMiddleware struct {
}

func NewAuthenticationContextMiddleware() *authContextMiddleware {
	return &authContextMiddleware{}
}

func (acm *authContextMiddleware) PropagateAuthentication(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIdFromToken := r.Header.Get(oathkeeper.ClientIdFromTokenHeader)
		r = r.WithContext(PutIntoContext(r.Context(), ClientIdFromTokenKey, clientIdFromToken))

		clientIdFromCertificate := r.Header.Get(oathkeeper.ClientIdFromCertificateHeader)
		r = r.WithContext(PutIntoContext(r.Context(), ClientIdFromCertificateKey, clientIdFromCertificate))

		clientCertificateHash := r.Header.Get(oathkeeper.ClientCertificateHashHeader)
		r = r.WithContext(PutIntoContext(r.Context(), ClientCertificateHashKey, clientCertificateHash))

		handler.ServeHTTP(w, r)
	})
}
