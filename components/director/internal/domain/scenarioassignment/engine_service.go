package scenarioassignment

import "github.com/kyma-incubator/compass/components/director/internal/model"

type assignmentEngineService struct{}

func NewAssignmentEngineService() *assignmentEngineService {
	return &assignmentEngineService{}
}

func (assignmentEngineService) EnsureScenarioAssignedToRuntimesMatchingSelector(in model.AutomaticScenarioAssignment) error {
	// TODO: Implement it

	// add scenario to runtimes, which have label matching selector
	return nil
}

func (assignmentEngineService) UnassignScenarioFromRuntimesMatchingSelector(in model.AutomaticScenarioAssignment) error {
	// TODO: Implement it

	// remove scenario from runtimes, which have label matching selector
	return nil
}

func (assignmentEngineService) UnassignScenariosFromRuntimesMatchingSelector(in []*model.AutomaticScenarioAssignment) error {
	// TODO: Implement it

	// remove scenarios from runtimes, which have label matching selector
	return nil
}
