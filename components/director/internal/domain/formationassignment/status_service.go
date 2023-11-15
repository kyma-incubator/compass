package formationassignment

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/domain/statusreport"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationassignment"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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
func (fau *formationAssignmentStatusService) UpdateWithConstraints(ctx context.Context, notificationStatusReport *statusreport.NotificationStatusReport, fa *model.FormationAssignment, operation model.FormationOperation) error {
	id := fa.ID

	log.C(ctx).Infof("Updating formation assignment with ID: %q", id)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	joinPointDetails, err := fau.faNotificationService.PrepareDetailsForNotificationStatusReturned(ctx, tenantID, fa, operation, notificationStatusReport)
	if err != nil {
		return errors.Wrap(err, "while preparing details for NotificationStatusReturned")
	}
	joinPointDetails.Location = formationconstraint.PreNotificationStatusReturned
	if err := fau.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PreOperation)
	}

	stateFromReport := notificationStatusReport.State
	fa.State = stateFromReport

	if isErrorState(model.FormationAssignmentState(stateFromReport)) {
		assignmentError := json.RawMessage{}
		if notificationStatusReport.Error != "" {
			assignmentErrorWrapper := AssignmentErrorWrapper{AssignmentError{
				Message:   notificationStatusReport.Error,
				ErrorCode: ClientError,
			}}
			marshaled, err := json.Marshal(assignmentErrorWrapper)
			if err != nil {
				return errors.Wrapf(err, "While preparing error message for assignment with ID %q", fa.ID)
			}
			assignmentError = marshaled
		}
		fa.Error = assignmentError
	}

	if !isErrorState(model.FormationAssignmentState(stateFromReport)) {
		// todo alternative is to clear the error and overwrite the config only if the new one is not empty
		ResetAssignmentConfigAndError(fa)
		configFromReport := notificationStatusReport.Configuration
		// todo if there is a config in the form of \"\" or {} do we still want to update it
		if configFromReport != nil && !formationconstraintpkg.IsConfigEmpty(string(configFromReport)) {
			fa.Value = configFromReport
		}
	}

	if err = fau.repo.Update(ctx, fa); err != nil {
		if apperrors.IsUnauthorizedError(err) {
			return apperrors.NewNotFoundError(resource.FormationAssignment, id)
		}
		return errors.Wrapf(err, "while updating formation assignment with ID: %q", id)
	}

	joinPointDetails.Location = formationconstraint.PostNotificationStatusReturned
	if err := fau.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PostOperation)
	}

	log.C(ctx).Infof("Assignment with ID %s set to state %s", id, fa.State)
	return nil
}

// DeleteWithConstraints deletes a Formation Assignment matching ID `id` and enforces NotificationStatusReturned constraints before and after delete.
func (fau *formationAssignmentStatusService) DeleteWithConstraints(ctx context.Context, id string, notificationStatusReport *statusreport.NotificationStatusReport) error {
	log.C(ctx).Infof("Deleting formation assignment with ID: %q", id)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	fa, err := fau.repo.Get(ctx, id, tenantID)
	if err != nil {
		return errors.Wrapf(err, "while getting formation assignment with id %q for tenant with id %q", id, tenantID)
	}

	joinPointDetails, err := fau.faNotificationService.PrepareDetailsForNotificationStatusReturned(ctx, tenantID, fa, model.UnassignFormation, notificationStatusReport)
	if err != nil {
		return errors.Wrap(err, "while preparing details for NotificationStatusReturned")
	}
	joinPointDetails.Location = formationconstraint.PreNotificationStatusReturned
	if err = fau.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PreOperation)
	}

	if err = fau.repo.Delete(ctx, id, tenantID); err != nil {
		if apperrors.IsUnauthorizedError(err) {
			return apperrors.NewNotFoundError(resource.FormationAssignment, id)
		}
		return errors.Wrapf(err, "while deleting formation assignment with ID: %q", id)
	}

	joinPointDetails.Location = formationconstraint.PostNotificationStatusReturned
	if err = fau.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PostOperation)
	}

	return nil
}

func isErrorState(state model.FormationAssignmentState) bool {
	return state == model.CreateErrorAssignmentState || state == model.DeleteErrorAssignmentState
}
