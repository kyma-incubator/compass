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
	// PreGenerateFormationAssignmentNotifications represents the location before GenerateFormationAssignmentNotifications operation execution
	PreGenerateFormationAssignmentNotifications = JoinPointLocation{
		OperationName:  model.GenerateFormationAssignmentNotificationOperation,
		ConstraintType: model.PreOperation,
	}
	// PostGenerateFormationAssignmentNotifications represents the location after GenerateFormationAssignmentNotifications operation execution
	PostGenerateFormationAssignmentNotifications = JoinPointLocation{
		OperationName:  model.GenerateFormationAssignmentNotificationOperation,
		ConstraintType: model.PostOperation,
	}
	// PreGenerateFormationNotifications represents the location before GenerateFormationNotifications operation execution
	PreGenerateFormationNotifications = JoinPointLocation{
		OperationName:  model.GenerateFormationNotificationOperation,
		ConstraintType: model.PreOperation,
	}
	// PostGenerateFormationNotifications represents the location after GenerateFormationNotifications operation execution
	PostGenerateFormationNotifications = JoinPointLocation{
		OperationName:  model.GenerateFormationNotificationOperation,
		ConstraintType: model.PostOperation,
	}
	// PreSendNotification represents the location before SendNotification operation execution
	PreSendNotification = JoinPointLocation{
		OperationName:  model.SendNotificationOperation,
		ConstraintType: model.PreOperation,
	}
	// PostSendNotification represents the location after SendNotification operation execution
	PostSendNotification = JoinPointLocation{
		OperationName:  model.SendNotificationOperation,
		ConstraintType: model.PostOperation,
	}
	// PreNotificationStatusReturned represents the location before NotificationStatusReturned operation execution
	PreNotificationStatusReturned = JoinPointLocation{
		OperationName:  model.NotificationStatusReturned,
		ConstraintType: model.PreOperation,
	}
	// PostNotificationStatusReturned represents the location after NotificationStatusReturned operation execution
	PostNotificationStatusReturned = JoinPointLocation{
		OperationName:  model.NotificationStatusReturned,
		ConstraintType: model.PostOperation,
	}
)
