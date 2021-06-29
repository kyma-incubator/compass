package systemfetcher

import (
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

func HandleFunc(defaultTenant string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		filter := req.URL.Query().Get("$filter")
		rw.WriteHeader(http.StatusOK)

		if !strings.Contains(filter, defaultTenant) {
			_, err := rw.Write([]byte(`[]`))
			if err != nil {
				httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
			}
			return
		}

		_, err := rw.Write([]byte(`[{
			"systemNumber": "1",
			"displayName": "name1",
			"productDescription": "description",
			"type": "type1",
			"prop": "val1",
			"baseUrl": "",
			"infrastructureProvider": "",
			"additionalUrls": {},
			"additionalAttributes": {}
		},{
			"systemNumber": "2",
			"displayName": "name2",
			"productDescription": "description",
			"type": "type2",
			"baseUrl": "",
			"infrastructureProvider": "",
			"additionalUrls": {},
			"additionalAttributes": {}
		}]`))
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}
