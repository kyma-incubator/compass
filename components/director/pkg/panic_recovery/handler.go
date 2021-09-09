package panicrecovery

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// NewPanicRecoveryMiddleware returns a standard mux middleware that provides panic recovery
func NewPanicRecoveryMiddleware() func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					encoder := json.NewEncoder(w)
					encoder.SetEscapeHTML(false)

					if err := encoder.Encode(apperrors.NewInternalError("Unrecovered panic")); err != nil {
						log.C(r.Context()).WithError(err).Errorf("Panic recovery failed: %s", err)
					}
					debug.PrintStack()
					log.C(r.Context()).WithField("error", err).Errorf("Recovered panic: %s", err)
				}
			}()

			handler.ServeHTTP(w, r)
		})
	}
}
