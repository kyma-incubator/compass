package health

import (
	"net/http"

	"code.cloudfoundry.org/lager"
)

func LivenessHandler(logger lager.Logger) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		logger.Info("liveness check ok")
		w.WriteHeader(http.StatusOK)
	}
}

func ReadinessHandler(logger lager.Logger) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		resp, err := http.Get("http://localhost:8080/cluster/v2/catalog")
		if err != nil {
			logger.Error("while sending request on readiness check", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		if resp.StatusCode != http.StatusOK {
			logger.Info("got unexpected status on readiness check")
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}
