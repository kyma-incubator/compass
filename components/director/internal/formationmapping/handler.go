package formationmapping

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
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

type malformedRequest struct {
	status int
	msg    string
}

// FormationAssignmentRequestBody contains the request input of the formation assignment async status request
type FormationAssignmentRequestBody struct {
	State         model.FormationAssignmentState `json:"state,omitempty"`
	Configuration json.RawMessage                `json:"configuration,omitempty"`
	Error         string                         `json:"error,omitempty"`
}

// FormationRequestBody contains the request input of the formation async status request
type FormationRequestBody struct {
	State model.FormationState `json:"state"`
	Error string               `json:"error,omitempty"`
}

// Handler is the base struct definition of the FormationMappingHandler
type Handler struct {
	transact              persistence.Transactioner
	faConverter           formationAssignmentConverter
	faService             FormationAssignmentService
	faNotificationService FormationAssignmentNotificationService
	formationService      formationService
}

// NewFormationMappingHandler creates a formation mapping Handler
func NewFormationMappingHandler(transact persistence.Transactioner, faConverter formationAssignmentConverter, faService FormationAssignmentService, faNotificationService FormationAssignmentNotificationService, formationService formationService) *Handler {
	return &Handler{
		transact:              transact,
		faConverter:           faConverter,
		faService:             faService,
		faNotificationService: faNotificationService,
		formationService:      formationService,
	}
}

// UpdateFormationAssignmentStatus handles asynchronous formation assignment status updates
func (h *Handler) UpdateFormationAssignmentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)
	errResp := errors.Errorf("An unexpected error occurred while processing the request. X-Request-Id: %s", correlationID)

	var reqBody FormationAssignmentRequestBody
	err := decodeJSONBody(w, r, &reqBody)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			log.C(ctx).Error(mr.msg)
			respondWithError(ctx, w, mr.status, mr)
		} else {
			log.C(ctx).Error(err.Error())
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		}
		return
	}

	log.C(ctx).Info("Validating formation assignment request body...")
	if err = reqBody.Validate(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating the request body")
		respondWithError(ctx, w, http.StatusBadRequest, errors.Errorf("Request Body contains invalid input: %q. X-Request-Id: %s", err.Error(), correlationID))
		return
	}

	routeVars := mux.Vars(r)
	formationID := routeVars[FormationIDParam]
	formationAssignmentID := routeVars[FormationAssignmentIDParam]

	if formationID == "" || formationAssignmentID == "" {
		log.C(ctx).Errorf("Missing required parameters: %q or/and %q", FormationIDParam, FormationAssignmentIDParam)
		respondWithError(ctx, w, http.StatusBadRequest, errors.Errorf("Not all of the required parameters are provided. X-Request-Id: %s", correlationID))
		return
	}

	tx, err := h.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("unable to establish connection with database")
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	fa, err := h.faService.GetGlobalByIDAndFormationID(ctx, formationAssignmentID, formationID)
	if err != nil {
		log.C(ctx).Error(err)
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	ctx = tenant.SaveToContext(ctx, fa.TenantID, "")

	formation, err := h.formationService.Get(ctx, formationID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while getting formation from formation assignment with ID: %q", fa.FormationID)
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	if len(reqBody.State) > 0 && formation.State != model.ReadyFormationState {
		log.C(ctx).WithError(err).Errorf("Cannot update formation assignment for formation with ID %q as formation is not in %q state. X-Request-Id: %s", fa.FormationID, model.ReadyFormationState, correlationID)
		respondWithError(ctx, w, http.StatusBadRequest, errResp)
		return
	}

	if fa.Source == fa.Target && fa.SourceType == fa.TargetType {
		errResp := errors.Errorf("Cannot update formation assignment with source %q and target %q. X-Request-Id: %s", fa.Source, fa.Target, correlationID)
		respondWithError(ctx, w, http.StatusBadRequest, errResp)
		return
	}

	if fa.State == string(model.DeletingAssignmentState) {
		log.C(ctx).Infof("Processing formation assignment asynchronous status update for %q operation...", model.UnassignFormation)
		isFADeleted, err := h.processAsynchronousFormationAssignmentUnassign(ctx, fa, reqBody)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while processing asynchronously formation assignment for %q operation", model.UnassignFormation)
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		if err = tx.Commit(); err != nil {
			log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		if isFADeleted {
			if err = h.processAsynchronousFormationUnassign(ctx, formation, fa); err != nil {
				log.C(ctx).WithError(err).Error("An error occurred while unassigning from formation")
				respondWithError(ctx, w, http.StatusInternalServerError, errResp)
				return
			}
		}

		log.C(ctx).Infof("The formation assignment with ID: %q was successfully processed asynchronously for %q operation", formationAssignmentID, model.UnassignFormation)
		httputils.Respond(w, http.StatusOK)
		return
	}

	if len(reqBody.State) > 0 && reqBody.State != model.CreateErrorAssignmentState && reqBody.State != model.ReadyAssignmentState && reqBody.State != model.ConfigPendingAssignmentState {
		log.C(ctx).Errorf("An invalid state: %q is provided for %q operation", reqBody.State, model.AssignFormation)
		respondWithError(ctx, w, http.StatusBadRequest, errors.Errorf("An invalid state: %s is provided for %s operation. X-Request-Id: %s", reqBody.State, model.AssignFormation, correlationID))
		return
	}

	if reqBody.State == model.CreateErrorAssignmentState {
		err = h.faService.SetAssignmentToErrorState(ctx, fa, reqBody.Error, formationassignment.ClientError, reqBody.State)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("while updating error state to: %s for formation assignment with ID: %q", reqBody.State, formationAssignmentID)
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
	}

	if len(reqBody.State) > 0 {
		fa.State = string(reqBody.State)
	} else {
		log.C(ctx).Infof("State is not provided, proceeding with the current state of the FA %q", fa.State)
	}

	if len(reqBody.Configuration) > 0 {
		fa.Value = reqBody.Configuration
	}

	log.C(ctx).Infof("Updating formation assignment with ID: %q and formation ID: %q with state: %q", formationAssignmentID, formationID, fa.State)
	err = h.faService.Update(ctx, formationAssignmentID, h.faConverter.ToInput(fa))
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while updating formation assignment with ID: %q and formation ID: %q with state: %q", formationAssignmentID, formationID, fa.State)
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	if len(reqBody.Configuration) == 0 { // do not generate formation assignment notifications when configuration is not provided
		if err = tx.Commit(); err != nil {
			log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		}

		log.C(ctx).Infof("The formation assignment with ID: %q and formation ID: %q was successfully updated with state: %q", formationAssignmentID, formationID, fa.State)
		httputils.Respond(w, http.StatusOK)
		return
	}

	log.C(ctx).Infof("Generating formation assignment notifications for ID: %q and formation ID: %q", fa.ID, fa.FormationID)
	notificationReq, err := h.faNotificationService.GenerateFormationAssignmentNotification(ctx, fa)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while generating formation assignment notifications for ID: %q and formation ID: %q", formationAssignmentID, formationID)
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	reverseFA, err := h.faService.GetReverseBySourceAndTarget(ctx, fa.FormationID, fa.Source, fa.Target)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while getting reverse formation assignment by source: %q and target: %q", fa.Source, fa.Target)
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	log.C(ctx).Infof("Generating reverse formation assignment notifications for ID: %q and formation ID: %q", reverseFA.ID, reverseFA.FormationID)
	reverseNotificationReq, err := h.faNotificationService.GenerateFormationAssignmentNotification(ctx, reverseFA)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while generating reverse formation assignment notifications for ID: %q and formation ID: %q", formationAssignmentID, formationID)
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
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
		Assignment:        &reverseFAReqMapping, // the status update call is a response to the original notification that's why here we switch the assignment and reverse assignment
		ReverseAssignment: &faReqMapping,
	}

	log.C(ctx).Infof("Processing formation assignment pairs and their notifications")
	_, err = h.faService.ProcessFormationAssignmentPair(ctx, &assignmentPair)
	if err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while processing formation assignment pairs and their notifications")
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
	}

	log.C(ctx).Info("The formation assignment notifications are successfully processed")
	httputils.Respond(w, http.StatusOK)
}

// UpdateFormationStatus handles asynchronous formation status updates
func (h *Handler) UpdateFormationStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)
	errResp := errors.Errorf("An unexpected error occurred while processing the request. X-Request-Id: %s", correlationID)

	var reqBody FormationRequestBody
	err := decodeJSONBody(w, r, &reqBody)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			log.C(ctx).Error(mr.msg)
			respondWithError(ctx, w, mr.status, mr)
		} else {
			log.C(ctx).Error(err.Error())
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		}
		return
	}

	log.C(ctx).Info("Validating formation request body...")
	if err = reqBody.Validate(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating the request body")
		respondWithError(ctx, w, http.StatusBadRequest, errors.Errorf("Request Body contains invalid input: %q. X-Request-Id: %s", err.Error(), correlationID))
		return
	}

	routeVars := mux.Vars(r)
	formationID := routeVars[FormationIDParam]
	if formationID == "" {
		log.C(ctx).Errorf("Missing required parameters: %q", FormationIDParam)
		respondWithError(ctx, w, http.StatusBadRequest, errors.Errorf("Not all of the required parameters are provided. X-Request-Id: %s", correlationID))
		return
	}

	tx, err := h.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("unable to establish connection with database")
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	f, err := h.formationService.GetGlobalByID(ctx, formationID)
	if err != nil {
		log.C(ctx).Error(err)
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	ctx = tenant.SaveToContext(ctx, f.TenantID, "")

	if f.State == model.DeletingFormationState {
		log.C(ctx).Infof("Processing asynchronous formation status update for %q operation...", model.DeleteFormation)
		err = h.processAsynchronousFormationDelete(ctx, f, reqBody)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while processing asynchronous formation status update for %q operation. X-Request-Id: %s", model.DeleteFormation, correlationID)
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		if err = tx.Commit(); err != nil {
			log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		log.C(ctx).Infof("The status update for formation with ID: %q was successfully processed asynchronously for %q operation", formationID, model.DeleteFormation)
		httputils.Respond(w, http.StatusOK)
		return
	}

	log.C(ctx).Infof("Processing asynchronous formation status update for %q operation...", model.CreateFormation)
	err = h.processAsynchronousFormationCreate(ctx, f, reqBody)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while processing asynchronous formation status update for %q operation. X-Request-Id: %s", model.CreateFormation, correlationID)
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	log.C(ctx).Infof("The status update for formation with ID: %q was successfully processed asynchronously for %q operation", formationID, model.CreateFormation)
	httputils.Respond(w, http.StatusOK)
}

func (h *Handler) processAsynchronousFormationUnassign(ctx context.Context, formation *model.Formation, fa *model.FormationAssignment) error {
	unassignTx, err := h.transact.Begin()
	if err != nil {
		return errors.Wrapf(err, "while betinning transaction")
	}
	defer h.transact.RollbackUnlessCommitted(ctx, unassignTx)

	unassignCtx := persistence.SaveToContext(ctx, unassignTx)

	if err = h.unassignObjectFromFormationWhenThereAreNoFormationAssignments(unassignCtx, fa, formation, fa.Source, fa.SourceType); err != nil {
		return errors.Wrapf(err, "while unassigning object with type: %q and ID: %q", fa.SourceType, fa.Source)
	}

	if err = h.unassignObjectFromFormationWhenThereAreNoFormationAssignments(unassignCtx, fa, formation, fa.Target, fa.TargetType); err != nil {
		return errors.Wrapf(err, "while unassigning object with type: %q and ID: %q", fa.TargetType, fa.Target)
	}

	if err = unassignTx.Commit(); err != nil {
		return errors.Wrapf(err, "while committing transaction")
	}

	return nil
}

func (h *Handler) unassignObjectFromFormationWhenThereAreNoFormationAssignments(ctx context.Context, fa *model.FormationAssignment, formation *model.Formation, objectID string, objectType model.FormationAssignmentType) error {
	formationAssignmentsForObject, err := h.faService.ListFormationAssignmentsForObjectID(ctx, fa.FormationID, objectID)
	if err != nil {
		return errors.Wrapf(err, "while listing formation assignments for object with type: %q and ID: %q", objectType, objectID)
	}

	// if there are no formation assignments left after the deletion, execute unassign formation for the object
	if len(formationAssignmentsForObject) == 0 {
		log.C(ctx).Infof("Unassining formation with name: %q for object with ID: %q and type: %q", formation.Name, objectID, objectType)
		f, err := h.formationService.UnassignFormation(ctx, fa.TenantID, objectID, graphql.FormationObjectType(objectType), *formation)
		if err != nil {
			return errors.Wrapf(err, "while unassigning formation with name: %q for object ID: %q and type: %q", formation.Name, objectID, objectType)
		}
		log.C(ctx).Infof("Object with type: %q and ID: %q was successfully unassigned from formation with name: %q", objectType, objectID, f.Name)
	}
	return nil
}

// Validate validates the formation assignment's request body input
func (b FormationAssignmentRequestBody) Validate() error {
	var fieldRules []*validation.FieldRules
	fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(model.ReadyAssignmentState, model.CreateErrorAssignmentState, model.DeleteErrorAssignmentState, model.ConfigPendingAssignmentState)))

	if b.Error != "" {
		fieldRules = make([]*validation.FieldRules, 0)
		fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(model.CreateErrorAssignmentState, model.DeleteErrorAssignmentState)))
		fieldRules = append(fieldRules, validation.Field(&b.Configuration, validation.Empty))
		return validation.ValidateStruct(&b, fieldRules...)
	} else if len(b.Configuration) > 0 {
		fieldRules = make([]*validation.FieldRules, 0)
		fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(model.ReadyAssignmentState, model.ConfigPendingAssignmentState)))
		fieldRules = append(fieldRules, validation.Field(&b.Error, validation.Empty))
		return validation.ValidateStruct(&b, fieldRules...)
	}

	return validation.ValidateStruct(&b, fieldRules...)
}

// Validate validates the formation's request body input
func (b FormationRequestBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.State,
			validation.When(len(b.Error) == 0, validation.In(model.ReadyFormationState, model.CreateErrorFormationState, model.DeleteErrorFormationState)).
				Else(validation.In(model.CreateErrorFormationState, model.DeleteErrorFormationState))))
}

// processAsynchronousFormationAssignmentUnassign handles the async unassign formation assignment status update
func (h *Handler) processAsynchronousFormationAssignmentUnassign(ctx context.Context, fa *model.FormationAssignment, reqBody FormationAssignmentRequestBody) (bool, error) {
	if reqBody.State != model.DeleteErrorAssignmentState && reqBody.State != model.ReadyAssignmentState {
		return false, errors.Errorf("An invalid state: %q is provided for %q operation", reqBody.State, model.UnassignFormation)
	}

	if reqBody.State == model.DeleteErrorAssignmentState {
		err := h.faService.SetAssignmentToErrorState(ctx, fa, reqBody.Error, formationassignment.ClientError, reqBody.State)
		if err != nil {
			return false, errors.Wrapf(err, "while updating error state to: %s for formation assignment with ID: %q", reqBody.State, fa.ID)
		}
		return false, nil
	}

	err := h.faService.Delete(ctx, fa.ID)
	if err != nil {
		return false, errors.Wrapf(err, "while deleting formation assignment with ID: %q", fa.ID)
	}

	return true, nil
}

// processAsynchronousFormationDelete handles the async delete formation status update
func (h *Handler) processAsynchronousFormationDelete(ctx context.Context, formation *model.Formation, reqBody FormationRequestBody) error {
	if reqBody.State != model.ReadyFormationState && reqBody.State != model.DeleteErrorFormationState {
		return errors.Errorf("An invalid state: %q is provided for %q operation", reqBody.State, model.DeleteFormation)
	}

	if reqBody.State == model.DeleteErrorFormationState {
		err := h.formationService.SetFormationToErrorState(ctx, formation, reqBody.Error, formationassignment.ClientError, reqBody.State)
		if err != nil {
			return errors.Wrapf(err, "while updating error state to: %s for formation with ID: %q", reqBody.State, formation.ID)
		}
		return nil
	}

	log.C(ctx).Infof("Deleting formation with ID: %q and name: %q", formation.ID, formation.Name)
	err := h.formationService.DeleteFormationEntityAndScenarios(ctx, formation.TenantID, formation.Name)
	if err != nil {
		return errors.Wrapf(err, "while deleting formation with ID: %q and name: %q", formation.ID, formation.Name)
	}

	return nil
}

// processAsynchronousFormationCreate handles the async create formation status update
func (h *Handler) processAsynchronousFormationCreate(ctx context.Context, formation *model.Formation, reqBody FormationRequestBody) error {
	if reqBody.State != model.ReadyFormationState && reqBody.State != model.CreateErrorFormationState {
		return errors.Errorf("An invalid state: %q is provided for %q operation", reqBody.State, model.CreateFormation)
	}

	if reqBody.State == model.CreateErrorFormationState {
		err := h.formationService.SetFormationToErrorState(ctx, formation, reqBody.Error, formationassignment.ClientError, reqBody.State)
		if err != nil {
			return errors.Wrapf(err, "while updating error state to: %s for formation with ID: %q", reqBody.State, formation.ID)
		}
		return nil
	}

	log.C(ctx).Infof("Updating formation with ID: %q and name: %q to: %q state", formation.ID, formation.Name, reqBody.State)
	formation.State = model.ReadyFormationState
	err := h.formationService.Update(ctx, formation)
	if err != nil {
		return errors.Wrapf(err, "while updating formation with ID: %q to: %q state", formation.ID, model.ReadyFormationState)
	}

	log.C(ctx).Infof("Resynchronizing formation with ID: %q and name: %q", formation.ID, formation.Name)
	_, err = h.formationService.ResynchronizeFormationNotifications(ctx, formation.ID)
	if err != nil {
		return errors.Wrapf(err, "while resynchronize formation notifications for formation with ID: %q", formation.ID)
	}

	return nil
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
