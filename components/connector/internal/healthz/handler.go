package healthz

import (
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"net/http"

	"github.com/pkg/errors"
)

func NewHTTPHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		_, err := writer.Write([]byte("ok"))
		if err != nil {
			log.C(request.Context()).Errorf(errors.Wrapf(err, "while writing to response body").Error())
		}
	}
}
