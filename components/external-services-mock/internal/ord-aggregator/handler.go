package ord_aggregator

import (
	"bytes"
	"fmt"
	"github.com/tidwall/sjson"
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

func HandleFuncOrdConfig(baseURLOverride, accessStrategy string) func(rw http.ResponseWriter, req *http.Request) {
	return HandleFuncOrdConfigWithDocPath(baseURLOverride, "/open-resource-discovery/v1/documents/example1", accessStrategy)
}

func HandleFuncOrdConfigWithDocPath(baseURLOverride, docPath, accessStrategy string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var baseURLFormat string
		if len(baseURLOverride) > 0 {
			baseURLFormat = fmt.Sprintf(`"baseUrl": "%s",`, baseURLOverride)
		}

		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte(fmt.Sprintf(ordConfig, baseURLFormat, docPath, accessStrategy)))
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}

func HandleFuncOrdDocument(expectedBaseURL, specsAccessStrategy string) func(rw http.ResponseWriter, req *http.Request) {
	return HandleFuncOrdDocumentWithAdditionalContent(expectedBaseURL, specsAccessStrategy, "", "")
}

func HandleFuncOrdDocumentWithAdditionalContent(expectedBaseURL, specsAccessStrategy, additionalEntities, additionalProperties string) func(rw http.ResponseWriter, req *http.Request) {
	randomSuffix := fmt.Sprintf("-%s", randSeq(10))
	return func(rw http.ResponseWriter, req *http.Request) {
		t, err := template.New("").Parse(ordDocument)
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while creating template"), http.StatusInternalServerError)
		}

		data := map[string]string{
			"randomSuffix":         randomSuffix,
			"baseURL":              expectedBaseURL,
			"specsAccessStrategy":  specsAccessStrategy,
			"additionalEntities":   additionalEntities,
			"additionalProperties": additionalProperties,
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

func HandleFuncInvalidOrdDocument(expectedBaseURL, specsAccessStrategy string) func(rw http.ResponseWriter, req *http.Request) {
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

		result, err := sjson.SetBytes(res.Bytes(), "packages[0].shortDescription", "Invalid/also invalid/final invalid short description symbols!")
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while setting invalid value in template"), http.StatusInternalServerError)
		}

		rw.WriteHeader(http.StatusOK)
		_, err = rw.Write(result)
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
