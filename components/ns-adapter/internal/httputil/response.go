package httputil

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"net/http"
)

func RespondWithError(ctx context.Context, w http.ResponseWriter, status int, err error) {
	log.C(ctx).WithError(err).Errorf("Responding with error: %v", err)
	w.Header().Add(httputils.HeaderContentType, httputils.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := ErrorResponse{Errors: Error{Code: status, Message: err.Error()}}
	encodingErr := json.NewEncoder(w).Encode(errorResponse.Error())
	if encodingErr != nil {
		log.C(ctx).WithError(err).Errorf("Failed to encode error response: %v", err)
	}
}