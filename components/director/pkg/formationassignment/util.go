package formationassignment

import "github.com/kyma-incubator/compass/components/director/internal/model"

// IsConfigEmpty checks for different "empty" json values that could be in the formation assignment configuration
func IsConfigEmpty(configuration string) bool {
	if configuration == "" || configuration == "{}" || configuration == "\"\"" || configuration == "[]" || configuration == "null" {
		return true
	}

	return false
}

func DetermineFormationOperationFromLatestAssignmentOperation(assignmentOperationType model.AssignmentOperationType) model.FormationOperation {
	if assignmentOperationType == model.Assign {
		return model.AssignFormation
	}
	return model.UnassignFormation
}
