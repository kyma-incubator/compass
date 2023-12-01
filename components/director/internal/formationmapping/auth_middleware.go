package formationmapping

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	tenantpkg "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

const (
	// FormationIDParam is formation URL path parameter placeholder
	FormationIDParam = "ucl-formation-id"
	// FormationAssignmentIDParam is formation assignment URL path parameter placeholder
	FormationAssignmentIDParam = "ucl-assignment-id"
	// ClientIDFromCertificateHeader contains the name of the header containing the client id from the certificate
	ClientIDFromCertificateHeader = "Client-Id-From-Certificate"
)

// FormationAssignmentService is responsible for the service-layer FormationAssignment operations
//
//go:generate mockery --name=FormationAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentService interface {
	GetGlobalByIDAndFormationID(ctx context.Context, formationAssignmentID, formationID string) (*model.FormationAssignment, error)
	GetReverseBySourceAndTarget(ctx context.Context, formationID, sourceID, targetID string) (*model.FormationAssignment, error)
	ProcessFormationAssignmentPair(ctx context.Context, mappingPair *formationassignment.AssignmentMappingPairWithOperation) (bool, error)
	Delete(ctx context.Context, id string) error
	ListFormationAssignmentsForObjectID(ctx context.Context, formationID, objectID string) ([]*model.FormationAssignment, error)
	SetAssignmentToErrorState(ctx context.Context, assignment *model.FormationAssignment, errorMessage string, errorCode formationassignment.AssignmentErrorCode, state model.FormationAssignmentState) error
	Update(ctx context.Context, id string, fa *model.FormationAssignment) error
}

// FormationAssignmentNotificationService represents the formation assignment notification service for generating notifications
//
//go:generate mockery --name=FormationAssignmentNotificationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentNotificationService interface {
	GenerateFormationAssignmentPair(ctx context.Context, fa, reverseFA *model.FormationAssignment, operation model.FormationOperation) (*formationassignment.AssignmentMappingPairWithOperation, error)
}

// formationService is responsible for the service-layer Formation operations
//
//go:generate mockery --exported --name=formationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationService interface {
	UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
	Get(ctx context.Context, id string) (*model.Formation, error)
	GetGlobalByID(ctx context.Context, id string) (*model.Formation, error)
	ResynchronizeFormationNotifications(ctx context.Context, formationID string, reset bool) (*model.Formation, error)
}

//go:generate mockery --exported --name=formationStatusService --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationStatusService interface {
	UpdateWithConstraints(ctx context.Context, formation *model.Formation, operation model.FormationOperation) error
	SetFormationToErrorStateWithConstraints(ctx context.Context, formation *model.Formation, errorMessage string, errorCode formationassignment.AssignmentErrorCode, state model.FormationState, operation model.FormationOperation) error
	DeleteFormationEntityAndScenariosWithConstraints(ctx context.Context, tnt string, formation *model.Formation) error
}

// RuntimeRepository is responsible for the repo-layer runtime operations
//
//go:generate mockery --name=RuntimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeRepository interface {
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

// RuntimeContextRepository is responsible for the repo-layer runtime context operations
//
//go:generate mockery --name=RuntimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeContextRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
}

// ApplicationRepository is responsible for the repo-layer application operations
//
//go:generate mockery --name=ApplicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

// TenantRepository is responsible for the repo-layer tenant operations
//
//go:generate mockery --name=TenantRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantRepository interface {
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// ApplicationTemplateRepository is responsible for the repo-layer application template operations
//
//go:generate mockery --name=ApplicationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateRepository interface {
	Exists(ctx context.Context, id string) (bool, error)
}

// LabelRepository is responsible for the repo-layer label operations
//
//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepository interface {
	ListForGlobalObject(ctx context.Context, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

// FormationRepository is responsible for the repo-layer formation operations
//
//go:generate mockery --name=FormationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationRepository interface {
	GetGlobalByID(ctx context.Context, id string) (*model.Formation, error)
}

// FormationTemplateRepository is responsible for the repo-layer formation template operations
//
//go:generate mockery --name=FormationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateRepository interface {
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
}

// ErrorResponse structure used for the JSON encoded response
type ErrorResponse struct {
	Message string `json:"error"`
}

// Authenticator struct containing all dependencies to verify the request authenticity
type Authenticator struct {
	transact                   persistence.Transactioner
	faService                  FormationAssignmentService
	runtimeRepo                RuntimeRepository
	runtimeContextRepo         RuntimeContextRepository
	appRepo                    ApplicationRepository
	appTemplateRepo            ApplicationTemplateRepository
	labelRepo                  LabelRepository
	formationRepo              FormationRepository
	formationTemplateRepo      FormationTemplateRepository
	tenantRepo                 TenantRepository
	globalSubaccountIDLabelKey string
	uclCertOUSubaccountID      string
}

// NewFormationMappingAuthenticator creates a new Authenticator
func NewFormationMappingAuthenticator(
	transact persistence.Transactioner,
	faService FormationAssignmentService,
	runtimeRepo RuntimeRepository,
	runtimeContextRepo RuntimeContextRepository,
	appRepo ApplicationRepository,
	appTemplateRepo ApplicationTemplateRepository,
	labelRepo LabelRepository,
	formationRepo FormationRepository,
	formationTemplateRepo FormationTemplateRepository,
	tenantRepo TenantRepository,
	globalSubaccountIDLabelKey string,
	uclCertOUSubaccountID string,
) *Authenticator {
	return &Authenticator{
		transact:                   transact,
		faService:                  faService,
		runtimeRepo:                runtimeRepo,
		runtimeContextRepo:         runtimeContextRepo,
		appRepo:                    appRepo,
		appTemplateRepo:            appTemplateRepo,
		labelRepo:                  labelRepo,
		formationRepo:              formationRepo,
		formationTemplateRepo:      formationTemplateRepo,
		tenantRepo:                 tenantRepo,
		globalSubaccountIDLabelKey: globalSubaccountIDLabelKey,
		uclCertOUSubaccountID:      uclCertOUSubaccountID,
	}
}

// FormationAssignmentHandler is a handler middleware that executes authorization check for the formation assignments requests reporting status
func (a *Authenticator) FormationAssignmentHandler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			correlationID := correlation.CorrelationIDFromContext(ctx)

			if r.Method != http.MethodPatch {
				w.WriteHeader(http.StatusMethodNotAllowed)
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

			clientID := r.Header.Get(ClientIDFromCertificateHeader)
			if clientID == "" {
				log.C(ctx).Errorf("Failed to find client ID from header: %s", ClientIDFromCertificateHeader)
				respondWithError(ctx, w, http.StatusBadRequest, errors.New("tenant not found in the request"))
			}

			isAuthorized, statusCode, err := a.isFormationAssignmentAuthorized(ctx, formationAssignmentID, formationID, clientID)
			if err != nil {
				log.C(ctx).Error(err.Error())
				errResp := errors.Errorf("An unexpected error occurred while processing the request. X-Request-Id: %s", correlationID)
				if statusCode == http.StatusNotFound && apperrors.IsNotFoundError(err) {
					errResp = errors.Errorf("Formation assignment with ID %q for formation %q was not found. X-Request-Id: %s", formationAssignmentID, formationID, correlationID)
				}
				respondWithError(ctx, w, statusCode, errResp)
				return
			}

			if !isAuthorized {
				httputils.Respond(w, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FormationHandler is a handler middleware that executes authorization check for the formation requests reporting status
func (a *Authenticator) FormationHandler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			correlationID := correlation.CorrelationIDFromContext(ctx)

			if r.Method != http.MethodPatch {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			routeVars := mux.Vars(r)
			formationID := routeVars[FormationIDParam]

			if formationID == "" {
				log.C(ctx).Errorf("Missing required parameters: %q", FormationIDParam)
				respondWithError(ctx, w, http.StatusBadRequest, errors.New("Not all of the required parameters are provided"))
				return
			}

			isAuthorized, statusCode, err := a.isFormationAuthorized(ctx, formationID)
			if err != nil {
				log.C(ctx).Error(err.Error())
				errResp := errors.Errorf("An unexpected error occurred while processing the request. X-Request-Id: %s", correlationID)
				if statusCode == http.StatusNotFound && apperrors.IsNotFoundError(err) {
					errResp = errors.Errorf("Formation with ID %q was not found. X-Request-Id: %s", formationID, correlationID)
				}
				respondWithError(ctx, w, statusCode, errResp)
				return
			}

			if !isAuthorized {
				httputils.Respond(w, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Authenticator) isFormationAuthorized(ctx context.Context, formationID string) (bool, int, error) {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrap(err, "while fetching consumer info from context")
	}
	consumerID := consumerInfo.ConsumerID
	consumerType := consumerInfo.ConsumerType
	log.C(ctx).Infof("Consumer with ID: %q and type: %q is trying to update formation with ID: %q", consumerID, consumerType, formationID)

	tx, err := a.transact.Begin()
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrap(err, "Unable to establish connection with database")
	}
	defer a.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	f, err := a.formationRepo.GetGlobalByID(ctx, formationID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, http.StatusNotFound, errors.Wrapf(err, "while getting formation with ID: %q globally", formationID)
		}
		return false, http.StatusInternalServerError, errors.Wrapf(err, "while getting formation with ID: %q globally", formationID)
	}

	ft, err := a.formationTemplateRepo.Get(ctx, f.FormationTemplateID)
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrapf(err, "while getting formation template with ID: %q", f.FormationTemplateID)
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("An error occurred while closing database transaction: %s", err.Error())
		return false, http.StatusInternalServerError, errors.Wrap(err, "unable to finalize database operation")
	}

	for _, id := range ft.LeadingProductIDs {
		if id == consumerID {
			log.C(ctx).Infof("Consumer with ID: %q is contained in the leading product IDs list from formation template with ID: %q and name: %q", consumerID, ft.ID, ft.Name)
			return true, http.StatusOK, nil
		}
	}

	log.C(ctx).Infof("Consumer with ID: %q did not match any of the leading product IDs from formation template with ID: %q and name: %q", consumerID, ft.ID, ft.Name)
	return false, http.StatusUnauthorized, nil
}

// isFormationAssignmentAuthorized verify through custom logic the caller is authorized to update the formation assignment status
func (a *Authenticator) isFormationAssignmentAuthorized(ctx context.Context, formationAssignmentID, formationID, clientID string) (bool, int, error) {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrap(err, "while fetching consumer info from context")
	}
	consumerID := consumerInfo.ConsumerID
	consumerType := consumerInfo.ConsumerType

	tx, err := a.transact.Begin()
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrap(err, "Unable to establish connection with database")
	}
	defer a.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if a.uclCertOUSubaccountID == clientID {
		if err := tx.Commit(); err != nil {
			return false, http.StatusInternalServerError, errors.Wrap(err, "while closing database transaction")
		}

		log.C(ctx).Infof("The caller with ID: %s is UCL and is allowed to update formation assignments", clientID)
		return true, http.StatusOK, nil
	}

	fa, err := a.faService.GetGlobalByIDAndFormationID(ctx, formationAssignmentID, formationID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, http.StatusNotFound, err
		}
		return false, http.StatusInternalServerError, err
	}

	if fa.TargetType == model.FormationAssignmentTypeApplication {
		tnt, err := a.tenantRepo.Get(ctx, fa.TenantID)
		if err != nil {
			return false, http.StatusInternalServerError, errors.Wrapf(err, "while getting tenant with ID: %q", fa.TenantID)
		}

		if consumerType == consumer.BusinessIntegration && tnt.Type == tenantpkg.ResourceGroup {
			if err := tx.Commit(); err != nil {
				return false, http.StatusInternalServerError, errors.Wrap(err, "while closing database transaction")
			}

			log.C(ctx).Infof("The caller with ID: %s and type: %s is allowed to update formation assignments in tenants of type %s", consumerID, consumerType, tnt.Type)
			return true, http.StatusOK, nil
		}

		app, err := a.appRepo.GetByID(ctx, fa.TenantID, fa.Target)
		if err != nil {
			return false, http.StatusInternalServerError, errors.Wrapf(err, "while getting application with ID: %q", fa.Target)
		}
		log.C(ctx).Infof("Successfully got application with ID: %q", fa.Target)

		// If the consumer is integration system validate the formation assignment type is application that can be managed by the integration system caller
		if consumerType == consumer.IntegrationSystem && app.IntegrationSystemID != nil && *app.IntegrationSystemID == consumerID {
			if err := tx.Commit(); err != nil {
				return false, http.StatusInternalServerError, errors.Wrap(err, "while closing database transaction")
			}

			log.C(ctx).Infof("The caller with ID: %q and type: %q manages the target of the formation assignment with ID: %q and type: %q that is being updated", consumerID, consumerType, fa.Target, fa.TargetType)
			return true, http.StatusOK, nil
		}

		if app.ApplicationTemplateID != nil && *app.ApplicationTemplateID == consumerID {
			if err := tx.Commit(); err != nil {
				return false, http.StatusInternalServerError, errors.Wrap(err, "while closing database transaction")
			}

			log.C(ctx).Infof("The caller with ID: %q and type: %q is the parent of the target of the formation assignment with ID: %q and type: %q that is being updated", consumerID, consumerType, fa.Target, fa.TargetType)
			return true, http.StatusOK, nil
		}

		consumerTenantPair, err := tenant.LoadTenantPairFromContext(ctx)
		if err != nil {
			return false, http.StatusInternalServerError, errors.Wrap(err, "while loading tenant pair from context")
		}
		consumerInternalTenantID := consumerTenantPair.InternalID
		consumerExternalTenantID := consumerTenantPair.ExternalID

		log.C(ctx).Infof("Tenant with internal ID: %q and external ID: %q for consumer with type: %q is trying to update formation assignment with ID: %q for formation with ID: %q about source: %q and source type: %q, and target: %q and target type: %q", consumerInternalTenantID, consumerExternalTenantID, consumerType, fa.ID, fa.FormationID, fa.Source, fa.SourceType, fa.Target, fa.TargetType)

		// Verify if the caller has owner access to the target of the formation assignment with type application that is being updated
		exists, err := a.appRepo.OwnerExists(ctx, consumerInternalTenantID, fa.Target)
		if err != nil {
			return false, http.StatusInternalServerError, errors.Wrapf(err, "an error occurred while verifying caller with internal tenant ID: %q has owner access to the target of the formation assignment with ID: %q and type: %q that is being updated", consumerInternalTenantID, fa.Target, fa.TargetType)
		}

		if exists {
			if err := tx.Commit(); err != nil {
				log.C(ctx).Errorf("An error occurred while closing database transaction: %s", err.Error())
				return false, http.StatusInternalServerError, errors.Wrap(err, "Unable to finalize database operation")
			}

			log.C(ctx).Infof("The caller with internal tenant ID: %q has owner access to the target of the formation assignment with ID: %q and type: %q that is being updated", consumerInternalTenantID, fa.Target, fa.TargetType)
			return true, http.StatusOK, nil
		}
		log.C(ctx).Warningf("The caller with internal tenant ID: %q has NOT direct owner access to the target of the formation assignment with ID: %q and type: %q that is being updated. Checking if the application is made through subscription...", consumerInternalTenantID, fa.Target, fa.TargetType)

		// Validate if the application is registered through subscription and the caller has owner access to the application template of that application
		return a.validateSubscriptionProvider(ctx, tx, app.ApplicationTemplateID, consumerExternalTenantID, fa.Target, string(fa.TargetType))
	}

	consumerTenantPair, err := tenant.LoadTenantPairFromContext(ctx)
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrap(err, "while loading tenant pair from context")
	}
	consumerInternalTenantID := consumerTenantPair.InternalID

	log.C(ctx).Infof("Tenant with internal ID: %q and external ID: %q for consumer with type: %q is trying to update formation assignment with ID: %q for formation with ID: %q about source: %q and source type: %q, and target: %q and target type: %q", consumerInternalTenantID, consumerTenantPair.ExternalID, consumerType, fa.ID, fa.FormationID, fa.Source, fa.SourceType, fa.Target, fa.TargetType)
	if fa.TargetType == model.FormationAssignmentTypeRuntime {
		exists, err := a.runtimeRepo.OwnerExists(ctx, consumerInternalTenantID, fa.Target)
		if err != nil {
			return false, http.StatusUnauthorized, errors.Wrapf(err, "while verifying caller with internal tenant ID: %q has owner access to the target of the formation assignment with ID: %q and type: %q that is being updated", consumerInternalTenantID, fa.Target, fa.TargetType)
		}

		if exists {
			if err := tx.Commit(); err != nil {
				log.C(ctx).Errorf("An error occurred while closing database transaction: %s", err.Error())
				return false, http.StatusInternalServerError, errors.Wrap(err, "Unable to finalize database operation")
			}

			log.C(ctx).Infof("The caller with internal tenant ID: %q has owner access to the target of the formation assignment with ID: %q and type: %q that is being updated", consumerInternalTenantID, fa.Target, fa.TargetType)
			return true, http.StatusOK, nil
		}

		return false, http.StatusUnauthorized, nil
	}

	if fa.TargetType == model.FormationAssignmentTypeRuntimeContext {
		rtmCtx, err := a.runtimeContextRepo.GetByID(ctx, fa.TenantID, fa.Target)
		if err != nil {
			return false, http.StatusInternalServerError, errors.Wrapf(err, "while getting runtime context with ID: %q", fa.Target)
		}

		exists, err := a.runtimeRepo.OwnerExists(ctx, consumerInternalTenantID, rtmCtx.RuntimeID)
		if err != nil {
			return false, http.StatusUnauthorized, errors.Wrapf(err, "while verifying caller with internal tenant ID: %q has owner access to the target's parent of the formation assignment with ID: %q and type: %q that is being updated", consumerInternalTenantID, fa.Target, fa.TargetType)
		}

		if exists {
			if err := tx.Commit(); err != nil {
				log.C(ctx).Errorf("An error occurred while closing database transaction: %s", err.Error())
				return false, http.StatusInternalServerError, errors.Wrap(err, "Unable to finalize database operation")
			}

			log.C(ctx).Infof("The caller with internal tenant ID: %q has owner access to the target's parent of the formation assignment with ID: %q and type: %q that is being updated", consumerInternalTenantID, fa.Target, fa.TargetType)
			return true, http.StatusOK, nil
		}

		return false, http.StatusUnauthorized, nil
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).Errorf("An error occurred while closing database transaction: %s", err.Error())
		return false, http.StatusInternalServerError, errors.Wrap(err, "Unable to finalize database operation")
	}

	return false, http.StatusUnauthorized, nil
}

// validateSubscriptionProvider validates if the application is registered through subscription and the caller has owner access to the application template
func (a *Authenticator) validateSubscriptionProvider(ctx context.Context, tx persistence.PersistenceTx, appTemplateID *string, consumerExternalTenantID, faTarget, faTargetType string) (bool, int, error) {
	if appTemplateID == nil || (appTemplateID != nil && *appTemplateID == "") {
		log.C(ctx).Warning("Application template ID should not be nil or empty")
		return false, http.StatusUnauthorized, nil
	}

	appTemplateExists, err := a.appTemplateRepo.Exists(ctx, *appTemplateID)
	if err != nil {
		return false, http.StatusUnauthorized, errors.Wrapf(err, "while checking application template existence for ID: %q", *appTemplateID)
	}

	if !appTemplateExists {
		return false, http.StatusUnauthorized, errors.Wrapf(err, "application template with ID: %q doesn't exist", *appTemplateID)
	}

	labels, err := a.labelRepo.ListForGlobalObject(ctx, model.AppTemplateLabelableObject, *appTemplateID)
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrapf(err, "while getting labels for application template with ID: %q", *appTemplateID)
	}

	consumerSubaccountLbl, consumerSubaccountLblExists := labels[a.globalSubaccountIDLabelKey]

	if !consumerSubaccountLblExists {
		return false, http.StatusUnauthorized, errors.Errorf("%q label should exist as part of the provider's application template", a.globalSubaccountIDLabelKey)
	}

	consumerSubaccountLblValue, ok := consumerSubaccountLbl.Value.(string)
	if !ok {
		return false, http.StatusUnauthorized, errors.Errorf("unexpected type of %q label, expect: string, got: %T", a.globalSubaccountIDLabelKey, consumerSubaccountLbl.Value)
	}

	if consumerExternalTenantID == consumerSubaccountLblValue {
		if err := tx.Commit(); err != nil {
			log.C(ctx).Errorf("An error occurred while closing database transaction: %s", err.Error())
			return false, http.StatusInternalServerError, errors.Wrap(err, "Unable to finalize database operation")
		}

		log.C(ctx).Infof("The caller with external ID: %q has owner access to the target's parent of the formation assignment with ID: %q and type: %q that is being updated", consumerExternalTenantID, faTarget, faTargetType)
		return true, http.StatusOK, nil
	}

	return false, http.StatusUnauthorized, nil
}

// respondWithError writes a http response using with the JSON encoded error wrapped in an ErrorResponse struct
func respondWithError(ctx context.Context, w http.ResponseWriter, status int, err error) {
	log.C(ctx).Errorf("Responding with error: %v", err)
	w.Header().Add(httputils.HeaderContentTypeKey, httputils.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := ErrorResponse{Message: err.Error()}
	encodingErr := json.NewEncoder(w).Encode(errorResponse)
	if encodingErr != nil {
		log.C(ctx).WithError(err).Errorf("Failed to encode error response: %v", err)
	}
}
