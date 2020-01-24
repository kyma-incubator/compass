package eventing

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"strings"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

const (
	RuntimeEventingURLLabel      = "runtime/event_service_url"
	EmptyEventingURL             = ""
	RuntimeDefaultEventingLabelf = "%s/defaultEventing"
)

//go:generate mockery -name=RuntimeRepository -output=automock -outpkg=automock -case=underscore
type RuntimeRepository interface {
	GetByFiltersAndID(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (*model.Runtime, error)
	GetOldestForFilters(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) (*model.Runtime, error)
	List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error)
}

//go:generate mockery -name=LabelRepository -output=automock -outpkg=automock -case=underscore
type LabelRepository interface {
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	DeleteByKey(ctx context.Context, tenant string, key string) error
	Upsert(ctx context.Context, label *model.Label) error
}

type service struct {
	runtimeRepo RuntimeRepository
	labelRepo   LabelRepository
}

func NewService(runtimeRepo RuntimeRepository, labelRepo LabelRepository) *service {
	return &service{
		runtimeRepo: runtimeRepo,
		labelRepo:   labelRepo,
	}
}

func (s *service) CleanupAfterUnregisteringApplication(ctx context.Context, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	labelKey := getDefaultEventingForAppLabelKey(appID)
	err = s.labelRepo.DeleteByKey(ctx, tenantID, labelKey)
	if err != nil {
		return nil, errors.Wrapf(err, "while deleting labels [key=%s]", labelKey)
	}

	return model.NewApplicationEventingConfiguration(EmptyEventingURL), nil
}

func (s *service) SetForApplication(ctx context.Context, runtimeID uuid.UUID, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
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

	if err := s.setRuntimeForAppEventing(ctx, *runtime, appID); err != nil {
		return nil, errors.Wrap(err, "while setting the runtime as default for eveting for application")
	}

	runtimeEventingCfg, err := s.GetForRuntime(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing configuration for runtime")
	}

	return model.NewApplicationEventingConfiguration(runtimeEventingCfg.DefaultURL), nil
}

func (s *service) UnsetForApplication(ctx context.Context, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	runtime, found, err := s.unsetForApplication(ctx, tenantID, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while deleting default eventing for application")
	}

	if !found {
		return model.NewApplicationEventingConfiguration(EmptyEventingURL), nil
	}

	runtimeID, err := uuid.Parse(runtime.ID)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing runtime ID as UUID")
	}

	runtimeEventingCfg, err := s.GetForRuntime(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing configuration for runtime")
	}

	return model.NewApplicationEventingConfiguration(runtimeEventingCfg.DefaultURL), nil
}

func (s *service) GetForApplication(ctx context.Context, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
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
			if err := s.setRuntimeForAppEventing(ctx, *runtime, appID); err != nil {
				return nil, errors.Wrap(err, "while setting the runtime as default for eveting for application")
			}
		}
	}

	if runtime == nil {
		return model.NewApplicationEventingConfiguration(EmptyEventingURL), nil
	}

	runtimeID, err := uuid.Parse(runtime.ID)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing runtime ID as UUID")
	}

	runtimeEventingCfg, err := s.GetForRuntime(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing configuration for runtime")
	}

	return model.NewApplicationEventingConfiguration(runtimeEventingCfg.DefaultURL), nil
}

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

		return model.NewRuntimeEventingConfiguration(EmptyEventingURL), nil
	}

	if label != nil {
		var ok bool
		if eventingURL, ok = label.Value.(string); !ok {
			return nil, fmt.Errorf("unable to cast label [key=%s, runtimeID=%s] value as a string", RuntimeEventingURLLabel, runtimeID)
		}
	}

	return model.NewRuntimeEventingConfiguration(eventingURL), nil
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
	runtimesPage, err := s.runtimeRepo.List(ctx, tenantID, labelFilterForRuntime, 1, cursor)
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
	runtimeScenariosFilter, hasScenarios, err := s.getScenariosFilter(ctx, tenantID, appID)
	if err != nil {
		return nil, false, errors.Wrap(err, "while getting application scenarios")
	}

	if !hasScenarios {
		return nil, false, errors.New("application does not belong to scenarios")
	}

	runtime, err := s.runtimeRepo.GetByFiltersAndID(ctx, tenantID, runtimeID.String(), runtimeScenariosFilter)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return nil, false, errors.Wrap(err, fmt.Sprintf("while getting the runtime [ID=%s] with scenarios with filter", runtimeID))
		}

		return nil, false, nil
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
	runtimeScenariosFilter, hasScenarios, err := s.getScenariosFilter(ctx, tenantID, appID)
	if err != nil {
		return nil, false, errors.Wrap(err, fmt.Sprintf("while getting application scenarios"))
	}

	if !hasScenarios {
		return nil, false, nil
	}

	runtime, err := s.runtimeRepo.GetOldestForFilters(ctx, tenantID, runtimeScenariosFilter)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return nil, false, errors.Wrap(err, fmt.Sprintf("while getting the oldest runtime for application [ID=%s] scenarios with filter", appID))
		}

		return nil, false, nil
	}

	return runtime, true, nil
}

func (s *service) getScenariosFilter(ctx context.Context, tenantID string, appID uuid.UUID) ([]*labelfilter.LabelFilter, bool, error) {
	appScenariosLabel, err := s.labelRepo.GetByKey(ctx, tenantID, model.ApplicationLabelableObject, appID.String(), model.ScenariosKey)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return nil, false, errors.Wrap(err, fmt.Sprintf("while getting the label [key=%s] for application [ID=%s]", model.ScenariosKey, appID))
		}

		return nil, false, nil
	}

	scenarios, err := label.ValueToStringsSlice(appScenariosLabel.Value)
	if err != nil {
		return nil, false, errors.Wrap(err, fmt.Sprintf("while converting label [key=%s] value to a slice of strings", model.ScenariosKey))
	}

	scenariosQuery := buildQueryForScenarios(scenarios)
	runtimeScenariosFilter := []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(model.ScenariosKey, scenariosQuery)}

	return runtimeScenariosFilter, true, nil
}

func (s *service) setRuntimeForAppEventing(ctx context.Context, runtime model.Runtime, appID uuid.UUID) error {
	defaultEventingForAppLabel := model.NewLabelForRuntime(runtime, getDefaultEventingForAppLabelKey(appID), "true")
	if err := s.labelRepo.Upsert(ctx, defaultEventingForAppLabel); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while labeling the runtime [ID=%s] as default for eventing for application [ID=%s]", runtime.ID, appID))
	}

	return nil
}

func buildQueryForScenarios(scenarios []string) string {
	var queryBuilder strings.Builder
	for idx, scenario := range scenarios {
		if idx > 0 {
			queryBuilder.WriteString(` || `)
		}

		queryBuilder.WriteString(fmt.Sprintf(`@ == "%s"`, scenario))
	}
	query := fmt.Sprintf(`$[*] ? ( %s )`, queryBuilder.String())

	return query
}

func getDefaultEventingForAppLabelKey(appID uuid.UUID) string {
	return fmt.Sprintf(RuntimeDefaultEventingLabelf, appID.String())
}
