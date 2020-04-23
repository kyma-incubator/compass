package httphelpers

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
)

const (
	HeaderContentTypeKey   = "Content-Type"
	HeaderContentTypeValue = "application/json;charset=UTF-8"
)

func WriteError(writer http.ResponseWriter, err error, statusCode int) {
	writer.Header().Set(HeaderContentTypeKey, HeaderContentTypeValue)

	response := model.ErrorResponse{
		Error: err.Error(),
	}

	value, err := json.Marshal(&response)
	if err != nil {
		//TODO: cleanup
		panic(err)
	}
	http.Error(writer, string(value), statusCode)
}
