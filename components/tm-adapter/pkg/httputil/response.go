package httputil

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/api/types"
	"net/http"
)

const (
	ContentTypeApplicationJSON = "application/json"
	HeaderContentTypeKey       = "Content-Type"
)

// RespondWithError writes a http response using with the JSON encoded error wrapped in an ErrorResponse struct
func RespondWithError(ctx context.Context, w http.ResponseWriter, status int, err error) {
	log.C(ctx).Errorf("Responding with error: %v", err)
	w.Header().Add(httputils.HeaderContentTypeKey, httputils.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := types.ErrorResponse{Message: err.Error()}
	encodingErr := json.NewEncoder(w).Encode(errorResponse)
	if encodingErr != nil {
		log.C(ctx).WithError(err).Errorf("Failed to encode error response: %v", err)
	}
}

// RespondWithBody writes a http response using with the JSON encoded data as payload
func RespondWithBody(ctx context.Context, w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add(HeaderContentTypeKey, ContentTypeApplicationJSON)
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to encode response body: %v", err)
	}
}

func Respond(w http.ResponseWriter, status int) {
	w.Header().Add(HeaderContentTypeKey, ContentTypeApplicationJSON)
	w.WriteHeader(status)
}
