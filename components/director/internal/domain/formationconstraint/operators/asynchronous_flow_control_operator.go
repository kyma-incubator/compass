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
	// AsynchronousFlowControlOperator represents the asynchronous flow control operator
	AsynchronousFlowControlOperator = "AsynchronousFlowControl"
)

// AsynchronousFlowControlOperatorInput is input constructor for AsynchronousFlowControlOperator. It returns empty OperatorInput
func AsynchronousFlowControlOperatorInput() OperatorInput {
	return &formationconstraint.AsynchronousFlowControlOperatorInput{}
}

// AsynchronousFlowControlOperator is an operator that based on different conditions behaves like the redirect operator, it redirects the formation assignment notification.
// In other cases it mutates the state, in order to control the flow of the engine, so that the assignment doesn't get deleted too early,
// and it resends the notification to the redirection endpoint, so that it can finish the cleanup.
// It introduces new deleting states.
func (e *ConstraintEngine) AsynchronousFlowControlOperator(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Starting executing operator: %s", AsynchronousFlowControlOperator)

	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).WithField(logrus.ErrorKey, err).Panic("recovered panic")
			debug.PrintStack()
		}
	}()

	ri, ok := input.(*formationconstraint.AsynchronousFlowControlOperatorInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator: %s", AsynchronousFlowControlOperator)
	}

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q and subtype: %q for location with constraint type: %q and operation name: %q during %q operation", ri.ResourceType, ri.ResourceSubtype, ri.Location.ConstraintType, ri.Location.OperationName, ri.Operation)

	if ri.Operation == model.AssignFormation && ri.Location.OperationName == model.SendNotificationOperation && ri.Location.ConstraintType == model.PreOperation {
		return e.RedirectNotification(ctx, &ri.RedirectNotificationInput)
	}
	if ri.Operation == model.UnassignFormation && ri.Location.OperationName == model.SendNotificationOperation && ri.Location.ConstraintType == model.PreOperation {
		formationAssignment, err := RetrieveFormationAssignmentPointer(ctx, ri.FAMemoryAddress)
		if err != nil {
			return false, err
		}
		if formationAssignment.State == string(model.InstanceCreatorDeletingAssignmentState) ||
			formationAssignment.State == string(model.InstanceCreatorDeleteErrorAssignmentState) {
			log.C(ctx).Infof("Tenant mapping participant processing unassign notification has alredy finished, redirecting notification for assignment %q with state %q to instance creator", formationAssignment.ID, formationAssignment.State)
			ri.ShouldRedirect = true
			return e.RedirectNotification(ctx, &ri.RedirectNotificationInput)
		}
		return true, nil
	}

	if ri.Location.OperationName == model.NotificationStatusReturned && ri.Location.ConstraintType == model.PreOperation {
		formationAssignment, err := RetrieveFormationAssignmentPointer(ctx, ri.FAMemoryAddress)
		if err != nil {
			return false, err
		}
		statusReport, err := RetrieveNotificationStatusReportPointer(ctx, ri.NotificationStatusReportMemoryAddress)
		if err != nil {
			return false, err
		}
		if ri.Operation == model.AssignFormation {
			if statusReport.State == string(model.ReadyAssignmentState) && !isNotificationStatusReportConfigEmpty(statusReport) {
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
				reverseAssignment, err := RetrieveFormationAssignmentPointer(ctx, ri.ReverseFAMemoryAddress)
				if err != nil {
					log.C(ctx).Warnf(errors.Wrapf(err, "Reverse assignment not found").Error())
				}
				log.C(ctx).Infof("Tenant mapping participant finished processing unassign notification successfully for assignment with ID %q, changing state to %q", formationAssignment.ID, model.InstanceCreatorDeletingAssignmentState)
				formationAssignment.State = string(model.InstanceCreatorDeletingAssignmentState)
				statusReport.State = string(model.InstanceCreatorDeletingAssignmentState)
				if err = e.formationAssignmentRepo.Update(ctx, formationAssignment); err != nil {
					return false, errors.Wrapf(err, "while updating formation assignment with ID %q", formationAssignment.ID)
				}
				log.C(ctx).Infof("Generating formation assignment notification for assignent with ID %q", formationAssignment.ID)
				assignmentPair, err := e.formationAssignmentNotificationSvc.GenerateFormationAssignmentPair(ctx, formationAssignment, reverseAssignment, model.UnassignFormation)
				if err != nil {
					return false, errors.Wrapf(err, "while generating formation assignment notification")
				}
				log.C(ctx).Infof("Sending notification to instance creator")
				_, err = e.formationAssignmentService.CleanupFormationAssignment(ctx, assignmentPair)
				if err != nil {
					return false, err
				}
				return true, nil
			}

			if formationAssignment.State == string(model.DeletingAssignmentState) && statusReport.State == string(model.DeleteErrorAssignmentState) {
				return true, nil
			}
			if formationAssignment.State == string(model.InstanceCreatorDeletingAssignmentState) && statusReport.State == string(model.ReadyAssignmentState) {
				log.C(ctx).Infof("Instance creator reported %q, proceeding with deletion of formation assignment with ID %q", statusReport.State, formationAssignment.ID)
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
