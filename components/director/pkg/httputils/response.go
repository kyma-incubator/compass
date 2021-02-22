package httputils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

func RespondWithBody(ctx context.Context, w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add(HeaderContentType, ContentTypeApplicationJSON)
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to decode error response")
	}
}

func RespondWithError(ctx context.Context, w http.ResponseWriter, status int, err error) {
	log.C(ctx).WithError(err).Error("Responding with error")
	w.Header().Add(HeaderContentType, ContentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := ErrorResponse{[]Error{{Message: err.Error()}}}
	encodingErr := json.NewEncoder(w).Encode(errorResponse)
	if encodingErr != nil {
		log.C(ctx).WithError(err).Error("Failed to encode error response")
	}
}
