package ord_aggregator

import (
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

func HandleFuncOrdConfig(baseURLOverride, accessStrategy string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var baseURLFormat string
		if len(baseURLOverride) > 0 {
			baseURLFormat = fmt.Sprintf(`"baseUrl": "%s",`, baseURLOverride)
		}

		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte(fmt.Sprintf(ordConfig, baseURLFormat, accessStrategy)))
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}

func HandleFuncOrdDocument(expectedBaseURL string, specsAccessStrategy string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte(fmt.Sprintf(ordDocument, expectedBaseURL, specsAccessStrategy)))
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}
