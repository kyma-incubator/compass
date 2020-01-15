package reqerror

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	HeaderContentTypeKey = "Content-Type"
	HeaderContentTypeValue = "application/json;charset=UTF-8"
)

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func WriteError(writer http.ResponseWriter, err error, appErrorCode int) {
	WriteErrorMessage(writer, err.Error(), appErrorCode)
}

func WriteErrorMessage(writer http.ResponseWriter, errMessage string, appErrorCode int) {
	writer.Header().Set(HeaderContentTypeKey, HeaderContentTypeValue)
	writer.WriteHeader(errorCodeToHTTPStatus(appErrorCode))

	response := ErrorResponse{
		Error: errMessage,
		Code: appErrorCode,
	}

	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		log.Error(errors.Wrapf(err, "while encoding JSON response body"))
	}
}

func errorCodeToHTTPStatus(code int) int {
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
	default:
		return http.StatusInternalServerError
	}
}
