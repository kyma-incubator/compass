package apispec

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
)

const (
	JSONFormat       = "json"
	YAMLFormat       = "yaml"
	XMLFormat        = "xml"
	formatQueryParam = "format"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return b
}

func randSpec(spec string) []byte {
	return []byte(fmt.Sprintf(spec, randString(10)))
}

func HandleFunc(rw http.ResponseWriter, req *http.Request) {
	handleResp(rw, req)
}

func FlappingHandleFunc() func(rw http.ResponseWriter, req *http.Request) {
	var toggle bool
	lock := sync.Mutex{}
	return func(rw http.ResponseWriter, req *http.Request) {
		lock.Lock()

		defer func() {
			toggle = !toggle
			lock.Unlock()
		}()

		if !toggle {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		handleResp(rw, req)
	}
}

func handleResp(rw http.ResponseWriter, req *http.Request) {
	formatQueryParamValue := req.URL.Query().Get(formatQueryParam)

	var err error
	switch formatQueryParamValue {
	case JSONFormat:
		rw.WriteHeader(http.StatusOK)
		_, err = rw.Write(randSpec(specJSONTemplate))
	case XMLFormat:
		rw.WriteHeader(http.StatusOK)
		_, err = rw.Write(randSpec(specXMLTemplate))
	case YAMLFormat:
		rw.WriteHeader(http.StatusOK)
		_, err = rw.Write(randSpec(specYAMLTemplate))
	default:
		httphelpers.WriteError(rw, errors.Errorf("Request query parameter %q is not provided. Provide one of the following: %q, %q, %q", formatQueryParam, JSONFormat, YAMLFormat, XMLFormat), http.StatusInternalServerError)
		return
	}

	if err != nil {
		httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
	}
}
