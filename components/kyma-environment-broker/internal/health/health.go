package health

import (
	"net/http"

	"code.cloudfoundry.org/lager"
)

func LivenessHandler(logger lager.Logger) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("ok"))
		if err != nil {
			logger.Error("while writing to response body", err)
		}
	}
}
