package scenarioassignment

import (
	"context"

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
	for _, label := range labels {
		scenarios, ok := label.Value.([]interface{})
		if !ok {
			return errors.Errorf("scenarios value is invalid type: %t", label.Value)
		}

		label.Value = append(scenarios, in.ScenarioName)
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
