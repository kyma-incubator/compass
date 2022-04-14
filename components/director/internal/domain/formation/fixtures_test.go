package formation

import "github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"

func UnusedLabelService() func() *automock.LabelService {
	return func() *automock.LabelService {
		return &automock.LabelService{}
	}
}

func UnusedLabelRepo() func() *automock.LabelRepository {
	return func() *automock.LabelRepository {
		return &automock.LabelRepository{}
	}
}

func UnusedASAService() func() *automock.AutomaticFormationAssignmentService {
	return func() *automock.AutomaticFormationAssignmentService {
		return &automock.AutomaticFormationAssignmentService{}
	}
}

func UnusedEngine() func() *automock.ScenarioAssignmentEngine {
	return func() *automock.ScenarioAssignmentEngine {
		return &automock.ScenarioAssignmentEngine{}
	}
}

func UnusedLabelDefService() func() *automock.LabelDefService {
	return func() *automock.LabelDefService {
		return &automock.LabelDefService{}
	}
}
