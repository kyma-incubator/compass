package httphelpers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
)

const (
	HeaderContentTypeKey   = "Content-Type"
	HeaderContentTypeValue = "application/json;charset=UTF-8"
)

func WriteError(writer http.ResponseWriter, errMsg error, statusCode int) {
	writer.Header().Set(HeaderContentTypeKey, HeaderContentTypeValue)

	response := model.ErrorResponse{
		Error: errMsg.Error(),
	}

	value, err := json.Marshal(&response)
	if err != nil {
		log.Fatalf("while wriiting error message: %s, while marshalling %s ", errMsg.Error(), err.Error())
	}
	http.Error(writer, string(value), statusCode)
}
