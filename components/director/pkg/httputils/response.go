package httputils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// RespondWithBody writes a http response using with the JSON encoded data as payload
func RespondWithBody(ctx context.Context, w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add(HeaderContentTypeKey, ContentTypeApplicationJSON)
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to decode error response: %v", err)
	}
}

// RespondWithError writes a http response using with the JSON encoded error wrapped in an Error struct
func RespondWithError(ctx context.Context, w http.ResponseWriter, status int, err error) {
	log.C(ctx).WithError(err).Errorf("Responding with error: %v", err)
	w.Header().Add(HeaderContentTypeKey, ContentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := ErrorResponse{[]Error{{Message: err.Error()}}}
	encodingErr := json.NewEncoder(w).Encode(errorResponse)
	if encodingErr != nil {
		log.C(ctx).WithError(err).Errorf("Failed to encode error response: %v", err)
	}
}

// Respond writes a http response only with status, without body
func Respond(w http.ResponseWriter, status int) {
	w.Header().Add(HeaderContentTypeKey, ContentTypeApplicationJSON)
	w.WriteHeader(status)
}
