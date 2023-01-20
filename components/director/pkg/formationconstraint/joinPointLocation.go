package formationconstraint

import "github.com/kyma-incubator/compass/components/director/internal/model"

// JoinPointLocation contains information to distinguish join points
type JoinPointLocation struct {
	OperationName  model.TargetOperation
	ConstraintType model.FormationConstraintType
}

var (
	PreAssign = JoinPointLocation{
		OperationName:  model.AssignFormationOperation,
		ConstraintType: model.PreOperation,
	}
	PostAssign = JoinPointLocation{
		OperationName:  model.AssignFormationOperation,
		ConstraintType: model.PostOperation,
	}
	PreUnassign = JoinPointLocation{
		OperationName:  model.UnassignFormationOperation,
		ConstraintType: model.PreOperation,
	}
	PostUnassign = JoinPointLocation{
		OperationName:  model.UnassignFormationOperation,
		ConstraintType: model.PostOperation,
	}
	PreCreate = JoinPointLocation{
		OperationName:  model.CreateFormationOperation,
		ConstraintType: model.PreOperation,
	}
	PostCreate = JoinPointLocation{
		OperationName:  model.CreateFormationOperation,
		ConstraintType: model.PostOperation,
	}
	PreDelete = JoinPointLocation{
		OperationName:  model.DeleteFormationOperation,
		ConstraintType: model.PreOperation,
	}
	PostDelete = JoinPointLocation{
		OperationName:  model.DeleteFormationOperation,
		ConstraintType: model.PostOperation,
	}
	PreGenerateNotifications = JoinPointLocation{
		OperationName:  model.GenerateNotificationOperation,
		ConstraintType: model.PreOperation,
	}
	PostGenerateNotifications = JoinPointLocation{
		OperationName:  model.GenerateNotificationOperation,
		ConstraintType: model.PostOperation,
	}
)
