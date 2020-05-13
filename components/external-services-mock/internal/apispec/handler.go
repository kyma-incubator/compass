package apispec

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return b
}

func randSpec() []byte {
	return []byte(fmt.Sprintf(specTemplate, randString(10)))
}

func HandleFunc(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write(randSpec())
	if err != nil {
		httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
	}
}
