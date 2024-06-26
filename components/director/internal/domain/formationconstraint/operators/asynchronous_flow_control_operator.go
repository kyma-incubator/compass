package operators

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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

	formationAssignment, err := RetrieveFormationAssignmentPointer(ctx, ri.FAMemoryAddress)
	if err != nil {
		return false, err
	}

	consumerTenant, err := e.getConsumerTenant(ctx, formationAssignment)
	if err != nil || consumerTenant == "" {
		if ri.FailOnNonBTPParticipants {
			return false, errors.Errorf("Instance creator is not supported on non-BTP participants")
		}
		log.C(ctx).Infof("Instance creator is not supported on non-BTP participants")
		return true, nil
	}

	if ri.Location.OperationName == model.SendNotificationOperation && ri.Location.ConstraintType == model.PreOperation {
		w, err := RetrieveWebhookPointerFromMemoryAddress(ctx, ri.WebhookMemoryAddress)
		if err != nil {
			return false, err
		}

		if w.Mode != nil && *w.Mode == graphql.WebhookModeSync {
			if ri.FailOnSyncParticipants {
				return false, errors.Errorf("Instance creator is not supported on synchronous participants")
			}
			return true, nil
		}

		if ri.Operation == model.AssignFormation {
			return e.RedirectNotification(ctx, &ri.RedirectNotificationInput)
		}
		if ri.Operation == model.UnassignFormation {
			latestAssignmentOperation, err := e.assignmentOperationService.GetLatestOperation(ctx, formationAssignment.ID, formationAssignment.FormationID)
			if err != nil {
				return false, err
			}

			if latestAssignmentOperation.Type == model.InstanceCreatorUnassign {
				log.C(ctx).Infof("Tenant mapping participant processing unassign notification has alredy finished, redirecting notification for assignment %q with state %q to instance creator", formationAssignment.ID, formationAssignment.State)
				ri.ShouldRedirect = true
				return e.RedirectNotification(ctx, &ri.RedirectNotificationInput)
			}
			return true, nil
		}
	}

	if ri.Location.OperationName == model.NotificationStatusReturned && ri.Location.ConstraintType == model.PreOperation {
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
				consumerInfo, err := consumer.LoadFromContext(ctx)
				if err != nil {
					return false, errors.Wrap(err, "while fetching consumer info from context")
				}

				// This handles the case when there is the following race condition:
				// 1. First unassign sends a notification to the participant, the participant responds with READY and the assignment is in INSTANCE_CREATOR_DELETING
				// 2. The second is started after the first and unassign sees the assignment as READY initially, but upon update it sees INSTANCE_CREATOR_DELETING and reverts it back to DELETING.
				// Theoretically the instance creator won't respond unless it has been notified.
				if consumerInfo.Type == consumer.InstanceCreator {
					log.C(ctx).Infof("Instance creator reported %q, proceeding with deletion of formation assignment with ID %q", statusReport.State, formationAssignment.ID)
					return true, nil
				}

				reverseAssignment, err := RetrieveFormationAssignmentPointer(ctx, ri.ReverseFAMemoryAddress)
				if err != nil {
					log.C(ctx).Warnf(errors.Wrapf(err, "Reverse assignment not found").Error())
				}

				log.C(ctx).Infof("Tenant mapping participant finished processing unassign notification successfully for assignment with ID %q, will create new %q Assignment Operation", formationAssignment.ID, model.InstanceCreatorUnassign)
				statusReport.State = string(model.DeletingAssignmentState) // set to DELETING state so that in CleanupFormationAssignment -> DeleteWithConstraints we don't delete the FA

				log.C(ctx).Infof("Generating formation assignment notification for assignent with ID %q", formationAssignment.ID)
				assignmentPair, err := e.formationAssignmentNotificationSvc.GenerateFormationAssignmentPair(ctx, formationAssignment, reverseAssignment, model.UnassignFormation)
				if err != nil {
					return false, errors.Wrapf(err, "while generating formation assignment notification")
				}
				w := assignmentPair.AssignmentReqMapping.Request.Webhook
				if w.Mode != nil && *w.Mode == graphql.WebhookModeSync {
					if ri.FailOnSyncParticipants {
						return false, errors.Errorf("Instance creator is not supported on synchronous participants")
					}
					return true, nil
				}

				_, err = e.assignmentOperationService.Create(ctx, &model.AssignmentOperationInput{
					Type:                  model.InstanceCreatorUnassign,
					FormationAssignmentID: formationAssignment.ID,
					FormationID:           formationAssignment.FormationID,
					TriggeredBy:           model.UnassignObject,
				})
				if err != nil {
					return false, errors.Wrapf(err, "while creating %s Operation for assignment with ID: %s", model.InstanceCreatorUnassign, formationAssignment.ID)
				}

				log.C(ctx).Infof("Sending notification to instance creator")
				_, err = e.formationAssignmentService.CleanupFormationAssignment(ctx, assignmentPair)
				if err != nil {
					return false, err
				}
				return true, nil
			}

			latestAssignmentOperation, err := e.assignmentOperationService.GetLatestOperation(ctx, formationAssignment.ID, formationAssignment.FormationID)
			if err != nil {
				return false, err
			}

			if latestAssignmentOperation.Type == model.InstanceCreatorUnassign && statusReport.State == string(model.DeleteErrorFormationState) {
				log.C(ctx).Infof("Instance creator reported %q for formation assignment with ID %q", statusReport.State, formationAssignment.ID)
				return true, nil
			}

			if formationAssignment.State == string(model.DeletingAssignmentState) && statusReport.State == string(model.DeleteErrorAssignmentState) {
				return true, nil
			}
		}
	}

	return true, nil
}

func (e *ConstraintEngine) getConsumerTenant(ctx context.Context, formationAssignment *model.FormationAssignment) (string, error) {
	labelableObjType, err := determineLabelableObjectType(formationAssignment.TargetType)
	if err != nil {
		return "", err
	}

	labels, err := e.labelRepo.ListForObject(ctx, formationAssignment.TenantID, labelableObjType, formationAssignment.Target)
	if err != nil {
		return "", errors.Wrapf(err, "while getting labels for %s with ID: %q", formationAssignment.TargetType, formationAssignment.Target)
	}

	globalSubaccIDLbl, globalSubaccIDExists := labels[GlobalSubaccountLabelKey]
	if !globalSubaccIDExists {
		return "", errors.Errorf("%q label does not exists for: %q with ID: %q", GlobalSubaccountLabelKey, formationAssignment.TargetType, formationAssignment.Target)
	}

	globalSubaccIDLblValue, ok := globalSubaccIDLbl.Value.(string)
	if !ok {
		return "", errors.Errorf("unexpected type of %q label, expect: string, got: %T", GlobalSubaccountLabelKey, globalSubaccIDLbl.Value)
	}

	return globalSubaccIDLblValue, nil
}

func determineLabelableObjectType(assignmentType model.FormationAssignmentType) (model.LabelableObject, error) {
	switch assignmentType {
	case model.FormationAssignmentTypeApplication:
		return model.ApplicationLabelableObject, nil
	case model.FormationAssignmentTypeRuntime:
		return model.RuntimeLabelableObject, nil
	case model.FormationAssignmentTypeRuntimeContext:
		return model.RuntimeContextLabelableObject, nil
	default:
		return "", errors.Errorf("Couldn't determine the label-able object type from assignment type: %q", assignmentType)
	}
}
