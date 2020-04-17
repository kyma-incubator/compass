package scenarioassignment

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=LabelRepository -output=automock -outpkg=automock -case=underscore
type LabelRepository interface {
	GetRuntimeScenariosWhereLabelsMatchSelector(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]model.Label, error)
	GetRuntimesIDsByKeyAndValue(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]string, error)
	GetScenarioLabelsForRuntimes(ctx context.Context, tenantID string, runtimesIDs []string) ([]model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
}

//go:generate mockery -name=LabelUpsertService -output=automock -outpkg=automock -case=underscore
type LabelUpsertService interface {
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

type engine struct {
	labelRepo              LabelRepository
	scenarioAssignmentRepo Repository
	labelService           LabelUpsertService
}

func NewEngine(labelService LabelUpsertService, labelRepo LabelRepository, scenarioAssignmentRepo Repository) *engine {
	return &engine{
		labelRepo:              labelRepo,
		scenarioAssignmentRepo: scenarioAssignmentRepo,
		labelService:           labelService,
	}
}

func (e *engine) EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	runtimesIDs, err := e.labelRepo.GetRuntimesIDsByKeyAndValue(ctx, in.Tenant, in.Selector.Key, in.Selector.Value)
	if err != nil {
		return errors.Wrapf(err, "while fetching runtimes id which match given selector:%+v", in)
	}

	if len(runtimesIDs) == 0 {
		return nil
	}

	labels, err := e.labelRepo.GetScenarioLabelsForRuntimes(ctx, in.Tenant, runtimesIDs)
	if err != nil {
		return errors.Wrap(err, "while fetching scenarios labels for matched runtimes")
	}
	err = e.upsertMergedScenarios(ctx, labels, in.Tenant, in.ScenarioName, e.uniqueScenarios)
	if err != nil {
		return errors.Wrap(err, "while upserting merged scenarios to runtimes")
	}

	rtmWithoutScenarios := make(map[string]interface{})
	for _, matchedID := range runtimesIDs {
		rtmWithoutScenarios[matchedID] = new(interface{})
	}

	for _, scenarioLabel := range labels {
		_, ok := rtmWithoutScenarios[scenarioLabel.ObjectID]
		if ok {
			delete(rtmWithoutScenarios, scenarioLabel.ObjectID)
		}
	}

	return e.createScenarios(ctx, rtmWithoutScenarios, in.Tenant, in.ScenarioName)
}

func (e *engine) RemoveAssignedScenario(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	labels, err := e.labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(ctx, in.Tenant, in.Selector.Key, in.Selector.Value)
	if err != nil {
		return errors.Wrap(err, "while getting runtimes scenarios which match given selector")
	}
	return e.upsertMergedScenarios(ctx, labels, in.Tenant, in.ScenarioName, e.removeScenario)
}

func (e *engine) RemoveAssignedScenarios(ctx context.Context, in []*model.AutomaticScenarioAssignment) error {
	for _, asa := range in {
		err := e.RemoveAssignedScenario(ctx, *asa)
		if err != nil {
			return errors.Wrapf(err, "while deleting automatic scenario assigment: %s", asa.ScenarioName)
		}
	}
	return nil
}

func (e engine) GetScenariosForSelectorLabels(ctx context.Context, inputLabels map[string]string) ([]string, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	scenariosSet := make(map[string]struct{})

	for k, v := range inputLabels {
		selector := model.LabelSelector{
			Key:   k,
			Value: v,
		}

		scenarioAssignments, err := e.scenarioAssignmentRepo.ListForSelector(ctx, selector, tenantID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting Automatic Scenario Assignments for selector [key: %s, val: %s]", k, v)
		}

		for _, sa := range scenarioAssignments {
			scenariosSet[sa.ScenarioName] = struct{}{}
		}

	}

	scenarios := make([]string, 0)
	for k := range scenariosSet {
		scenarios = append(scenarios, k)
	}
	return scenarios, nil
}

func (e engine) MergeScenariosFromInputLabelsAndAssignments(ctx context.Context, inputLabels map[string]interface{}) ([]interface{}, error) {
	scenariosSet := make(map[string]struct{})

	possibleSelectorLabels := e.convertMapStringInterfaceToMapStringString(inputLabels)

	scenariosFromAssignments, err := e.GetScenariosForSelectorLabels(ctx, possibleSelectorLabels)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting scenarios for selector labels")
	}

	for _, scenario := range scenariosFromAssignments {
		scenariosSet[scenario] = struct{}{}
	}

	scenariosFromInput, isScenarioLabelInInput := inputLabels[model.ScenariosKey]

	if isScenarioLabelInInput {
		scenariosFromInputInterfaceSlice, ok := scenariosFromInput.([]interface{})
		if !ok {
			return nil, errors.New("while converting scenarios label to an interface slice")
		}

		for _, scenario := range scenariosFromInputInterfaceSlice {
			scenariosSet[fmt.Sprint(scenario)] = struct{}{}
		}

	}

	scenarios := make([]interface{}, 0)
	for k := range scenariosSet {
		scenarios = append(scenarios, k)
	}
	return scenarios, nil
}

func (e engine) MergeScenarios(baseScenarios, scenariosToDelete, scenariosToAdd []interface{}) []interface{} {
	scenariosSet := map[interface{}]struct{}{}
	for _, scenario := range baseScenarios {
		scenariosSet[scenario] = struct{}{}
	}

	for _, scenario := range scenariosToDelete {
		delete(scenariosSet, scenario)
	}

	for _, scenario := range scenariosToAdd {
		scenariosSet[scenario] = struct{}{}
	}

	scenarios := make([]interface{}, 0)
	for k := range scenariosSet {
		scenarios = append(scenarios, k)
	}
	return scenarios
}

func (e *engine) createScenarios(ctx context.Context, rtmIDs map[string]interface{}, tenantID, scenario string) error {
	for rtmID, _ := range rtmIDs {
		label := &model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []interface{}{scenario},
			ObjectID:   rtmID,
			ObjectType: model.RuntimeLabelableObject,
		}

		err := e.labelService.UpsertLabel(ctx, tenantID, label)
		if err != nil {
			return errors.Wrap(err, "while inserting new scenarios for runtime")
		}

	}
	return nil
}

func (e *engine) upsertMergedScenarios(ctx context.Context, labels []model.Label, tenantID, scenarioName string, mergeFn func(scenarios []interface{}, diffScenario string) ([]interface{}, error)) error {
	for _, label := range labels {
		scenarios, ok := label.Value.([]interface{})
		if !ok {
			return errors.Errorf("scenarios value is invalid type: %t", label.Value)
		}

		newScenarios, err := mergeFn(scenarios, scenarioName)
		if err != nil {
			return errors.Wrap(err, "while merging scenarios")
		}
		err = e.updateScenarioLabel(ctx, tenantID, label, newScenarios)
		if err != nil {
			return errors.Wrap(err, "while updating scenarios label")
		}
	}
	return nil
}

func (e *engine) updateScenarioLabel(ctx context.Context, tenantID string, label model.Label, scenarios []interface{}) error {
	if len(scenarios) == 0 {
		return e.labelRepo.Delete(ctx, tenantID, model.RuntimeLabelableObject, label.ObjectID, model.ScenariosKey)
	} else {
		labelInput := model.LabelInput{
			Key:        label.Key,
			Value:      scenarios,
			ObjectID:   label.ObjectID,
			ObjectType: label.ObjectType,
		}
		return e.labelService.UpsertLabel(ctx, tenantID, &labelInput)
	}
}

func (e *engine) uniqueScenarios(scenarios []interface{}, newScenario string) ([]interface{}, error) {
	set := make(map[interface{}]struct{})

	for _, scenario := range scenarios {
		output, ok := scenario.(string)
		if !ok {
			return nil, errors.New("scenario is not a string")
		}
		set[output] = struct{}{}
	}
	set[newScenario] = struct{}{}

	var uniqueScenarios []interface{}
	for scenario, _ := range set {
		uniqueScenarios = append(uniqueScenarios, scenario)
	}

	return uniqueScenarios, nil
}

func (e *engine) removeScenario(scenarios []interface{}, toRemove string) ([]interface{}, error) {
	var newScenarios []interface{}
	for _, scenario := range scenarios {
		output, ok := scenario.(string)
		if !ok {
			return nil, errors.New("item in scenarios is not a string")
		}

		if output != toRemove {
			newScenarios = append(newScenarios, output)
		}
	}
	return newScenarios, nil
}

func (e engine) convertMapStringInterfaceToMapStringString(inputLabels map[string]interface{}) map[string]string {
	convertedLabels := make(map[string]string)

	for k, v := range inputLabels {
		val, ok := v.(string)
		if ok {
			convertedLabels[k] = val
		}
	}

	return convertedLabels
}
