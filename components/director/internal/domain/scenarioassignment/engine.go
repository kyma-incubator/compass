package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=LabelRepository -output=automock -outpkg=automock -case=underscore
type LabelRepository interface {
	GetRuntimeScenariosWhereLabelsMatchSelector(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]model.Label, error)
	Upsert(ctx context.Context, label *model.Label) error
}

type engine struct {
	labelRepo LabelRepository
}

func NewEngine(labelRepo LabelRepository) *engine {
	return &engine{labelRepo: labelRepo}
}

func (e *engine) EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	labels, err := e.labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(ctx, in.Tenant, in.Selector.Key, in.Selector.Value)
	if err != nil {
		return errors.Wrap(err, "while getting runtimes scenarios which match given selector")
	}
	return e.mergeScenarios(ctx, labels, in.ScenarioName, e.uniqueScenarios)
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
func (e *engine) mergeScenarios(ctx context.Context, labels []model.Label, scenarioName string, mergeFn func(scenarios []interface{}, diffScenario string) ([]string, error)) error {
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
