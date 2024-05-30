package formationmapping

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/internal/domain/statusreport"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationassignmentpkg "github.com/kyma-incubator/compass/components/director/pkg/formationassignment"
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

//go:generate mockery --exported --name=formationAssignmentStatusService --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationAssignmentStatusService interface {
	UpdateWithConstraints(ctx context.Context, notificationStatusReport *statusreport.NotificationStatusReport, fa *model.FormationAssignment, operation model.FormationOperation) error
	DeleteWithConstraints(ctx context.Context, id string, notificationStatusReport *statusreport.NotificationStatusReport) error
}

//go:generate mockery --exported --name=assignmentOperationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type assignmentOperationService interface {
	Create(ctx context.Context, in *model.AssignmentOperationInput) (string, error)
	Finish(ctx context.Context, assignmentID, formationID string) error
	FinishByID(ctx context.Context, operationID string) error
	GetLatestOperation(ctx context.Context, assignmentID, formationID string) (*model.AssignmentOperation, error)
	GetByID(ctx context.Context, operationID string) (*model.AssignmentOperation, error)
}

type RequestBody interface {
	GetState() model.FormationAssignmentState
	GetConfiguration() json.RawMessage
	GetError() string

	SetState(state model.FormationAssignmentState)
}

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

// AssignmentOperationRequestBody contains the request input of the formation assignment with operation async status request
type AssignmentOperationRequestBody struct {
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
	transact                   persistence.Transactioner
	faService                  FormationAssignmentService
	faStatusService            formationAssignmentStatusService
	faNotificationService      FormationAssignmentNotificationService
	assignmentOperationService assignmentOperationService
	formationService           formationService
	formationStatusService     formationStatusService
}

// NewFormationMappingHandler creates a formation mapping Handler
func NewFormationMappingHandler(transact persistence.Transactioner, faService FormationAssignmentService, faStatusService formationAssignmentStatusService, faNotificationService FormationAssignmentNotificationService, assignmentOperationService assignmentOperationService, formationService formationService, formationStatusService formationStatusService) *Handler {
	return &Handler{
		transact:                   transact,
		faService:                  faService,
		faStatusService:            faStatusService,
		faNotificationService:      faNotificationService,
		assignmentOperationService: assignmentOperationService,
		formationService:           formationService,
		formationStatusService:     formationStatusService,
	}
}

// ResetFormationAssignmentStatus handles formation assignment status updates
func (h *Handler) ResetFormationAssignmentStatus(w http.ResponseWriter, r *http.Request) {
	h.updateFormationAssignmentStatus(w, r, true)
}

// UpdateFormationAssignmentStatus handles formation assignment status updates
func (h *Handler) UpdateFormationAssignmentStatus(w http.ResponseWriter, r *http.Request) {
	h.updateFormationAssignmentStatus(w, r, false)
}

func (h *Handler) updateFormationAssignmentStatus(w http.ResponseWriter, r *http.Request, reset bool) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)
	traceID := correlation.TraceIDFromContext(ctx)
	spanID := correlation.SpanIDFromContext(ctx)
	parentSpanID := correlation.ParentSpanIDFromContext(ctx)

	errResp := errors.Errorf("An unexpected error occurred while processing the request. X-Request-Id: %s", correlationID)

	var assignmentReqBody FormationAssignmentRequestBody
	err := decodeJSONBody(w, r, &assignmentReqBody)
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

	if formationassignmentpkg.IsConfigEmpty(string(assignmentReqBody.Configuration)) {
		assignmentReqBody.Configuration = nil
	}

	log.C(ctx).Info("Validating formation assignment request body...")
	if err = assignmentReqBody.Validate(ctx); err != nil {
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

	logger := log.C(ctx).WithField(log.FieldFormationID, formationID)
	logger = logger.WithField(log.FieldFormationAssignmentID, formationAssignmentID)
	ctx = log.ContextWithLogger(ctx, logger)

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
		if apperrors.IsNotFoundError(err) {
			errResp := errors.Errorf("Formation assignment with ID %q was not found. X-Request-Id: %s", formationAssignmentID, correlationID)
			respondWithError(ctx, w, http.StatusNotFound, errResp)
			return
		}
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	ctx = tenant.SaveToContext(ctx, fa.TenantID, "")

	formation, err := h.formationService.Get(ctx, formationID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while getting formation from formation assignment with ID: %q", fa.FormationID)
		if apperrors.IsNotFoundError(err) {
			errResp := errors.Errorf("Formation with ID %q was not found. X-Request-Id: %s", formationID, correlationID)
			respondWithError(ctx, w, http.StatusNotFound, errResp)
			return
		}
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	// Only customers of type BI can push configurations with empty state even if the formation is not yet ready.
	// This mechanism is used to set up initial configuration before sending the first wave of notifications
	if len(assignmentReqBody.State) > 0 && formation.State != model.ReadyFormationState {
		log.C(ctx).WithError(err).Errorf("Cannot update formation assignment for formation with ID %q as formation is not in %q state. X-Request-Id: %s", fa.FormationID, model.ReadyFormationState, correlationID)
		respondWithError(ctx, w, http.StatusBadRequest, errResp)
		return
	}

	latestAssignmentOperation, err := h.assignmentOperationService.GetLatestOperation(ctx, fa.ID, fa.FormationID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error when getting latest operation for assignment with ID: %s. X-Request-ID: %s", fa.ID, correlationID)
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	formationOperation := formationassignmentpkg.DetermineFormationOperationFromLatestAssignmentOperation(latestAssignmentOperation.Type)
	notificationStatusReport := newNotificationStatusReportFromRequestBody(assignmentReqBody, fa, latestAssignmentOperation.Type)
	originalStateFromStatusReport := notificationStatusReport.State
	if !isStateSupportedForOperation(ctx, model.FormationAssignmentState(originalStateFromStatusReport), formationOperation, formation.State, reset) {
		log.C(ctx).Errorf("An invalid state: %q is provided for %q operation with reset option %t", originalStateFromStatusReport, formationOperation, reset)
		errResp := errors.Errorf("An invalid state: %s is provided for %s operation. X-Request-Id: %s", originalStateFromStatusReport, formationOperation, correlationID)
		respondWithError(ctx, w, http.StatusBadRequest, errResp)
		return
	}

	changeToRegularReadyStateInStatusReport(notificationStatusReport)
	stateFromStatusReport := notificationStatusReport.State

	if reset {
		if stateFromStatusReport != string(model.ReadyAssignmentState) && stateFromStatusReport != string(model.ConfigPendingAssignmentState) {
			errResp := errors.Errorf("Cannot reset formation assignment with source %q and target %q to state %s. X-Request-Id: %s", fa.Source, fa.Target, assignmentReqBody.State, correlationID)
			respondWithError(ctx, w, http.StatusBadRequest, errResp)
			return
		}

		if formationassignmentpkg.IsConfigEmpty(string(notificationStatusReport.Configuration)) {
			errResp := errors.Errorf("Cannot reset formation assignment with source %q and target %q to state %s because provided configuration is empty. X-Request-Id: %s", fa.Source, fa.Target, assignmentReqBody.State, correlationID)
			respondWithError(ctx, w, http.StatusBadRequest, errResp)
			return
		}

		if fa.State != string(model.ReadyAssignmentState) {
			errResp := errors.Errorf("Cannot reset formation assignment with source %q and target %q because assignment is not in %q state. X-Request-Id: %s", fa.Source, fa.Target, model.ReadyAssignmentState, correlationID)
			respondWithError(ctx, w, http.StatusBadRequest, errResp)
			return
		}
		reverseFA, err := h.faService.GetReverseBySourceAndTarget(ctx, fa.FormationID, fa.Source, fa.Target)
		if err != nil {
			log.C(ctx).Error(err)
			if apperrors.IsNotFoundError(err) {
				errResp := errors.Errorf("Cannot reset formation assignment with source %q and target %q because reverse assignment is missing. X-Request-Id: %s", fa.Source, fa.Target, correlationID)
				respondWithError(ctx, w, http.StatusBadRequest, errResp)
				return
			}
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
		if reverseFA.State != string(model.ReadyAssignmentState) {
			errResp := errors.Errorf("Cannot reset formation assignment with source %q and target %q because reverse assignment is not in %q state. X-Request-Id: %s", reverseFA.Source, reverseFA.Target, model.ReadyAssignmentState, correlationID)
			respondWithError(ctx, w, http.StatusBadRequest, errResp)
			return
		}

		log.C(ctx).Infof("Resetting formation assignment with ID: %s to state: %s", fa.ID, stateFromStatusReport)
		fa.State = stateFromStatusReport
		if err = h.faService.Update(ctx, fa.ID, fa); err != nil {
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
		log.C(ctx).Infof("Creating %s Operation for formation assignment with ID: %s triggered by reset on the status API", model.Assign, fa.ID)
		if _, err = h.assignmentOperationService.Create(ctx, &model.AssignmentOperationInput{
			Type:                  model.Assign,
			FormationAssignmentID: fa.ID,
			FormationID:           fa.FormationID,
			TriggeredBy:           model.ResetAssignment,
		}); err != nil {
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		log.C(ctx).Infof("Resetting reverse formation assignment with ID: %s to state: %s", reverseFA.ID, model.InitialAssignmentState)
		reverseFA.State = string(model.InitialAssignmentState)
		if err = h.faService.Update(ctx, reverseFA.ID, reverseFA); err != nil {
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
		log.C(ctx).Infof("Creating %s Operation for reverse formation assignment with ID: %s triggered by reset on the status API", model.Assign, fa.ID)
		if _, err = h.assignmentOperationService.Create(ctx, &model.AssignmentOperationInput{
			Type:                  model.Assign,
			FormationAssignmentID: reverseFA.ID,
			FormationID:           reverseFA.FormationID,
			TriggeredBy:           model.ResetAssignment,
		}); err != nil {
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
	}

	if formationOperation == model.UnassignFormation {
		log.C(ctx).Infof("Processing status update for formation assignment with ID: %s during %q operation", fa.ID, model.UnassignFormation)
		isFADeleted, err := h.processFormationAssignmentUnassignStatusUpdate(ctx, fa, notificationStatusReport, latestAssignmentOperation.Type)

		if commitErr := tx.Commit(); commitErr != nil {
			log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while processing formation assignment status update for %q operation", model.UnassignFormation)
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		if isFADeleted {
			if err = h.processFormationUnassign(ctx, formation, fa); err != nil {
				log.C(ctx).WithError(err).Error("An error occurred while unassigning from formation")
				respondWithError(ctx, w, http.StatusInternalServerError, errResp)
				return
			}
		}

		log.C(ctx).Infof("The formation assignment with ID: %q was successfully processed for %q operation", formationAssignmentID, model.UnassignFormation)
		httputils.Respond(w, http.StatusOK)
		return
	}

	log.C(ctx).Infof("Processing status update for formation assignment with ID: %s during %q operation", fa.ID, model.AssignFormation)
	shouldProcessNotifications, errorResponse := h.processFormationAssignmentAssignStatusUpdate(ctx, fa, notificationStatusReport, correlationID)
	if errorResponse != nil {
		respondWithError(ctx, w, errorResponse.statusCode, errors.New(errorResponse.errorMessage))
		return
	}

	if fa.State == string(model.ReadyAssignmentState) {
		log.C(ctx).Infof("Finish %s Operation for assignment with ID: %s during status report", model.Assign, fa.ID)
		if finishOpErr := h.assignmentOperationService.Finish(ctx, fa.ID, fa.FormationID); finishOpErr != nil {
			log.C(ctx).WithError(finishOpErr).Errorf("An error occurred while finishing %s Operation for formation assignment with ID: %q", model.Assign, fa.ID)
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}
	log.C(ctx).Infof("The formation assignment with ID: %q and formation ID: %q was successfully updated with state: %q", formationAssignmentID, formationID, fa.State)

	if formationassignmentpkg.IsConfigEmpty(string(notificationStatusReport.Configuration)) { // do not generate formation assignment notifications when configuration is empty
		log.C(ctx).Info("The configuration in the request body is empty. Formation assignment notification won't be generated")
		httputils.Respond(w, http.StatusOK)
		return
	}

	if shouldProcessNotifications {
		// The formation assignment notifications processing is independent of the status update request handling.
		// That's why we're executing it in a go routine and in parallel to this returning a response to the client
		go h.processFormationAssignmentNotifications(fa, correlationID, traceID, spanID, parentSpanID)
	}

	httputils.Respond(w, http.StatusOK)
}

// ResetAssignmentOperationStatus handles formation assignment status updates for a given formation operation
func (h *Handler) ResetAssignmentOperationStatus(w http.ResponseWriter, r *http.Request) {
	h.updateAssignmentOperationStatus(w, r, true)
}

// UpdateAssignmentOperationStatus handles formation assignment status updates for a given formation operation
func (h *Handler) UpdateAssignmentOperationStatus(w http.ResponseWriter, r *http.Request) {
	h.updateAssignmentOperationStatus(w, r, false)
}

func (h *Handler) updateAssignmentOperationStatus(w http.ResponseWriter, r *http.Request, reset bool) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)
	traceID := correlation.TraceIDFromContext(ctx)
	spanID := correlation.SpanIDFromContext(ctx)
	parentSpanID := correlation.ParentSpanIDFromContext(ctx)

	errResp := errors.Errorf("An unexpected error occurred while processing the request. X-Request-Id: %s", correlationID)

	var assignmentOpReqBody AssignmentOperationRequestBody
	err := decodeJSONBody(w, r, &assignmentOpReqBody)
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

	if formationassignmentpkg.IsConfigEmpty(string(assignmentOpReqBody.Configuration)) {
		assignmentOpReqBody.Configuration = nil
	}

	log.C(ctx).Info("Validating assignment operation request body...")
	if err = assignmentOpReqBody.Validate(ctx); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating the request body")
		respondWithError(ctx, w, http.StatusBadRequest, errors.Errorf("Request Body contains invalid input: %q. X-Request-Id: %s", err.Error(), correlationID))
		return
	}

	routeVars := mux.Vars(r)
	formationID := routeVars[FormationIDParam]
	formationAssignmentID := routeVars[FormationAssignmentIDParam]
	formationOperationID := routeVars[FormationOperationIDParam]

	if formationID == "" || formationAssignmentID == "" || (!reset && formationOperationID == "") {
		log.C(ctx).Errorf("Missing required parameters: %q or/and %q or/and %q triggered by reset set to %t", FormationIDParam, FormationAssignmentIDParam, FormationOperationIDParam, reset)
		respondWithError(ctx, w, http.StatusBadRequest, errors.Errorf("Not all of the required parameters are provided. X-Request-Id: %s", correlationID))
		return
	}

	logger := log.C(ctx).WithField(log.FieldFormationID, formationID)
	logger = logger.WithField(log.FieldFormationAssignmentID, formationAssignmentID)
	if !reset {
		logger.WithField(log.FieldFormationOperationID, formationOperationID)
	}
	ctx = log.ContextWithLogger(ctx, logger)

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
		if apperrors.IsNotFoundError(err) {
			errResp := errors.Errorf("Formation assignment with ID %q was not found. X-Request-Id: %s", formationAssignmentID, correlationID)
			respondWithError(ctx, w, http.StatusNotFound, errResp)
			return
		}
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	ctx = tenant.SaveToContext(ctx, fa.TenantID, "")

	formation, err := h.formationService.Get(ctx, formationID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while getting formation from formation assignment with ID: %q", fa.FormationID)
		if apperrors.IsNotFoundError(err) {
			errResp := errors.Errorf("Formation with ID %q was not found. X-Request-Id: %s", formationID, correlationID)
			respondWithError(ctx, w, http.StatusNotFound, errResp)
			return
		}
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	// Only customers of type BI can push configurations with empty state even if the formation is not yet ready.
	// This mechanism is used to set up initial configuration before sending the first wave of notifications
	if len(assignmentOpReqBody.State) > 0 && formation.State != model.ReadyFormationState {
		log.C(ctx).WithError(err).Errorf("Cannot update formation assignment for formation with ID %q as formation is not in %q state. X-Request-Id: %s", fa.FormationID, model.ReadyFormationState, correlationID)
		respondWithError(ctx, w, http.StatusBadRequest, errResp)
		return
	}

	var assignmentOperation *model.AssignmentOperation
	if reset {
		assignmentOperation, err = h.assignmentOperationService.GetLatestOperation(ctx, fa.ID, fa.FormationID)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Error when getting latest operation for assignment with ID: %s. X-Request-ID: %s", fa.ID, correlationID)
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		logger.WithField(log.FieldFormationOperationID, formationOperationID)
		ctx = log.ContextWithLogger(ctx, logger)
	} else {
		assignmentOperation, err = h.assignmentOperationService.GetByID(ctx, formationOperationID)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Error when getting assignment operation with ID: %s. X-Request-ID: %s", formationOperationID, correlationID)
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
	}

	// TODO:: Remove after "Simplified states" Task
	translateSimplifiedToOldState(ctx, assignmentOpReqBody, assignmentOperation.Type)

	formationOperation := formationassignmentpkg.DetermineFormationOperationFromLatestAssignmentOperation(assignmentOperation.Type)
	notificationStatusReport := newNotificationStatusReportFromRequestBody(assignmentOpReqBody, fa, assignmentOperation.Type)
	originalStateFromStatusReport := notificationStatusReport.State
	if !isStateSupportedForOperation(ctx, model.FormationAssignmentState(originalStateFromStatusReport), formationOperation, formation.State, reset) {
		log.C(ctx).Errorf("An invalid state: %q is provided for %q operation with reset option %t", originalStateFromStatusReport, formationOperation, reset)
		errResp := errors.Errorf("An invalid state: %s is provided for %s operation. X-Request-Id: %s", originalStateFromStatusReport, formationOperation, correlationID)
		respondWithError(ctx, w, http.StatusBadRequest, errResp)
		return
	}

	changeToRegularReadyStateInStatusReport(notificationStatusReport)
	stateFromStatusReport := notificationStatusReport.State

	if reset {
		if stateFromStatusReport != string(model.ReadyAssignmentState) && stateFromStatusReport != string(model.ConfigPendingAssignmentState) {
			errResp := errors.Errorf("Cannot reset formation assignment with source %q and target %q to state %s. X-Request-Id: %s", fa.Source, fa.Target, assignmentOpReqBody.State, correlationID)
			respondWithError(ctx, w, http.StatusBadRequest, errResp)
			return
		}

		if formationassignmentpkg.IsConfigEmpty(string(notificationStatusReport.Configuration)) {
			errResp := errors.Errorf("Cannot reset formation assignment with source %q and target %q to state %s because provided configuration is empty. X-Request-Id: %s", fa.Source, fa.Target, assignmentOpReqBody.State, correlationID)
			respondWithError(ctx, w, http.StatusBadRequest, errResp)
			return
		}

		if fa.State != string(model.ReadyAssignmentState) {
			errResp := errors.Errorf("Cannot reset formation assignment with source %q and target %q because assignment is not in %q state. X-Request-Id: %s", fa.Source, fa.Target, model.ReadyAssignmentState, correlationID)
			respondWithError(ctx, w, http.StatusBadRequest, errResp)
			return
		}
		reverseFA, err := h.faService.GetReverseBySourceAndTarget(ctx, fa.FormationID, fa.Source, fa.Target)
		if err != nil {
			log.C(ctx).Error(err)
			if apperrors.IsNotFoundError(err) {
				errResp := errors.Errorf("Cannot reset formation assignment with source %q and target %q because reverse assignment is missing. X-Request-Id: %s", fa.Source, fa.Target, correlationID)
				respondWithError(ctx, w, http.StatusBadRequest, errResp)
				return
			}
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
		if reverseFA.State != string(model.ReadyAssignmentState) {
			errResp := errors.Errorf("Cannot reset formation assignment with source %q and target %q because reverse assignment is not in %q state. X-Request-Id: %s", reverseFA.Source, reverseFA.Target, model.ReadyAssignmentState, correlationID)
			respondWithError(ctx, w, http.StatusBadRequest, errResp)
			return
		}

		log.C(ctx).Infof("Resetting formation assignment with ID: %s to state: %s", fa.ID, stateFromStatusReport)
		fa.State = stateFromStatusReport
		if err = h.faService.Update(ctx, fa.ID, fa); err != nil {
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
		log.C(ctx).Infof("Creating %s Operation for formation assignment with ID: %s  v", model.Assign, fa.ID)
		if _, err = h.assignmentOperationService.Create(ctx, &model.AssignmentOperationInput{
			Type:                  model.Assign,
			FormationAssignmentID: fa.ID,
			FormationID:           fa.FormationID,
			TriggeredBy:           model.ResetAssignment,
		}); err != nil {
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		log.C(ctx).Infof("Resetting reverse formation assignment with ID: %s to state: %s", reverseFA.ID, model.InitialAssignmentState)
		reverseFA.State = string(model.InitialAssignmentState)
		if err = h.faService.Update(ctx, reverseFA.ID, reverseFA); err != nil {
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
		log.C(ctx).Infof("Creating %s Operation for reverse formation assignment with ID: %s triggered by reset on the status API", model.Assign, fa.ID)
		if _, err = h.assignmentOperationService.Create(ctx, &model.AssignmentOperationInput{
			Type:                  model.Assign,
			FormationAssignmentID: reverseFA.ID,
			FormationID:           reverseFA.FormationID,
			TriggeredBy:           model.ResetAssignment,
		}); err != nil {
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
	}

	if formationOperation == model.UnassignFormation {
		log.C(ctx).Infof("Processing status update for formation assignment with ID: %s during %q operation", fa.ID, model.UnassignFormation)
		isFADeleted, err := h.processFormationAssignmentUnassignStatusUpdate(ctx, fa, notificationStatusReport, assignmentOperation.Type)

		if commitErr := tx.Commit(); commitErr != nil {
			log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while processing formation assignment status update for %q operation", model.UnassignFormation)
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		if isFADeleted {
			if err = h.processFormationUnassign(ctx, formation, fa); err != nil {
				log.C(ctx).WithError(err).Error("An error occurred while unassigning from formation")
				respondWithError(ctx, w, http.StatusInternalServerError, errResp)
				return
			}
		}

		log.C(ctx).Infof("The formation assignment with ID: %q was successfully processed for %q operation", formationAssignmentID, model.UnassignFormation)
		httputils.Respond(w, http.StatusOK)
		return
	}

	log.C(ctx).Infof("Processing status update for formation assignment with ID: %s during %q operation", fa.ID, model.AssignFormation)
	shouldProcessNotifications, errorResponse := h.processFormationAssignmentAssignStatusUpdate(ctx, fa, notificationStatusReport, correlationID)
	if errorResponse != nil {
		respondWithError(ctx, w, errorResponse.statusCode, errors.New(errorResponse.errorMessage))
		return
	}

	if fa.State == string(model.ReadyAssignmentState) {
		log.C(ctx).Infof("Finish %s Operation with ID: %s during status report", model.Assign, formationOperationID)
		if finishOpErr := h.assignmentOperationService.FinishByID(ctx, formationOperationID); finishOpErr != nil {
			log.C(ctx).WithError(finishOpErr).Errorf("An error occurred while finishing %s Operation with ID: %q", model.Assign, formationOperationID)
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}
	log.C(ctx).Infof("The formation assignment with ID: %q and formation ID: %q was successfully updated with state: %q", formationAssignmentID, formationID, fa.State)

	if formationassignmentpkg.IsConfigEmpty(string(notificationStatusReport.Configuration)) { // do not generate formation assignment notifications when configuration is empty
		log.C(ctx).Info("The configuration in the request body is empty. Formation assignment notification won't be generated")
		httputils.Respond(w, http.StatusOK)
		return
	}

	if shouldProcessNotifications {
		// The formation assignment notifications processing is independent of the status update request handling.
		// That's why we're executing it in a go routine and in parallel to this returning a response to the client
		go h.processFormationAssignmentNotifications(fa, correlationID, traceID, spanID, parentSpanID)
	}

	httputils.Respond(w, http.StatusOK)
}

// UpdateFormationStatus handles formation status updates
func (h *Handler) UpdateFormationStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)
	traceID := correlation.TraceIDFromContext(ctx)
	spanID := correlation.SpanIDFromContext(ctx)
	parentSpanID := correlation.ParentSpanIDFromContext(ctx)

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

	logger := log.C(ctx).WithField(log.FieldFormationID, formationID)
	ctx = log.ContextWithLogger(ctx, logger)

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
		log.C(ctx).Infof("Processing formation status update for %q operation...", model.DeleteFormation)
		err = h.processFormationDeleteStatusUpdate(ctx, f, reqBody)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while processing formation status update for %q operation. X-Request-Id: %s", model.DeleteFormation, correlationID)
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		if err = tx.Commit(); err != nil {
			log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
			respondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		log.C(ctx).Infof("The status update for formation with ID: %q was successfully processed for %q operation", formationID, model.DeleteFormation)
		httputils.Respond(w, http.StatusOK)
		return
	}

	log.C(ctx).Infof("Processing formation status update for %q operation...", model.CreateFormation)
	shouldResync, err := h.processFormationCreateStatusUpdate(ctx, f, reqBody)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while processing formation status update for %q operation. X-Request-Id: %s", model.CreateFormation, correlationID)
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
		respondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}
	log.C(ctx).Infof("The status update for formation with ID: %q was successfully processed for %q operation", formationID, model.CreateFormation)

	if shouldResync {
		// The formation notifications processing is independent of the status update request handling.
		// That's why we're executing it in a go routine and in parallel to this returning a response to the client
		go h.processFormationNotifications(f, correlationID, traceID, spanID, parentSpanID)
	}

	httputils.Respond(w, http.StatusOK)
}

func (h *Handler) processFormationUnassign(ctx context.Context, formation *model.Formation, fa *model.FormationAssignment) error {
	unassignTx, err := h.transact.Begin()
	if err != nil {
		return errors.Wrapf(err, "while beginning transaction")
	}
	defer h.transact.RollbackUnlessCommitted(ctx, unassignTx)

	unassignCtx := persistence.SaveToContext(ctx, unassignTx)

	if err = h.unassignObjectFromFormationWhenThereAreNoFormationAssignments(unassignCtx, fa, formation, fa.Source, fa.SourceType); err != nil {
		return errors.Wrapf(err, "while unassigning object with type: %q and ID: %q", fa.SourceType, fa.Source)
	}

	// Skip second unassign from formation in case the formation assignment is self-referenced one,
	// and if happened to be the last one. It will be handled above.
	if fa.Source != fa.Target {
		if err = h.unassignObjectFromFormationWhenThereAreNoFormationAssignments(unassignCtx, fa, formation, fa.Target, fa.TargetType); err != nil {
			return errors.Wrapf(err, "while unassigning object with type: %q and ID: %q", fa.TargetType, fa.Target)
		}
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
		err = h.formationService.UnassignFromScenarioLabel(ctx, fa.TenantID, objectID, graphql.FormationObjectType(objectType), formation)
		if err != nil {
			return errors.Wrapf(err, "while unassigning formation with name: %q for object ID: %q and type: %q", formation.Name, objectID, objectType)
		}
		log.C(ctx).Infof("Object with type: %q and ID: %q was successfully unassigned from formation with name: %q", objectType, objectID, formation.Name)
	}
	return nil
}

// Validate validates the formation assignment's request body input
func (b FormationAssignmentRequestBody) Validate(ctx context.Context) error {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching consumer info from context")
	}
	consumerType := consumerInfo.Type

	var fieldRules []*validation.FieldRules
	fieldRules = append(
		fieldRules,
		validation.Field(&b.State,
			validation.Required.When(consumerType != consumer.BusinessIntegration && consumerType != consumer.User),
			validation.When(b.Error != "", validation.In(model.CreateErrorAssignmentState, model.DeleteErrorAssignmentState)),
			validation.When(len(b.Configuration) > 0, validation.In(model.ReadyAssignmentState, model.ConfigPendingAssignmentState, model.CreateReadyFormationAssignmentState, model.DeleteReadyFormationAssignmentState)),
			// in case of empty error and configuration
			validation.In(model.ReadyAssignmentState, model.CreateErrorAssignmentState, model.DeleteErrorAssignmentState, model.ConfigPendingAssignmentState, model.CreateReadyFormationAssignmentState, model.DeleteReadyFormationAssignmentState),
		),
		validation.Field(&b.Configuration, validation.When(b.Error != "", validation.Empty)),
		validation.Field(&b.Error, validation.When(len(b.Configuration) > 0, validation.Empty)),
	)

	return validation.ValidateStruct(&b, fieldRules...)
}

func (b FormationAssignmentRequestBody) GetState() model.FormationAssignmentState {
	return b.State
}

func (b FormationAssignmentRequestBody) GetConfiguration() json.RawMessage {
	return b.Configuration
}

func (b FormationAssignmentRequestBody) GetError() string {
	return b.Error
}

func (b FormationAssignmentRequestBody) SetState(state model.FormationAssignmentState) {
	b.State = state
}

// Validate validates the assignment operation's request body input
func (b AssignmentOperationRequestBody) Validate(ctx context.Context) error {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching consumer info from context")
	}
	consumerType := consumerInfo.Type

	var fieldRules []*validation.FieldRules
	fieldRules = append(
		fieldRules,
		validation.Field(&b.State,
			validation.Required.When(consumerType != consumer.BusinessIntegration && consumerType != consumer.User),
			validation.When(b.Error != "", validation.In(model.ErrorAssignmentState)),
			validation.When(len(b.Configuration) > 0, validation.In(model.ReadyAssignmentState, model.ConfigPendingAssignmentState)),
			// in case of empty error and configuration
			validation.In(model.ReadyAssignmentState, model.ErrorAssignmentState, model.ConfigPendingAssignmentState),
		),
		validation.Field(&b.Configuration, validation.When(b.Error != "", validation.Empty)),
		validation.Field(&b.Error, validation.When(len(b.Configuration) > 0, validation.Empty)),
	)

	return validation.ValidateStruct(&b, fieldRules...)
}

func (b AssignmentOperationRequestBody) GetState() model.FormationAssignmentState {
	return b.State
}

func (b AssignmentOperationRequestBody) GetConfiguration() json.RawMessage {
	return b.Configuration
}

func (b AssignmentOperationRequestBody) GetError() string {
	return b.Error
}

func (b AssignmentOperationRequestBody) SetState(state model.FormationAssignmentState) {
	b.State = state
}

// Validate validates the formation's request body input
func (b FormationRequestBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.State,
			validation.When(len(b.Error) == 0, validation.In(model.ReadyFormationState, model.CreateErrorFormationState, model.DeleteErrorFormationState)).
				Else(validation.In(model.CreateErrorFormationState, model.DeleteErrorFormationState))))
}

// processFormationAssignmentUnassignStatusUpdate handles the async unassign formation assignment status update
func (h *Handler) processFormationAssignmentUnassignStatusUpdate(ctx context.Context, fa *model.FormationAssignment, statusReport *statusreport.NotificationStatusReport, latestAssignmentOperationType model.AssignmentOperationType) (bool, error) {
	stateFromStatusReport := model.FormationAssignmentState(statusReport.State)

	if latestAssignmentOperationType == model.InstanceCreatorUnassign {
		consumerInfo, err := consumer.LoadFromContext(ctx)
		if err != nil {
			return false, err
		}

		if consumerInfo.Type != consumer.InstanceCreator {
			return false, nil
		}
	}

	if stateFromStatusReport == model.DeleteErrorAssignmentState {
		if err := h.faStatusService.UpdateWithConstraints(ctx, statusReport, fa, model.UnassignFormation); err != nil {
			return false, errors.Wrapf(err, "while updating error state to: %s for formation assignment with ID: %q", stateFromStatusReport, fa.ID)
		}
		return false, nil
	}

	if err := h.faStatusService.DeleteWithConstraints(ctx, fa.ID, statusReport); err != nil {
		log.C(ctx).WithError(err).Infof("An error occurred while deleting the assignment with ID: %q with constraints", fa.ID)
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("Assignment with ID %q has already been deleted", fa.ID)
			return true, nil
		}

		if updateError := h.faService.SetAssignmentToErrorState(ctx, fa, err.Error(), formationassignment.TechnicalError, model.DeleteErrorAssignmentState); updateError != nil {
			return false, errors.Wrapf(updateError, "while updating error state: %s", errors.Wrapf(err, "while deleting formation assignment with id %q", fa.ID).Error())
		}

		return false, errors.Wrapf(err, "while deleting formation assignment with ID: %q", fa.ID)
	}

	return true, nil
}

func (h *Handler) processFormationAssignmentAssignStatusUpdate(ctx context.Context, fa *model.FormationAssignment, statusReport *statusreport.NotificationStatusReport, correlationID string) (bool, *responseError) {
	log.C(ctx).Infof("Updating formation assignment with ID: %q, formation ID: %q and state: %q", fa.ID, fa.FormationID, fa.State)
	if err := h.faStatusService.UpdateWithConstraints(ctx, statusReport, fa, model.AssignFormation); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while updating formation assignment with ID: %q and formation ID: %q with state: %q", fa.ID, fa.FormationID, fa.State)
		return false, &responseError{
			statusCode:   http.StatusInternalServerError,
			errorMessage: fmt.Sprintf("An unexpected error occurred while processing the request. X-Request-Id: %s", correlationID),
		}
	}

	state := model.FormationAssignmentState(statusReport.State)
	shouldSendReverseNotification := state != model.CreateErrorAssignmentState && state != model.InitialAssignmentState

	return shouldSendReverseNotification, nil
}

// processFormationDeleteStatusUpdate handles the async delete formation status update
func (h *Handler) processFormationDeleteStatusUpdate(ctx context.Context, formation *model.Formation, reqBody FormationRequestBody) error {
	if reqBody.State != model.ReadyFormationState && reqBody.State != model.DeleteErrorFormationState {
		return errors.Errorf("An invalid state: %q is provided for %q operation", reqBody.State, model.DeleteFormation)
	}

	if reqBody.State == model.DeleteErrorFormationState {
		if err := h.formationStatusService.SetFormationToErrorStateWithConstraints(ctx, formation, reqBody.Error, formationassignment.ClientError, reqBody.State, model.DeleteFormation); err != nil {
			return errors.Wrapf(err, "while updating error state to: %s for formation with ID: %q", reqBody.State, formation.ID)
		}
		return nil
	}

	log.C(ctx).Infof("Deleting formation with ID: %q and name: %q", formation.ID, formation.Name)
	if err := h.formationStatusService.DeleteFormationEntityAndScenariosWithConstraints(ctx, formation.TenantID, formation); err != nil {
		return errors.Wrapf(err, "while deleting formation with ID: %q and name: %q", formation.ID, formation.Name)
	}

	return nil
}

// processFormationCreateStatusUpdate handles the async create formation status update
func (h *Handler) processFormationCreateStatusUpdate(ctx context.Context, formation *model.Formation, reqBody FormationRequestBody) (bool, error) {
	if formation.State == model.DraftFormationState {
		return false, errors.Errorf("Formations in state: %q do not support status updates", formation.State)
	}

	if reqBody.State != model.ReadyFormationState && reqBody.State != model.CreateErrorFormationState {
		return false, errors.Errorf("An invalid state: %q is provided for %q operation", reqBody.State, model.CreateFormation)
	}

	if reqBody.State == model.CreateErrorFormationState {
		if err := h.formationStatusService.SetFormationToErrorStateWithConstraints(ctx, formation, reqBody.Error, formationassignment.ClientError, reqBody.State, model.CreateFormation); err != nil {
			return false, errors.Wrapf(err, "while updating error state to: %s for formation with ID: %q", reqBody.State, formation.ID)
		}
		return false, nil
	}

	log.C(ctx).Infof("Updating formation with ID: %q and name: %q to: %q state", formation.ID, formation.Name, reqBody.State)
	formation.State = model.ReadyFormationState
	if err := h.formationStatusService.UpdateWithConstraints(ctx, formation, model.CreateFormation); err != nil {
		return false, errors.Wrapf(err, "while updating formation with ID: %q to: %q state", formation.ID, model.ReadyFormationState)
	}

	return true, nil
}

func (h *Handler) processFormationAssignmentNotifications(fa *model.FormationAssignment, correlationID, traceID, spanID, parentSpanID string) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	ctx = tenant.SaveToContext(ctx, fa.TenantID, "")
	ctx = correlation.AddCorrelationIDsToContext(ctx, correlationID, traceID, spanID, parentSpanID)

	logger := log.AddCorrelationIDsToLogger(ctx, correlationID, traceID, spanID, parentSpanID)
	logger = logger.WithField(log.FieldFormationID, fa.FormationID)
	logger = logger.WithField(log.FieldFormationAssignmentID, fa.ID)
	ctx = log.ContextWithLogger(ctx, logger)

	log.C(ctx).Info("Configuration is provided in the request body. Starting formation assignment asynchronous notifications processing...")

	tx, err := h.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("unable to establish connection with database")
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	reverseFA, err := h.faService.GetReverseBySourceAndTarget(ctx, fa.FormationID, fa.Source, fa.Target)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while getting reverse formation assignment by source: %q and target: %q", fa.Source, fa.Target)
		return
	}

	assignmentPair, err := h.faNotificationService.GenerateFormationAssignmentPair(ctx, reverseFA, fa, model.AssignFormation)
	if err != nil {
		return
	}

	if assignmentPair.AssignmentReqMapping.Request == nil && assignmentPair.ReverseAssignmentReqMapping.Request == nil {
		log.C(ctx).Info("No formation assignment notification is generated. Returning...")
		return
	}

	log.C(ctx).Infof("Processing formation assignment pair and its notifications...")
	_, err = h.faService.ProcessFormationAssignmentPair(ctx, assignmentPair)
	if err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while processing formation assignment pair and its notifications")
		if updateError := h.faService.SetAssignmentToErrorState(ctx, reverseFA, err.Error(), formationassignment.TechnicalError, model.CreateErrorAssignmentState); updateError != nil {
			log.C(ctx).WithError(updateError).Errorf("An error occurred while updating formation assignment with ID: %s to error state", reverseFA.ID)
			return
		}
		if err = tx.Commit(); err != nil {
			log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
		}
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
		return
	}

	log.C(ctx).Info("Finished formation assignment asynchronous notifications processing")
}

func (h *Handler) processFormationNotifications(f *model.Formation, correlationID, traceID, spanID, parentSpanID string) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	ctx = tenant.SaveToContext(ctx, f.TenantID, "")
	ctx = correlation.AddCorrelationIDsToContext(ctx, correlationID, traceID, spanID, parentSpanID)

	logger := log.AddCorrelationIDsToLogger(ctx, correlationID, traceID, spanID, parentSpanID)
	logger = logger.WithField(log.FieldFormationID, f.ID)
	ctx = log.ContextWithLogger(ctx, logger)

	tx, err := h.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("unable to establish connection with database")
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Starting asynchronous resynchronization for formation with ID: %q and name: %q...", f.ID, f.Name)
	if _, err := h.formationService.ResynchronizeFormationNotifications(ctx, f.ID, false); err != nil {
		log.C(ctx).WithError(err).Errorf("while resynchronize formation notifications for formation with ID: %q", f.ID)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while closing database transaction")
		return
	}

	log.C(ctx).Infof("Finished asynchronous formation resynchronization processing for formation with ID: %q and name: %q", f.ID, f.Name)
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

type responseError struct {
	statusCode   int
	errorMessage string
}

func newNotificationStatusReportFromRequestBody(requestBody RequestBody, fa *model.FormationAssignment, assignmentOperationType model.AssignmentOperationType) *statusreport.NotificationStatusReport {
	return statusreport.NewNotificationStatusReport(requestBody.GetConfiguration(), calculateState(requestBody, fa, assignmentOperationType), requestBody.GetError())
}

func calculateState(requestBody RequestBody, fa *model.FormationAssignment, assignmentOperationType model.AssignmentOperationType) string {
	if requestBody.GetState() != "" {
		return string(requestBody.GetState())
	}

	if requestBody.GetError() == "" {
		return fa.State
	}

	formationOperation := formationassignmentpkg.DetermineFormationOperationFromLatestAssignmentOperation(assignmentOperationType)

	if formationOperation == model.AssignFormation {
		return string(model.CreateErrorAssignmentState)
	}

	return string(model.DeleteErrorAssignmentState)
}

func changeToRegularReadyStateInStatusReport(statusReport *statusreport.NotificationStatusReport) {
	if statusReport.State == string(model.CreateReadyFormationAssignmentState) || statusReport.State == string(model.DeleteReadyFormationAssignmentState) {
		statusReport.State = string(model.ReadyAssignmentState)
	}
}

func isSupportedStateForReset(state model.FormationAssignmentState) bool {
	return state == model.ReadyAssignmentState || state == model.ConfigPendingAssignmentState
}

func isSupportedStateForStatusUpdateWithAssignOperation(state model.FormationAssignmentState) bool {
	return state == model.CreateErrorAssignmentState ||
		state == model.ReadyAssignmentState ||
		state == model.CreateReadyFormationAssignmentState ||
		state == model.ConfigPendingAssignmentState
}

func isSupportedStateForStatusUpdateWithUnassignOperation(state model.FormationAssignmentState) bool {
	return state == model.DeleteErrorAssignmentState ||
		state == model.ReadyAssignmentState ||
		state == model.DeleteReadyFormationAssignmentState
}

func isStateSupportedForOperation(ctx context.Context, state model.FormationAssignmentState, operation model.FormationOperation, formationState model.FormationState, isReset bool) bool {
	isSupportedForOperation := false

	if operation == model.AssignFormation {
		isSupportedForOperation = isSupportedStateForStatusUpdateWithAssignOperation(state)
	}

	if operation == model.UnassignFormation {
		isSupportedForOperation = isSupportedStateForStatusUpdateWithUnassignOperation(state)
	}

	if isReset {
		return isSupportedForOperation && isSupportedStateForReset(state)
	}

	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return isSupportedForOperation
	}

	if consumerInfo.Type == consumer.User {
		isSupportedForOperation = isSupportedForOperation || (formationState == model.DraftFormationState && operation == model.AssignFormation && state == model.InitialAssignmentState)
	}

	if consumerInfo.Type == consumer.BusinessIntegration {
		isSupportedForOperation = isSupportedForOperation || (operation == model.AssignFormation && state == model.InitialAssignmentState)
	}

	return isSupportedForOperation
}

// TODO:: Remove with "Simplified states" task
func translateSimplifiedToOldState(ctx context.Context, reqBody RequestBody, assignmentOperationType model.AssignmentOperationType) {
	if reqBody.GetState() == model.ErrorAssignmentState {
		if assignmentOperationType == model.Assign {
			log.C(ctx).Debugf("Translate simplified state %q to old state %q", reqBody.GetState(), "CreateErrorAssignmentState")
			reqBody.SetState(model.CreateErrorAssignmentState)
		} else {
			log.C(ctx).Debugf("Translate simplified state %q to old state %q", reqBody.GetState(), "DeleteErrorAssignmentState")
			reqBody.SetState(model.DeleteErrorAssignmentState)
		}
	}
}
