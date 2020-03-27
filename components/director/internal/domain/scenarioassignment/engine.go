package scenarioassignment

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=LabelRepository -output=automock -outpkg=automock -case=underscore
type LabelRepository interface {
	GetRuntimeScenariosWhereLabelsMatchSelector(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]model.Label, error)
	Upsert(ctx context.Context, label *model.Label) error
}

type engine struct {
	labelRepo              LabelRepository
	scenarioAssignmentRepo Repository
}

func NewEngine(labelRepo LabelRepository, scenarioAssignmentRepo Repository) *engine {
	return &engine{
		labelRepo:              labelRepo,
		scenarioAssignmentRepo: scenarioAssignmentRepo,
	}
}

func (e *engine) EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	labels, err := e.labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(ctx, in.Tenant, in.Selector.Key, in.Selector.Value)
	if err != nil {
		return errors.Wrap(err, "while getting runtimes scenarios which match given selector")
	}
	return e.upsertMergedScenarios(ctx, labels, in.ScenarioName, e.uniqueScenarios)
}

func (e *engine) uniqueScenarios(scenarios []interface{}, newScenario string) ([]string, error) {
	set := make(map[string]struct{})
	set[newScenario] = struct{}{}

	for _, scenario := range scenarios {
		output, ok := scenario.(string)
		if !ok {
			return nil, errors.New("scenario is not a string")
		}
		set[output] = struct{}{}
	}

	return str.MapToSlice(set), nil
}
func (e *engine) upsertMergedScenarios(ctx context.Context, labels []model.Label, scenarioName string, mergeFn func(scenarios []interface{}, diffScenario string) ([]string, error)) error {
	for _, label := range labels {
		scenarios, ok := label.Value.([]interface{})
		if !ok {
			return errors.Errorf("scenarios value is invalid type: %t", label.Value)
		}

		output, err := mergeFn(scenarios, scenarioName)
		if err != nil {
			return errors.Wrap(err, "while merging scenarios")
		}
		label.Value = output
		err = e.labelRepo.Upsert(ctx, &label)
		if err != nil {
			return errors.Wrapf(err, "while updating runtime label: %s", label.ObjectID)
		}
	}
	return nil
}

func (engine) RemoveAssignedScenario(in model.AutomaticScenarioAssignment) error {
	// TODO: Implement it

	// remove scenario from runtimes, which have label matching selector
	return nil
}

func (engine) RemoveAssignedScenarios(in []*model.AutomaticScenarioAssignment) error {
	// TODO: Implement it

	// remove scenarios from runtimes, which have label matching selector
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

func (e engine) MergeScenariosFromInputAndAssignmentsFromInput(ctx context.Context, inputLabels map[string]interface{}) ([]interface{}, error) {
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

func (e engine) ComputeScenarios(oldScenariosLabel, previousScenariosFromAssignments, newScenariosFromAssignments []interface{}) []interface{} {
	scenariosSet := map[interface{}]struct{}{}
	for _, scenario := range oldScenariosLabel {
		scenariosSet[scenario] = struct{}{}
	}

	for _, scenario := range previousScenariosFromAssignments {
		delete(scenariosSet, scenario)
	}

	for _, scenario := range newScenariosFromAssignments {
		scenariosSet[scenario] = struct{}{}
	}

	scenarios := make([]interface{}, 0)
	for k := range scenariosSet {
		scenarios = append(scenarios, k)
	}

	return scenarios
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

func (e engine) convertInterfaceSliceToStringSlice(in []interface{}) []string {
	out := []string{}
	for _, val := range in {
		out = append(out, fmt.Sprint(val))
	}
	return out
}

func (e engine) convertStringSliceToInterfaceSlice(in []string) []interface{} {
	out := []interface{}{}
	for _, val := range in {
		out = append(out, val)
	}
	return out
}
