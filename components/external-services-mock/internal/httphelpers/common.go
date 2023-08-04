package httphelpers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	AuthorizationHeaderKey           = "Authorization"
	ContentTypeHeaderKey             = "Content-Type"
	ContentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeApplicationJSON       = "application/json;charset=UTF-8"
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
		log.D().Fatalf("while writing error message: %s, while marshalling %s ", errMsg.Error(), err.Error())
	}
	http.Error(writer, string(value), statusCode)
}

func RespondWithError(ctx context.Context, writer http.ResponseWriter, logErr error, respErrMsg, correlationID string, statusCode int) {
	log.C(ctx).Error(logErr)
	WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", respErrMsg, correlationID), statusCode)
}
