package eventing

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	"k8s.io/utils/strings/slices"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

const (
	isNormalizedLabel = "isNormalized"
	// RuntimeEventingURLLabel missing godoc
	RuntimeEventingURLLabel = "runtime_eventServiceUrl"
	// EmptyEventingURL missing godoc
	EmptyEventingURL = ""
	// RuntimeDefaultEventingLabelf missing godoc
	RuntimeDefaultEventingLabelf = "%s_defaultEventing"
)

// RuntimeRepository missing godoc
//
//go:generate mockery --name=RuntimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
	GetOldestFromIDs(ctx context.Context, tenant string, runtimeIDs []string) (*model.Runtime, error)
	List(ctx context.Context, tenant string, runtimeIDs []string, filters []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error)
}

// LabelRepository missing godoc
//
//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepository interface {
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	DeleteByKey(ctx context.Context, tenant string, key string) error
	Upsert(ctx context.Context, tenant string, label *model.Label) error
}

// FormationService is responsible for managing formations. Provides functionality for listing all formations by object ID and listing all IDs of objects that are part of the specified formations.
//
//go:generate mockery --name=FormationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationService interface {
	ListFormationsForObject(ctx context.Context, objectID string) ([]*model.Formation, error)
	ListObjectIDsOfTypeForFormations(ctx context.Context, tenantID string, formationNames []string, objectType model.FormationAssignmentType) ([]string, error)
}

type service struct {
	appNameNormalizer normalizer.Normalizator
	runtimeRepo       RuntimeRepository
	labelRepo         LabelRepository
	formationService  FormationService
}

// NewService missing godoc
func NewService(appNameNormalizer normalizer.Normalizator, runtimeRepo RuntimeRepository, labelRepo LabelRepository, formationService FormationService) *service {
	return &service{
		appNameNormalizer: appNameNormalizer,
		runtimeRepo:       runtimeRepo,
		labelRepo:         labelRepo,
		formationService:  formationService,
	}
}

// CleanupAfterUnregisteringApplication missing godoc
func (s *service) CleanupAfterUnregisteringApplication(ctx context.Context, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	labelKey := getDefaultEventingForAppLabelKey(appID)
	err = s.labelRepo.DeleteByKey(ctx, tenantID, labelKey)
	if err != nil {
		return nil, errors.Wrapf(err, "while deleting Labels for Application with id %s", appID)
	}

	return model.NewEmptyApplicationEventingConfig()
}

// SetForApplication missing godoc
func (s *service) SetForApplication(ctx context.Context, runtimeID uuid.UUID, app model.Application) (*model.ApplicationEventingConfiguration, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	appID, err := uuid.Parse(app.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing application ID: %s", app.ID)
	}

	_, _, err = s.unsetForApplication(ctx, tenantID, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while deleting default eventing for application")
	}

	runtime, found, err := s.getRuntimeForApplicationScenarios(ctx, tenantID, runtimeID, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting the runtime")
	}

	if !found {
		return nil, fmt.Errorf("does not find the given runtime [ID=%s] assigned to the application scenarios", runtimeID)
	}

	if err := s.setRuntimeForAppEventing(ctx, tenantID, *runtime, appID); err != nil {
		return nil, errors.Wrap(err, "while setting the runtime as default for eventing for application")
	}

	runtimeEventingCfg, err := s.GetForRuntime(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing configuration for runtime")
	}

	shouldNormalize, err := s.shouldNormalizeApplicationName(ctx, tenantID, runtime)
	if err != nil {
		return nil, errors.Wrap(err, "while determining whether application name should be normalized in runtime eventing URL")
	}

	appName := app.Name
	if shouldNormalize {
		appName = s.appNameNormalizer.Normalize(app.Name)
	}

	return model.NewApplicationEventingConfiguration(runtimeEventingCfg.DefaultURL, appName)
}

// UnsetForApplication missing godoc
func (s *service) UnsetForApplication(ctx context.Context, app model.Application) (*model.ApplicationEventingConfiguration, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	appID, err := uuid.Parse(app.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing application ID: %s", app.ID)
	}

	runtime, found, err := s.unsetForApplication(ctx, tenantID, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while deleting default eventing for application")
	}

	if !found {
		return model.NewEmptyApplicationEventingConfig()
	}

	runtimeID, err := uuid.Parse(runtime.ID)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing runtime ID as UUID")
	}

	runtimeEventingCfg, err := s.GetForRuntime(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing configuration for runtime")
	}

	shouldNormalize, err := s.shouldNormalizeApplicationName(ctx, tenantID, runtime)
	if err != nil {
		return nil, errors.Wrap(err, "while determining whether application name should be normalized in runtime eventing URL")
	}

	appName := app.Name
	if shouldNormalize {
		appName = s.appNameNormalizer.Normalize(app.Name)
	}

	return model.NewApplicationEventingConfiguration(runtimeEventingCfg.DefaultURL, appName)
}

// GetForApplication missing godoc
func (s *service) GetForApplication(ctx context.Context, app model.Application) (*model.ApplicationEventingConfiguration, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	appID, err := uuid.Parse(app.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing application ID: %s", app.ID)
	}

	var defaultVerified, foundDefault, foundOldest bool
	runtime, foundDefault, err := s.getDefaultRuntimeForAppEventing(ctx, tenantID, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting default runtime for app eventing")
	}

	if foundDefault {
		if defaultVerified, err = s.ensureScenariosOrDeleteLabel(ctx, tenantID, *runtime, appID); err != nil {
			return nil, errors.Wrap(err, "while ensuring the scenarios assigned to the runtime and application")
		}
	}

	if !defaultVerified {
		runtime, foundOldest, err = s.getOldestRuntime(ctx, tenantID, appID)
		if err != nil {
			return nil, errors.Wrap(err, "while getting the oldest runtime for scenarios")
		}

		if foundOldest {
			if err := s.setRuntimeForAppEventing(ctx, tenantID, *runtime, appID); err != nil {
				return nil, errors.Wrap(err, "while setting the runtime as default for eventing for application")
			}
		}
	}

	if runtime == nil {
		return model.NewEmptyApplicationEventingConfig()
	}

	runtimeID, err := uuid.Parse(runtime.ID)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing runtime ID as UUID")
	}

	runtimeEventingCfg, err := s.GetForRuntime(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing configuration for runtime")
	}

	shouldNormalize, err := s.shouldNormalizeApplicationName(ctx, tenantID, runtime)
	if err != nil {
		return nil, errors.Wrap(err, "while determining whether application name should be normalized in runtime eventing URL")
	}

	appName := app.Name
	if shouldNormalize {
		appName = s.appNameNormalizer.Normalize(app.Name)
	}

	if app.SystemNumber != nil {
		appName += "-" + *app.SystemNumber
	}

	return model.NewApplicationEventingConfiguration(runtimeEventingCfg.DefaultURL, appName)
}

// GetForRuntime missing godoc
func (s *service) GetForRuntime(ctx context.Context, runtimeID uuid.UUID) (*model.RuntimeEventingConfiguration, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	var eventingURL string
	label, err := s.labelRepo.GetByKey(ctx, tenantID, model.RuntimeLabelableObject, runtimeID.String(), RuntimeEventingURLLabel)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return nil, errors.Wrap(err, fmt.Sprintf("while getting the label [key=%s] for runtime [ID=%s]", RuntimeEventingURLLabel, runtimeID))
		}

		return model.NewRuntimeEventingConfiguration(EmptyEventingURL)
	}

	if label != nil {
		var ok bool
		if eventingURL, ok = label.Value.(string); !ok {
			return nil, fmt.Errorf("unable to cast label [key=%s, runtimeID=%s] value as a string", RuntimeEventingURLLabel, runtimeID)
		}
	}

	return model.NewRuntimeEventingConfiguration(eventingURL)
}

func (s *service) shouldNormalizeApplicationName(ctx context.Context, tenant string, runtime *model.Runtime) (bool, error) {
	label, err := s.labelRepo.GetByKey(ctx, tenant, model.RuntimeLabelableObject, runtime.ID, isNormalizedLabel)
	notFoundErr := apperrors.IsNotFoundError(err)
	if !notFoundErr && err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("while getting the label [key=%s] for runtime [ID=%s]", isNormalizedLabel, runtime.ID))
	}

	return notFoundErr || label.Value == "true", nil
}

func (s *service) unsetForApplication(ctx context.Context, tenantID string, appID uuid.UUID) (*model.Runtime, bool, error) {
	runtime, foundDefault, err := s.getDefaultRuntimeForAppEventing(ctx, tenantID, appID)
	if err != nil {
		return nil, false, errors.Wrap(err, "while getting default runtime for app eventing")
	}

	if !foundDefault {
		return nil, foundDefault, nil
	}

	runtimeID, err := uuid.Parse(runtime.ID)
	if err != nil {
		return nil, foundDefault, errors.Wrap(err, "while parsing runtime ID as UUID")
	}

	labelKey := getDefaultEventingForAppLabelKey(appID)
	err = s.deleteLabelFromRuntime(ctx, tenantID, labelKey, runtimeID)
	if err != nil {
		return nil, foundDefault, errors.Wrap(err, "while deleting label")
	}

	return runtime, foundDefault, nil
}

func (s *service) getDefaultRuntimeForAppEventing(ctx context.Context, tenantID string, appID uuid.UUID) (*model.Runtime, bool, error) {
	labelKey := getDefaultEventingForAppLabelKey(appID)
	labelFilterForRuntime := []*labelfilter.LabelFilter{labelfilter.NewForKey(labelKey)}

	var cursor string
	runtimesPage, err := s.runtimeRepo.List(ctx, tenantID, []string{}, labelFilterForRuntime, 1, cursor)
	if err != nil {
		return nil, false, errors.Wrap(err, fmt.Sprintf("while fetching runtimes with label [key=%s]", labelKey))
	}

	if runtimesPage.TotalCount == 0 {
		return nil, false, nil
	}

	if runtimesPage.TotalCount > 1 {
		return nil, false, fmt.Errorf("got multpile runtimes labeled [key=%s] as default for eventing", labelKey)
	}

	runtime := runtimesPage.Data[0]

	return runtime, true, nil
}

func (s *service) ensureScenariosOrDeleteLabel(ctx context.Context, tenantID string, runtime model.Runtime, appID uuid.UUID) (bool, error) {
	runtimeID, err := uuid.Parse(runtime.ID)
	if err != nil {
		return false, errors.Wrap(err, "while parsing runtime ID as UUID")
	}

	_, belongsToScenarios, err := s.getRuntimeForApplicationScenarios(ctx, tenantID, runtimeID, appID)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("while verifing whether runtime [ID=%s] belongs to the application scenarios", runtimeID))
	}

	if !belongsToScenarios {
		labelKey := getDefaultEventingForAppLabelKey(appID)
		if err = s.deleteLabelFromRuntime(ctx, tenantID, labelKey, runtimeID); err != nil {
			return false, errors.Wrap(err, "when deleting current default runtime for the application because of scenarios mismatch")
		}
	}

	return belongsToScenarios, nil
}

func (s *service) getRuntimeForApplicationScenarios(ctx context.Context, tenantID string, runtimeID, appID uuid.UUID) (*model.Runtime, bool, error) {
	runtimeIDs, err := s.getRuntimeIDsInFormationWithApp(ctx, tenantID, appID.String())
	if err != nil {
		return nil, false, errors.Wrapf(err, "while getting runtimes IDs in formation with application: %s", appID)
	}

	if !slices.Contains(runtimeIDs, runtimeID.String()) {
		return nil, false, nil
	}

	runtime, err := s.runtimeRepo.GetByID(ctx, tenantID, runtimeID.String())
	if err != nil {
		return nil, false, errors.Wrap(err, fmt.Sprintf("while getting the runtime [ID=%s] with scenarios with filter", runtimeID))
	}

	return runtime, true, nil
}

func (s *service) deleteLabelFromRuntime(ctx context.Context, tenantID, labelKey string, runtimeID uuid.UUID) error {
	if err := s.labelRepo.Delete(ctx, tenantID, model.RuntimeLabelableObject, runtimeID.String(), labelKey); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while deleting label [key=%s, runtimeID=%s]", labelKey, runtimeID))
	}

	return nil
}

func (s *service) getOldestRuntime(ctx context.Context, tenantID string, appID uuid.UUID) (*model.Runtime, bool, error) {
	runtimeIDs, err := s.getRuntimeIDsInFormationWithApp(ctx, tenantID, appID.String())
	if err != nil {
		return nil, false, errors.Wrapf(err, "while getting runtimes IDs in formation with application: %s", appID)
	}

	if len(runtimeIDs) == 0 {
		return nil, false, nil
	}

	runtime, err := s.runtimeRepo.GetOldestFromIDs(ctx, tenantID, runtimeIDs)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return nil, false, errors.Wrap(err, fmt.Sprintf("while getting the oldest runtime for application [ID=%s] scenarios with filter", appID))
		}

		return nil, false, nil
	}

	return runtime, true, nil
}

func (s *service) setRuntimeForAppEventing(ctx context.Context, tenant string, runtime model.Runtime, appID uuid.UUID) error {
	defaultEventingForAppLabel := model.NewLabelForRuntime(runtime.ID, tenant, getDefaultEventingForAppLabelKey(appID), "true")
	if err := s.labelRepo.Upsert(ctx, tenant, defaultEventingForAppLabel); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while labeling the runtime [ID=%s] as default for eventing for application [ID=%s]", runtime.ID, appID))
	}

	return nil
}

func getDefaultEventingForAppLabelKey(appID uuid.UUID) string {
	return fmt.Sprintf(RuntimeDefaultEventingLabelf, appID.String())
}

func (s *service) getRuntimeIDsInFormationWithApp(ctx context.Context, tenantID, appID string) ([]string, error) {
	formationsForApplication, err := s.formationService.ListFormationsForObject(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Formations for Application with ID: %s", appID)
	}

	if len(formationsForApplication) == 0 {
		return nil, nil
	}

	formationNames := make([]string, 0, len(formationsForApplication))
	for _, formation := range formationsForApplication {
		formationNames = append(formationNames, formation.Name)
	}

	runtimeIDs, err := s.formationService.ListObjectIDsOfTypeForFormations(ctx, tenantID, formationNames, model.FormationAssignmentTypeRuntime)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Runtimes for Formations: %v", formationNames)
	}

	return runtimeIDs, nil
}
