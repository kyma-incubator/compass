package formationmapping

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-openapi/runtime/middleware/header"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

const (
	FormationIDParam           = "ucl-formation-id"
	FormationAssignmentIDParam = "ucl-assignment-id"
)

type ConfigurationState string

const (
	ConfigurationStateReady         ConfigurationState = "READY"
	ConfigurationStateCreateError   ConfigurationState = "CREATE_ERROR"
	ConfigurationStateDeleteError   ConfigurationState = "DELETE_ERROR"
	ConfigurationStateConfigPending ConfigurationState = "CONFIG_PENDING"
)

// FormationAssignmentService is responsible for the service-layer FormationAssignment operations
//go:generate mockery --name=FormationAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentService interface {
	GetGlobalByID(ctx context.Context, id string) (*model.FormationAssignment, error)
}

//go:generate mockery --exported --name=runtimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

//go:generate mockery --exported --name=runtimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeContextRepository interface {
	GetGlobalByID(ctx context.Context, id string) (*model.RuntimeContext, error)
}

// ApplicationRepository missing godoc
//go:generate mockery --name=ApplicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationRepository interface {
	GetGlobalByID(ctx context.Context, id string) (*model.Application, error)
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

type ErrorResponse struct {
	Message string `json:"error"`
}

type malformedRequest struct {
	status int
	msg    string
}

type RequestBody struct {
	State         ConfigurationState `json:"state"`
	Configuration json.RawMessage    `json:"configuration"`
	Error         string             `json:"error"`
}

type Handler struct {
	faService          FormationAssignmentService
	appRepo            ApplicationRepository
	runtimeRepo        runtimeRepository
	runtimeContextRepo runtimeContextRepository
}

func NewFormationMappingHandler(faService FormationAssignmentService, appRepo ApplicationRepository, runtimeRepo runtimeRepository, runtimeContextRepo runtimeContextRepository) *Handler {
	return &Handler{
		faService:          faService,
		appRepo:            appRepo,
		runtimeRepo:        runtimeRepo,
		runtimeContextRepo: runtimeContextRepo,
	}

}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// todo:: delete
	//if r.Header.Get("Content-Type") != "" {
	//	value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
	//	if value != "application/json" {
	//		msg := "Content-Type header is not application/json"
	//		httputils.RespondWithError(ctx, w, http.StatusUnsupportedMediaType, ErrorResponse{Message: msg})
	//	}
	//}

	routeVars := mux.Vars(r)
	formationID := routeVars[FormationIDParam]
	formationAssignmentID := routeVars[FormationAssignmentIDParam]

	if formationID == "" || formationAssignmentID == "" {
		log.C(ctx).Errorf("Missing required parameters: %q or/and %q", FormationIDParam, FormationAssignmentIDParam)
		httputils.RespondWithError(ctx, w, http.StatusBadRequest, ErrorResponse{Message: "Not all of the required parameters are provided"})
	}

	// todo::: delete
	//r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	//var reqBody RequestBody
	//err := json.NewDecoder(r.Body).Decode(&reqBody)
	//if err != nil {
	//	log.C(ctx).Errorf("An unexpected error occurred while decoding request body: %s", err)
	//	httputils.RespondWithError(ctx, w, http.StatusInternalServerError, ErrorResponse{Message: "An unexpected error occurred"})
	//}

	var reqBody RequestBody
	err := decodeJSONBody(w, r, &reqBody)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			log.C(ctx).Error(mr.msg)
			httputils.RespondWithError(ctx, w, mr.status, ErrorResponse{Message: mr.msg})
		} else {
			log.C(ctx).Error(err.Error())
			httputils.RespondWithError(ctx, w, http.StatusInternalServerError, ErrorResponse{Message: "An unexpected error occurred while processing the request"})
		}
		return
	}

	if err := reqBody.Validate(); err != nil {
		log.C(ctx).Errorf("An error occurred while validating the request body: %s", err)
		httputils.RespondWithError(ctx, w, http.StatusBadRequest, ErrorResponse{Message: fmt.Sprintf("Request Body contains invalid input: %s", err)})
		return
	}

	isAuthorized, err, statusCode := h.isAuthorized(ctx, formationAssignmentID)
	if err != nil {
		log.C(ctx).Error(err.Error())
		httputils.RespondWithError(ctx, w, statusCode, ErrorResponse{Message: "An unexpected error occurred while processing the request"})
		return
	}

	if !isAuthorized {
		httputils.Respond(w, http.StatusUnauthorized)
		return
	}

	// todo::: implement business logic here

	httputils.Respond(w, http.StatusOK)
}

// Validate validates the request body input
func (b RequestBody) Validate() error {
	var fieldRules []*validation.FieldRules

	fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(ConfigurationStateReady, ConfigurationStateCreateError, ConfigurationStateDeleteError, ConfigurationStateConfigPending)))

	if b.Error != "" {
		fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(ConfigurationStateCreateError)))
	}

	if len(b.Configuration) > 0 {
		fieldRules = append(fieldRules, validation.Field(&b.State, validation.In(ConfigurationStateReady, ConfigurationStateConfigPending)))
	}

	return validation.ValidateStruct(&b, fieldRules...)
}

func (b RequestBody) ImprovedValidate() error {
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
		return errors.New("The Request Body cannot contains only State")
	}
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get(httputils.HeaderContentTypeKey) != "" {
		value, _ := header.ParseValueAndParams(r.Header, httputils.HeaderContentTypeKey)
		if value != httputils.ContentTypeApplicationJSON {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
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
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}

func (h *Handler) isAuthorized(ctx context.Context, formationAssignmentID string) (bool, error, int) {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while fetching consumer info from context"), http.StatusInternalServerError
	}
	consumerID := consumerInfo.ConsumerID
	consumerType := consumerInfo.ConsumerType

	fa, err := h.faService.GetGlobalByID(ctx, formationAssignmentID)

	if err != nil {
		return false, errors.Wrapf(err, "while getting formation assignment with ID: %q globally", formationAssignmentID), http.StatusInternalServerError
	}
	log.C(ctx).Infof("Found formation assignment with ID: %q for formation with ID: %q about source: %q and source type: %q", fa.ID, fa.FormationID, fa.Source, fa.SourceType)

	if fa.TargetType == model.FormationAssignmentTypeApplication {
		log.C(ctx).Infof("The formation assignment that is being update has type: %s and ID: %q", model.FormationAssignmentTypeApplication, fa.Target)
		app, err := h.appRepo.GetGlobalByID(ctx, fa.Target)
		if err != nil {
			return false, errors.Wrapf(err, "while getting application with ID: %q globally", fa.Target), http.StatusInternalServerError
		}

		// If the consumer is integration system validate the formation assignment type is application that can be managed by the integration system caller
		if consumerType == consumer.IntegrationSystem && app.IntegrationSystemID != nil && *app.IntegrationSystemID == consumerID {
			log.C(ctx).Info("The integration system caller has access to the formation assignment target application that is being updated")
			return true, nil, http.StatusOK
		}

		//// Validate if application is registered through subscription, the caller has owner access to that application
		//var labels map[string]interface{}
		//if err := json.Unmarshal(app.OrdLabels, &labels); err != nil {
		//	return false, errors.Wrapf(err, "while unmarshaling labels for application with ID: %q", app.ID), http.StatusInternalServerError
		//}
		//
		//if _, ok := labels["xsappnameKey"]; !ok { // the presence of the provider label is guarantee the application is created through subscription
		//	return false, errors.Errorf("application with ID: %q does not contains provider label: %q", app.ID, "xsappnameKey"), http.StatusInternalServerError // todo::: extract label
		//}
		//
		//exists, err := h.appRepo.OwnerExists(ctx, consumerID, fa.Target) // todo::: tenants?
		//if err != nil {
		//	return false, errors.Wrapf(err, "while checking if application with ID: %q has owner access to: %q", fa.Target, consumerID), http.StatusUnauthorized
		//}
		//
		//if exists {
		//	log.C(ctx).Info("The application is created through subscription and the caller has access to it")
		//	return true, nil, http.StatusOK
		//}

		// Validate if application is registered through subscription, the caller has owner access to that application
		return h.validateSubscriptionApplication(ctx, app, consumerID, fa.Target)
	}

	if fa.TargetType == model.FormationAssignmentTypeRuntime && (consumerType == consumer.Runtime || consumerType == consumer.ExternalCertificate) {
		log.C(ctx).Infof("The formation assignment that is being update has type: %s and ID: %q", model.FormationAssignmentTypeRuntime, fa.Target)
		exists, err := h.runtimeRepo.OwnerExists(ctx, consumerID, fa.Target)
		if err != nil {
			return false, errors.Wrapf(err, "while checking if runtime with ID: %q exists and has owner access to: %q", fa.Target, consumerID), http.StatusUnauthorized
		}

		if exists {
			log.C(ctx).Info("The caller has access to the formation assignment target runtime that is being updated")
			return true, nil, http.StatusOK
		}
	}

	if fa.TargetType == model.FormationAssignmentTypeRuntimeContext && (consumerType == consumer.Runtime || consumerType == consumer.ExternalCertificate) {
		log.C(ctx).Infof("The formation assignment that is being update has type: %s and ID: %q", model.FormationAssignmentTypeRuntimeContext, fa.Target)
		// check if the consumerID is owner of the fa.Target's parent - if consumerID is owner of the application/runtime context's parent(app template/runtime)
		// fa.Target could be either runtimeCtxID or appID
		log.C(ctx).Infof("Trying to get runtime context with ID: %q from formation assignment with ID: %q", fa.Target, fa.ID)
		rtmCtx, err := h.runtimeContextRepo.GetGlobalByID(ctx, fa.Target)
		if err != nil {
			return false, errors.Wrapf(err, "while getting runtime context with ID: %q globally", fa.Target), http.StatusInternalServerError
		}

		// todo:: do we need this request? can we directly use the rtmCtx.RuntimeID to check the runtime owner?
		runtime, err := h.runtimeRepo.GetByID(ctx, consumerID, rtmCtx.RuntimeID)
		if err != nil {
			return false, errors.Wrapf(err, "while getting runtime with ID: %q from runtime context with ID: %q", fa.Target, rtmCtx.ID), http.StatusInternalServerError
		}

		exists, err := h.runtimeRepo.OwnerExists(ctx, consumerID, runtime.ID)
		if err != nil {
			return false, errors.Wrapf(err, "while checking if runtime with ID: %q exists and has owner access to: %q", fa.Target, consumerID), http.StatusUnauthorized
		}

		if exists {
			return true, nil, http.StatusOK
		}
	}

	return false, nil, http.StatusUnauthorized
}

func (h *Handler) validateSubscriptionApplication(ctx context.Context, app *model.Application, consumerID, faTarget string) (bool, error, int) {
	var labels map[string]interface{}
	if err := json.Unmarshal(app.OrdLabels, &labels); err != nil {
		return false, errors.Wrapf(err, "while unmarshaling labels for application with ID: %q", app.ID), http.StatusInternalServerError
	}

	if _, ok := labels["xsappnameKey"]; !ok { // the presence of the provider label is guarantee the application is created through subscription
		return false, errors.Errorf("application with ID: %q does not contains provider label: %q", app.ID, "xsappnameKey"), http.StatusInternalServerError // todo::: extract label
	}

	exists, err := h.appRepo.OwnerExists(ctx, consumerID, faTarget) // todo::: tenants?
	if err != nil {
		return false, errors.Wrapf(err, "while checking if application with ID: %q has owner access to: %q", faTarget, consumerID), http.StatusUnauthorized
	}

	if exists {
		log.C(ctx).Info("The application is created through subscription and the caller has access to it")
		return true, nil, http.StatusOK
	}

	return false, nil, http.StatusUnauthorized
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func (e ErrorResponse) Error() string {
	return e.Message
}
