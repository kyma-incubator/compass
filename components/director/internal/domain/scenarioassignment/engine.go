package scenarioassignment

import "github.com/kyma-incubator/compass/components/director/internal/model"

type engine struct{}

func NewEngine() *engine {
	return &engine{}
}

func (engine) EnsureScenarioAssigned(in model.AutomaticScenarioAssignment) error {
	// TODO: Implement it

	// add scenario to runtimes, which have label matching selector
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
