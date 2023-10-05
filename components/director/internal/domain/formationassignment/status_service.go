package formationassignment

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// formationAssignmentStatusService service encapsulates all the specifics around persisting the state reported by notification receiver for a formation assignment
type formationAssignmentStatusService struct {
	repo                  FormationAssignmentRepository
	constraintEngine      constraintEngine
	faNotificationService faNotificationService
}

// NewFormationAssignmentStatusService creates formation assignment status service
func NewFormationAssignmentStatusService(repo FormationAssignmentRepository, constraintEngine constraintEngine, faNotificationService faNotificationService) *formationAssignmentStatusService {
	return &formationAssignmentStatusService{
		repo:                  repo,
		constraintEngine:      constraintEngine,
		faNotificationService: faNotificationService,
	}
}

// UpdateWithConstraints updates a Formation Assignment and enforces NotificationStatusReturned constraints before and after the update
func (fau *formationAssignmentStatusService) UpdateWithConstraints(ctx context.Context, fa *model.FormationAssignment, operation model.FormationOperation) error {
	id := fa.ID

	log.C(ctx).Infof("Updating formation assignment with ID: %q", id)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	faFromDB, err := fau.repo.Get(ctx, id, tenantID)
	if err != nil {
		return errors.Wrapf(err, "while ensuring formation assignment with ID: %q exists", id)
	}

	joinPointDetails, err := fau.faNotificationService.PrepareDetailsForNotificationStatusReturned(ctx, tenantID, fa, operation, faFromDB.State)
	if err != nil {
		return errors.Wrap(err, "while preparing details for NotificationStatusReturned")
	}
	joinPointDetails.Location = formationconstraint.PreNotificationStatusReturned
	if err := fau.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PreOperation)
	}

	if err = fau.repo.Update(ctx, fa); err != nil {
		return errors.Wrapf(err, "while updating formation assignment with ID: %q", id)
	}

	joinPointDetails.Location = formationconstraint.PostNotificationStatusReturned
	if err := fau.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PostOperation)
	}

	return nil
}

// SetAssignmentToErrorStateWithConstraints updates Formation Assignment state to error state using the errorMessage, errorCode and state parameters.
// Also, it enforces NotificationStatusReturned constraints before and after the update.
func (fau *formationAssignmentStatusService) SetAssignmentToErrorStateWithConstraints(ctx context.Context, assignment *model.FormationAssignment, errorMessage string, errorCode AssignmentErrorCode, state model.FormationAssignmentState, operation model.FormationOperation) error {
	assignment.State = string(state)
	assignmentError := AssignmentErrorWrapper{AssignmentError{
		Message:   errorMessage,
		ErrorCode: errorCode,
	}}
	marshaled, err := json.Marshal(assignmentError)
	if err != nil {
		return errors.Wrapf(err, "While preparing error message for assignment with ID %q", assignment.ID)
	}
	assignment.Error = marshaled
	if err := fau.UpdateWithConstraints(ctx, assignment, operation); err != nil {
		return errors.Wrapf(err, "While updating formation assignment with id %q", assignment.ID)
	}
	log.C(ctx).Infof("Assignment with ID %s set to state %s", assignment.ID, assignment.State)
	return nil
}

// DeleteWithConstraints deletes a Formation Assignment matching ID `id` and enforces NotificationStatusReturned constraints before and after delete.
func (fau *formationAssignmentStatusService) DeleteWithConstraints(ctx context.Context, id string) error {
	log.C(ctx).Infof("Deleting formation assignment with ID: %q", id)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	fa, err := fau.repo.Get(ctx, id, tenantID)
	if err != nil {
		return errors.Wrapf(err, "while getting formation assignment with id %q for tenant with id %q", id, tenantID)
	}
	faStateFromDB := fa.State

	fa.State = string(model.ReadyAssignmentState)
	fa.Value = nil
	if err := fau.repo.Update(ctx, fa); err != nil {
		return errors.Wrapf(err, "while updating formation asssignment with ID: %s to: %q state", id, model.ReadyAssignmentState)
	}

	joinPointDetails, err := fau.faNotificationService.PrepareDetailsForNotificationStatusReturned(ctx, tenantID, fa, model.UnassignFormation, faStateFromDB)
	if err != nil {
		return errors.Wrap(err, "while preparing details for NotificationStatusReturned")
	}
	joinPointDetails.Location = formationconstraint.PreNotificationStatusReturned
	if err = fau.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PreOperation)
	}

	if err = fau.repo.Delete(ctx, id, tenantID); err != nil {
		return errors.Wrapf(err, "while deleting formation assignment with ID: %q", id)
	}

	joinPointDetails.Location = formationconstraint.PostNotificationStatusReturned
	if err = fau.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PostOperation)
	}

	return nil
}
