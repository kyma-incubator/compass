package panic_recovery

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"github.com/pkg/errors"
	"net/http"
	"runtime/debug"
)

// NewRecoveryMiddleware returns a standard mux middleware that provides panic recovery
func NewRecoveryMiddleware() func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					encoder := json.NewEncoder(w)
					encoder.SetEscapeHTML(false)

					if err := encoder.Encode(apiresponses.ErrorResponse{Description: "Internal Server Error"}); err != nil {
						err = errors.Wrap(err, "while encoding response")
						log.C(r.Context()).WithError(err).Error("panic recovery failure")
					}
					debug.PrintStack()
					log.C(r.Context()).WithField("error", err).Error("recovered panic")
				}
			}()

			handler.ServeHTTP(w, r)
		})
	}
}
