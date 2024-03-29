package res

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const unauthorizedErrorMessage = "insufficient scopes provided"

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func WriteError(writer http.ResponseWriter, err error, appErrorCode int) {
	WriteErrorMessage(writer, err.Error(), appErrorCode)
}

func WriteAppError(writer http.ResponseWriter, err error) {
	errCode := apperrors.CodeInternal
	appErr, ok := err.(apperrors.AppError)
	if ok {
		errCode = appErr.Code()
	}
	WriteErrorMessage(writer, err.Error(), errCode)
}

func WriteErrorMessage(writer http.ResponseWriter, errMessage string, appErrorCode int) {
	log.WithFields(log.Fields{
		"errMessage":   errMessage,
		"appErrorCode": appErrorCode,
	}).Infof("writing error...")
	writer.Header().Set(HeaderContentTypeKey, HeaderContentTypeValue)
	writer.WriteHeader(errorCodeToHTTPStatus(errMessage, appErrorCode))

	response := ErrorResponse{
		Error: errMessage,
		Code:  errorCodeToHTTPStatus(errMessage, appErrorCode),
	}

	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		log.Error(errors.Wrapf(err, "while encoding JSON response body"))
	}
}

/**
Copied from https://github.com/kyma-project/kyma/tree/main/components/application-registry
*/

func errorCodeToHTTPStatus(errMessage string, code int) int {
	if strings.Contains(errMessage, unauthorizedErrorMessage) {
		return http.StatusUnauthorized
	}

	switch code {
	case apperrors.CodeInternal:
		return http.StatusInternalServerError
	case apperrors.CodeNotFound:
		return http.StatusNotFound
	case apperrors.CodeAlreadyExists:
		return http.StatusConflict
	case apperrors.CodeWrongInput:
		return http.StatusBadRequest
	case apperrors.CodeUpstreamServerCallFailed:
		return http.StatusBadGateway
	case apperrors.CodeForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
