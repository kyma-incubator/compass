package formationconstraint

import "github.com/kyma-incubator/compass/components/director/internal/model"

// JoinPointLocation contains information to distinguish join points
type JoinPointLocation struct {
	OperationName  model.TargetOperation
	ConstraintType model.FormationConstraintType
}

var (
	// PreAssign represents the location before AssignFormation operation execution
	PreAssign = JoinPointLocation{
		OperationName:  model.AssignFormationOperation,
		ConstraintType: model.PreOperation,
	}
	// PostAssign represents the location after AssignFormation operation execution
	PostAssign = JoinPointLocation{
		OperationName:  model.AssignFormationOperation,
		ConstraintType: model.PostOperation,
	}
	// PreUnassign represents the location before UnassignFormation operation execution
	PreUnassign = JoinPointLocation{
		OperationName:  model.UnassignFormationOperation,
		ConstraintType: model.PreOperation,
	}
	// PostUnassign represents the location after UnassignFormation operation execution
	PostUnassign = JoinPointLocation{
		OperationName:  model.UnassignFormationOperation,
		ConstraintType: model.PostOperation,
	}
	// PreCreate represents the location before CreateFormation operation execution
	PreCreate = JoinPointLocation{
		OperationName:  model.CreateFormationOperation,
		ConstraintType: model.PreOperation,
	}
	// PostCreate represents the location after CreateFormation operation execution
	PostCreate = JoinPointLocation{
		OperationName:  model.CreateFormationOperation,
		ConstraintType: model.PostOperation,
	}
	// PreDelete represents the location before DeleteFormation operation execution
	PreDelete = JoinPointLocation{
		OperationName:  model.DeleteFormationOperation,
		ConstraintType: model.PreOperation,
	}
	// PostDelete represents the location after DeleteFormation operation execution
	PostDelete = JoinPointLocation{
		OperationName:  model.DeleteFormationOperation,
		ConstraintType: model.PostOperation,
	}
	// PreGenerateNotifications represents the location before GenerateNotification operation execution
	PreGenerateNotifications = JoinPointLocation{
		OperationName:  model.GenerateNotificationOperation,
		ConstraintType: model.PreOperation,
	}
	// PostGenerateNotifications represents the location after GenerateNotification operation execution
	PostGenerateNotifications = JoinPointLocation{
		OperationName:  model.GenerateNotificationOperation,
		ConstraintType: model.PostOperation,
	}
)
