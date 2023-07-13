package httphelpers

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	AuthorizationHeaderKey           = "Authorization"
	ContentTypeHeaderKey             = "Content-Type"
	ContentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeApplicationJSON       = "application/json"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteError(writer http.ResponseWriter, errMsg error, statusCode int) {
	writer.Header().Set(ContentTypeHeaderKey, ContentTypeApplicationJSON)

	response := ErrorResponse{
		Error: errMsg.Error(),
	}

	value, err := json.Marshal(&response)
	if err != nil {
		log.Fatalf("while writing error message: %s, while marshalling %s ", errMsg.Error(), err.Error())
	}
	http.Error(writer, string(value), statusCode)
}
