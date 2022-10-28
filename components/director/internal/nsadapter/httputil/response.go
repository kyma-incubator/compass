package httputil

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// RespondWithError sends error response with the provided status code and error message
func RespondWithError(ctx context.Context, w http.ResponseWriter, status int, err error) {
	log.C(ctx).WithError(err).Errorf("Responding with error: %v", err)
	w.Header().Add(httputils.HeaderContentTypeKey, httputils.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := ErrorResponse{Error: err}
	encodingErr := json.NewEncoder(w).Encode(errorResponse)
	if encodingErr != nil {
		log.C(ctx).WithError(err).Errorf("Failed to encode error response: %v", err)
	}
}
