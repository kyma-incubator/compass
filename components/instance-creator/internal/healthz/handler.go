package healthz

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"
)

// NewHTTPHandler returns new handler function to process readiness and liveness probes
func NewHTTPHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, err := writer.Write([]byte("ok"))
		if err != nil {
			log.C(request.Context()).Errorf(errors.Wrapf(err, "while writing to response body").Error())
		}
	}
}
