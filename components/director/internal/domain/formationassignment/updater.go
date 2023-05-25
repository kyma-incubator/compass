package formationassignment

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

type formationAssignmentUpdaterService struct {
	repo                        FormationAssignmentRepository
	constraintEngine            constraintEngine
	formationRepository         formationRepository
	formationTemplateRepository formationTemplateRepository
}

func NewFormationAssignmentUpdaterService(repo FormationAssignmentRepository, constraintEngine constraintEngine, formationRepository formationRepository, formationTemplateRepository formationTemplateRepository) *formationAssignmentUpdaterService {
	return &formationAssignmentUpdaterService{
		repo:                        repo,
		constraintEngine:            constraintEngine,
		formationRepository:         formationRepository,
		formationTemplateRepository: formationTemplateRepository,
	}
}

func (fau *formationAssignmentUpdaterService) Update(ctx context.Context, fa *model.FormationAssignment, operation model.FormationOperation) error {
	id := fa.ID

	log.C(ctx).Infof("Updating formation assignment with ID: %q", id)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	joinPointDetails, err := fau.prepareDetailsForNotificationStatusReturned(ctx, tenantID, fa, operation)
	if err != nil {
		return errors.Wrap(err, "while preparing details for NotificationStatusReturned")
	}
	joinPointDetails.Location = formationconstraint.PreNotificationStatusReturned
	if err := fau.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PreOperation)
	}

	if exists, err := fau.repo.Exists(ctx, id, tenantID); err != nil {
		return errors.Wrapf(err, "while ensuring formation assignment with ID: %q exists", id)
	} else if !exists {
		return apperrors.NewNotFoundError(resource.FormationAssignment, id)
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

func (fau *formationAssignmentUpdaterService) SetAssignmentToErrorState(ctx context.Context, assignment *model.FormationAssignment, errorMessage string, errorCode AssignmentErrorCode, state model.FormationAssignmentState, operation model.FormationOperation) error {
	assignment.State = string(state)
	assignmentError := AssignmentErrorWrapper{AssignmentError{
		Message:   errorMessage,
		ErrorCode: errorCode,
	}}
	marshaled, err := json.Marshal(assignmentError)
	if err != nil {
		return errors.Wrapf(err, "While preparing error message for assignment with ID %q", assignment.ID)
	}
	assignment.Value = marshaled
	if err := fau.Update(ctx, assignment, operation); err != nil {
		return errors.Wrapf(err, "While updating formation assignment with id %q", assignment.ID)
	}
	log.C(ctx).Infof("Assignment with ID %s set to state %s", assignment.ID, assignment.State)
	return nil
}

func (fau *formationAssignmentUpdaterService) getReverseBySourceAndTarget(ctx context.Context, formationID, sourceID, targetID string) (*model.FormationAssignment, error) {
	log.C(ctx).Infof("Getting reverse formation assignment for formation ID: %q and source: %q and target: %q", formationID, sourceID, targetID)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	reverseFA, err := fau.repo.GetReverseBySourceAndTarget(ctx, tenantID, formationID, sourceID, targetID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting reverse formation assignment for formation ID: %q and source: %q and target: %q", formationID, sourceID, targetID)
	}

	return reverseFA, nil
}

func (fau *formationAssignmentUpdaterService) prepareDetailsForNotificationStatusReturned(ctx context.Context, tenantID string, fa *model.FormationAssignment, operation model.FormationOperation) (*formationconstraint.NotificationStatusReturnedOperationDetails, error) {
	formation, err := fau.formationRepository.Get(ctx, fa.FormationID, tenantID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation with ID %q in tenant %q: %v", fa.FormationID, tenantID, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation with ID %q in tenant %q", fa.FormationID, tenantID)
	}

	template, err := fau.formationTemplateRepository.Get(ctx, formation.FormationTemplateID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation template by ID: %q: %v", formation.FormationTemplateID, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation template by ID: %q", formation.FormationTemplateID)
	}

	reverseFa, err := fau.getReverseBySourceAndTarget(ctx, formation.ID, fa.Source, fa.Target)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			log.C(ctx).Errorf("An error occurred while getting reverse formation assignment: %v", err)
			return nil, errors.Wrap(err, "An error occurred while getting reverse formation assignment")
		}
		log.C(ctx).Debugf("Reverse assignment with source %q and target %q in formation with ID %q is not found.", fa.Target, fa.Source, formation.ID)
	}

	return &formationconstraint.NotificationStatusReturnedOperationDetails{
		ResourceType:               model.FormationResourceType,
		ResourceSubtype:            template.Name,
		Operation:                  operation,
		FormationAssignment:        fa,
		ReverseFormationAssignment: reverseFa,
		Formation:                  formation,
		FormationTemplate:          template,
	}, nil
}
