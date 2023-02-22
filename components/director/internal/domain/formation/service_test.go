package formation_test

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"

	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceList(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	pageSize := 100
	cursor := "start"

	expectedFormationPage := &model.FormationPage{
		Data: []*model.Formation{&modelFormation},
		PageInfo: &pagination.Page{
			StartCursor: cursor,
			EndCursor:   "",
			HasNextPage: false,
		},
		TotalCount: 1,
	}

	testCases := []struct {
		Name                  string
		FormationRepoFn       func() *automock.FormationRepository
		InputID               string
		InputPageSize         int
		ExpectedFormationPage *model.FormationPage
		ExpectedErrMessage    string
	}{
		{
			Name: "Success",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("List", ctx, TntInternalID, pageSize, cursor).Return(expectedFormationPage, nil).Once()
				return repo
			},
			InputID:               FormationID,
			InputPageSize:         pageSize,
			ExpectedFormationPage: expectedFormationPage,
			ExpectedErrMessage:    "",
		},
		{
			Name: "Returns error when can't list formations",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("List", ctx, TntInternalID, pageSize, cursor).Return(nil, testErr).Once()
				return repo
			},
			InputID:               FormationID,
			InputPageSize:         pageSize,
			ExpectedFormationPage: nil,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name:                  "Returns error when page size is not between 1 and 200",
			InputID:               FormationID,
			InputPageSize:         300,
			ExpectedFormationPage: nil,
			ExpectedErrMessage:    "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			formationRepo := unusedFormationRepo()
			if testCase.FormationRepoFn != nil {
				formationRepo = testCase.FormationRepoFn()
			}

			svc := formation.NewService(nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			actual, err := svc.List(ctx, testCase.InputPageSize, cursor)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormationPage, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			formationRepo.AssertExpectations(t)
		})
	}
}

func TestServiceGet(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	testCases := []struct {
		Name               string
		FormationRepoFn    func() *automock.FormationRepository
		InputID            string
		ExpectedFormation  *model.Formation
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(&modelFormation, nil).Once()
				return repo
			},
			InputID:            FormationID,
			ExpectedFormation:  &modelFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when can't get the formation",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            FormationID,
			ExpectedFormation:  nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			formationRepo := testCase.FormationRepoFn()

			svc := formation.NewService(nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			actual, err := svc.Get(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			formationRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetFormationByName(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	testCases := []struct {
		Name               string
		FormationRepoFn    func() *automock.FormationRepository
		Input              string
		ExpectedFormation  *model.Formation
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, TntInternalID).Return(&modelFormation, nil).Once()
				return repo
			},
			Input:              testFormationName,
			ExpectedFormation:  &modelFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when can't get the formation",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, TntInternalID).Return(nil, testErr).Once()
				return repo
			},
			Input:              testFormationName,
			ExpectedFormation:  nil,
			ExpectedErrMessage: "An error occurred while getting formation by name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			formationRepo := testCase.FormationRepoFn()

			svc := formation.NewService(nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			actual, err := svc.GetFormationByName(ctx, testCase.Input, TntInternalID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			formationRepo.AssertExpectations(t)
		})
	}
}

func TestServiceCreateFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	in := model.Formation{
		Name: testFormationName,
	}

	expectedFormation := &model.Formation{
		ID:                  FormationID,
		TenantID:            TntInternalID,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
		State:               model.ReadyFormationState,
	}

	formationWithReadyState := fixFormationModelWithState(model.ReadyFormationState)
	formationWithCreateErrorStateAndClientAssignmentError := fixFormationModelWithStateAndAssignmentError(t, model.CreateErrorFormationState, testErr.Error(), formationassignment.ClientError)
	formationWithCreateErrorStateAndTechnicalAssignmentError := fixFormationModelWithStateAndAssignmentError(t, model.CreateErrorFormationState, testErr.Error(), formationassignment.TechnicalError)

	testSchema, err := labeldef.NewSchemaForFormations([]string{testScenario})
	assert.NoError(t, err)
	testSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, testSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{testScenario, testFormationName})
	assert.NoError(t, err)
	newSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, newSchema)

	emptySchemaLblDef := fixScenariosLabelDefinition(TntInternalID, testSchemaLblDef)
	emptySchemaLblDef.Schema = nil

	testCases := []struct {
		Name                    string
		UUIDServiceFn           func() *automock.UuidService
		LabelDefRepositoryFn    func() *automock.LabelDefRepository
		LabelDefServiceFn       func() *automock.LabelDefService
		NotificationsSvcFn      func() *automock.NotificationsService
		FormationTemplateRepoFn func() *automock.FormationTemplateRepository
		FormationRepoFn         func() *automock.FormationRepository
		ConstraintEngineFn      func() *automock.ConstraintEngine
		webhookRepoFn           func() *automock.WebhookRepository
		TemplateName            string
		ExpectedFormation       *model.Formation
		ExpectedErrMessage      string
	}{
		{
			Name: "success when no labeldef exists",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, TntInternalID, []string{testFormationName}).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.CreateFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success when labeldef exists",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.CreateFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error when labeldef is missing and can not create it",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, TntInternalID, []string{testFormationName}).Return(testErr)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when labeldef is missing and create formation fails",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, TntInternalID, []string{testFormationName}).Return(nil)
				return labelDefService
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(testErr).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "An error occurred while creating formation with name",
		},
		{
			Name: "error when can not get labeldef",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(nil, testErr)
				return labelDefRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			LabelDefServiceFn:  unusedLabelDefService,
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "while getting `scenarios` label definition: Test error",
		},
		{
			Name: "error when labeldef's schema is missing",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&emptySchemaLblDef, nil)
				return labelDefRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			LabelDefServiceFn:  unusedLabelDefService,
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "missing schema",
		},
		{
			Name: "error when validating existing labels against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, testSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when validating automatic scenario assignment against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, testSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, testSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "while validating Scenario Assignments against a new schema",
		},
		{
			Name: "error when update with version fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(testErr)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, testSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, testSchemaLblDef.Key).Return(nil)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when getting formation template by name fails",
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(nil, testErr).Once()
				return formationTemplateRepoMock
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "An error occurred while getting formation template by name",
		},
		{
			Name: "error when creating formation fails",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(testErr).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "An error occurred while creating formation with name: \"test-formation\"",
		},
		{
			Name: "error while enforcing constraint pre operation",
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "While enforcing constraints for target operation \"CREATE_FORMATION\" and constraint type \"PRE\": Test error",
		},
		{
			Name: "error when listing formation template's webhooks fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(nil, testErr).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "when listing formation lifecycle webhooks for formation template with ID",
		},
		{
			Name: "error while enforcing constraints post operation",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, TntInternalID, []string{testFormationName}).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.CreateFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postCreateLocation, createFormationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "While enforcing constraints for target operation \"CREATE_FORMATION\" and constraint type \"POST\": Test error",
		},
		{
			Name: "Success when there is formation notifications",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, formationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.CreateFormation).Return(formationNotificationRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationRequest).Return(formationNotificationWebhookSuccessResponse, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(nil).Once()
				formationRepoMock.On("Update", ctx, formationWithReadyState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedFormation:  formationWithReadyState,
			ExpectedErrMessage: "",
		},
		{
			Name: "Error when generating formation notification fails",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, formationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.CreateFormation).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedFormation:  nil,
			ExpectedErrMessage: "while generating notifications for formation with ID",
		},
		{
			Name: "Error when sending formation notifications fails",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, formationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.CreateFormation).Return(formationNotificationRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationRequest).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(nil).Once()
				formationRepoMock.On("Update", ctx, formationWithCreateErrorStateAndTechnicalAssignmentError).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedFormation:  nil,
			ExpectedErrMessage: "while sending notification for formation with ID",
		},
		{
			Name: "Error when sending formation notifications fails and subsequently formation update fails",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, formationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.CreateFormation).Return(formationNotificationRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationRequest).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(nil).Once()
				formationRepoMock.On("Update", ctx, formationWithCreateErrorStateAndTechnicalAssignmentError).Return(testErr).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedFormation:  nil,
			ExpectedErrMessage: "while updating error state",
		},
		{
			Name: "Error when there is formation notifications, webhook response status is incorrect and formation update fails",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, formationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.CreateFormation).Return(formationNotificationRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationRequest).Return(formationNotificationWebhookErrorResponse, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(nil).Once()
				formationRepoMock.On("Update", ctx, formationWithCreateErrorStateAndClientAssignmentError).Return(testErr).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedFormation:  nil,
			ExpectedErrMessage: "while updating error state for formation with ID",
		},
		{
			Name: "Error when there is formation notifications but webhook response status is incorrect",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, formationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.CreateFormation).Return(formationNotificationRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationRequest).Return(formationNotificationWebhookErrorResponse, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(nil).Once()
				formationRepoMock.On("Update", ctx, formationWithCreateErrorStateAndClientAssignmentError).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedFormation:  nil,
			ExpectedErrMessage: fmt.Sprintf("Received error from formation webhook response: %s", testErr.Error()),
		},
		{
			Name: "Error when there is formation notifications, webhook response status is correct but the formation update fails",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, formationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.CreateFormation).Return(formationNotificationRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationRequest).Return(formationNotificationWebhookSuccessResponse, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithReadyState).Return(nil).Once()
				formationRepoMock.On("Update", ctx, formationWithReadyState).Return(testErr).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedFormation:  nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			uidService := unusedUUIDService()
			if testCase.UUIDServiceFn != nil {
				uidService = testCase.UUIDServiceFn()
			}
			labelDefRepo := unusedLabelDefRepository()
			if testCase.LabelDefServiceFn != nil {
				labelDefRepo = testCase.LabelDefRepositoryFn()
			}
			labelDefService := unusedLabelDefService()
			if testCase.LabelDefServiceFn != nil {
				labelDefService = testCase.LabelDefServiceFn()
			}

			notificationsService := unusedNotificationsService()
			if testCase.NotificationsSvcFn != nil {
				notificationsService = testCase.NotificationsSvcFn()
			}

			formationRepo := unusedFormationRepo()
			if testCase.FormationRepoFn != nil {
				formationRepo = testCase.FormationRepoFn()
			}
			formationTemplateRepo := unusedFormationTemplateRepo()
			if testCase.FormationTemplateRepoFn != nil {
				formationTemplateRepo = testCase.FormationTemplateRepoFn()
			}
			constraintEngine := unusedConstraintEngine()
			if testCase.ConstraintEngineFn != nil {
				constraintEngine = testCase.ConstraintEngineFn()
			}

			webhookRepo := unusedWebhookRepository()
			if testCase.webhookRepoFn != nil {
				webhookRepo = testCase.webhookRepoFn()
			}

			svc := formation.NewService(nil, labelDefRepo, nil, formationRepo, formationTemplateRepo, nil, uidService, labelDefService, nil, nil, nil, nil, nil, nil, nil, notificationsService, constraintEngine, webhookRepo, runtimeType, applicationType)

			// WHEN
			actual, err := svc.CreateFormation(ctx, TntInternalID, in, testCase.TemplateName)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, uidService, labelDefRepo, labelDefService, notificationsService, formationRepo, formationTemplateRepo, constraintEngine, webhookRepo)
		})
	}
}

func TestServiceDeleteFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	in := model.Formation{
		Name: testFormationName,
	}

	formationWithCreateErrorStateAndTechnicalAssignmentError := fixFormationModelWithStateAndAssignmentError(t, model.DeleteErrorFormationState, testErr.Error(), formationassignment.TechnicalError)

	expectedFormation := &model.Formation{
		ID:                  FormationID,
		TenantID:            TntInternalID,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
		State:               model.ReadyFormationState,
	}

	formationWithReadyState := fixFormationModelWithState(model.ReadyFormationState)

	testSchema, err := labeldef.NewSchemaForFormations([]string{testScenario, testFormationName})
	assert.NoError(t, err)
	testSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, testSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{testScenario})
	assert.NoError(t, err)
	newSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, newSchema)

	nilSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, testSchema)
	nilSchemaLblDef.Schema = nil

	testCases := []struct {
		Name                    string
		LabelDefRepositoryFn    func() *automock.LabelDefRepository
		LabelDefServiceFn       func() *automock.LabelDefService
		NotificationsSvcFn      func() *automock.NotificationsService
		FormationRepoFn         func() *automock.FormationRepository
		FormationTemplateRepoFn func() *automock.FormationTemplateRepository
		ConstraintEngineFn      func() *automock.ConstraintEngine
		webhookRepoFn           func() *automock.WebhookRepository
		InputFormation          model.Formation
		ExpectedFormation       *model.Formation
		ExpectedErrMessage      string
	}{
		{
			Name: "success",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("DeleteByName", ctx, TntInternalID, testFormationName).Return(nil).Once()
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error when can not get labeldef",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(nil, testErr).Once()
				return labelDefRepo
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedErrMessage: "while getting `scenarios` label definition: Test error",
		},
		{
			Name: "error when labeldef's schema is missing",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&nilSchemaLblDef, nil)
				return labelDefRepo
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedErrMessage: "missing schema",
		},
		{
			Name: "error when validating existing labels against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(testErr)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when validating automatic scenario assignment against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(testErr)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedErrMessage: "while validating Scenario Assignments against a new schema: Test error",
		},
		{
			Name: "error when update with version fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&newSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(testErr)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, newSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, newSchemaLblDef.Key).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when can't get formation by name",
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(nil, testErr).Once()
				return formationRepoMock
			},
			InputFormation:     in,
			ExpectedFormation:  nil,
			ExpectedErrMessage: "While deleting formation: An error occurred while getting formation by name: \"test-formation\": Test error",
		},
		{
			Name: "error when deleting formation by name fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				formationRepoMock.On("DeleteByName", ctx, TntInternalID, testFormationName).Return(testErr).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedErrMessage: "An error occurred while deleting formation with name: \"test-formation\": Test error",
		},
		{
			Name: "error while enforcing constraints pre operation",
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			InputFormation:     in,
			ExpectedErrMessage: "While enforcing constraints for target operation \"DELETE_FORMATION\" and constraint type \"PRE\": Test error",
		},
		{
			Name: "error while enforcing constraint post operation",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("DeleteByName", ctx, TntInternalID, testFormationName).Return(nil).Once()
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedErrMessage: "While enforcing constraints for target operation \"DELETE_FORMATION\" and constraint type \"POST\": Test error",
		},
		{
			Name: "Error when listing formation template's webhooks fails",
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(nil, testErr).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedFormation:  nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when generating formation notifications fails",
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.DeleteFormation).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedFormation:  nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when processing formation notifications fails",
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, FormationTemplateID, model.DeleteFormation).Return(formationNotificationRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationRequest).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Update", ctx, formationWithCreateErrorStateAndTechnicalAssignmentError).Return(testErr).Once()
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedFormation:  nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			labelDefRepo := unusedLabelDefRepository()
			if testCase.LabelDefRepositoryFn != nil {
				labelDefRepo = testCase.LabelDefRepositoryFn()
			}

			labelDefService := unusedLabelDefService()
			if testCase.LabelDefServiceFn != nil {
				labelDefService = testCase.LabelDefServiceFn()
			}

			notificationsService := unusedNotificationsService()
			if testCase.NotificationsSvcFn != nil {
				notificationsService = testCase.NotificationsSvcFn()
			}

			formationRepo := unusedFormationRepo()
			if testCase.FormationRepoFn != nil {
				formationRepo = testCase.FormationRepoFn()
			}

			formationTemplateRepo := unusedFormationTemplateRepo()
			if testCase.FormationTemplateRepoFn != nil {
				formationTemplateRepo = testCase.FormationTemplateRepoFn()
			}

			constraintEngine := unusedConstraintEngine()
			if testCase.ConstraintEngineFn != nil {
				constraintEngine = testCase.ConstraintEngineFn()
			}

			webhookRepo := unusedWebhookRepository()
			if testCase.webhookRepoFn != nil {
				webhookRepo = testCase.webhookRepoFn()
			}

			svc := formation.NewService(nil, labelDefRepo, nil, formationRepo, formationTemplateRepo, nil, nil, labelDefService, nil, nil, nil, nil, nil, nil, nil, notificationsService, constraintEngine, webhookRepo, runtimeType, applicationType)

			// WHEN
			actual, err := svc.DeleteFormation(ctx, TntInternalID, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, labelDefRepo, labelDefService, notificationsService, formationRepo, formationTemplateRepo, constraintEngine)
		})
	}
}

func TestService_DeleteManyASAForSameTargetTenant(t *testing.T) {
	ctx := fixCtxWithTenant()

	scenarioNameA := "scenario-A"
	scenarioNameB := "scenario-B"
	models := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   scenarioNameA,
			TargetTenantID: TargetTenantID,
		},
		{
			ScenarioName:   scenarioNameB,
			TargetTenantID: TargetTenantID,
		},
	}

	formations := []*model.Formation{
		{
			ID:                  FormationID,
			TenantID:            tenantID.String(),
			FormationTemplateID: FormationTemplateID,
			Name:                scenarioNameA,
		},
		{
			ID:                  FormationID,
			TenantID:            tenantID.String(),
			FormationTemplateID: FormationTemplateID,
			Name:                scenarioNameB,
		},
	}

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForTargetTenant", ctx, tenantID.String(), TargetTenantID).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(make([]*model.Runtime, 0), nil).Twice()

		runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(make([]*model.Runtime, 0), nil).Twice()

		formationRepo := &automock.FormationRepository{}
		formationRepo.On("GetByName", ctx, scenarioNameA, "").Return(formations[0], nil).Once()
		formationRepo.On("GetByName", ctx, scenarioNameB, "").Return(formations[1], nil).Once()

		formationTemplateRepo := &automock.FormationTemplateRepository{}
		formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
		formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()

		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, formationRepo, formationTemplateRepo)

		svc := formation.NewService(nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.NoError(t, err)
	})

	t.Run("return error when removing assigned scenarios fails", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForTargetTenant", ctx, tenantID.String(), TargetTenantID).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(nil, fixError())

		formationRepo := &automock.FormationRepository{}

		formationRepo.On("GetByName", ctx, scenarioNameA, "").Return(formations[0], nil).Once()

		formationTemplateRepo := &automock.FormationTemplateRepository{}
		formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()

		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, formationRepo, formationTemplateRepo)

		svc := formation.NewService(nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("return error when input slice is empty", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, []*model.AutomaticScenarioAssignment{})

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected at least one item in Assignments slice")
	})

	t.Run("return error when input slice contains assignments with different selectors", func(t *testing.T) {
		// GIVEN
		modelsWithDifferentSelectors := []*model.AutomaticScenarioAssignment{
			{
				ScenarioName:   scenarioNameA,
				TargetTenantID: TargetTenantID,
			},
			{
				ScenarioName:   scenarioNameB,
				TargetTenantID: "differentTargetTenantID",
			},
		}

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, modelsWithDifferentSelectors)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "all input items have to have the same target tenant")
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}

		mockRepo.On("DeleteForTargetTenant", ctx, tenantID.String(), TargetTenantID).Return(fixError()).Once()

		defer mock.AssertExpectationsForObjects(t, mockRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignments: %s", ErrMsg))
	})

	t.Run("returns error when empty tenant", func(t *testing.T) {
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
		err := svc.DeleteManyASAForSameTargetTenant(context.TODO(), models)
		require.EqualError(t, err, "cannot read tenant from context")
	})
}

func TestService_CreateAutomaticScenarioAssignment(t *testing.T) {
	ctx := fixCtxWithTenant()

	testCases := []struct {
		Name               string
		LabelDefServiceFn  func() *automock.LabelDefService
		AsaRepoFn          func() *automock.AutomaticFormationAssignmentRepository
		AsaEngineFN        func() *automock.AsaEngine
		InputASA           model.AutomaticScenarioAssignment
		ExpectedASA        model.AutomaticScenarioAssignment
		ExpectedErrMessage string
	}{
		{
			Name: "happy path",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(nil).Once()
				return mockRepo
			},
			AsaEngineFN: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("EnsureScenarioAssigned", ctx, fixModel(testFormationName), mock.Anything).Return(nil).Once()
				return engine
			},
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        fixModel(testFormationName),
			ExpectedErrMessage: "",
		},
		{
			Name: "returns error on getting available scenarios from label definition",
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}
				labelDefSvc.On("GetAvailableScenarios", mock.Anything, tenantID.String()).Return(nil, fixError()).Once()
				return labelDefSvc
			},
			AsaRepoFn:          unusedASARepo,
			AsaEngineFN:        unusedASAEngine,
			InputASA:           fixModel(ScenarioName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: "while getting available scenarios: some error",
		},
		{
			Name: "returns error on creating asa",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(fixError()).Once()
				return mockRepo
			},
			AsaEngineFN:        unusedASAEngine,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: "while persisting Assignment",
		},

		{Name: "returns error on creating asa",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(apperrors.NewNotUniqueError(resource.AutomaticScenarioAssigment)).Once()
				return mockRepo
			},
			AsaEngineFN:        unusedASAEngine,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: "a given scenario already has an assignment",
		},
		{
			Name: "returns error on ensure scenario assigned",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(nil).Once()
				return mockRepo
			},
			AsaEngineFN: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("EnsureScenarioAssigned", ctx, fixModel(testFormationName), mock.Anything).Return(fixError()).Once()
				return engine
			},
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: "while assigning scenario to runtimes matching selector",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			asaRepo := testCase.AsaRepoFn()
			tenantSvc := &automock.TenantService{}
			labelDefService := testCase.LabelDefServiceFn()
			asaEngine := testCase.AsaEngineFN()

			svc := formation.NewServiceWithAsaEngine(nil, nil, nil, nil, nil, nil, nil, labelDefService, asaRepo, nil, tenantSvc, nil, nil, nil, nil, nil, nil, runtimeType, applicationType, asaEngine)

			// WHEN
			actual, err := svc.CreateAutomaticScenarioAssignment(ctx, testCase.InputASA)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedASA, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Equal(t, testCase.ExpectedASA, actual)
			}

			mock.AssertExpectationsForObjects(t, tenantSvc, asaRepo, labelDefService, asaEngine)
		})
	}

	t.Run("returns error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(context.TODO(), fixModel(ScenarioName))

		// THEN
		assert.EqualError(t, err, "cannot read tenant from context")
	})
}

func TestService_DeleteAutomaticScenarioAssignment(t *testing.T) {
	ctx := fixCtxWithTenant()

	testErr := errors.New("test err")

	testCases := []struct {
		Name               string
		AsaRepoFn          func() *automock.AutomaticFormationAssignmentRepository
		AsaEngineFN        func() *automock.AsaEngine
		InputASA           model.AutomaticScenarioAssignment
		ExpectedASA        model.AutomaticScenarioAssignment
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(nil).Once()
				return mockRepo
			},
			AsaEngineFN: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.Mock.On("RemoveAssignedScenario", ctx, fixModel(testFormationName), mock.Anything).Return(nil).Once()
				return engine
			},
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        fixModel(testFormationName),
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when deleting ASA for scenario name fails",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(testErr).Once()
				return mockRepo
			},
			AsaEngineFN:        unusedASAEngine,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: "while deleting the Assignment",
		},
		{
			Name: "Returns error when removing assigned scenario",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(nil).Once()
				return mockRepo
			},
			AsaEngineFN: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.Mock.On("RemoveAssignedScenario", ctx, fixModel(testFormationName), mock.Anything).Return(testErr).Once()
				return engine
			},
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: "while unassigning scenario from runtimes",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			asaRepo := testCase.AsaRepoFn()
			asaEngine := testCase.AsaEngineFN()
			tenantSvc := &automock.TenantService{}

			svc := formation.NewServiceWithAsaEngine(nil, nil, nil, nil, nil, nil, nil, nil, asaRepo, nil, tenantSvc, nil, nil, nil, nil, nil, nil, runtimeType, applicationType, asaEngine)

			// WHEN
			err := svc.DeleteAutomaticScenarioAssignment(ctx, testCase.InputASA)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, tenantSvc, asaRepo, asaEngine)
		})
	}

	t.Run("returns error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(context.TODO(), fixModel(ScenarioName))

		// THEN
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_MergeScenariosFromInputLabelsAndAssignments(t *testing.T) {
	// GIVEN
	ctx := fixCtxWithTenant()

	differentTargetTenant := "differentTargetTenant"
	runtimeID := "runtimeID"
	labelKey := "key"
	labelValue := "val"

	scenario := "SCENARIO"

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   ScenarioName,
			Tenant:         tenantID.String(),
			TargetTenantID: TargetTenantID,
		},
		{
			ScenarioName:   ScenarioName2,
			Tenant:         tenantID.String(),
			TargetTenantID: differentTargetTenant,
		},
	}

	formations := []*model.Formation{
		{
			ID:                  FormationID,
			TenantID:            tenantID.String(),
			FormationTemplateID: FormationTemplateID,
			Name:                ScenarioName,
		},
		{
			ID:                  FormationID,
			TenantID:            tenantID.String(),
			FormationTemplateID: FormationTemplateID,
			Name:                ScenarioName2,
		},
	}

	testCases := []struct {
		Name                    string
		AsaRepoFn               func() *automock.AutomaticFormationAssignmentRepository
		RuntimeContextRepoFn    func() *automock.RuntimeContextRepository
		RuntimeRepoFn           func() *automock.RuntimeRepository
		FormationRepoFn         func() *automock.FormationRepository
		FormationTemplateRepoFn func() *automock.FormationTemplateRepository
		InputLabels             map[string]interface{}
		ExpectedScenarios       []interface{}
		ExpectedErrMessage      string
	}{
		{
			Name: "Success",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, tenantID.String()).Return(assignments, nil)
				return asaRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, runtimeID).Return(false, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, TargetTenantID, runtimeID, runtimeLblFilters).Return(true, nil).Once()
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, differentTargetTenant, runtimeID, runtimeLblFilters).Return(false, nil).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, tenantID.String()).Return(formations[0], nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName2, tenantID.String()).Return(formations[1], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			InputLabels: map[string]interface{}{
				labelKey: labelValue,
			},
			ExpectedScenarios:  []interface{}{ScenarioName},
			ExpectedErrMessage: "",
		},
		{
			Name: "Success if scenarios label is in input",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, tenantID.String()).Return(assignments, nil)
				return asaRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, runtimeID).Return(false, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, TargetTenantID, runtimeID, runtimeLblFilters).Return(true, nil).Once()
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, differentTargetTenant, runtimeID, runtimeLblFilters).Return(false, nil).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, tenantID.String()).Return(formations[0], nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName2, tenantID.String()).Return(formations[1], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			InputLabels: map[string]interface{}{
				labelKey:           labelValue,
				model.ScenariosKey: []interface{}{scenario},
			},
			ExpectedScenarios:  []interface{}{ScenarioName, scenario},
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when checking if ASA is matching to runtime fails",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, tenantID.String()).Return(assignments, nil)
				return asaRepo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, TargetTenantID, runtimeID, runtimeLblFilters).Return(false, testErr).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, tenantID.String()).Return(formations[0], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			InputLabels: map[string]interface{}{
				labelKey: labelValue,
			},
			ExpectedScenarios:  []interface{}{},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when scenarios from input are not interface slice",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, tenantID.String()).Return(assignments, nil)
				return asaRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, runtimeID).Return(false, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, TargetTenantID, runtimeID, runtimeLblFilters).Return(true, nil).Once()
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, differentTargetTenant, runtimeID, runtimeLblFilters).Return(false, nil).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, tenantID.String()).Return(formations[0], nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName2, tenantID.String()).Return(formations[1], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			InputLabels: map[string]interface{}{
				labelKey:           labelValue,
				model.ScenariosKey: []string{scenario},
			},
			ExpectedScenarios:  []interface{}{},
			ExpectedErrMessage: "while converting scenarios label",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			asaRepo := testCase.AsaRepoFn()
			runtimeRepo := testCase.RuntimeRepoFn()
			runtimeContextRepo := testCase.RuntimeContextRepoFn()
			formationRepo := testCase.FormationRepoFn()
			formationTemplateRepo := testCase.FormationTemplateRepoFn()

			svc := formation.NewService(nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			actualScenarios, err := svc.MergeScenariosFromInputLabelsAndAssignments(ctx, testCase.InputLabels, runtimeID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				require.ElementsMatch(t, testCase.ExpectedScenarios, actualScenarios)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo, formationTemplateRepo, formationRepo)
		})
	}
}

func TestService_GetFormationsForObject(t *testing.T) {
	id := "rtmID"
	testErr := "testErr"

	scenarios := []interface{}{"scenario1", "scenario2"}

	labelInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		ObjectID:   id,
		ObjectType: model.RuntimeLabelableObject,
	}

	label := &model.Label{
		ID:         id,
		Key:        "scenarios",
		Value:      scenarios,
		ObjectID:   id,
		ObjectType: model.RuntimeLabelableObject,
	}

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), labelInput).Return(label, nil).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, labelService, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		formations, err := svc.GetFormationsForObject(ctx, tenantID.String(), model.RuntimeLabelableObject, id)

		// THEN
		require.NoError(t, err)
		require.ElementsMatch(t, formations, scenarios)
		mock.AssertExpectationsForObjects(t, labelService)
	})

	t.Run("Returns error while getting label", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), labelInput).Return(nil, errors.New(testErr)).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, labelService, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		formations, err := svc.GetFormationsForObject(ctx, tenantID.String(), model.RuntimeLabelableObject, id)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "while fetching scenario label for")
		require.Nil(t, formations)
		mock.AssertExpectationsForObjects(t, labelService)
	})
}

func TestServiceResynchronizeFormationNotifications(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	allStates := []string{string(model.InitialAssignmentState),
		string(model.DeletingAssignmentState),
		string(model.CreateErrorAssignmentState),
		string(model.DeleteErrorAssignmentState)}

	testErr := errors.New("test error")

	testFormation := &model.Formation{ID: FormationID, Name: "fid"}

	formationAssignments := []*model.FormationAssignment{
		{
			ID:          "id1",
			FormationID: FormationID,
			Source:      "RT1",
			SourceType:  model.FormationAssignmentTypeRuntime,
			Target:      "A1",
			TargetType:  model.FormationAssignmentTypeApplication,
			State:       string(model.InitialAssignmentState),
		},
		{
			ID:          "id2",
			FormationID: FormationID,
			Source:      "RTCTX1",
			SourceType:  model.FormationAssignmentTypeRuntimeContext,
			Target:      "A1",
			TargetType:  model.FormationAssignmentTypeApplication,
			State:       string(model.CreateErrorFormationState),
		},
		{
			ID:          "id3",
			FormationID: FormationID,
			Source:      "RT1",
			SourceType:  model.FormationAssignmentTypeRuntime,
			Target:      "RTCTX1",
			TargetType:  model.FormationAssignmentTypeRuntimeContext,
			State:       string(model.DeletingFormationState),
		},
		{
			ID:          "id4",
			FormationID: FormationID,
			Source:      "RTCTX1",
			SourceType:  model.FormationAssignmentTypeRuntimeContext,
			Target:      "RTCTX1",
			TargetType:  model.FormationAssignmentTypeRuntimeContext,
			State:       string(model.DeleteErrorFormationState),
		},
	}
	reverseAssignment := &model.FormationAssignment{
		ID:          "id1",
		FormationID: FormationID,
		Source:      "A1",
		SourceType:  model.FormationAssignmentTypeApplication,
		Target:      "RT1",
		TargetType:  model.FormationAssignmentTypeRuntime,
		State:       string(model.ReadyAssignmentState),
	}

	notificationsForAssignments := []*webhookclient.FormationAssignmentNotificationRequest{
		{
			Webhook: graphql.Webhook{
				ID: "faID1",
			},
		},
		{
			Webhook: graphql.Webhook{
				ID: "faID2",
			},
		},
		{
			Webhook: graphql.Webhook{
				ID: "faID3",
			},
		},
		{
			Webhook: graphql.Webhook{
				ID: "faID4",
			},
		},
	}
	var formationAssignmentPairs []*formationassignment.AssignmentMappingPair
	for i := range formationAssignments {
		formationAssignmentPairs = append(formationAssignmentPairs, &formationassignment.AssignmentMappingPair{
			Assignment: &formationassignment.FormationAssignmentRequestMapping{
				Request:             notificationsForAssignments[i],
				FormationAssignment: formationAssignments[i],
			},
			ReverseAssignment: &formationassignment.FormationAssignmentRequestMapping{
				Request:             nil,
				FormationAssignment: nil,
			},
		})
	}

	testCases := []struct {
		Name                                     string
		LabelServiceFn                           func() *automock.LabelService
		LabelRepoFn                              func() *automock.LabelRepository
		AsaEngineFN                              func() *automock.AsaEngine
		FormationRepositoryFn                    func() *automock.FormationRepository
		FormationAssignmentNotificationServiceFN func() *automock.FormationAssignmentNotificationsService
		NotificationServiceFN                    func() *automock.NotificationsService
		FormationAssignmentServiceFn             func() *automock.FormationAssignmentService
		RuntimeContextRepoFn                     func() *automock.RuntimeContextRepository
		ExpectedErrMessage                       string
	}{
		{
			Name: "success when resynchronization is successful and there are leftover formation assignments",
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", ctx, TntInternalID, FormationID, allStates).Return(formationAssignments, nil)

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", ctx, FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, ""))
				}

				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[0]).Return(false, nil)
				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[1]).Return(false, nil)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[2]).Return(false, nil)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[3]).Return(false, nil)

				svc.On("ListFormationAssignmentsForObjectID", ctx, FormationID, formationAssignments[3].Source).Return([]*model.FormationAssignment{{ID: "id6"}}, nil)
				svc.On("ListFormationAssignmentsForObjectID", ctx, FormationID, formationAssignments[3].Target).Return([]*model.FormationAssignment{{ID: "id7"}}, nil)
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[0]).Return(notificationsForAssignments[0], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[1]).Return(notificationsForAssignments[1], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[2]).Return(notificationsForAssignments[2], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[3]).Return(notificationsForAssignments[3], nil)
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(nil, nil)
				return repo
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "success when resynchronization is successful and there no left formation assignments should unassign",
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", ctx, TntInternalID, FormationID, allStates).Return(formationAssignments, nil)

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", ctx, FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, ""))
				}

				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[0]).Return(false, nil)
				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[1]).Return(false, nil)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[2]).Return(false, nil)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[3]).Return(false, nil)

				svc.On("ListFormationAssignmentsForObjectID", ctx, FormationID, formationAssignments[3].Source).Return([]*model.FormationAssignment{{ID: "id6"}}, nil)
				svc.On("ListFormationAssignmentsForObjectID", ctx, FormationID, formationAssignments[3].Target).Return([]*model.FormationAssignment{}, nil)
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[0]).Return(notificationsForAssignments[0], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[1]).Return(notificationsForAssignments[1], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[2]).Return(notificationsForAssignments[2], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[3]).Return(notificationsForAssignments[3], nil)
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(testFormation, nil)
				return repo
			},
			AsaEngineFN: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, "RTCTX1", testFormation.Name, graphql.FormationObjectTypeRuntimeContext).Return(true, nil)
				return engine
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "returns error when failing to unassign from formation after resynchronizing",
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", ctx, TntInternalID, FormationID, allStates).Return(formationAssignments, nil)

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", ctx, FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, ""))
				}

				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[0]).Return(false, nil)
				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[1]).Return(false, nil)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[2]).Return(false, nil)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[3]).Return(false, nil)

				svc.On("ListFormationAssignmentsForObjectID", ctx, FormationID, formationAssignments[3].Target).Return([]*model.FormationAssignment{}, nil)

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[0]).Return(notificationsForAssignments[0], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[1]).Return(notificationsForAssignments[1], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[2]).Return(notificationsForAssignments[2], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[3]).Return(notificationsForAssignments[3], nil)
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(testFormation, nil)
				return repo
			},
			AsaEngineFN: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, "RTCTX1", testFormation.Name, graphql.FormationObjectTypeRuntimeContext).Return(false, testErr)
				return engine
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when failing to list formation assignments for participant",
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", ctx, TntInternalID, FormationID, allStates).Return(formationAssignments, nil)

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", ctx, FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, ""))
				}

				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[0]).Return(false, nil)
				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[1]).Return(false, nil)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[2]).Return(false, nil)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[3]).Return(false, nil)

				svc.On("ListFormationAssignmentsForObjectID", ctx, FormationID, mock.Anything).Return(nil, testErr)
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[0]).Return(notificationsForAssignments[0], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[1]).Return(notificationsForAssignments[1], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[2]).Return(notificationsForAssignments[2], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[3]).Return(notificationsForAssignments[3], nil)
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(testFormation, nil)
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when failing to get formation",
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", ctx, TntInternalID, FormationID, allStates).Return(formationAssignments, nil)

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", ctx, FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, ""))
				}

				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[0]).Return(false, nil)
				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[1]).Return(false, nil)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[2]).Return(false, nil)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[3]).Return(false, nil)

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[0]).Return(notificationsForAssignments[0], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[1]).Return(notificationsForAssignments[1], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[2]).Return(notificationsForAssignments[2], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[3]).Return(notificationsForAssignments[3], nil)
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(nil, testErr)
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when failing processing formation assignments fails",
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", ctx, TntInternalID, FormationID, allStates).Return(formationAssignments, nil)

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", ctx, FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, ""))
				}

				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[0]).Return(false, testErr)
				svc.On("ProcessFormationAssignmentPair", ctx, formationAssignmentPairs[1]).Return(false, testErr)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[2]).Return(false, testErr)
				svc.On("CleanupFormationAssignment", ctx, formationAssignmentPairs[3]).Return(false, testErr)

				svc.On("ListFormationAssignmentsForObjectID", ctx, FormationID, formationAssignments[3].Source).Return([]*model.FormationAssignment{{ID: "id6"}}, nil)
				svc.On("ListFormationAssignmentsForObjectID", ctx, FormationID, formationAssignments[3].Target).Return([]*model.FormationAssignment{}, nil)
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[0]).Return(notificationsForAssignments[0], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[1]).Return(notificationsForAssignments[1], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[2]).Return(notificationsForAssignments[2], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[3]).Return(notificationsForAssignments[3], nil)
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(testFormation, nil)
				return repo
			},
			AsaEngineFN: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, "RTCTX1", testFormation.Name, graphql.FormationObjectTypeRuntimeContext).Return(true, nil)
				return engine
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when getting reverse formation assignment",
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", ctx, TntInternalID, FormationID, allStates).Return([]*model.FormationAssignment{formationAssignments[0]}, nil)

				svc.On("GetReverseBySourceAndTarget", ctx, FormationID, formationAssignments[0].Source, formationAssignments[0].Target).Return(nil, testErr)

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[0]).Return(notificationsForAssignments[0], nil)
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when generating reverse formation assignment notification",
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", ctx, TntInternalID, FormationID, allStates).Return([]*model.FormationAssignment{formationAssignments[0]}, nil)

				svc.On("GetReverseBySourceAndTarget", ctx, FormationID, formationAssignments[0].Source, formationAssignments[0].Target).Return(reverseAssignment, nil)

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[0]).Return(notificationsForAssignments[0], nil)
				svc.On("GenerateFormationAssignmentNotification", ctx, reverseAssignment).Return(nil, testErr)
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when generating formation assignment notification",
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", ctx, TntInternalID, FormationID, allStates).Return([]*model.FormationAssignment{formationAssignments[0]}, nil)

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", ctx, formationAssignments[0]).Return(nil, testErr)
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when listing formation assignments with states",
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", ctx, TntInternalID, FormationID, allStates).Return(nil, testErr)

				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			labelService := unusedLabelService()
			if testCase.LabelServiceFn != nil {
				labelService = testCase.LabelServiceFn()
			}
			runtimeContextRepo := unusedRuntimeContextRepo()
			if testCase.RuntimeContextRepoFn != nil {
				runtimeContextRepo = testCase.RuntimeContextRepoFn()
			}
			formationRepo := unusedFormationRepo()
			if testCase.FormationRepositoryFn != nil {
				formationRepo = testCase.FormationRepositoryFn()
			}
			labelRepo := unusedLabelRepo()
			if testCase.LabelRepoFn != nil {
				labelRepo = testCase.LabelRepoFn()
			}
			notificationsSvc := unusedNotificationsService()
			if testCase.NotificationServiceFN != nil {
				notificationsSvc = testCase.NotificationServiceFN()
			}
			formationAssignmentSvc := unusedFormationAssignmentService()
			if testCase.FormationAssignmentServiceFn != nil {
				formationAssignmentSvc = testCase.FormationAssignmentServiceFn()
			}
			asaEngine := unusedASAEngine()
			if testCase.AsaEngineFN != nil {
				asaEngine = testCase.AsaEngineFN()
			}
			formationAssignmentNotificationService := unusedFormationAssignmentNotificationService()
			if testCase.FormationAssignmentNotificationServiceFN != nil {
				formationAssignmentNotificationService = testCase.FormationAssignmentNotificationServiceFN()
			}
			svc := formation.NewServiceWithAsaEngine(nil, nil, labelRepo, formationRepo, nil, labelService, nil, nil, nil, nil, nil, nil, runtimeContextRepo, formationAssignmentSvc, formationAssignmentNotificationService, notificationsSvc, nil, runtimeType, applicationType, asaEngine)

			// WHEN
			err := svc.ResynchronizeFormationNotifications(ctx, FormationID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
			mock.AssertExpectationsForObjects(t, labelService, runtimeContextRepo, formationRepo, labelRepo, notificationsSvc, formationAssignmentSvc, formationAssignmentNotificationService)
		})
		t.Run("returns error when empty tenant", func(t *testing.T) {
			svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
			err := svc.ResynchronizeFormationNotifications(context.TODO(), FormationID)
			require.Contains(t, err.Error(), "cannot read tenant from context")
		})
	}
}
