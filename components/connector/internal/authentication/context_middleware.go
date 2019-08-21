package authentication

import (
	"net/http"
)

const (
	ConnectorTokenHeader string = "Connector-Token"
)

type authContextMiddleware struct {
	headerParser CertificateHeaderParser
}

func NewAuthenticationContextMiddleware(headerParser CertificateHeaderParser) *authContextMiddleware {
	return &authContextMiddleware{
		headerParser: headerParser,
	}
}

func (acm *authContextMiddleware) PropagateAuthentication(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(ConnectorTokenHeader)
		r = r.WithContext(PutInContext(r.Context(), ConnectorTokenKey, token))

		subject, hash, found := acm.headerParser.GetCertificateData(r)
		if found {
			r = r.WithContext(PutInContext(r.Context(), CertificateCommonNameKey, subject))
			r = r.WithContext(PutInContext(r.Context(), CertificateHashKey, hash))
		}

		handler.ServeHTTP(w, r)
	})
}
