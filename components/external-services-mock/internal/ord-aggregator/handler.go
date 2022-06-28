package ord_aggregator

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"text/template"
	"time"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func HandleFuncOrdConfig(baseURLOverride, accessStrategy string, isMultiTenant bool) func(rw http.ResponseWriter, req *http.Request) {
	return HandleFuncOrdConfigWithDocPath(baseURLOverride, "/open-resource-discovery/v1/documents/example1", accessStrategy, isMultiTenant)
}

func HandleFuncOrdConfigWithDocPath(baseURLOverride, docPath, accessStrategy string, isMultiTenant bool) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var baseURLFormat string
		if len(baseURLOverride) > 0 {
			baseURLFormat = fmt.Sprintf(`"baseUrl": "%s",`, baseURLOverride)
		}

		if tnt := req.Header["tenant"]; isMultiTenant && len(tnt) == 0 {
			httphelpers.WriteError(rw, errors.New("tenant header is missing"), http.StatusInternalServerError)
		}

		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte(fmt.Sprintf(ordConfig, baseURLFormat, docPath, accessStrategy)))
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}

func HandleFuncOrdDocument(expectedBaseURL string, specsAccessStrategy string) func(rw http.ResponseWriter, req *http.Request) {
	randomSuffix := fmt.Sprintf("-%s", randSeq(10))
	return func(rw http.ResponseWriter, req *http.Request) {
		t, err := template.New("").Parse(ordDocument)
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while creating template"), http.StatusInternalServerError)
		}

		data := map[string]string{
			"randomSuffix":        randomSuffix,
			"baseURL":             expectedBaseURL,
			"specsAccessStrategy": specsAccessStrategy,
		}

		res := new(bytes.Buffer)
		if err = t.Execute(res, data); err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while executing template"), http.StatusInternalServerError)
		}

		rw.WriteHeader(http.StatusOK)
		_, err = rw.Write(res.Bytes())
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
