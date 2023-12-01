package operators

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// RedirectNotificationCleanupOperator represents the redirect notification operator
	RedirectNotificationCleanupOperator = "RedirectNotificationCleanup"
)

// RedirectNotificationCleanupOperatorInput is input constructor for RedirectNotificationCleanupOperator. It returns empty OperatorInput
func RedirectNotificationCleanupOperatorInput() OperatorInput {
	return &formationconstraint.RedirectNotificationInput{}
}

// RedirectNotificationCleanupOperator is an operator that based on different condition could redirect the formation assignment notification
func (e *ConstraintEngine) RedirectNotificationCleanupOperator(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Starting executing operator: %s", RedirectNotificationCleanupOperator)

	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).WithField(logrus.ErrorKey, err).Panic("recovered panic")
			debug.PrintStack()
		}
	}()

	ri, ok := input.(*formationconstraint.RedirectNotificationCleanupInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator: %s", RedirectNotificationCleanupOperator)
	}

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q and subtype: %q for location with constraint type: %q and operation name: %q during %q operation", ri.ResourceType, ri.ResourceSubtype, ri.Location.ConstraintType, ri.Location.OperationName, ri.Operation)

	formationAssignment, err := RetrieveFormationAssignmentPointer(ctx, ri.FAMemoryAddress)
	if err != nil {
		return false, err
	}
	reverseAssignment, err := RetrieveFormationAssignmentPointer(ctx, ri.ReverseFAMemoryAddress)
	if err != nil {
		return false, err
	}
	if ri.Operation == model.AssignFormation && ri.Location.OperationName == model.SendNotificationOperation && ri.Location.ConstraintType == model.PreOperation {
		return e.RedirectNotification(ctx, ri)
	}
	if ri.Operation == model.UnassignFormation && ri.Location.OperationName == model.SendNotificationOperation && ri.Location.ConstraintType == model.PreOperation{
		if formationAssignment.State == string(model.InstanceCreatorDeletingAssignmentState) ||
			formationAssignment.State == string(model.InstanceCreatorDeleteErrorAssignmentState) {
			return e.RedirectNotification(ctx, ri)
		}
		return true, nil
	}

	if ri.Location.OperationName == model.NotificationStatusReturned && ri.Location.ConstraintType == model.PreOperation {
		statusReport, err := RetrieveNotificationStatusReportPointer(ctx, ri.NotificationStatusReportMemoryAddress)
		if err != nil {
			return false, err
		}
		if ri.Operation == model.AssignFormation {
			if statusReport.State == string(model.ReadyAssignmentState) {
				var assignmentConfig Configuration
				if err = json.Unmarshal(statusReport.Configuration, &assignmentConfig); err != nil {
					return false, errors.Wrapf(err, "while unmarshalling tenant mapping response configuration for assignment with ID: %q", formationAssignment.ID)
				}
				if assignmentConfig.Credentials.InboundCommunicationDetails != nil {
					statusReport.State = string(model.ConfigPendingAssignmentState)
				}
			}
			return true, nil
		}
		if ri.Operation == model.UnassignFormation {
			if formationAssignment.State == string(model.DeletingAssignmentState) && statusReport.State == string(model.ReadyAssignmentState) {
				formationAssignment.State = string(model.InstanceCreatorDeletingAssignmentState)
				if err = e.formationAssignmentRepo.Update(ctx, formationAssignment); err != nil {
					return false, errors.Wrapf(err, "while updating formation assignment with ID %q", formationAssignment.ID)
				}
				assignmentPair, err := e.formationAssignmentNotificationSvc.GenerateFormationAssignmentPair(ctx, formationAssignment, reverseAssignment, model.UnassignFormation)
				if err != nil {
					return false, errors.Wrapf(err, "while generating formation assignment notification")
				}
				_, err =  e.formationAssignmentService.CleanupFormationAssignment(ctx, assignmentPair)
				if err != nil {
					return false, err
				}
				return false, nil
			}

			if formationAssignment.State == string(model.DeletingAssignmentState) && statusReport.State == string(model.DeleteErrorAssignmentState) {
				return true, nil
			}
			if formationAssignment.State == string(model.InstanceCreatorDeletingAssignmentState) && statusReport.State == string(model.ReadyAssignmentState) {
				return true, nil
			}
			if formationAssignment.State == string(model.InstanceCreatorDeletingAssignmentState) && statusReport.State == string(model.DeleteErrorFormationState) {
				statusReport.State = string(model.InstanceCreatorDeleteErrorAssignmentState)
				return true, nil
			}
		}
	}

	return true, nil
}
