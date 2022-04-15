package formation

import "github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"

func UnusedLabelService() *automock.LabelService {
	return &automock.LabelService{}
}

func UnusedLabelRepo() *automock.LabelRepository {
	return &automock.LabelRepository{}
}

func UnusedASAService() *automock.AutomaticFormationAssignmentService {
	return &automock.AutomaticFormationAssignmentService{}
}

func UnusedEngine() *automock.ScenarioAssignmentEngine {
	return &automock.ScenarioAssignmentEngine{}
}

func UnusedLabelDefService() *automock.LabelDefService {
	return &automock.LabelDefService{}
}
