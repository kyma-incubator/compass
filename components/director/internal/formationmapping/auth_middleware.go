package formationmapping

import (
	"context"
	"encoding/json"
	"net/http"

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

// FormationAssignmentService is responsible for the service-layer FormationAssignment operations
//go:generate mockery --exported --name=FormationAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentService interface {
	GetGlobalByID(ctx context.Context, id string) (*model.FormationAssignment, error)
}

// RuntimeRepository is responsible for the repo-layer runtime operations
//go:generate mockery --exported --name=RuntimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeRepository interface {
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

// RuntimeContextRepository is responsible for the repo-layer runtime context operations
//go:generate mockery --exported --name=RuntimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeContextRepository interface {
	GetGlobalByID(ctx context.Context, id string) (*model.RuntimeContext, error)
}

// ApplicationRepository is responsible for the repo-layer application operations
//go:generate mockery --exported --name=ApplicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationRepository interface {
	GetGlobalByID(ctx context.Context, id string) (*model.Application, error)
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

// ApplicationTemplateRepository is responsible for the repo-layer application template operations
//go:generate mockery --exported --name=ApplicationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateRepository interface {
	Exists(ctx context.Context, id string) (bool, error)
}

// LabelRepository is responsible for the repo-layer label operations
//go:generate mockery --exported --name=LabelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepository interface {
	ListForGlobalObject(ctx context.Context, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

// errorResponse structure used for the JSON encoded response
type errorResponse struct {
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
			}

			isAuthorized, statusCode, err := a.isAuthorized(ctx, formationAssignmentID)
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
func (a *Authenticator) isAuthorized(ctx context.Context, formationAssignmentID string) (bool, int, error) {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrap(err, "while fetching consumer info from context")
	}
	consumerType := consumerInfo.ConsumerType

	tx, err := a.transact.Begin()
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrap(err, "Unable to establish connection with database")
	}
	defer a.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	fa, err := a.faService.GetGlobalByID(ctx, formationAssignmentID)
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrapf(err, "while getting formation assignment with ID: %q globally", formationAssignmentID)
	}

	consumerTenantPair, err := tenant.LoadTenantPairFromContext(ctx)
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrap(err, "while loading tenant pair from context")
	}
	consumerInternalTenantID := consumerTenantPair.InternalID
	consumerExternalTenantID := consumerTenantPair.ExternalID

	log.C(ctx).Infof("Consumer with internal ID: %q and external ID: %q with type: %q is trying to update formation assignment with ID: %q for formation with ID: %q about source: %q and source type: %q, and target: %q and target type: %q", consumerInternalTenantID, consumerExternalTenantID, consumerType, fa.ID, fa.FormationID, fa.Source, fa.SourceType, fa.Target, fa.TargetType)
	if fa.TargetType == model.FormationAssignmentTypeApplication {
		log.C(ctx).Infof("The formation assignment that is being update has type: %s and ID: %q", model.FormationAssignmentTypeApplication, fa.Target)

		app, err := a.appRepo.GetGlobalByID(ctx, fa.Target)
		if err != nil {
			return false, http.StatusInternalServerError, errors.Wrapf(err, "while getting application with ID: %q globally", fa.Target)
		}
		log.C(ctx).Infof("Successfully got application with ID: %q globally", fa.Target)

		// If the consumer is integration system validate the formation assignment type is application that can be managed by the integration system caller
		if consumerType == consumer.IntegrationSystem && app.IntegrationSystemID != nil && *app.IntegrationSystemID == consumerInfo.ConsumerID {
			log.C(ctx).Infof("The caller with ID: %q and type: %q has owner access to the formation assignment with target: %q and target type: %q that is being updated", consumerInfo.ConsumerID, consumerType, fa.Target, fa.TargetType)
			return true, http.StatusOK, nil
		}

		// Verify if the caller has owner access to the formation assignment with type application that is being updated
		exists, err := a.appRepo.OwnerExists(ctx, consumerInternalTenantID, fa.Target)
		if err != nil {
			log.C(ctx).Warningf("an error occurred while verifying caller with ID: %q and type: %q has owner access to formation assignment with type: %q and target ID: %q", consumerInternalTenantID, consumerType, fa.TargetType, fa.Target)
		}

		if exists {
			log.C(ctx).Infof("The caller with ID: %q and type: %q has owner access to the formation assignment with target: %q and target type: %q that is being updated", consumerInternalTenantID, consumerType, fa.Target, fa.TargetType)
			return true, http.StatusOK, nil
		}

		// Validate if the application is registered through subscription and the caller has owner access to that application
		return a.validateSubscriptionProvider(ctx, app.ApplicationTemplateID, consumerExternalTenantID, string(consumerType), fa.Target, string(fa.TargetType))
	}

	if fa.TargetType == model.FormationAssignmentTypeRuntime && (consumerType == consumer.Runtime || consumerType == consumer.ExternalCertificate || consumerType == consumer.SuperAdmin) { // consumer.SuperAdmin is needed for the local testing setup
		log.C(ctx).Infof("The formation assignment that is being update has type: %s and ID: %q", model.FormationAssignmentTypeRuntime, fa.Target)

		exists, err := a.runtimeRepo.OwnerExists(ctx, consumerInternalTenantID, fa.Target)
		if err != nil {
			return false, http.StatusUnauthorized, errors.Wrapf(err, "while verifying caller with ID: %q and type: %q has owner access to formation assignment with type: %q and target ID: %q", consumerInternalTenantID, consumerType, fa.TargetType, fa.Target)
		}

		if exists {
			log.C(ctx).Infof("The caller with ID: %q and type: %q has owner access to the formation assignment with target: %q and target type: %q that is being updated", consumerInternalTenantID, consumerType, fa.Target, fa.TargetType)
			return true, http.StatusOK, nil
		}
	}

	if fa.TargetType == model.FormationAssignmentTypeRuntimeContext && (consumerType == consumer.Runtime || consumerType == consumer.ExternalCertificate || consumerType == consumer.SuperAdmin) { // consumer.SuperAdmin is needed for the local testing setup
		log.C(ctx).Infof("The formation assignment that is being update has type: %s and ID: %q", model.FormationAssignmentTypeRuntimeContext, fa.Target)

		log.C(ctx).Debugf("Getting runtime context with ID: %q from formation assignment with ID: %q", fa.Target, fa.ID)
		rtmCtx, err := a.runtimeContextRepo.GetGlobalByID(ctx, fa.Target)
		if err != nil {
			return false, http.StatusInternalServerError, errors.Wrapf(err, "while getting runtime context with ID: %q globally", fa.Target)
		}

		exists, err := a.runtimeRepo.OwnerExists(ctx, consumerInternalTenantID, rtmCtx.RuntimeID)
		if err != nil {
			return false, http.StatusUnauthorized, errors.Wrapf(err, "while verifying caller with ID: %q and type: %q has owner access to the parent of the formation assignment with type: %q and target ID: %q", consumerInternalTenantID, consumerType, fa.TargetType, fa.Target)
		}

		if exists {
			log.C(ctx).Infof("The caller with ID: %q and type: %q has owner access to the parent of the formation assignment with target: %q and target type: %q that is being updated", consumerInternalTenantID, consumerType, fa.Target, fa.TargetType)
			return true, http.StatusOK, nil
		}
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).Errorf("An error occurred while closing database transaction: %s", err.Error())
		return false, http.StatusInternalServerError, errors.Wrap(err, "Unable to finalize database operation")
	}

	return false, http.StatusUnauthorized, nil
}

// validateSubscriptionProvider validates if the application is registered through subscription and the caller has owner access to that application
func (a *Authenticator) validateSubscriptionProvider(ctx context.Context, appTemplateID *string, consumerExternalTenantID, consumerType, faTarget, faTargetType string) (bool, int, error) {
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
		log.C(ctx).Infof("The caller with external ID: %q and type: %q has owner access to the formation assignment with target: %q and target type: %q that is being updated", consumerExternalTenantID, consumerType, faTarget, faTargetType)
		return true, http.StatusOK, nil
	}

	return false, http.StatusUnauthorized, nil
}

// respondWithError writes a http response using with the JSON encoded error wrapped in an ErrorResponse struct
func respondWithError(ctx context.Context, w http.ResponseWriter, status int, err error) {
	log.C(ctx).WithError(err).Errorf("Responding with error: %v", err)
	w.Header().Add(httputils.HeaderContentTypeKey, httputils.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := errorResponse{Message: err.Error()}
	encodingErr := json.NewEncoder(w).Encode(errorResponse)
	if encodingErr != nil {
		log.C(ctx).WithError(err).Errorf("Failed to encode error response: %v", err)
	}
}
