package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
)

func NewAuthorizationMiddleware() authorizationHeadersMiddleware {
	return authorizationHeadersMiddleware{}
}

type authorizationHeadersMiddleware struct {
}

func (c authorizationHeadersMiddleware) GetAuthorizationHeaders(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers, err := extractHeaders(r)
		if err != nil {
			reqerror.WriteError(w, err, apperrors.CodeForbidden)

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

	return map[string]string{}, errors.New("invalid token or certificate")
}

func (ah AuthorizationHeaders) GetSystemAuthID() string {
	clientIDFromToken := ah[oathkeeper.ClientIdFromTokenHeader]

	if clientIDFromToken != "" {
		return clientIDFromToken
	}

	return ah[oathkeeper.ClientIdFromCertificateHeader]
}
