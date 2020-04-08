package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type engine struct {
	asaRepo Repository
}

func NewEngine(asaRepo Repository) *engine {
	return &engine{asaRepo: asaRepo}
}

func (e *engine) EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	return e.asaRepo.EnsureScenarioAssigned(ctx, in)
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
