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

func UnusedLabelDefServiceFn() *automock.LabelDefService {
	return &automock.LabelDefService{}
}

func UnusedASARepo() *automock.AutomaticFormationAssignmentRepository {
	return &automock.AutomaticFormationAssignmentRepository{}
}

func UnusedRuntimeRepo() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}

func UnusedEngine() *automock.ScenarioAssignmentEngine {
	return &automock.ScenarioAssignmentEngine{}
}

func UnusedLabelDefService() *automock.LabelDefService {
	return &automock.LabelDefService{}
}
