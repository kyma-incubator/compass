package formation_test

import (
	"context"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

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

			svc := formation.NewService(nil, nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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
			InputID:           FormationID,
			ExpectedFormation: &modelFormation,
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

			svc := formation.NewService(nil, nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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

func TestService_ListFormationsForObject(t *testing.T) {
	ctx := context.TODO()

	testCases := []struct {
		Name                     string
		FormationAssignmentSvcFn func() *automock.FormationAssignmentService
		FormationRepoFn          func() *automock.FormationRepository
		Input                    string
		ExpectedFormations       []*model.Formation
		ExpectedErrMessage       string
	}{
		{
			Name: "Success",
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("ListAllForObjectGlobal", ctx, ApplicationID).Return(formationAssignments, nil).Once()
				return svc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListByIDs", ctx, mock.MatchedBy(func(formationIDs []string) bool {
					return assert.ElementsMatch(t, formationIDs, []string{FormationID, FormationID2})
				})).Return(modelFormations, nil).Once()
				return repo
			},
			Input:              ApplicationID,
			ExpectedFormations: modelFormations,
		},
		{
			Name: "Success when no assignments are returned",
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("ListAllForObjectGlobal", ctx, ApplicationID).Return(nil, nil).Once()
				return svc
			},
			Input:              ApplicationID,
			ExpectedFormations: nil,
		},
		{
			Name: "Error when listing formations",
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("ListAllForObjectGlobal", ctx, ApplicationID).Return(formationAssignments, nil).Once()
				return svc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListByIDs", ctx, mock.MatchedBy(func(formationIDs []string) bool {
					return assert.ElementsMatch(t, formationIDs, []string{FormationID, FormationID2})
				})).Return(nil, testErr).Once()
				return repo
			},
			Input:              ApplicationID,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when listing formation assignments",
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("ListAllForObjectGlobal", ctx, ApplicationID).Return(nil, testErr).Once()
				return svc
			},
			Input:              ApplicationID,
			ExpectedErrMessage: fmt.Sprintf("while listing formations assignments for participant with ID %s", ApplicationID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			formationAssignmentService := &automock.FormationAssignmentService{}
			if testCase.FormationAssignmentSvcFn != nil {
				formationAssignmentService = testCase.FormationAssignmentSvcFn()
			}
			formationRepo := &automock.FormationRepository{}
			if testCase.FormationRepoFn != nil {
				formationRepo = testCase.FormationRepoFn()
			}

			svc := formation.NewService(nil, nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, formationAssignmentService, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			actual, err := svc.ListFormationsForObject(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormations, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, formationRepo, formationAssignmentService)
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
			Input:             testFormationName,
			ExpectedFormation: &modelFormation,
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

			svc := formation.NewService(nil, nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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

func TestService_GetGlobalByID(t *testing.T) {
	ctx := context.TODO()
	ctxWithTenant := tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	testCases := []struct {
		Name               string
		FormationRepoFn    func() *automock.FormationRepository
		ExpectedFormation  *model.Formation
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetGlobalByID", ctxWithTenant, FormationID).Return(&modelFormation, nil).Once()
				return repo
			},
			ExpectedFormation: &modelFormation,
		},
		{
			Name: "Error when getting formation globally fails",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetGlobalByID", ctxWithTenant, FormationID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedFormation:  nil,
			ExpectedErrMessage: fmt.Sprintf("An error occurred while getting formation by ID: %q globally", FormationID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			formationRepo := testCase.FormationRepoFn()
			defer formationRepo.AssertExpectations(t)

			svc := formation.NewService(nil, nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			actual, err := svc.GetGlobalByID(ctxWithTenant, FormationID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	ctx := context.TODO()
	ctxWithTenant := tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	testCases := []struct {
		Name               string
		FormationRepoFn    func() *automock.FormationRepository
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Update", ctxWithTenant, &modelFormation).Return(nil).Once()
				return repo
			},
		},
		{
			Name: "Error when updating formation fails",
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Update", ctxWithTenant, &modelFormation).Return(testErr).Once()
				return repo
			},
			ExpectedErrMessage: fmt.Sprintf("An error occurred while updating formation with ID: %q", FormationID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			formationRepo := testCase.FormationRepoFn()
			defer formationRepo.AssertExpectations(t)

			svc := formation.NewService(nil, nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			err := svc.Update(ctxWithTenant, &modelFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestServiceCreateFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	in := model.Formation{
		Name: testFormationName,
	}

	expectedFormation := fixFormationModelWithState(model.ReadyFormationState)
	expectedFormationInInitialState := fixFormationModelWithState(model.InitialFormationState)
	expectedFormationInDraftState := fixFormationModelWithState(model.DraftFormationState)

	formationWithReadyState := fixFormationModelWithState(model.ReadyFormationState)
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
		FormationInput          *model.Formation
		UUIDServiceFn           func() *automock.UuidService
		LabelDefRepositoryFn    func() *automock.LabelDefRepository
		LabelDefServiceFn       func() *automock.LabelDefService
		NotificationsSvcFn      func() *automock.NotificationsService
		FormationTemplateRepoFn func() *automock.FormationTemplateRepository
		FormationRepoFn         func() *automock.FormationRepository
		ConstraintEngineFn      func() *automock.ConstraintEngine
		webhookRepoFn           func() *automock.WebhookRepository
		StatusServiceFn         func() *automock.StatusService
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
				labelDefRepo.On("GetByKey", ctxWithTenantAndLoggerMatcher(), TntInternalID, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctxWithTenantAndLoggerMatcher(), TntInternalID, []string{testFormationName}).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctxWithTenantAndLoggerMatcher(), testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctxWithTenantAndLoggerMatcher(), formationWithReadyState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctxWithTenantAndLoggerMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:      testFormationTemplateName,
			ExpectedFormation: expectedFormation,
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
				labelDefRepo.On("GetByKey", ctxWithTenantAndLoggerMatcher(), TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctxWithTenantAndLoggerMatcher(), newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctxWithTenantAndLoggerMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctxWithTenantAndLoggerMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctxWithTenantAndLoggerMatcher(), testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctxWithTenantAndLoggerMatcher(), formationWithReadyState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctxWithTenantAndLoggerMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:      testFormationTemplateName,
			ExpectedFormation: expectedFormation,
		},
		{
			Name: "success when state is provided externally - initial state",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctxWithTenantAndLoggerMatcher(), TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctxWithTenantAndLoggerMatcher(), newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctxWithTenantAndLoggerMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctxWithTenantAndLoggerMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), emptyFormationLifecycleWebhooks, TntInternalID, formationWithInitialState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctxWithTenantAndLoggerMatcher(), testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctxWithTenantAndLoggerMatcher(), formationWithInitialState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctxWithTenantAndLoggerMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			FormationInput: &model.Formation{
				Name:  testFormationName,
				State: model.InitialFormationState,
			},
			TemplateName:      testFormationTemplateName,
			ExpectedFormation: expectedFormationInInitialState,
		},
		{
			Name: "success when state is provided externally - draft state",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctxWithTenantAndLoggerMatcher(), TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctxWithTenantAndLoggerMatcher(), newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctxWithTenantAndLoggerMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctxWithTenantAndLoggerMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctxWithTenantAndLoggerMatcher(), testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctxWithTenantAndLoggerMatcher(), formationWithDraftState).Return(nil).Once()
				return formationRepoMock
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctxWithTenantAndLoggerMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			FormationInput: &model.Formation{
				Name:  testFormationName,
				State: model.DraftFormationState,
			},
			TemplateName:      testFormationTemplateName,
			ExpectedFormation: expectedFormationInDraftState,
		},
		{
			Name: "error when state is provided externally - draft state, while enforcing constraints post operation",
			UUIDServiceFn: func() *automock.UuidService {
				uuidService := &automock.UuidService{}
				uuidService.On("Generate").Return(fixUUID())
				return uuidService
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctxWithTenantAndLoggerMatcher(), TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctxWithTenantAndLoggerMatcher(), newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctxWithTenantAndLoggerMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctxWithTenantAndLoggerMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctxWithTenantAndLoggerMatcher(), testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctxWithTenantAndLoggerMatcher(), formationWithDraftState).Return(nil).Once()
				return formationRepoMock
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctxWithTenantAndLoggerMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postCreateLocation, createFormationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			FormationInput: &model.Formation{
				Name:  testFormationName,
				State: model.DraftFormationState,
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when state is provided externally - draft state, while enforcing constraints pre operation",
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
			FormationInput: &model.Formation{
				Name:  testFormationName,
				State: model.DraftFormationState,
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: testErr.Error(),
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
			ExpectedErrMessage: fmt.Sprintf("An error occurred while creating formation with name: %q", testFormationName),
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
			ExpectedErrMessage: "while enforcing constraints for target operation \"CREATE_FORMATION\" and constraint type \"PRE\": Test error",
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
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(emptyFormationNotificationRequests, nil).Once()
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
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postCreateLocation, createFormationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "while enforcing constraints for target operation \"CREATE_FORMATION\" and constraint type \"POST\": Test error",
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
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), formationLifecycleSyncWebhooks, TntInternalID, formationWithInitialState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", ctxWithTenantAndLoggerMatcher(), formationNotificationSyncCreateRequest).Return(formationNotificationWebhookSuccessResponse, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithInitialState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("UpdateWithConstraints", ctxWithTenantAndLoggerMatcher(), formationWithReadyState, model.CreateFormation).Return(nil).Once()
				return svc
			},
			TemplateName:      testFormationTemplateName,
			ExpectedFormation: formationWithReadyState,
		},
		{
			Name: "Success when there is formation notifications with ASYNC_CALLBACK",
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
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), formationLifecycleAsyncWebhooks, TntInternalID, formationWithInitialState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationAsyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", ctxWithTenantAndLoggerMatcher(), formationNotificationAsyncCreateRequest).Return(formationNotificationWebhookSuccessResponse, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithInitialState).Return(nil).Once()
				formationRepoMock.On("Update", ctxWithTenantAndLoggerMatcher(), formationWithInitialState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleAsyncWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:      testFormationTemplateName,
			ExpectedFormation: formationWithInitialState,
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
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), formationLifecycleSyncWebhooks, TntInternalID, formationWithInitialState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", ctxWithTenantAndLoggerMatcher(), formationNotificationSyncCreateRequest).Return(formationNotificationWebhookErrorResponse, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithInitialState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("SetFormationToErrorStateWithConstraints", ctxWithTenantAndLoggerMatcher(), formationWithInitialState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorFormationState, model.CreateFormation).Return(nil).Once()
				return svc
			},
			TemplateName:      testFormationTemplateName,
			ExpectedFormation: formationWithInitialState,
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
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), formationLifecycleSyncWebhooks, TntInternalID, formationWithInitialState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithInitialState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
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
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), formationLifecycleSyncWebhooks, TntInternalID, formationWithInitialState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", ctxWithTenantAndLoggerMatcher(), formationNotificationSyncCreateRequest).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithInitialState).Return(nil).Once()
				formationRepoMock.On("Update", ctxWithTenantAndLoggerMatcher(), formationWithCreateErrorStateAndTechnicalAssignmentError).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
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
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), formationLifecycleSyncWebhooks, TntInternalID, formationWithInitialState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", ctxWithTenantAndLoggerMatcher(), formationNotificationSyncCreateRequest).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithInitialState).Return(nil).Once()
				formationRepoMock.On("Update", ctxWithTenantAndLoggerMatcher(), formationWithCreateErrorStateAndTechnicalAssignmentError).Return(testErr).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			TemplateName:       testFormationTemplateName,
			ExpectedFormation:  nil,
			ExpectedErrMessage: "while updating error state",
		},
		{
			Name: "Error when there is formation notifications, webhook response status is incorrect and setting formation to error state with constraints fails",
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
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), formationLifecycleSyncWebhooks, TntInternalID, formationWithInitialState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", ctxWithTenantAndLoggerMatcher(), formationNotificationSyncCreateRequest).Return(formationNotificationWebhookErrorResponse, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithInitialState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("SetFormationToErrorStateWithConstraints", ctxWithTenantAndLoggerMatcher(), formationWithInitialState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorFormationState, model.CreateFormation).Return(testErr).Once()
				return svc
			},
			TemplateName:       testFormationTemplateName,
			ExpectedFormation:  nil,
			ExpectedErrMessage: "while updating error state for formation with ID",
		},
		{
			Name: "Error when there is formation notifications, webhook response status is correct but the formation update with constraints fails",
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
				notificationSvc.On("GenerateFormationNotifications", ctxWithTenantAndLoggerMatcher(), formationLifecycleSyncWebhooks, TntInternalID, formationWithInitialState, testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", ctxWithTenantAndLoggerMatcher(), formationNotificationSyncCreateRequest).Return(formationNotificationWebhookSuccessResponse, nil).Once()
				return notificationSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, TntInternalID).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, formationWithInitialState).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("UpdateWithConstraints", ctxWithTenantAndLoggerMatcher(), formationWithReadyState, model.CreateFormation).Return(testErr).Once()
				return svc
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			input := testCase.FormationInput
			if input == nil {
				input = &in
			}

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

			statusSvc := &automock.StatusService{}
			if testCase.StatusServiceFn != nil {
				statusSvc = testCase.StatusServiceFn()
			}

			svc := formation.NewService(nil, nil, labelDefRepo, nil, formationRepo, formationTemplateRepo, nil, uidService, labelDefService, nil, nil, nil, nil, nil, nil, nil, nil, notificationsService, constraintEngine, webhookRepo, statusSvc, runtimeType, applicationType)

			// WHEN
			actual, err := svc.CreateFormation(ctx, TntInternalID, *input, testCase.TemplateName)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, uidService, labelDefRepo, labelDefService, notificationsService, formationRepo, formationTemplateRepo, constraintEngine, webhookRepo, statusSvc)
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

	expectedFormation := fixFormationModelWithState(model.ReadyFormationState)
	expectedFormationInitialState := fixFormationModelWithState(model.InitialFormationState)
	expectedFormationDraftState := fixFormationModelWithState(model.DraftFormationState)
	expectedFormation2 := fixFormationModelWithState(model.ReadyFormationState)
	formationWithDeletingState := fixFormationModelWithState(model.DeletingFormationState)

	formationWithReadyState := fixFormationModelWithState(model.ReadyFormationState)
	formationWithDraftState := fixFormationModelWithState(model.DraftFormationState)

	testSchema, err := labeldef.NewSchemaForFormations([]string{testScenario, testFormationName})
	assert.NoError(t, err)
	testSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, testSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{testScenario})
	assert.NoError(t, err)
	newSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, newSchema)

	nilSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, testSchema)
	nilSchemaLblDef.Schema = nil

	testCases := []struct {
		Name                             string
		LabelDefRepositoryFn             func() *automock.LabelDefRepository
		LabelDefServiceFn                func() *automock.LabelDefService
		NotificationsSvcFn               func() *automock.NotificationsService
		FormationRepoFn                  func() *automock.FormationRepository
		FormationAssignmentSvcFn         func() *automock.FormationAssignmentService
		FormationTemplateRepoFn          func() *automock.FormationTemplateRepository
		AutomaticScenarioAssignmentSvcFn func() *automock.AutomaticFormationAssignmentService
		ConstraintEngineFn               func() *automock.ConstraintEngine
		webhookRepoFn                    func() *automock.WebhookRepository
		InputFormation                   model.Formation
		ExpectedFormation                *model.Formation
		ExpectedErrMessage               string
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("DeleteByName", ctx, TntInternalID, testFormationName).Return(nil).Once()
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
			InputFormation:    in,
			ExpectedFormation: expectedFormation,
		},
		{
			Name: "success for draft formation should not execute notifications",
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithDraftState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("DeleteByName", ctx, TntInternalID, testFormationName).Return(nil).Once()
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormationDraftState, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
			InputFormation:    in,
			ExpectedFormation: expectedFormationDraftState,
		},
		{
			Name: "success for initial formation should not execute notifications when there are no webhooks",
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, expectedFormationInitialState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("DeleteByName", ctx, TntInternalID, testFormationName).Return(nil).Once()
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormationInitialState, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
			InputFormation:    in,
			ExpectedFormation: expectedFormationInitialState,
		},
		{
			Name: "success when formation has async webhook",
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, formationLifecycleAsyncWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(formationNotificationAsyncDeleteRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationAsyncDeleteRequest).Return(fixFormationNotificationWebhookResponse(200, 200, nil), nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(fixFormationModelWithState(model.ReadyFormationState), nil).Once()
				formationRepoMock.On("Update", ctx, fixFormationModelWithState(model.DeletingFormationState)).Return(nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleAsyncWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:    in,
			ExpectedFormation: fixFormationModelWithState(model.DeletingFormationState),
		},
		{
			Name: "error when formation has async webhook and updating to deleting state fails",
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, formationLifecycleAsyncWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(formationNotificationAsyncDeleteRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationAsyncDeleteRequest).Return(fixFormationNotificationWebhookResponse(200, 200, nil), nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(fixFormationModelWithState(model.ReadyFormationState), nil).Once()
				formationRepoMock.On("Update", ctx, fixFormationModelWithState(model.DeletingFormationState)).Return(testErr).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleAsyncWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Success when formation has async webhook and updating to deleting state returns unauthorized error",
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, formationLifecycleAsyncWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(formationNotificationAsyncDeleteRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationAsyncDeleteRequest).Return(fixFormationNotificationWebhookResponse(200, 200, nil), nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(fixFormationModelWithState(model.ReadyFormationState), nil).Once()
				formationRepoMock.On("Update", ctx, fixFormationModelWithState(model.DeletingFormationState)).Return(unauthorizedError).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleAsyncWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:    in,
			ExpectedFormation: formationWithDeletingState,
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
			ExpectedErrMessage: fmt.Sprintf("while deleting formation: An error occurred while getting formation by name: %q: Test error", testFormationName),
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				formationRepoMock.On("DeleteByName", ctx, TntInternalID, testFormationName).Return(testErr).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
			ExpectedErrMessage: fmt.Sprintf("An error occurred while deleting formation with name: %q: Test error", testFormationName),
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
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preDeleteLocation, deleteFormationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			InputFormation:     in,
			ExpectedErrMessage: "while enforcing constraints for target operation \"DELETE_FORMATION\" and constraint type \"PRE\": Test error",
		},
		{
			Name: "error because formation is not empty",
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
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return([]*model.FormationAssignment{{
					ID:     FormationAssignmentID,
					Source: FormationAssignmentSource,
					Target: FormationAssignmentTarget,
				}}, nil)
				return formationAssignmentSvc
			},
			InputFormation:     in,
			ExpectedErrMessage: "because it is not empty",
		},
		{
			Name: "error while getting formation assignments",
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
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(nil, testErr)
				return formationAssignmentSvc
			},
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(emptyFormationNotificationRequests, nil).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("DeleteByName", ctx, TntInternalID, testFormationName).Return(nil).Once()
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
			ExpectedErrMessage: "while enforcing constraints for target operation \"DELETE_FORMATION\" and constraint type \"POST\": Test error",
		},
		{
			Name: "Error when listing formation template's webhooks fails",
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(formationNotificationAsyncDeleteRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationAsyncDeleteRequest).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Update", ctx, formationWithCreateErrorStateAndTechnicalAssignmentError).Return(testErr).Once()
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
			Name: "Error when processing formation notifications fails and formation update returns unauthorized error",
			NotificationsSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", ctx, emptyFormationLifecycleWebhooks, TntInternalID, formationWithReadyState, testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(formationNotificationAsyncDeleteRequests, nil).Once()
				notificationSvc.On("SendNotification", ctx, formationNotificationAsyncDeleteRequest).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Update", ctx, formationWithCreateErrorStateAndTechnicalAssignmentError).Return(unauthorizedError).Once()
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation2, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, nil).Once()
				return svc
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
			Name: "Error when there is subaccount assigned to the formation",
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation2, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(fixModel(in.Name), nil).Once()
				return svc
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedFormation:  nil,
			ExpectedErrMessage: fmt.Sprintf("cannot delete formation with ID %q, because there is still a subaccount part of it", expectedFormation2.ID),
		},
		{
			Name: "Error when checking for assigned subaccount",
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation2, nil).Once()
				return formationRepoMock
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GetAssignmentsForFormation", ctx, TntInternalID, FormationID).Return(emptyFormationAssignments, nil)
				return formationAssignmentSvc
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			AutomaticScenarioAssignmentSvcFn: func() *automock.AutomaticFormationAssignmentService {
				svc := &automock.AutomaticFormationAssignmentService{}
				svc.On("GetForScenarioName", ctx, expectedFormation.Name).Return(nil, testErr).Once()
				return svc
			},
			webhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(emptyFormationLifecycleWebhooks, nil).Once()
				return webhookRepo
			},
			InputFormation:     in,
			ExpectedFormation:  nil,
			ExpectedErrMessage: fmt.Sprintf("while getting automatic scenario assignment for formation with name %q", expectedFormation2.Name),
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
			asaService := unusedASAService()
			if testCase.AutomaticScenarioAssignmentSvcFn != nil {
				asaService = testCase.AutomaticScenarioAssignmentSvcFn()
			}
			constraintEngine := unusedConstraintEngine()
			if testCase.ConstraintEngineFn != nil {
				constraintEngine = testCase.ConstraintEngineFn()
			}
			webhookRepo := unusedWebhookRepository()
			if testCase.webhookRepoFn != nil {
				webhookRepo = testCase.webhookRepoFn()
			}
			formationAssignmentService := unusedFormationAssignmentService()
			if testCase.FormationAssignmentSvcFn != nil {
				formationAssignmentService = testCase.FormationAssignmentSvcFn()
			}

			svc := formation.NewService(nil, nil, labelDefRepo, nil, formationRepo, formationTemplateRepo, nil, nil, labelDefService, nil, asaService, nil, nil, nil, formationAssignmentService, nil, nil, notificationsService, constraintEngine, webhookRepo, nil, runtimeType, applicationType)

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

			mock.AssertExpectationsForObjects(t, labelDefRepo, labelDefService, notificationsService, formationRepo, formationTemplateRepo, asaService, constraintEngine)
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

		svc := formation.NewService(nil, nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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

		svc := formation.NewService(nil, nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("return error when input slice is empty", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
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

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignments: %s", ErrMsg))
	})

	t.Run("returns error when empty tenant", func(t *testing.T) {
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
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
		InputASA           *model.AutomaticScenarioAssignment
		ExpectedASA        *model.AutomaticScenarioAssignment
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
			InputASA:    fixModel(testFormationName),
			ExpectedASA: fixModel(testFormationName),
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
			ExpectedASA:        nil,
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
			ExpectedASA:        nil,
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
			ExpectedASA:        nil,
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
			ExpectedASA:        nil,
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

			svc := formation.NewServiceWithAsaEngine(nil, nil, nil, nil, nil, nil, nil, nil, labelDefService, asaRepo, nil, tenantSvc, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType, asaEngine, nil, nil)

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
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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
		InputASA           *model.AutomaticScenarioAssignment
		ExpectedASA        *model.AutomaticScenarioAssignment
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
				engine.Mock.On("UnassignFormationComingFromASA", ctx, fixModel(testFormationName), mock.Anything).Return(nil).Once()
				return engine
			},
			InputASA:    fixModel(testFormationName),
			ExpectedASA: fixModel(testFormationName),
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
			ExpectedASA:        nil,
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
				engine.Mock.On("UnassignFormationComingFromASA", ctx, fixModel(testFormationName), mock.Anything).Return(testErr).Once()
				return engine
			},
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        nil,
			ExpectedErrMessage: "while unassigning scenario from runtimes",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			asaRepo := testCase.AsaRepoFn()
			asaEngine := testCase.AsaEngineFN()
			tenantSvc := &automock.TenantService{}

			svc := formation.NewServiceWithAsaEngine(nil, nil, nil, nil, nil, nil, nil, nil, nil, asaRepo, nil, tenantSvc, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType, asaEngine, nil, nil)

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
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(context.TODO(), fixModel(ScenarioName))

		// THEN
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestServiceResynchronizeFormationNotifications(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	transactionError := errors.New("transaction error")
	txGen := txtest.NewTransactionContextGenerator(transactionError)

	allStates := model.ResynchronizableFormationAssignmentStates

	testFormation := fixFormationModelWithState(model.ReadyFormationState)
	formationInCreateErrorState := fixFormationModelWithStateAndAssignmentError(t, model.CreateErrorFormationState, testErr.Error(), formationassignment.ClientError)
	formationInCreateErrorStateTechnicalError := fixFormationModelWithStateAndAssignmentError(t, model.CreateErrorFormationState, testErr.Error(), formationassignment.TechnicalError)

	runtimeLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(TntInternalID),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName, secondTestFormationName},
		ObjectID:   RuntimeContextID,
		ObjectType: model.RuntimeContextLabelableObject,
		Version:    0,
	}
	runtimeLblInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   RuntimeContextID,
		ObjectType: model.RuntimeContextLabelableObject,
		Version:    0,
	}

	fa1 := fixFormationAssignmentModelWithParameters("id1", FormationID, RuntimeID, ApplicationID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication, model.InitialAssignmentState)
	fa2 := fixFormationAssignmentModelWithParameters("id2", FormationID, RuntimeContextID, ApplicationID, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeApplication, model.CreateErrorAssignmentState)
	fa4 := fixFormationAssignmentModelWithParameters("id4", FormationID, RuntimeContextID, RuntimeContextID, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeRuntimeContext, model.DeleteErrorAssignmentState)
	fa3 := fixFormationAssignmentModelWithParameters("id3", FormationID, RuntimeID, RuntimeContextID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeRuntimeContext, model.DeletingAssignmentState)
	formationAssignments := []*model.FormationAssignment{fa1, fa2, fa3, fa4}

	assignmentOperation1 := mock.MatchedBy(func(op *model.AssignmentOperationInput) bool {
		return op.Type == model.Assign && op.FormationAssignmentID == "id1" && op.FormationID == FormationID && op.TriggeredBy == model.ResetAssignment
	})
	assignmentOperation2 := mock.MatchedBy(func(op *model.AssignmentOperationInput) bool {
		return op.Type == model.Assign && op.FormationAssignmentID == "id2" && op.FormationID == FormationID && op.TriggeredBy == model.ResetAssignment
	})
	assignmentOperation3 := mock.MatchedBy(func(op *model.AssignmentOperationInput) bool {
		return op.Type == model.Assign && op.FormationAssignmentID == "id3" && op.FormationID == FormationID && op.TriggeredBy == model.ResetAssignment
	})
	assignmentOperation4 := mock.MatchedBy(func(op *model.AssignmentOperationInput) bool {
		return op.Type == model.Assign && op.FormationAssignmentID == "id4" && op.FormationID == FormationID && op.TriggeredBy == model.ResetAssignment
	})
	assignmentOperationWithAssignType := fixAssignmentOperationModelWithTypeAndTrigger(model.Assign, model.AssignObject)
	assignmentOperationWithUnassignType := fixAssignmentOperationModelWithTypeAndTrigger(model.Unassign, model.UnassignObject)

	formationAssignmentsInDeletingState := cloneFormationAssignments(formationAssignments)
	setAssignmentsToState(model.DeletingAssignmentState, formationAssignmentsInDeletingState...)

	formationAssignmentsInInitialState := cloneFormationAssignments(formationAssignments)
	setAssignmentsToState(model.InitialAssignmentState, formationAssignmentsInInitialState...)

	formationAssignmentsInReadyAndOneCreateErrorStates := cloneFormationAssignments(formationAssignments)
	setAssignmentsToState(model.ReadyAssignmentState, formationAssignmentsInReadyAndOneCreateErrorStates...)
	formationAssignmentsInReadyAndOneCreateErrorStates[3].State = string(model.CreateErrorAssignmentState)

	reverseAssignment := &model.FormationAssignment{
		ID:          "id1",
		FormationID: FormationID,
		Source:      ApplicationID,
		SourceType:  model.FormationAssignmentTypeApplication,
		Target:      RuntimeID,
		TargetType:  model.FormationAssignmentTypeRuntime,
		State:       string(model.ReadyAssignmentState),
	}

	webhookModeAsyncCallback := graphql.WebhookModeAsyncCallback
	notificationsForAssignments := []*webhookclient.FormationAssignmentNotificationRequest{
		{
			Webhook: &graphql.Webhook{
				ID: WebhookID,
			},
		},
		{
			Webhook: &graphql.Webhook{
				ID: Webhook2ID,
			},
		},
		{
			Webhook: &graphql.Webhook{
				ID:   Webhook3ID,
				Mode: &webhookModeAsyncCallback,
			},
		},
		{
			Webhook: &graphql.Webhook{
				ID: Webhook4ID,
			},
		},
	}

	var formationAssignmentPairs = make([]*formationassignment.AssignmentMappingPairWithOperation, 0, len(formationAssignments))
	for i := range formationAssignments {
		formationAssignmentPairs = append(formationAssignmentPairs, fixFormationAssignmentPairWithNoReverseAssignment(notificationsForAssignments[i], formationAssignments[i]))
	}

	var formationAssignmentInitialPairs = make([]*formationassignment.AssignmentMappingPairWithOperation, 0, len(formationAssignments))
	for i := range formationAssignmentsInInitialState {
		formationAssignmentInitialPairs = append(formationAssignmentInitialPairs, fixFormationAssignmentPairWithNoReverseAssignment(notificationsForAssignments[i], formationAssignmentsInInitialState[i]))
	}

	var formationAssignmentReadyAndCreateErrorPairs = make([]*formationassignment.AssignmentMappingPairWithOperation, 0, len(formationAssignmentsInReadyAndOneCreateErrorStates))
	for i := range formationAssignmentsInReadyAndOneCreateErrorStates {
		formationAssignmentReadyAndCreateErrorPairs = append(formationAssignmentReadyAndCreateErrorPairs, fixFormationAssignmentPairWithNoReverseAssignment(nil, formationAssignmentsInReadyAndOneCreateErrorStates[i]))
		formationAssignmentReadyAndCreateErrorPairs[i].Operation = model.AssignFormation
	}

	testSchema, err := labeldef.NewSchemaForFormations([]string{testScenario, testFormationName})
	assert.NoError(t, err)
	testSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, testSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{testScenario})
	assert.NoError(t, err)
	newSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, newSchema)

	nilSchemaLblDef := fixScenariosLabelDefinition(TntInternalID, testSchema)
	nilSchemaLblDef.Schema = nil

	testCases := []struct {
		Name                                     string
		FormationAssignments                     []*model.FormationAssignment
		ShouldReset                              bool
		TxFn                                     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		LabelServiceFn                           func() *automock.LabelService
		LabelRepoFn                              func() *automock.LabelRepository
		FormationRepositoryFn                    func() *automock.FormationRepository
		FormationTemplateRepositoryFn            func() *automock.FormationTemplateRepository
		FormationAssignmentNotificationServiceFN func() *automock.FormationAssignmentNotificationsService
		NotificationServiceFN                    func() *automock.NotificationsService
		FormationAssignmentServiceFn             func() *automock.FormationAssignmentService
		WebhookRepoFn                            func() *automock.WebhookRepository
		RuntimeContextRepoFn                     func() *automock.RuntimeContextRepository
		LabelDefRepositoryFn                     func() *automock.LabelDefRepository
		LabelDefServiceFn                        func() *automock.LabelDefService
		StatusServiceFn                          func() *automock.StatusService
		AssignmentOperationServiceFn             func() *automock.AssignmentOperationService
		ExpectedErrMessage                       string
	}{
		// Business logic tests for tenant mapping notifications only
		{
			Name:                 "success when resynchronization is successful and there are leftover formation assignments",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, nil).Once()

				svc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[3].Source).Return([]*model.FormationAssignment{{ID: "id6"}}, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
		},
		{
			Name:                 "success when resynchronization is successful and there is formation assignment in create error state and error but no webhook",
			FormationAssignments: formationAssignmentsInReadyAndOneCreateErrorStates,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignmentsInReadyAndOneCreateErrorStates, nil).Once()

				for _, fa := range formationAssignmentsInReadyAndOneCreateErrorStates {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentReadyAndCreateErrorPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentReadyAndCreateErrorPairs[1]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentReadyAndCreateErrorPairs[2]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentReadyAndCreateErrorPairs[3]).Return(false, nil).Once()

				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignmentsInReadyAndOneCreateErrorStates[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignmentsInReadyAndOneCreateErrorStates[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignmentsInReadyAndOneCreateErrorStates[2].ID, formationAssignmentsInInitialState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignmentsInReadyAndOneCreateErrorStates[3].ID, formationAssignmentsInInitialState[3]).Return(nil).Once()
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignmentsInReadyAndOneCreateErrorStates[0], model.AssignFormation).Return(nil, nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignmentsInReadyAndOneCreateErrorStates[1], model.AssignFormation).Return(nil, nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignmentsInReadyAndOneCreateErrorStates[2], model.AssignFormation).Return(nil, nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignmentsInReadyAndOneCreateErrorStates[3], model.AssignFormation).Return(nil, nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
		},
		{
			Name:                 "success when resynchronization is successful and there are NO left formation assignments should unassign",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, nil).Once()

				svc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[3].Source).Return(nil, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(runtimeLbl, nil).Once()
				svc.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeContextID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
		},
		{
			Name:                 "returns error when failing unassign formation",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, nil).Once()

				svc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[3].Source).Return(nil, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(nil, testErr).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
		},
		{
			Name:                 "returns error when failing to commit transaction after sending notifications",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenFailsOnCommit(1)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name:                 "returns error when failing to begin transaction after sending notifications",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenFailsOnBegin(2)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name:                 "returns error when failing to commit transaction after checking for unassign",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenFailsOnCommit(2)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, nil).Once()

				svc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[3].Source).Return([]*model.FormationAssignment{{ID: "id6"}}, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name:                 "returns error when failing to unassign from formation after resynchronizing",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, nil).Once()

				svc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[3].Target).Return([]*model.FormationAssignment{}, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(nil, testErr).Once()
				return svc
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:                 "returns error when failing to list formation assignments for participant",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, nil).Once()

				svc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), FormationID, mock.Anything).Return(nil, testErr).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when failing to get formation",
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when updating formation assignment to deleting state in case of async_callback webhook",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[0].Source, formationAssignments[0].Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()

				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(testErr).Once()
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:                 "Error when committing the first transaction",
			FormationAssignments: formationAssignments,
			TxFn:                 txGen.ThatFailsOnCommit,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name:                 "Error when the second transaction fail to begin",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenFailsOnBegin(1)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name:                 "returns error when failing processing formation assignments fails",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, testErr).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, testErr).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, testErr).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, testErr).Once()

				svc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[3].Source).Return([]*model.FormationAssignment{{ID: "id6"}}, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:                 "returns error when failing to get latest operation",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(nil, testErr).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:                 "returns error when failing to update operation",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[0].Source, formationAssignments[0].Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()

				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(testErr).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when getting reverse formation assignment",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return([]*model.FormationAssignment{formationAssignments[0]}, nil).Once()

				svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[0].Source, formationAssignments[0].Target).Return(nil, testErr).Once()

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when generating reverse formation assignment notification",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return([]*model.FormationAssignment{formationAssignments[0]}, nil)

				svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[0].Source, formationAssignments[0].Target).Return(reverseAssignment, nil)

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil)
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), reverseAssignment, model.AssignFormation).Return(nil, testErr)
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when generating formation assignment notification",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return([]*model.FormationAssignment{formationAssignments[0]}, nil).Once()

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(nil, testErr).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when listing formation assignments with states",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(nil, testErr).Once()

				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns error when failing to begin transaction",
			TxFn: txGen.ThatFailsOnBegin,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			ExpectedErrMessage: transactionError.Error(),
		},
		// Business logic tests for formation and tenant mapping notifications
		{
			Name:                 "success when both formation and formation assignment resynchronization are successful and there no left formation assignments should unassign",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(4)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, nil).Once()

				svc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[3].Source).Return([]*model.FormationAssignment{{ID: "id6"}}, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()
				return svc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", txtest.CtxWithDBMatcher(), formationLifecycleSyncWebhooks, TntInternalID, fixFormationModelWithState(model.InitialFormationState), testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", txtest.CtxWithDBMatcher(), formationNotificationSyncCreateRequest).Return(formationNotificationWebhookSuccessResponse, nil).Once()
				return notificationSvc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}

				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", txtest.CtxWithDBMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithState(model.InitialFormationState), nil).Once()
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), fixFormationModelWithState(model.ReadyFormationState), model.CreateFormation).Return(nil).Once()
				return svc
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithUnassignType, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResyncAssignment).Return(nil).Once()
				return svc
			},
		},
		// Business logic tests for formation notifications only
		{
			Name: "success when resynchronization is successful for formation notifications",
			TxFn: txGen.ThatSucceeds,
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", txtest.CtxWithDBMatcher(), formationLifecycleSyncWebhooks, TntInternalID, fixFormationModelWithState(model.InitialFormationState), testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", txtest.CtxWithDBMatcher(), formationNotificationSyncCreateRequest).Return(formationNotificationWebhookErrorResponse, nil).Once()
				return notificationSvc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", txtest.CtxWithDBMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithState(model.InitialFormationState), nil).Once()
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationID, TntInternalID).Return(formationInCreateErrorState, nil).Once()
				return repo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("SetFormationToErrorStateWithConstraints", txtest.CtxWithDBMatcher(), formationWithInitialState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorFormationState, model.CreateFormation).Return(nil).Once()
				return svc
			},
		},
		{
			Name: "error when resynchronization is successful for formation notifications but fails while committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", txtest.CtxWithDBMatcher(), formationLifecycleSyncWebhooks, TntInternalID, fixFormationModelWithState(model.InitialFormationState), testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", txtest.CtxWithDBMatcher(), formationNotificationSyncCreateRequest).Return(formationNotificationWebhookErrorResponse, nil).Once()
				return notificationSvc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", txtest.CtxWithDBMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithState(model.InitialFormationState), nil).Once()
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationID, TntInternalID).Return(formationInCreateErrorState, nil).Once()
				return repo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("SetFormationToErrorStateWithConstraints", txtest.CtxWithDBMatcher(), formationWithInitialState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorFormationState, model.CreateFormation).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name: "error when resynchronization is successful for formation notifications but fails while getting formation",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", txtest.CtxWithDBMatcher(), formationLifecycleSyncWebhooks, TntInternalID, fixFormationModelWithState(model.InitialFormationState), testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", txtest.CtxWithDBMatcher(), formationNotificationSyncCreateRequest).Return(formationNotificationWebhookErrorResponse, nil).Once()
				return notificationSvc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", txtest.CtxWithDBMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithState(model.InitialFormationState), nil).Once()
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationID, TntInternalID).Return(nil, testErr).Once()
				return repo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("SetFormationToErrorStateWithConstraints", txtest.CtxWithDBMatcher(), formationWithInitialState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorFormationState, model.CreateFormation).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when resynchronization is unsuccessful for formation notifications due to technical error",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", txtest.CtxWithDBMatcher(), formationLifecycleSyncWebhooks, TntInternalID, fixFormationModelWithState(model.InitialFormationState), testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", txtest.CtxWithDBMatcher(), formationNotificationSyncCreateRequest).Return(nil, testErr).Once()
				return notificationSvc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", txtest.CtxWithDBMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithState(model.InitialFormationState), nil).Once()
				repo.On("Update", txtest.CtxWithDBMatcher(), formationInCreateErrorStateTechnicalError).Return(nil).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when resynchronization is unsuccessful for formation notifications during generating notifications",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", txtest.CtxWithDBMatcher(), formationLifecycleSyncWebhooks, TntInternalID, fixFormationModelWithState(model.InitialFormationState), testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(nil, testErr).Once()
				return notificationSvc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", txtest.CtxWithDBMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithState(model.InitialFormationState), nil).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when resynchronization is unsuccessful for formation notifications during getting webhooks",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", txtest.CtxWithDBMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(nil, testErr).Once()
				return webhookRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithState(model.InitialFormationState), nil).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when resynchronization is unsuccessful for formation notifications during getting formation template",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(nil, testErr)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithState(model.InitialFormationState), nil).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when resynchronization is unsuccessful for formation notifications while beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithState(model.InitialFormationState), nil).Once()
				return repo
			},
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name: "success when resynchronization is successful for formation notifications with DELETE_ERROR state",
			TxFn: txGen.ThatSucceeds,
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", txtest.CtxWithDBMatcher(), TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", txtest.CtxWithDBMatcher(), newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", txtest.CtxWithDBMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", txtest.CtxWithDBMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", txtest.CtxWithDBMatcher(), formationLifecycleSyncWebhooks, TntInternalID, fixFormationModelWithStateAndAssignmentError(t, model.DeleteErrorFormationState, testErr.Error(), formationassignment.ClientError), testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(formationNotificationSyncDeleteRequests, nil).Once()
				notificationSvc.On("SendNotification", txtest.CtxWithDBMatcher(), formationNotificationSyncDeleteRequest).Return(formationNotificationWebhookSuccessResponse, nil).Once()
				return notificationSvc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", txtest.CtxWithDBMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithStateAndAssignmentError(t, model.DeleteErrorFormationState, testErr.Error(), formationassignment.ClientError), nil).Once()
				repo.On("DeleteByName", txtest.CtxWithDBMatcher(), TntInternalID, testFormationName).Return(nil).Once()
				return repo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), fixFormationModelWithState(model.ReadyFormationState), model.DeleteFormation).Return(nil).Once()
				return svc
			},
		},
		{
			Name: "error when resynchronization is successful for formation notifications with DELETE_ERROR during committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", txtest.CtxWithDBMatcher(), TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", txtest.CtxWithDBMatcher(), newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", txtest.CtxWithDBMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", txtest.CtxWithDBMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", txtest.CtxWithDBMatcher(), formationLifecycleSyncWebhooks, TntInternalID, fixFormationModelWithStateAndAssignmentError(t, model.DeleteErrorFormationState, testErr.Error(), formationassignment.ClientError), testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(formationNotificationSyncDeleteRequests, nil).Once()
				notificationSvc.On("SendNotification", txtest.CtxWithDBMatcher(), formationNotificationSyncDeleteRequest).Return(formationNotificationWebhookSuccessResponse, nil).Once()
				return notificationSvc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", txtest.CtxWithDBMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithStateAndAssignmentError(t, model.DeleteErrorFormationState, testErr.Error(), formationassignment.ClientError), nil).Once()
				repo.On("DeleteByName", txtest.CtxWithDBMatcher(), TntInternalID, testFormationName).Return(nil).Once()
				return repo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), fixFormationModelWithState(model.ReadyFormationState), model.DeleteFormation).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name: "error when resynchronization is successful for formation notifications with DELETE_ERROR during deletion of formation",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", txtest.CtxWithDBMatcher(), TntInternalID, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", txtest.CtxWithDBMatcher(), newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", txtest.CtxWithDBMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", txtest.CtxWithDBMatcher(), newSchema, TntInternalID, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", txtest.CtxWithDBMatcher(), formationLifecycleSyncWebhooks, TntInternalID, fixFormationModelWithStateAndAssignmentError(t, model.DeleteErrorFormationState, testErr.Error(), formationassignment.ClientError), testFormationTemplateName, FormationTemplateID, model.DeleteFormation).Return(formationNotificationSyncDeleteRequests, nil).Once()
				notificationSvc.On("SendNotification", txtest.CtxWithDBMatcher(), formationNotificationSyncDeleteRequest).Return(formationNotificationWebhookSuccessResponse, nil).Once()
				return notificationSvc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", txtest.CtxWithDBMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithStateAndAssignmentError(t, model.DeleteErrorFormationState, testErr.Error(), formationassignment.ClientError), nil).Once()
				repo.On("DeleteByName", txtest.CtxWithDBMatcher(), TntInternalID, testFormationName).Return(testErr).Once()
				return repo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), fixFormationModelWithState(model.ReadyFormationState), model.DeleteFormation).Return(nil).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		// Business logic tests for tenant mapping notifications with reset
		{
			Name:        "success when reset and resynchronize is successful and there are leftover formation assignments",
			ShouldReset: true,
			TxFn:        txGen.ThatSucceedsTwice,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignmentsInInitialState, nil).Once()
				svc.On("GetAssignmentsForFormation", txtest.CtxWithDBMatcher(), TntInternalID, FormationID).Return(cloneFormationAssignments(formationAssignments), nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Twice()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Twice()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInInitialState[2]).Return(nil).Twice()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInInitialState[3]).Return(nil).Twice()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentInitialPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentInitialPairs[1]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentInitialPairs[2]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentInitialPairs[3]).Return(false, nil).Once()

				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignmentsInInitialState[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignmentsInInitialState[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignmentsInInitialState[2], model.AssignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignmentsInInitialState[3], model.AssignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(fixFormationTemplateModelThatSupportsReset(), nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation1).Return("", nil).Once()
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation2).Return("", nil).Once()
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation3).Return("", nil).Once()
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation4).Return("", nil).Once()

				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()
				svc.On("GetLatestOperation", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID).Return(assignmentOperationWithAssignType, nil).Once()

				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, FormationID, model.ResetAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, FormationID, model.ResetAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, FormationID, model.ResetAssignment).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, FormationID, model.ResetAssignment).Return(nil).Once()
				return svc
			},
		},
		{
			Name:        "error when creating an operation",
			ShouldReset: true,
			TxFn:        txGen.ThatDoesntExpectCommit,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormation", txtest.CtxWithDBMatcher(), TntInternalID, FormationID).Return(cloneFormationAssignments(formationAssignments), nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(fixFormationTemplateModelThatSupportsReset(), nil).Once()
				return repo
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation1).Return("", testErr).Once()
				return svc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:        "error when reset and resynchronize when updating formation assignment to initial state fails",
			ShouldReset: true,
			TxFn:        txGen.ThatDoesntExpectCommit,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormation", txtest.CtxWithDBMatcher(), TntInternalID, FormationID).Return(cloneFormationAssignments([]*model.FormationAssignment{formationAssignments[0]}), nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(testErr).Once()

				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(fixFormationTemplateModelThatSupportsReset(), nil).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:        "error when reset and resynchronize when getting formation assignment for formation fails",
			ShouldReset: true,
			TxFn:        txGen.ThatDoesntExpectCommit,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormation", txtest.CtxWithDBMatcher(), TntInternalID, FormationID).Return(nil, testErr).Once()

				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(fixFormationTemplateModelThatSupportsReset(), nil).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:        "error when reset and resynchronize formation template does not support resetting",
			ShouldReset: true,
			TxFn:        txGen.ThatDoesntExpectCommit,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(fixFormationTemplateModel(), nil).Once()
				return repo
			},
			ExpectedErrMessage: fmt.Sprintf("formation template %q does not support resetting", testFormationTemplateName),
		},
		{
			Name:        "error when reset and resynchronize fails getting formation template",
			ShouldReset: true,
			TxFn:        txGen.ThatDoesntExpectCommit,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationID, TntInternalID).Return(testFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := txGen.ThatDoesntStartTransaction()
			if testCase.TxFn != nil {
				persist, transact = testCase.TxFn()
			}

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

			formationTemplateRepo := unusedFormationTemplateRepo()
			if testCase.FormationTemplateRepositoryFn != nil {
				formationTemplateRepo = testCase.FormationTemplateRepositoryFn()
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

			formationAssignmentNotificationService := unusedFormationAssignmentNotificationService()
			if testCase.FormationAssignmentNotificationServiceFN != nil {
				formationAssignmentNotificationService = testCase.FormationAssignmentNotificationServiceFN()
			}

			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFn != nil {
				webhookRepo = testCase.WebhookRepoFn()
			}

			labelDefRepo := unusedLabelDefRepository()
			if testCase.LabelDefRepositoryFn != nil {
				labelDefRepo = testCase.LabelDefRepositoryFn()
			}

			labelDefSvc := unusedLabelDefService()
			if testCase.LabelDefServiceFn != nil {
				labelDefSvc = testCase.LabelDefServiceFn()
			}

			statusService := &automock.StatusService{}
			if testCase.StatusServiceFn != nil {
				statusService = testCase.StatusServiceFn()
			}

			assignmentOperationService := &automock.AssignmentOperationService{}
			if testCase.AssignmentOperationServiceFn != nil {
				assignmentOperationService = testCase.AssignmentOperationServiceFn()
			}

			assignmentsBeforeModifications := make(map[string]*model.FormationAssignment)
			for _, a := range testCase.FormationAssignments {
				assignmentsBeforeModifications[a.ID] = a.Clone()
			}
			defer func() {
				for i, a := range testCase.FormationAssignments {
					testCase.FormationAssignments[i] = assignmentsBeforeModifications[a.ID]
				}
			}()

			svc := formation.NewServiceWithAsaEngine(transact, nil, labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelService, nil, labelDefSvc, nil, nil, nil, nil, runtimeContextRepo, formationAssignmentSvc, webhookRepo, formationAssignmentNotificationService, notificationsSvc, nil, runtimeType, applicationType, nil, statusService, assignmentOperationService)

			// WHEN
			_, err := svc.ResynchronizeFormationNotifications(ctx, FormationID, testCase.ShouldReset)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
			mock.AssertExpectationsForObjects(t, persist, transact, labelService, runtimeContextRepo, formationRepo, labelRepo, notificationsSvc, formationAssignmentSvc, formationAssignmentNotificationService, formationTemplateRepo, webhookRepo, statusService, assignmentOperationService)
		})
	}

	t.Run("returns error when empty tenant", func(t *testing.T) {
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
		_, err := svc.ResynchronizeFormationNotifications(context.TODO(), FormationID, false)
		require.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func cloneFormationAssignments(assignments []*model.FormationAssignment) []*model.FormationAssignment {
	clonedAssignments := make([]*model.FormationAssignment, 0, len(assignments))
	for _, a := range assignments {
		clonedAssignments = append(clonedAssignments, a.Clone())
	}
	return clonedAssignments
}

func setAssignmentsToState(state model.FormationAssignmentState, assignments ...*model.FormationAssignment) {
	for _, a := range assignments {
		a.State = string(state)
	}
}
