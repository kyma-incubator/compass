package httputils

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

func RespondWithBody(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add(HeaderContentType, ContentTypeApplicationJSON)
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		logrus.Error("Failed to decode error response")
	}
}

func RespondWithError(w http.ResponseWriter, status int, err error) {
	logrus.Error(err.Error())
	w.Header().Add(HeaderContentType, ContentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := ErrorResponse{err}
	encodingErr := json.NewEncoder(w).Encode(errorResponse)
	if encodingErr != nil {
		logrus.Error("Failed to encode error response")
	}
}
