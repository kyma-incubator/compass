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

// RequestBody contains the request input of the formation mapping async request
type RequestBody struct {
	State         model.FormationAssignmentState `json:"state"`
	Configuration json.RawMessage                `json:"configuration,omitempty"`
	Error         string                         `json:"error,omitempty"`
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

// UpdateStatus handles asynchronous formation mapping update operations
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)
	errResp := errors.Errorf("An unexpected error occurred while processing the request. X-Request-Id: %s", correlationID)

	var reqBody RequestBody
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

	log.C(ctx).Info("Validating request body...")
	if err = reqBody.Validate(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating the request body")
		respondWithError(ctx, w, http.StatusBadRequest, errors.Errorf("Request Body contains invalid input: %s. X-Request-Id: %s", err.Error(), correlationID))
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

	if fa.LastOperation == model.UnassignFormation {
		log.C(ctx).Infof("Processing formation assignment asynchronous status update for %q operation...", model.UnassignFormation)
		isFADeleted, err := h.processFormationAssignmentAsynchronousUnassign(ctx, fa, reqBody)
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
			if err = h.processFormationAsynchronousUnassign(ctx, fa); err != nil {
				log.C(ctx).WithError(err).Error("An error occurred while unassigning from formation")
				respondWithError(ctx, w, http.StatusInternalServerError, errResp)
				return
			}
		}

		log.C(ctx).Infof("The formation assignment with ID: %q was successfully processed asynchronously for %q operation", formationAssignmentID, model.UnassignFormation)
		httputils.Respond(w, http.StatusOK)
		return
	}

	if reqBody.State != model.CreateErrorAssignmentState && reqBody.State != model.ReadyAssignmentState && reqBody.State != model.ConfigPendingAssignmentState {
		log.C(ctx).Errorf("An invalid state: %q is provided for %s operation", reqBody.State, model.AssignFormation)
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

	fa.State = string(reqBody.State)
	if len(reqBody.Configuration) > 0 {
		fa.Value = reqBody.Configuration
	}

	log.C(ctx).Infof("Updating formation assignment with ID: %q and formation ID: %q with state: %q", formationAssignmentID, formationID, reqBody.State)
	err = h.faService.Update(ctx, formationAssignmentID, h.faConverter.ToInput(fa))
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while updating formation assignment with ID: %q and formation ID: %q to state: %q", formationAssignmentID, formationID, reqBody.State)
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	if len(reqBody.Configuration) == 0 { // do not generate formation assignment notifications when configuration is not provided
		if err = tx.Commit(); err != nil {
			log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		}

		log.C(ctx).Infof("The formation assignment with ID: %q and formation ID: %q was successfully updated with state: %q", formationAssignmentID, formationID, reqBody.State)
		httputils.Respond(w, http.StatusOK)
		return
	}

	log.C(ctx).Infof("Generating formation assignment notifications for ID: %q and formation ID: %q about last initiator ID: %q and type: %q", fa.ID, fa.FormationID, fa.LastOperationInitiator, fa.LastOperationInitiatorType)
	notificationReq, err := h.faNotificationService.GenerateNotification(ctx, fa)
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

	log.C(ctx).Infof("Generating reverse formation assignment notifications for ID: %q and formation ID: %q about last initiator ID: %q and type: %q", fa.ID, fa.FormationID, fa.LastOperationInitiator, fa.LastOperationInitiatorType)
	reverseNotificationReq, err := h.faNotificationService.GenerateNotification(ctx, reverseFA)
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
	err = h.faService.ProcessFormationAssignmentPair(ctx, &assignmentPair)
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

func (h *Handler) processFormationAsynchronousUnassign(ctx context.Context, fa *model.FormationAssignment) error {
	unassignTx, err := h.transact.Begin()
	if err != nil {
		return errors.Wrapf(err, "while betinning transaction")
	}
	defer h.transact.RollbackUnlessCommitted(ctx, unassignTx)
	unassignCtx := persistence.SaveToContext(ctx, unassignTx)
	if err = h.unassignObjectFromFormationWhenThereAreNoFormationAssignments(err, unassignCtx, fa, fa.Source, fa.SourceType); err != nil {
		return errors.Wrapf(err, "while unassigning object with type: %q and ID: %q", fa.SourceType, fa.Source)
	}
	if err = h.unassignObjectFromFormationWhenThereAreNoFormationAssignments(err, unassignCtx, fa, fa.Target, fa.TargetType); err != nil {
		return errors.Wrapf(err, "while unassigning object with type: %q and ID: %q", fa.TargetType, fa.Target)
	}
	if err = unassignTx.Commit(); err != nil {
		return errors.Wrapf(err, "while commiting transaction")
	}
	return nil
}

func (h *Handler) unassignObjectFromFormationWhenThereAreNoFormationAssignments(err error, unassignCtx context.Context, fa *model.FormationAssignment, objectID string, objectType model.FormationAssignmentType) error {
	formationAssignmentsForObject, err := h.faService.ListFormationAssignmentsForObjectID(unassignCtx, fa.FormationID, objectID)
	if err != nil {
		return errors.Wrapf(err, "while listing formation assignments for object with type: %q and ID: %q", objectType, objectID)
	}

	if len(formationAssignmentsForObject) == 0 { // if there are no formation assignments left after the deletion, execute formation unassign for the last operation initiator
		formation, err := h.formationService.Get(unassignCtx, fa.FormationID)
		if err != nil {
			return errors.Wrapf(err, "while getting formation from formation assignment with ID: %q", fa.FormationID)
		}

		log.C(unassignCtx).Infof("Unassining formation with name: %q for object with ID: %q and type: %q", formation.Name, objectID, objectType)
		f, err := h.formationService.UnassignFormation(unassignCtx, fa.TenantID, objectID, graphql.FormationObjectType(objectType), *formation)
		if err != nil {
			return errors.Wrapf(err, "while unassigning formation with name: %q for object ID: %q and type: %q", formation.Name, objectID, objectType)
		}
		log.C(unassignCtx).Infof("Object with type: %q and ID: %q was successfully unassigned from formation with name: %q", objectType, objectID, f.Name)
	}
	return nil
}

// Validate validates the request body input
func (b RequestBody) Validate() error {
	var fieldRules []*validation.FieldRules
	fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(model.ReadyAssignmentState, model.CreateErrorAssignmentState, model.DeleteErrorAssignmentState, model.ConfigPendingAssignmentState)))

	if b.Error != "" {
		fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(model.CreateErrorAssignmentState, model.DeleteErrorAssignmentState)))
		fieldRules = append(fieldRules, validation.Field(&b.Configuration, validation.Empty))
		return validation.ValidateStruct(&b, fieldRules...)
	} else if len(b.Configuration) > 0 {
		fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(model.ReadyAssignmentState, model.ConfigPendingAssignmentState)))
		fieldRules = append(fieldRules, validation.Field(&b.Error, validation.Empty))
		return validation.ValidateStruct(&b, fieldRules...)
	} else {
		return errors.New("The request body cannot contains only state")
	}
}

// processFormationAssignmentAsynchronousUnassign handles the async unassign formation assignment status update
func (h *Handler) processFormationAssignmentAsynchronousUnassign(ctx context.Context, fa *model.FormationAssignment, reqBody RequestBody) (bool, error) {
	if reqBody.State != model.DeleteErrorAssignmentState && reqBody.State != model.ReadyAssignmentState {
		return false, errors.Errorf("An invalid state: %q is provided for %s operation", reqBody.State, model.UnassignFormation)
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
