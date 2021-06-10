package label

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

type scenarioService struct {
	labelRepo LabelRepository
}

func NewScenarioService(labelRepo LabelRepository) *scenarioService {
	return &scenarioService{
		labelRepo: labelRepo,
	}
}

func (s *scenarioService) GetScenarioNamesForApplication(ctx context.Context, applicationID string) ([]string, error) {
	return s.getScenarioNamesForObject(ctx, model.ApplicationLabelableObject, applicationID)
}

func (s *scenarioService) GetScenarioNamesForRuntime(ctx context.Context, runtimeId string) ([]string, error) {
	return s.getScenarioNamesForObject(ctx, model.RuntimeLabelableObject, runtimeId)
}

func (s *scenarioService) GetRuntimeScenarioLabelsForAnyMatchingScenario(ctx context.Context, scenarios []string) ([]model.Label, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.labelRepo.ListByObjectTypeAndMatchAnyScenario(ctx, tnt, model.RuntimeLabelableObject, scenarios)
}

func (s *scenarioService) GetApplicationScenarioLabelsForAnyMatchingScenario(ctx context.Context, scenarios []string) ([]model.Label, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.labelRepo.ListByObjectTypeAndMatchAnyScenario(ctx, tnt, model.ApplicationLabelableObject, scenarios)
}

func (s *scenarioService) getScenarioNamesForObject(ctx context.Context, objectType model.LabelableObject, objectId string) ([]string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	log.C(ctx).Infof("Getting scenarios for %s with id %s", objectType, objectId)

	objLabel, err := s.labelRepo.GetByKey(ctx, tnt, objectType, objectId, model.ScenariosKey)

	if err != nil {
		if apperrors.ErrorCode(err) == apperrors.NotFound {
			log.C(ctx).Infof("No scenarios found for %s", objectType)
			return make([]string, 0), nil
		}
		return nil, errors.Wrapf(err, "while fetching scenarios for object with id: %s and type: %s", objectId, objectType)
	}

	scenarios, err := ValueToStringsSlice(objLabel.Value)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing %s label values", objectType)
	}

	return scenarios, nil
}

func (s *scenarioService) GetBundleInstanceAuthsScenarioLabels(ctx context.Context, appId, runtimeId string) ([]model.Label, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.labelRepo.GetBundleInstanceAuthsScenarioLabels(ctx, tnt, appId, runtimeId)
}

// MergeScenarios apply merge function on scenarios' value and create model.LabelInput. If no scenarios are left as a result of the merge function
// then model.LabelInput is nil
func MergeScenarios(label model.Label, newScenarios []string, mergeFn func(scenarios, diffScenario []string) []string) (*model.LabelInput, error) {
	if model.ScenariosKey != label.Key {
		return nil, errors.New("provided label is not scenario")
	}

	scenariosString, err := GetScenariosAsStringSlice(label)
	if err != nil {
		return nil, err
	}

	scenariosToUpsert := mergeFn(scenariosString, newScenarios)
	if len(scenariosToUpsert) == 0 {
		return nil, nil
	}

	return &model.LabelInput{
		Key:        label.Key,
		Value:      scenariosToUpsert,
		ObjectID:   label.ObjectID,
		ObjectType: label.ObjectType,
	}, nil
}
