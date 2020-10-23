package correlation

import (
	"context"
	"net/http"
)

const (
	ContextField       = "CorrelationID"
	RequestIDHeaderKey = "x-request-id"
)

var headerKeys = []string{
	RequestIDHeaderKey,
	"X-Request-ID",
	"X-Request-Id",
	"X-Correlation-ID",
	"X-CorrelationID",
	"X-ForRequest-ID"}

type ContextEnrichMiddleware struct{}

func NewContextEnrichMiddleware() *ContextEnrichMiddleware {
	return &ContextEnrichMiddleware{}
}

func (cem *ContextEnrichMiddleware) AttachCorrelationIDToContext(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var correlationID string
		for _, h := range headerKeys {
			if correlationID = r.Header.Get(h); correlationID != "" {
				break
			}
		}

		if correlationID == "" {
			correlationID = newID()
		}

		r = r.WithContext(context.WithValue(r.Context(), ContextField, correlationID))
		nextHandler.ServeHTTP(w, r)
	})
}
