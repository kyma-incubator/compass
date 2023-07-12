package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime/middleware/header"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strings"
)

// SuccessResponse structure used for JSON encoded success response
type SuccessResponse struct {
	State string `json:"state,omitempty"`
}

// ErrorResponse structure used for JSON encoded error response
type ErrorResponse struct {
	State   string `json:"state,omitempty"`
	Message string `json:"error"`
}

func decodeJSONBody(r *http.Request, dst interface{}) error {
	if r.Header.Get(httputils.HeaderContentTypeKey) != "" {
		if value, _ := header.ParseValueAndParams(r.Header, httputils.HeaderContentTypeKey); value != httputils.ContentTypeApplicationJSON {
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: "Content-Type header is not application/json"}
		}
	}

	r.Body = io.NopCloser(r.Body)

	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			return &malformedRequest{status: http.StatusBadRequest, msg: "Request body contains badly-formed JSON"}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			return &malformedRequest{status: http.StatusBadRequest, msg: "Request body must not be empty"}

		case err.Error() == "http: request body too large":
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: "Request body must not be larger than 1MB"}

		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return &malformedRequest{status: http.StatusBadRequest, msg: "Request body must only contain a single JSON object"}
	}

	return nil
}

// respondWithError writes a http response using with the JSON error wrapped in an ErrorResponse struct
func respondWithError(ctx context.Context, w http.ResponseWriter, status int, state string, err error) {
	log.C(ctx).Error(err.Error())
	w.Header().Add(httputils.HeaderContentTypeKey, httputils.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := ErrorResponse{State: state, Message: err.Error()}
	httputils.RespondWithBody(ctx, w, status, errorResponse)
}

// respondWithSuccess writes a http response using with the JSON success wrapped in an SuccessResponse struct
func respondWithSuccess(ctx context.Context, w http.ResponseWriter, state, msg string) {
	log.C(ctx).Info(msg)
	w.Header().Add(httputils.HeaderContentTypeKey, httputils.ContentTypeApplicationJSON)
	w.WriteHeader(successStatusCode)
	successResponse := SuccessResponse{State: state}
	httputils.RespondWithBody(ctx, w, successStatusCode, successResponse)
}
