package middlewares

import (
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"net/http"
)

func NewClientFromTokenMiddleware() authorizationHeadersMiddleware {
	return authorizationHeadersMiddleware{}
}

type authorizationHeadersMiddleware struct {
}

func (c authorizationHeadersMiddleware) GetAuthorizationHeaders(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers, err := extractHeaders(r)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		r = r.WithContext(context.WithValue(r.Context(), AuthorizationHeadersKey, AuthorizationHeaders(headers)))
		handler.ServeHTTP(w, r)
	})
}

func extractHeaders(r *http.Request) (map[string]string, error) {
	clientIdFromToken := r.Header.Get(oathkeeper.ClientIdFromTokenHeader)
	if clientIdFromToken != "" {
		return map[string]string{
			oathkeeper.ClientIdFromTokenHeader: clientIdFromToken,
		}, nil
	}

	clientIdFromCertificate := r.Header.Get(oathkeeper.ClientIdFromCertificateHeader)
	clientCertificateHash := r.Header.Get(oathkeeper.ClientCertificateHashHeader)
	if clientIdFromCertificate != "" && clientCertificateHash != "" {
		return map[string]string{
			oathkeeper.ClientIdFromCertificateHeader: clientIdFromCertificate,
			oathkeeper.ClientCertificateHashHeader:   clientCertificateHash,
		}, nil
	}

	return map[string]string{}, errors.New("authorization failed")
}

func (ah AuthorizationHeaders) GetClientID() string {
	clientIDFromToken := ah[oathkeeper.ClientIdFromTokenHeader]

	if clientIDFromToken != "" {
		return clientIDFromToken
	}

	return ah[oathkeeper.ClientIdFromCertificateHeader]
}
