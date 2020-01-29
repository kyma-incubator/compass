package util

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

const (
	ContentTypeApplicationJSON = "application/json"
	HeaderContentType          = "Content-Type"
)

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	Message string `json:"message"`
}

func RespondWithError(w http.ResponseWriter, status int, err error) {
	logrus.Error(err.Error())
	w.Header().Add(HeaderContentType, ContentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := ErrorResponse{[]Error{{Message: err.Error()}}}
	encodingErr := json.NewEncoder(w).Encode(errorResponse)
	if encodingErr != nil {
		logrus.Error("Failed to encode error response")
	}
}
