package formationmapping

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/gorilla/mux"
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
)

//go:generate mockery --exported --name=formationAssignmentConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationAssignmentConverter interface {
	ToInput(assignment *model.FormationAssignment) *model.FormationAssignmentInput
}

// FormationAssignmentService is responsible for the service-layer FormationAssignment operations
//go:generate mockery --name=FormationAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentService interface {
	GetGlobalByIDAndFormationID(ctx context.Context, formationAssignmentID, formationID string) (*model.FormationAssignment, error)
	GetReverseBySourceAndTarget(ctx context.Context, formationID, sourceID, targetID string) (*model.FormationAssignment, error)
	ProcessFormationAssignmentPair(ctx context.Context, mappingPair *formationassignment.AssignmentMappingPair) error
	Update(ctx context.Context, id string, in *model.FormationAssignmentInput) error
	Delete(ctx context.Context, id string) error
	ListFormationAssignmentsForObjectID(ctx context.Context, formationID, objectID string) ([]*model.FormationAssignment, error)
	SetAssignmentToErrorState(ctx context.Context, assignment *model.FormationAssignment, errorMessage string, errorCode formationassignment.AssignmentErrorCode, state model.FormationAssignmentState) error
}

// FormationAssignmentNotificationService represents the formation assignment notification service for generating notifications
//go:generate mockery --name=FormationAssignmentNotificationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentNotificationService interface {
	GenerateNotification(ctx context.Context, formationAssignment *model.FormationAssignment) (*webhookclient.NotificationRequest, error)
}

// formationService is responsible for the service-layer Formation operations
//go:generate mockery --exported --name=formationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationService interface {
	UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
	Get(ctx context.Context, id string) (*model.Formation, error)
}

// RuntimeRepository is responsible for the repo-layer runtime operations
//go:generate mockery --name=RuntimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeRepository interface {
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

// RuntimeContextRepository is responsible for the repo-layer runtime context operations
//go:generate mockery --name=RuntimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeContextRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
}

// ApplicationRepository is responsible for the repo-layer application operations
//go:generate mockery --name=ApplicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

// TenantRepository is responsible for the repo-layer tenant operations
//go:generate mockery --name=TenantRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantRepository interface {
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// ApplicationTemplateRepository is responsible for the repo-layer application template operations
//go:generate mockery --name=ApplicationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateRepository interface {
	Exists(ctx context.Context, id string) (bool, error)
}

// LabelRepository is responsible for the repo-layer label operations
//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepository interface {
	ListForGlobalObject(ctx context.Context, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
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
	selfRegDistinguishLabelKey string
	consumerSubaccountLabelKey string
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
	selfRegDistinguishLabelKey,
	consumerSubaccountLabelKey string,
) *Authenticator {
	return &Authenticator{
		transact:                   transact,
		faService:                  faService,
		runtimeRepo:                runtimeRepo,
		runtimeContextRepo:         runtimeContextRepo,
		appRepo:                    appRepo,
		appTemplateRepo:            appTemplateRepo,
		labelRepo:                  labelRepo,
		selfRegDistinguishLabelKey: selfRegDistinguishLabelKey,
		consumerSubaccountLabelKey: consumerSubaccountLabelKey,
	}
}

// Handler is a handler middleware that executes authorization check for the formation mapping requests
func (a *Authenticator) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

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

			isAuthorized, statusCode, err := a.isAuthorized(ctx, formationAssignmentID, formationID)
			if err != nil {
				log.C(ctx).Error(err.Error())
				respondWithError(ctx, w, statusCode, errors.New("An unexpected error occurred while processing the request"))
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

// isAuthorized verify through custom logic the caller is authorized to update the formation assignment status
func (a *Authenticator) isAuthorized(ctx context.Context, formationAssignmentID, formationID string) (bool, int, error) {
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

	fa, err := a.faService.GetGlobalByIDAndFormationID(ctx, formationAssignmentID, formationID)
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrapf(err, "while getting formation assignment with ID: %q and formation ID: %q globally", formationAssignmentID, formationID)
	}

	consumerTenantPair, err := tenant.LoadTenantPairFromContext(ctx)
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrap(err, "while loading tenant pair from context")
	}
	consumerInternalTenantID := consumerTenantPair.InternalID
	consumerExternalTenantID := consumerTenantPair.ExternalID

	log.C(ctx).Infof("Tenant with internal ID: %q and external ID: %q for consumer with type: %q is trying to update formation assignment with ID: %q for formation with ID: %q about source: %q and source type: %q, and target: %q and target type: %q", consumerInternalTenantID, consumerExternalTenantID, consumerType, fa.ID, fa.FormationID, fa.Source, fa.SourceType, fa.Target, fa.TargetType)
	if fa.TargetType == model.FormationAssignmentTypeApplication {
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

			log.C(ctx).Infof("The caller with ID: %q and type: %q has owner access to the target of the formation assignment with ID: %q and type: %q that is being updated", consumerID, consumerType, fa.Target, fa.TargetType)
			return true, http.StatusOK, nil
		}

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

	_, selfRegLblExists := labels[a.selfRegDistinguishLabelKey]
	consumerSubaccountLbl, consumerSubaccountLblExists := labels[a.consumerSubaccountLabelKey]

	if !selfRegLblExists || !consumerSubaccountLblExists {
		return false, http.StatusUnauthorized, errors.Errorf("both %q and %q labels should be provided as part of the provider's application template", a.selfRegDistinguishLabelKey, a.consumerSubaccountLabelKey)
	}

	consumerSubaccountLblValue, ok := consumerSubaccountLbl.Value.(string)
	if !ok {
		return false, http.StatusUnauthorized, errors.Errorf("unexpected type of %q label, expect: string, got: %T", a.consumerSubaccountLabelKey, consumerSubaccountLbl.Value)
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
