package formation_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	preDetails  = fixDetailsForNotificationStatusReturned("formationType", model.DeleteFormation, formationconstraint.PreNotificationStatusReturned, &modelFormation)
	postDetails = fixDetailsForNotificationStatusReturned("formationType", model.DeleteFormation, formationconstraint.PostNotificationStatusReturned, &modelFormation)
)

func TestUpdateWithConstraints(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		Name               string
		FormationRepoFn    func() *automock.FormationRepository
		NotificationsSvcFn func() *automock.NotificationsService
		ConstraintEngineFn func() *automock.ConstraintEngine
		InputFormation     *model.Formation
		FormationOperation model.FormationOperation
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Update", ctx, &modelFormation).Return(nil).Once()
				return repo
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, formationconstraint.PostNotificationStatusReturned, postDetails, postDetails.Formation.FormationTemplateID).Return(nil).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
		},
		{
			Name: "Returns error when enforcing post constraints fails",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Update", ctx, &modelFormation).Return(nil).Once()
				return repo
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, formationconstraint.PostNotificationStatusReturned, postDetails, postDetails.Formation.FormationTemplateID).Return(testErr).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when update fails",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Update", ctx, &modelFormation).Return(testErr).Once()
				return repo
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when enforcing pre constraints fails",
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(testErr).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when enforcing pre constraints fails",
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(nil, testErr).Once()
				return notificationsSvc
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			formationRepo := &automock.FormationRepository{}
			if testCase.FormationRepoFn != nil {
				formationRepo = testCase.FormationRepoFn()
			}
			notificationsSvc := &automock.NotificationsService{}
			if testCase.NotificationsSvcFn != nil {
				notificationsSvc = testCase.NotificationsSvcFn()
			}
			engine := &automock.ConstraintEngine{}
			if testCase.ConstraintEngineFn != nil {
				engine = testCase.ConstraintEngineFn()
			}

			svc := formation.NewFormationStatusService(formationRepo, nil, nil, notificationsSvc, engine)

			// WHEN
			err := svc.UpdateWithConstraints(ctx, testCase.InputFormation, testCase.FormationOperation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, formationRepo, notificationsSvc, engine)
		})
	}
}

func TestSetFormationToErrorStateWithConstraints(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		Name               string
		FormationRepoFn    func() *automock.FormationRepository
		NotificationsSvcFn func() *automock.NotificationsService
		ConstraintEngineFn func() *automock.ConstraintEngine
		InputFormation     *model.Formation
		FormationOperation model.FormationOperation
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Update", ctx, &modelFormation).Return(nil).Once()
				return repo
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, formationconstraint.PostNotificationStatusReturned, postDetails, postDetails.Formation.FormationTemplateID).Return(nil).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
		},
		{
			Name: "Returns error when update fails",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Update", ctx, &modelFormation).Return(nil).Once()
				return repo
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, formationconstraint.PostNotificationStatusReturned, postDetails, postDetails.Formation.FormationTemplateID).Return(testErr).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			formationRepo := &automock.FormationRepository{}
			if testCase.FormationRepoFn != nil {
				formationRepo = testCase.FormationRepoFn()
			}
			notificationsSvc := &automock.NotificationsService{}
			if testCase.NotificationsSvcFn != nil {
				notificationsSvc = testCase.NotificationsSvcFn()
			}
			engine := &automock.ConstraintEngine{}
			if testCase.ConstraintEngineFn != nil {
				engine = testCase.ConstraintEngineFn()
			}

			svc := formation.NewFormationStatusService(formationRepo, nil, nil, notificationsSvc, engine)

			// WHEN
			err := svc.SetFormationToErrorStateWithConstraints(ctx, testCase.InputFormation, ErrMsg, formationassignment.TechnicalError, model.CreateErrorFormationState, testCase.FormationOperation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, formationRepo, notificationsSvc, engine)
		})
	}
}

func TestDeleteFormationEntityAndScenariosWithConstraints(t *testing.T) {
	ctx := context.Background()

	testSchema, err := labeldef.NewSchemaForFormations([]string{testScenario, testFormationName})
	assert.NoError(t, err)
	testSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, testSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{testScenario})
	assert.NoError(t, err)
	newSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, newSchema)

	nilSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, testSchema)
	nilSchemaLblDef.Schema = nil

	testCases := []struct {
		Name                 string
		FormationRepoFn      func() *automock.FormationRepository
		LabelDefRepositoryFn func() *automock.LabelDefRepository
		LabelDefServiceFn    func() *automock.LabelDefService
		NotificationsSvcFn   func() *automock.NotificationsService
		ConstraintEngineFn   func() *automock.ConstraintEngine
		InputFormation       *model.Formation
		FormationOperation   model.FormationOperation
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("DeleteByName", ctx, TntInternalID, modelFormation.Name).Return(nil).Once()
				return repo
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil).Once()
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil).Once()
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil).Once()
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil).Once()
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, formationconstraint.PostNotificationStatusReturned, postDetails, postDetails.Formation.FormationTemplateID).Return(nil).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
		},
		{
			Name: "Returns error when enforcing post constraints fails",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("DeleteByName", ctx, TntInternalID, modelFormation.Name).Return(nil).Once()
				return repo
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil).Once()
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil).Once()
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil).Once()
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil).Once()
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, formationconstraint.PostNotificationStatusReturned, postDetails, postDetails.Formation.FormationTemplateID).Return(testErr).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when delete fails",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("DeleteByName", ctx, TntInternalID, modelFormation.Name).Return(testErr).Once()
				return repo
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil).Once()
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil).Once()
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil).Once()
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil).Once()
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when update label def fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil).Once()
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(testErr).Once()
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil).Once()
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil).Once()
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when validate asa against schema fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil).Once()
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil).Once()
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(testErr).Once()
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when validate existing labels against schema fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil).Once()
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(testErr).Once()
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when getting label def fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(nil, testErr).Once()
				return labelDefRepo
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(nil).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when enforcing pre constraints fails",
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(preDetails, nil).Once()
				return notificationsSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, formationconstraint.PreNotificationStatusReturned, preDetails, preDetails.Formation.FormationTemplateID).Return(testErr).Once()
				return engine
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when enforcing pre constraints fails",
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationsSvc := &automock.NotificationsService{}
				notificationsSvc.On("PrepareDetailsForNotificationStatusReturned", ctx, &modelFormation, model.DeleteFormation).Return(nil, testErr).Once()
				return notificationsSvc
			},
			InputFormation:     &modelFormation,
			FormationOperation: model.DeleteFormation,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			formationRepo := &automock.FormationRepository{}
			if testCase.FormationRepoFn != nil {
				formationRepo = testCase.FormationRepoFn()
			}
			labelDefRepo := &automock.LabelDefRepository{}
			if testCase.LabelDefRepositoryFn != nil {
				labelDefRepo = testCase.LabelDefRepositoryFn()
			}
			labelDefSvc := &automock.LabelDefService{}
			if testCase.LabelDefServiceFn != nil {
				labelDefSvc = testCase.LabelDefServiceFn()
			}
			notificationsSvc := &automock.NotificationsService{}
			if testCase.NotificationsSvcFn != nil {
				notificationsSvc = testCase.NotificationsSvcFn()
			}
			engine := &automock.ConstraintEngine{}
			if testCase.ConstraintEngineFn != nil {
				engine = testCase.ConstraintEngineFn()
			}

			svc := formation.NewFormationStatusService(formationRepo, labelDefRepo, labelDefSvc, notificationsSvc, engine)

			// WHEN
			err := svc.DeleteFormationEntityAndScenariosWithConstraints(ctx, TntInternalID, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, formationRepo, notificationsSvc, engine, labelDefRepo, labelDefSvc)
		})
	}
}
