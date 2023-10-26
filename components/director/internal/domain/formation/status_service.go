package formation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// formationStatusService service encapsulates all the specifics around persisting the state reported by notification receiver for a formation
type formationStatusService struct {
	formationRepository  FormationRepository
	labelDefRepository   labelDefRepository
	labelDefService      labelDefService
	notificationsService NotificationsService
	constraintEngine     constraintEngine
}

// NewFormationStatusService creates formation status service
func NewFormationStatusService(formationRepository FormationRepository, labelDefRepository labelDefRepository, labelDefService labelDefService, notificationsService NotificationsService, constraintEngine constraintEngine) *formationStatusService {
	return &formationStatusService{
		formationRepository:  formationRepository,
		labelDefRepository:   labelDefRepository,
		labelDefService:      labelDefService,
		notificationsService: notificationsService,
		constraintEngine:     constraintEngine,
	}
}

// UpdateWithConstraints updates formation and enforces NotificationStatusReturned constraints before and after update.
func (s *formationStatusService) UpdateWithConstraints(ctx context.Context, formation *model.Formation, operation model.FormationOperation) error {
	joinPointDetails, err := s.notificationsService.PrepareDetailsForNotificationStatusReturned(ctx, formation, operation)
	if err != nil {
		return errors.Wrap(err, "while preparing details for NotificationStatusReturned")
	}
	joinPointDetails.Location = formationconstraint.PreNotificationStatusReturned
	if err := s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PreOperation)
	}

	if err := s.formationRepository.Update(ctx, formation); err != nil {
		log.C(ctx).Errorf("An error occurred while updating formation with ID: %q", formation.ID)
		return errors.Wrapf(err, "An error occurred while updating formation with ID: %q", formation.ID)
	}

	joinPointDetails.Location = formationconstraint.PostNotificationStatusReturned
	if err := s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PostOperation)
	}

	return nil
}

// DeleteFormationEntityAndScenariosWithConstraints removes the formation name from scenarios label definitions and deletes the formation entity from the DB and enforces NotificationStatusReturned constraints before and after delete.
func (s *formationStatusService) DeleteFormationEntityAndScenariosWithConstraints(ctx context.Context, tnt string, formation *model.Formation) error {
	joinPointDetails, err := s.notificationsService.PrepareDetailsForNotificationStatusReturned(ctx, formation, model.DeleteFormation)
	if err != nil {
		return errors.Wrap(err, "while preparing details for NotificationStatusReturned")
	}
	joinPointDetails.Location = formationconstraint.PreNotificationStatusReturned
	if err := s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PreOperation)
	}

	if err := s.deleteFormationFromLabelDef(ctx, tnt, formation.Name); err != nil {
		return err
	}

	// TODO:: Currently we need to support both mechanisms of formation creation/deletion(through label definitions and Formations entity) for backwards compatibility
	if err := s.formationRepository.DeleteByName(ctx, tnt, formation.Name); err != nil {
		log.C(ctx).Errorf("An error occurred while deleting formation with name: %q", formation.Name)
		return errors.Wrapf(err, "An error occurred while deleting formation with name: %q", formation.Name)
	}

	joinPointDetails.Location = formationconstraint.PostNotificationStatusReturned
	if err := s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostNotificationStatusReturned, joinPointDetails, joinPointDetails.Formation.FormationTemplateID); err != nil {
		return errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PostOperation)
	}

	return nil
}

// SetFormationToErrorStateWithConstraints sets formation to error state and enforces NotificationStatusReturned constraints before and after update
func (s *formationStatusService) SetFormationToErrorStateWithConstraints(ctx context.Context, formation *model.Formation, errorMessage string, errorCode formationassignment.AssignmentErrorCode, state model.FormationState, operation model.FormationOperation) error {
	log.C(ctx).Infof("Setting formation with ID: %q to state: %q", formation.ID, state)
	formation.State = state

	formationError := formationassignment.AssignmentError{
		Message:   errorMessage,
		ErrorCode: errorCode,
	}

	marshaledErr, err := json.Marshal(formationError)
	if err != nil {
		return errors.Wrapf(err, "While preparing error message for formation with ID: %q", formation.ID)
	}
	formation.Error = marshaledErr

	if err := s.UpdateWithConstraints(ctx, formation, operation); err != nil {
		return err
	}
	return nil
}

func (s *formationStatusService) deleteFormationFromLabelDef(ctx context.Context, tnt, formationName string) error {
	def, err := s.labelDefRepository.GetByKey(ctx, tnt, model.ScenariosKey)
	if err != nil {
		return errors.Wrapf(err, "while getting `%s` label definition", model.ScenariosKey)
	}
	if def.Schema == nil {
		return fmt.Errorf("missing schema for `%s` label definition", model.ScenariosKey)
	}

	formationNames, err := labeldef.ParseFormationsFromSchema(def.Schema)
	if err != nil {
		return err
	}

	schema, err := labeldef.NewSchemaForFormations(deleteFormation(formationNames, formationName))
	if err != nil {
		return errors.Wrap(err, "while parsing scenarios")
	}

	if err = s.labelDefService.ValidateExistingLabelsAgainstSchema(ctx, schema, tnt, model.ScenariosKey); err != nil {
		return err
	}
	if err = s.labelDefService.ValidateAutomaticScenarioAssignmentAgainstSchema(ctx, schema, tnt, model.ScenariosKey); err != nil {
		return errors.Wrap(err, "while validating Scenario Assignments against a new schema")
	}

	return s.labelDefRepository.UpdateWithVersion(ctx, model.LabelDefinition{
		ID:      def.ID,
		Tenant:  tnt,
		Key:     model.ScenariosKey,
		Schema:  &schema,
		Version: def.Version,
	})
}
