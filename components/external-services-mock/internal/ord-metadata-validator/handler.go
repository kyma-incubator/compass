package ord_metadata_validator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

const (
	respErrorMsg = "An unexpected error occurred while processing the request"
)

// Handler is responsible to mock and handle any Destination Service requests
type Handler struct {
	ORDValidationErrors []model.ValidationResult
}

// NewHandler creates a new Handler
func NewHandler() *Handler {
	return &Handler{
		ORDValidationErrors: make([]model.ValidationResult, 0),
	}
}

func (h *Handler) CreateValidationErrors(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if r.Header.Get(httphelpers.ContentTypeHeaderKey) != httphelpers.ContentTypeApplicationJSON {
		respondWithHeader(ctx, writer, fmt.Sprintf("Unsupported media type, expected: %s got: %s", httphelpers.ContentTypeApplicationJSON, r.Header.Get(httphelpers.ContentTypeHeaderKey)), http.StatusUnsupportedMediaType)
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while reading destination request body"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}

	var validationErrors []model.ValidationResult
	if err := json.Unmarshal(bodyBytes, &validationErrors); err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while unmarshalling validation errors"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}

	h.ORDValidationErrors = validationErrors

	httputils.Respond(writer, http.StatusCreated)
}

func (h *Handler) DeleteValidationErrors(writer http.ResponseWriter, r *http.Request) {
	h.ORDValidationErrors = nil

	httputils.Respond(writer, http.StatusOK)
}

func (h *Handler) Validate(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if r.Header.Get(httphelpers.ContentTypeHeaderKey) != httphelpers.ContentTypeApplicationJSON {
		respondWithHeader(ctx, writer, fmt.Sprintf("Unsupported media type, expected: %s got: %s", httphelpers.ContentTypeApplicationJSON, r.Header.Get(httphelpers.ContentTypeHeaderKey)), http.StatusUnsupportedMediaType)
	}

	bodyBytes, err := json.Marshal(h.ORDValidationErrors)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while marshalling ORD Validation errors"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(bodyBytes)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while writing response"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}
}

func respondWithHeader(ctx context.Context, writer http.ResponseWriter, logErrMsg string, statusCode int) {
	log.C(ctx).Error(logErrMsg)
	writer.WriteHeader(statusCode)
	return
}
