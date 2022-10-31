package formationmapping

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/pkg/errors"

	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-openapi/runtime/middleware/header"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// ConfigurationState describes possible state values
type ConfigurationState string

const (
	// ConfigurationStateReady represents ready state of the configuration
	ConfigurationStateReady ConfigurationState = "READY"
	// ConfigurationStateCreateError represents error in the configuration creation
	ConfigurationStateCreateError ConfigurationState = "CREATE_ERROR"
	// ConfigurationStateDeleteError represents error in the configuration deletion
	ConfigurationStateDeleteError ConfigurationState = "DELETE_ERROR"
	// ConfigurationStateConfigPending represents pending state of the configuration
	ConfigurationStateConfigPending ConfigurationState = "CONFIG_PENDING"
)

type malformedRequest struct {
	status int
	msg    string
}

// RequestBody contains the request input of the formation mapping async request
type RequestBody struct {
	State         ConfigurationState `json:"state"`
	Configuration json.RawMessage    `json:"configuration,omitempty"`
	Error         string             `json:"error,omitempty"`
}

// Handler is the base struct definition of the FormationMappingHandler
type Handler struct {
	faService             FormationAssignmentService
	faNotificationService FormationAssignmentNotificationService
	formationRepository   FormationRepository
}

// NewFormationMappingHandler creates an empty formation mapping Handler
func NewFormationMappingHandler(faService FormationAssignmentService, faNotificationService FormationAssignmentNotificationService, formationRepository FormationRepository) *Handler {
	return &Handler{
		faService:             faService,
		faNotificationService: faNotificationService,
		formationRepository:   formationRepository,
	}
}

// UpdateStatus handles asynchronous formation mapping update operations
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var reqBody RequestBody
	err := decodeJSONBody(w, r, &reqBody)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			log.C(ctx).Error(mr.msg)
			respondWithError(ctx, w, mr.status, mr)
		} else {
			log.C(ctx).Error(err.Error())
			respondWithError(ctx, w, http.StatusInternalServerError, errors.New("An unexpected error occurred while processing the request"))
		}
		return
	}

	log.C(ctx).Info("Validating request body...")
	if err := reqBody.Validate(); err != nil {
		log.C(ctx).Errorf("An error occurred while validating the request body: %s", err)
		respondWithError(ctx, w, http.StatusBadRequest, errors.Errorf("Request Body contains invalid input: %s", err.Error()))
		return
	}

	routeVars := mux.Vars(r)
	formationID := routeVars[FormationIDParam]
	formationAssignmentID := routeVars[FormationAssignmentIDParam]

	if formationID == "" || formationAssignmentID == "" {
		log.C(ctx).Errorf("Missing required parameters: %q or/and %q", FormationIDParam, FormationAssignmentIDParam)
		respondWithError(ctx, w, http.StatusBadRequest, errors.New("Not all of the required parameters are provided"))
		return
	}

	log.C(ctx).Infof("Updating status of formation assignment with ID: %q for formation with ID: %q", formationAssignmentID, formationID)
	// todo:: implement business logic here

	fa, err := h.faService.GetGlobalByID(ctx, formationAssignmentID)
	if err != nil {
		respondWithError(ctx, w, http.StatusInternalServerError, errors.Errorf("while getting formation assignment with ID: %q globally", formationAssignmentID))
		return
	}

	reverseFA := fa
	reverseFA.Source = fa.Target
	reverseFA.SourceType = fa.TargetType
	reverseFA.Target = fa.Source
	reverseFA.TargetType = fa.SourceType

	formation, err := h.formationRepository.Get(ctx, fa.FormationID, fa.TenantID)
	if err != nil {
		respondWithError(ctx, w, http.StatusInternalServerError, errors.Errorf("while getting formation with ID: %q in tenant with ID: %q", fa.FormationID, fa.TenantID))
		return
	}

	notificationReq, err := h.faNotificationService.GenerateNotification(ctx, fa.TenantID, fa.Source, formation.Name, fa.SourceType, fa.TargetType)
	if err != nil {
		respondWithError(ctx, w, http.StatusInternalServerError, errors.Errorf("while generating formation assignment notification for object ID: %q and object type: %q for formation with name: %q and tenant with ID: %q", fa.Target, fa.TargetType, formation.Name, fa.TenantID))
		return
	}

	reverseNotificationReq, err := h.faNotificationService.GenerateNotification(ctx, fa.TenantID, fa.Target, formation.Name, fa.TargetType, fa.SourceType)
	if err != nil {
		respondWithError(ctx, w, http.StatusInternalServerError, errors.Errorf("while generating formation assignment notification for object ID: %q and object type: %q for formation with name: %q and tenant with ID: %q", fa.Source, fa.SourceType, formation.Name, fa.TenantID))
		return
	}

	faReqMapping := formationassignment.FormationAssignmentRequestMapping{
		Request:             notificationReq,
		FormationAssignment: fa,
	}

	reverseFAReqMapping := formationassignment.FormationAssignmentRequestMapping{
		Request:             reverseNotificationReq,
		FormationAssignment: reverseFA,
	}

	assignmentPair := formationassignment.AssignmentMappingPair{
		Assignment:        &reverseFAReqMapping,
		ReverseAssignment: &faReqMapping,
	}

	err = h.faService.UpdateFormationAssignment(ctx, &assignmentPair)
	if err != nil {
		respondWithError(ctx, w, http.StatusInternalServerError, errors.New("while updating formation assignment pair"))
		return
	}

	httputils.Respond(w, http.StatusOK)
}

// Validate validates the request body input
func (b RequestBody) Validate() error {
	var fieldRules []*validation.FieldRules
	fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(ConfigurationStateReady, ConfigurationStateCreateError, ConfigurationStateDeleteError, ConfigurationStateConfigPending)))

	if b.Error != "" {
		fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(ConfigurationStateCreateError)))
		fieldRules = append(fieldRules, validation.Field(&b.Configuration, validation.Empty))
		return validation.ValidateStruct(&b, fieldRules...)
	} else if len(b.Configuration) > 0 {
		fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(ConfigurationStateReady, ConfigurationStateConfigPending)))
		fieldRules = append(fieldRules, validation.Field(&b.Error, validation.Empty))
		return validation.ValidateStruct(&b, fieldRules...)
	} else {
		return errors.New("The request body cannot contains only state")
	}
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get(httputils.HeaderContentTypeKey) != "" {
		value, _ := header.ParseValueAndParams(r.Header, httputils.HeaderContentTypeKey)
		if value != httputils.ContentTypeApplicationJSON {
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: "Content-Type header is not application/json"}
		}
	}

	// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
	// response body. A request body larger than that will now result in
	// Decode() returning a "http: request body too large" error.
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
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

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return &malformedRequest{status: http.StatusBadRequest, msg: "Request body must only contain a single JSON object"}
	}

	return nil
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}
