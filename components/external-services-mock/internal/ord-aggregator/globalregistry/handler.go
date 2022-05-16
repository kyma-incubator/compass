package ord_global_registry

import (
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

func HandleFuncOrdConfig(certSecuredGlobalBaseURL string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		ordConfig := fmt.Sprintf(ordConfigTemplate, certSecuredGlobalBaseURL)
		_, err := rw.Write([]byte(ordConfig))
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}

func HandleFuncOrdDocument() func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte(ordDocument))
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}
