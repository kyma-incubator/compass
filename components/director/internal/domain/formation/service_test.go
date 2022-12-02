package formation_test

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/pkg/errors"

	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceList(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("Test error")

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
				repo.On("List", ctx, Tnt, pageSize, cursor).Return(expectedFormationPage, nil).Once()
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
				repo.On("List", ctx, Tnt, pageSize, cursor).Return(nil, testErr).Once()
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

			svc := formation.NewService(nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("Test error")

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
				repo.On("Get", ctx, FormationID, Tnt).Return(&modelFormation, nil).Once()
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
				repo.On("Get", ctx, FormationID, Tnt).Return(nil, testErr).Once()
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

			svc := formation.NewService(nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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

func TestServiceCreateFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("Test error")

	in := model.Formation{
		Name: testFormationName,
	}
	expected := &model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            Tnt,
	}

	testSchema, err := labeldef.NewSchemaForFormations([]string{testScenario})
	assert.NoError(t, err)
	testSchemaLblDef := fixScenariosLabelDefinition(Tnt, testSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{testScenario, testFormationName})
	assert.NoError(t, err)
	newSchemaLblDef := fixScenariosLabelDefinition(Tnt, newSchema)

	emptySchemaLblDef := fixScenariosLabelDefinition(Tnt, testSchemaLblDef)
	emptySchemaLblDef.Schema = nil

	templateName := "Side-by-side extensibility with Kyma"

	testCases := []struct {
		Name                    string
		UUIDServiceFn           func() *automock.UuidService
		LabelDefRepositoryFn    func() *automock.LabelDefRepository
		LabelDefServiceFn       func() *automock.LabelDefService
		FormationTemplateRepoFn func() *automock.FormationTemplateRepository
		FormationRepoFn         func() *automock.FormationRepository
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, Tnt, []string{testFormationName}).Return(nil)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByName", ctx, templateName).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, fixFormationModel()).Return(nil).Once()
				return formationRepoMock
			},
			TemplateName:       templateName,
			ExpectedFormation:  expected,
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByName", ctx, templateName).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, fixFormationModel()).Return(nil).Once()
				return formationRepoMock
			},
			TemplateName:       templateName,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error when labeldef is missing and can not create it",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, Tnt, []string{testFormationName}).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when labeldef is missing and create formation fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, Tnt, []string{testFormationName}).Return(nil)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByName", ctx, templateName).Return(nil, testErr).Once()
				return formationTemplateRepoMock
			},
			TemplateName:       templateName,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when can not get labeldef",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(nil, testErr)
				return labelDefRepo
			},
			LabelDefServiceFn:  unusedLabelDefService,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when labeldef's schema is missing",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&emptySchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn:  unusedLabelDefService,
			ExpectedErrMessage: "missing schema",
		},
		{
			Name: "error when validating existing labels against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, testSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when validating automatic scenario assignment against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, testSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, testSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when update with version fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(testErr)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, testSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, testSchemaLblDef.Key).Return(nil)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when getting formation template by name fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByName", ctx, templateName).Return(nil, testErr).Once()
				return formationTemplateRepoMock
			},
			TemplateName:       templateName,
			ExpectedErrMessage: testErr.Error(),
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByName", ctx, templateName).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, fixFormationModel()).Return(testErr).Once()
				return formationRepoMock
			},
			TemplateName:       templateName,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			uuidSvcMock := &automock.UuidService{}
			if testCase.UUIDServiceFn != nil {
				uuidSvcMock = testCase.UUIDServiceFn()
			}
			lblDefRepo := testCase.LabelDefRepositoryFn()
			lblDefService := testCase.LabelDefServiceFn()
			formationRepoMock := &automock.FormationRepository{}
			if testCase.FormationRepoFn != nil {
				formationRepoMock = testCase.FormationRepoFn()
			}
			formationTemplateRepoMock := &automock.FormationTemplateRepository{}
			if testCase.FormationTemplateRepoFn != nil {
				formationTemplateRepoMock = testCase.FormationTemplateRepoFn()
			}

			svc := formation.NewService(nil, lblDefRepo, nil, formationRepoMock, formationTemplateRepoMock, nil, uuidSvcMock, lblDefService, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			actual, err := svc.CreateFormation(ctx, Tnt, in, testCase.TemplateName)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, uuidSvcMock, lblDefRepo, lblDefService, formationRepoMock, formationTemplateRepoMock)
		})
	}
}

func TestServiceDeleteFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("Test error")

	in := model.Formation{
		Name: testFormationName,
	}

	expected := &model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            Tnt,
	}

	testSchema, err := labeldef.NewSchemaForFormations([]string{testScenario, testFormationName})
	assert.NoError(t, err)
	testSchemaLblDef := fixScenariosLabelDefinition(Tnt, testSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{testScenario})
	assert.NoError(t, err)
	newSchemaLblDef := fixScenariosLabelDefinition(Tnt, newSchema)

	nilSchemaLblDef := fixScenariosLabelDefinition(Tnt, testSchema)
	nilSchemaLblDef.Schema = nil

	testCases := []struct {
		Name                 string
		LabelDefRepositoryFn func() *automock.LabelDefRepository
		LabelDefServiceFn    func() *automock.LabelDefService
		FormationRepoFn      func() *automock.FormationRepository
		InputFormation       model.Formation
		ExpectedFormation    *model.Formation
		ExpectedErrMessage   string
	}{
		{
			Name: "success",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("DeleteByName", ctx, Tnt, testFormationName).Return(nil).Once()
				formationRepoMock.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepoMock
			},
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error when can not get labeldef",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(nil, testErr)
				return labelDefRepo
			},
			LabelDefServiceFn:  unusedLabelDefService,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when labeldef's schema is missing",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&nilSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn:  unusedLabelDefService,
			InputFormation:     in,
			ExpectedErrMessage: "missing schema",
		},
		{
			Name: "error when validating existing labels against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(testErr)
				return labelDefService
			},
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when validating automatic scenario assignment against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(testErr)
				return labelDefService
			},
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when update with version fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&newSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(testErr)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, newSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, newSchemaLblDef.Key).Return(nil)
				return labelDefService
			},
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when can't get formation by name",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("DeleteByName", ctx, Tnt, testFormationName).Return(nil).Once()
				formationRepoMock.On("GetByName", ctx, testFormationName, Tnt).Return(nil, testErr).Once()
				return formationRepoMock
			},
			InputFormation:     in,
			ExpectedFormation:  nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when deleting formation template by name fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				formationRepoMock.On("DeleteByName", ctx, Tnt, testFormationName).Return(testErr).Once()
				return formationRepoMock
			},
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			lblDefRepo := testCase.LabelDefRepositoryFn()
			lblDefService := testCase.LabelDefServiceFn()
			formationRepoMock := &automock.FormationRepository{}
			if testCase.FormationRepoFn != nil {
				formationRepoMock = testCase.FormationRepoFn()
			}
			svc := formation.NewService(nil, lblDefRepo, nil, formationRepoMock, nil, nil, nil, lblDefService, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			actual, err := svc.DeleteFormation(ctx, Tnt, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, lblDefRepo, lblDefService)
		})
	}
}

func TestService_CreateAutomaticScenarioAssignment(t *testing.T) {
	ctx := fixCtxWithTenant()

	testErr := errors.New("test err")

	tnt := tenantID.String()

	rtmIDs := []string{"123", "456", "789"}
	rtmNames := []string{"first", "second", "third"}

	runtimes := []*model.Runtime{
		{
			ID:   rtmIDs[0],
			Name: rtmNames[0],
		},
		{
			ID:   rtmIDs[1],
			Name: rtmNames[1],
		},
		{
			ID:   rtmIDs[2],
			Name: rtmNames[2],
		},
	}
	formationAssignments := []*model.FormationAssignment{
		{
			ID:          "fa1",
			FormationID: FormationID,
			Source:      "1",
		},
		{
			ID:          "fa2",
			FormationID: FormationID,
			Source:      "123",
		},
		{
			ID:          "fa3",
			FormationID: FormationID,
			Source:      "2",
		},
	}
	notifications := []*webhookclient.NotificationRequest{
		{
			Webhook: graphql.Webhook{
				ID: "wid1",
			},
		},
		{
			Webhook: graphql.Webhook{
				ID: "wid2",
			},
		},
		{
			Webhook: graphql.Webhook{
				ID: "wid3",
			},
		},
	}
	ownedRuntimes := []*model.Runtime{runtimes[0], runtimes[1]}

	runtimeLblInputs := []*model.LabelInput{
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmIDs[0],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmIDs[1],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
	}

	rtmContexts := []*model.RuntimeContext{
		{
			ID:        "1",
			RuntimeID: rtmIDs[0],
			Key:       "test",
			Value:     "test",
		},
		{
			ID:        "2",
			RuntimeID: rtmIDs[2],
			Key:       "test",
			Value:     "test",
		},
	}

	runtimeCtxLblInputs := []*model.LabelInput{
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmContexts[0].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmContexts[1].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
	}

	runtimeTypeLblInput := []*model.LabelInput{
		{
			Key:        runtimeType,
			ObjectID:   rtmIDs[0],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
		{
			Key:        runtimeType,
			ObjectID:   rtmIDs[1],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
		{
			Key:        runtimeType,
			ObjectID:   rtmIDs[2],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
	}

	expectedRtmCtxLabels := []*model.Label{
		{
			ID:         "1",
			Tenant:     &tnt,
			Key:        "scenarios",
			Value:      []interface{}{},
			ObjectID:   rtmContexts[0].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
		{
			ID:         "2",
			Tenant:     &tnt,
			Key:        "scenarios",
			Value:      []interface{}{},
			ObjectID:   rtmContexts[1].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
	}

	expectedRuntimeLabel := &model.Label{
		ID:         "1",
		Tenant:     &tnt,
		Key:        "scenarios",
		Value:      []interface{}{},
		ObjectID:   rtmIDs[0],
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeTypeLbl := &model.Label{
		ID:         "123",
		Key:        runtimeType,
		Value:      runtimeType,
		Tenant:     str.Ptr(Tnt),
		ObjectID:   rtmIDs[0],
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}

	testCases := []struct {
		Name                          string
		LabelDefServiceFn             func() *automock.LabelDefService
		AsaRepoFn                     func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFN                 func() *automock.RuntimeRepository
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		LabelRepoFN                   func() *automock.LabelRepository
		LabelServiceFn                func() *automock.LabelService
		NotificationServiceFN         func() *automock.NotificationsService
		FormationAssignmentServiceFn  func() *automock.FormationAssignmentService
		InputASA                      model.AutomaticScenarioAssignment
		ExpectedASA                   model.AutomaticScenarioAssignment
		ExpectedErrMessage            string
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
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(nil, apperrors.NewNotFoundError(resource.RuntimeContext, rtmContexts[0].ID)).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[2]).Return(rtmContexts[1], nil).Once()
				runtimeContextRepo.On("GetByID", ctx, tenantID.String(), rtmContexts[0].ID).Return(&model.RuntimeContext{
					RuntimeID: runtimes[0].ID,
				}, nil)
				runtimeContextRepo.On("GetByID", ctx, tenantID.String(), rtmContexts[1].ID).Return(&model.RuntimeContext{
					RuntimeID: runtimes[1].ID,
				}, nil)

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Times(3)

				return runtimeContextRepo
			},
			LabelRepoFN: unusedLabelRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[0].ID, &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[0]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[1].ID, &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[2]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmContexts[0].ID, graphql.FormationObjectTypeRuntimeContext, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[0]}, nil).Once()
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmIDs[0], graphql.FormationObjectTypeRuntime, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[1]}, nil).Once()
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmContexts[1].ID, graphql.FormationObjectTypeRuntimeContext, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[2]}, nil).Once()

				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[0]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[0]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[2]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[2]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(4)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Times(4)
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[0]).Return(runtimeTypeLbl, nil)
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[1]).Return(runtimeTypeLbl, nil)

				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[0]).Return(expectedRtmCtxLabels[0], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[0].ID, runtimeCtxLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[1]).Return(expectedRtmCtxLabels[1], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[1].ID, runtimeCtxLblInputs[1]).Return(nil)
				return svc
			},
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        fixModel(testFormationName),
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when assigning runtime context to formation fails",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("GetByID", ctx, tenantID.String(), rtmContexts[0].ID).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByID", ctx, tenantID.String(), rtmContexts[1].ID).Return(rtmContexts[1], nil).Once()

				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(nil, apperrors.NewNotFoundError(resource.RuntimeContext, rtmContexts[0].ID)).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[2]).Return(rtmContexts[1], nil).Once()

				runtimeContextRepo.On("GetByID", ctx, tenantID.String(), rtmContexts[0].ID).Return(&model.RuntimeContext{
					RuntimeID: runtimes[0].ID,
				}, nil)

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Twice()

				return runtimeContextRepo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[0].ID, &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[0]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmContexts[0].ID, graphql.FormationObjectTypeRuntimeContext, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[0]}, nil).Once()
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmIDs[0], graphql.FormationObjectTypeRuntime, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[1]}, nil).Once()

				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[0]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[0]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(4)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Times(4)
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[0]).Return(runtimeTypeLbl, nil)
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[2]).Return(runtimeTypeLbl, nil)
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[0]).Return(expectedRtmCtxLabels[0], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[0].ID, runtimeCtxLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[1]).Return(expectedRtmCtxLabels[1], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[1].ID, runtimeCtxLblInputs[1]).Return(testErr)
				return svc
			},
			LabelRepoFN:        unusedLabelRepo,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing runtime contexts for runtime fails",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(nil, testErr).Once()
				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(2)
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmIDs[0], graphql.FormationObjectTypeRuntime, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[1]}, nil).Once()

				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Times(2)
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[0]).Return(runtimeTypeLbl, nil)
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)
				return svc
			},
			LabelRepoFN:        unusedLabelRepo,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing all runtimes fails",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()
				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(2)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Times(2)
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmIDs[0], graphql.FormationObjectTypeRuntime, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[1]}, nil).Once()

				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[0]).Return(runtimeTypeLbl, nil)
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)
				return svc
			},
			LabelRepoFN:        unusedLabelRepo,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when assigning runtime to formation fails",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(2)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Times(2)
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[0]).Return(runtimeTypeLbl, nil)
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(nil, testErr)
				return svc
			},
			LabelRepoFN:        unusedLabelRepo,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when checking if runtime exists by id fails",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, testErr).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(1)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelRepoFN:        unusedLabelRepo,
			LabelServiceFn:     unusedLabelService,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing owned runtimes fails",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(1)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelRepoFN:        unusedLabelRepo,
			LabelServiceFn:     unusedLabelService,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error getting formation template by ID fails",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(1)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			LabelRepoFN:        unusedLabelRepo,
			LabelServiceFn:     unusedLabelService,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error getting formation by name fails",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{testFormationName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel(testFormationName)).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(nil, testErr).Times(1)
				return repo
			},
			LabelRepoFN:                   unusedLabelRepo,
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			LabelServiceFn:                unusedLabelService,
			InputASA:                      fixModel(testFormationName),
			ExpectedASA:                   model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:            testErr.Error(),
		},
		{
			Name: "returns error when scenario already has an assignment",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{ScenarioName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", mock.Anything, fixModel(ScenarioName)).Return(apperrors.NewNotUniqueError("")).Once()
				return mockRepo
			},
			RuntimeRepoFN:                 unusedRuntimeRepo,
			RuntimeContextRepoFn:          unusedRuntimeContextRepo,
			LabelServiceFn:                unusedLabelService,
			FormationRepositoryFn:         unusedFormationRepo,
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			LabelRepoFN:                   unusedLabelRepo,
			InputASA:                      fixModel(ScenarioName),
			ExpectedASA:                   model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:            "a given scenario already has an assignment",
		},
		{
			Name: "returns error when given scenario does not exist",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{"completely-different-scenario"})
			},
			AsaRepoFn:                     unusedASARepo,
			RuntimeRepoFN:                 unusedRuntimeRepo,
			RuntimeContextRepoFn:          unusedRuntimeContextRepo,
			LabelServiceFn:                unusedLabelService,
			FormationRepositoryFn:         unusedFormationRepo,
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			LabelRepoFN:                   unusedLabelRepo,
			InputASA:                      fixModel(ScenarioName),
			ExpectedASA:                   model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:            apperrors.NewNotFoundError(resource.AutomaticScenarioAssigment, fixModel(ScenarioName).ScenarioName).Error(),
		},
		{
			Name: "returns error on persisting in DB",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{ScenarioName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", mock.Anything, fixModel(ScenarioName)).Return(fixError()).Once()
				return mockRepo
			},
			RuntimeRepoFN:                 unusedRuntimeRepo,
			RuntimeContextRepoFn:          unusedRuntimeContextRepo,
			LabelServiceFn:                unusedLabelService,
			FormationRepositoryFn:         unusedFormationRepo,
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			LabelRepoFN:                   unusedLabelRepo,
			InputASA:                      fixModel(ScenarioName),
			ExpectedASA:                   model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:            "while persisting Assignment: some error",
		},
		{
			Name: "returns error on getting available scenarios from label definition",
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}
				labelDefSvc.On("GetAvailableScenarios", mock.Anything, tenantID.String()).Return(nil, fixError()).Once()
				return labelDefSvc
			},
			AsaRepoFn:                     unusedASARepo,
			RuntimeRepoFN:                 unusedRuntimeRepo,
			RuntimeContextRepoFn:          unusedRuntimeContextRepo,
			LabelServiceFn:                unusedLabelService,
			FormationRepositoryFn:         unusedFormationRepo,
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			LabelRepoFN:                   unusedLabelRepo,
			InputASA:                      fixModel(ScenarioName),
			ExpectedASA:                   model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:            "while getting available scenarios: some error",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			asaRepo := testCase.AsaRepoFn()
			tenantSvc := &automock.TenantService{}
			labelDefService := testCase.LabelDefServiceFn()
			runtimeRepo := testCase.RuntimeRepoFN()
			runtimeContextRepo := testCase.RuntimeContextRepoFn()
			formationRepo := testCase.FormationRepositoryFn()
			formationTemplateRepo := testCase.FormationTemplateRepositoryFn()
			lblService := testCase.LabelServiceFn()
			labelRepo := testCase.LabelRepoFN()

			notificationSvc := unusedNotificationsService()
			if testCase.NotificationServiceFN != nil {
				notificationSvc = testCase.NotificationServiceFN()
			}

			formationAssignmentSvc := unusedFormationAssignmentService()
			if testCase.FormationAssignmentServiceFn != nil {
				formationAssignmentSvc = testCase.FormationAssignmentServiceFn()
			}
			svc := formation.NewService(nil, nil, labelRepo, formationRepo, formationTemplateRepo, lblService, nil, labelDefService, asaRepo, nil, tenantSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, notificationSvc, runtimeType, applicationType)

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

			mock.AssertExpectationsForObjects(t, tenantSvc, asaRepo, labelDefService, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, lblService, labelRepo, notificationSvc, formationAssignmentSvc)
		})
	}

	t.Run("returns error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(context.TODO(), fixModel(ScenarioName))

		// THEN
		assert.EqualError(t, err, "cannot read tenant from context")
	})
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

		svc := formation.NewService(nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, runtimeType, applicationType)

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

		svc := formation.NewService(nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("return error when input slice is empty", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
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

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignments: %s", ErrMsg))
	})

	t.Run("returns error when empty tenant", func(t *testing.T) {
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
		err := svc.DeleteManyASAForSameTargetTenant(context.TODO(), models)
		require.EqualError(t, err, "cannot read tenant from context")
	})
}

func TestService_DeleteAutomaticScenarioAssignment(t *testing.T) {
	ctx := fixCtxWithTenant()

	testErr := errors.New("test err")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	rtmIDs := []string{"123", "456", "789"}
	rtmNames := []string{"first", "second", "third"}

	runtimes := []*model.Runtime{
		{
			ID:   rtmIDs[0],
			Name: rtmNames[0],
		},
		{
			ID:   rtmIDs[1],
			Name: rtmNames[1],
		},
		{
			ID:   rtmIDs[2],
			Name: rtmNames[2],
		},
	}
	formationAssignments := []*model.FormationAssignment{
		{
			ID:          "fa1",
			FormationID: FormationID,
			Source:      "1",
		},
		{
			ID:          "fa2",
			FormationID: FormationID,
			Source:      "123",
		},
		{
			ID:          "fa3",
			FormationID: FormationID,
			Source:      "2",
		},
	}
	notifications := []*webhookclient.NotificationRequest{
		{
			Webhook: graphql.Webhook{
				ID: "wid1",
			},
		},
		{
			Webhook: graphql.Webhook{
				ID: "wid2",
			},
		},
		{
			Webhook: graphql.Webhook{
				ID: "wid3",
			},
		},
	}
	ownedRuntimes := []*model.Runtime{runtimes[0], runtimes[1]}

	runtimeLblInputs := []*model.LabelInput{
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmIDs[0],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmIDs[1],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
	}

	rtmContexts := []*model.RuntimeContext{
		{
			ID:        "1",
			RuntimeID: rtmIDs[0],
			Key:       "test",
			Value:     "test",
		},
		{
			ID:        "2",
			RuntimeID: rtmIDs[2],
			Key:       "test",
			Value:     "test",
		},
	}

	runtimeCtxLblInputs := []*model.LabelInput{
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmContexts[0].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmContexts[1].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
	}
	tnt := tenantID.String()

	expectedRtmCtxLabels := []*model.Label{
		{
			ID:         "1",
			Tenant:     &tnt,
			Key:        "scenarios",
			Value:      []interface{}{testFormationName},
			ObjectID:   rtmContexts[0].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
		{
			ID:         "2",
			Tenant:     &tnt,
			Key:        "scenarios",
			Value:      []interface{}{testFormationName},
			ObjectID:   rtmContexts[1].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
	}

	expectedRuntimeLabel := &model.Label{
		ID:         "1",
		Tenant:     &tnt,
		Key:        "scenarios",
		Value:      []interface{}{testFormationName},
		ObjectID:   rtmIDs[0],
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}

	testCases := []struct {
		Name                          string
		LabelDefServiceFn             func() *automock.LabelDefService
		AsaRepoFn                     func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFN                 func() *automock.RuntimeRepository
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		LabelServiceFn                func() *automock.LabelService
		LabelRepositoryFn             func() *automock.LabelRepository
		NotificationSvcFn             func() *automock.NotificationsService
		FormationAssignmentSvcFn      func() *automock.FormationAssignmentService
		TransactionerFn               func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		InputASA                      model.AutomaticScenarioAssignment
		ExpectedASA                   model.AutomaticScenarioAssignment
		ExpectedErrMessage            string
	}{
		{
			Name:              "Success",
			LabelDefServiceFn: unusedLabelDefService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(nil).Once()
				mockRepo.On("ListAll", ctx, tenantID.String()).Return(make([]*model.AutomaticScenarioAssignment, 0), nil).Times(3)
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()

				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(nil, apperrors.NewNotFoundError(resource.RuntimeContext, rtmContexts[0].ID)).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[2]).Return(rtmContexts[1], nil).Once()

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Times(3)

				return runtimeContextRepo
			},
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(4)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil).Once()

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[0]).Return(expectedRtmCtxLabels[0], nil).Once()

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[1]).Return(expectedRtmCtxLabels[1], nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimes[0].ID, model.ScenariosKey).Return(nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeContextLabelableObject, rtmContexts[0].ID, model.ScenariosKey).Return(nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeContextLabelableObject, rtmContexts[1].ID, model.ScenariosKey).Return(nil).Once()

				return repo
			},
			NotificationSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[0].ID, &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[0]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[1].ID, &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[2]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmContexts[0].ID).Return([]*model.FormationAssignment{formationAssignments[0]}, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmIDs[0]).Return([]*model.FormationAssignment{formationAssignments[1]}, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmContexts[1].ID).Return([]*model.FormationAssignment{formationAssignments[2]}, nil)

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmContexts[0].ID).Return(nil, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmIDs[0]).Return(nil, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmContexts[1].ID).Return(nil, nil)

				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[0]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[0]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[2]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[2]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        fixModel(testFormationName),
			ExpectedErrMessage: "",
		},
		{
			Name:              "Returns error when unassigning runtime context fails",
			LabelDefServiceFn: unusedLabelDefService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(nil).Once()
				mockRepo.On("ListAll", ctx, tenantID.String()).Return(make([]*model.AutomaticScenarioAssignment, 0), nil).Times(3)
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(nil, apperrors.NewNotFoundError(resource.RuntimeContext, rtmContexts[0].ID)).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[2]).Return(rtmContexts[1], nil).Once()

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Times(2)

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(4)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			TransactionerFn: txGen.ThatSucceedsTwice,
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil).Once()

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[0]).Return(expectedRtmCtxLabels[0], nil).Once()

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[1]).Return(expectedRtmCtxLabels[1], nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimes[0].ID, model.ScenariosKey).Return(nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeContextLabelableObject, rtmContexts[0].ID, model.ScenariosKey).Return(nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeContextLabelableObject, rtmContexts[1].ID, model.ScenariosKey).Return(testErr).Once()

				return repo
			},
			NotificationSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[0].ID, &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[0]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmContexts[0].ID).Return([]*model.FormationAssignment{formationAssignments[0]}, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmIDs[0]).Return([]*model.FormationAssignment{formationAssignments[1]}, nil)

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmContexts[0].ID).Return(nil, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmIDs[0]).Return(nil, nil)

				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[0]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[0]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Returns error when listing runtime contexts for runtime fails",
			LabelDefServiceFn: unusedLabelDefService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(nil).Once()
				mockRepo.On("ListAll", ctx, tenantID.String()).Return(make([]*model.AutomaticScenarioAssignment, 0), nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(nil, testErr).Once()

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Once()

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Twice()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimes[0].ID, model.ScenariosKey).Return(nil).Once()
				return repo
			},
			NotificationSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmIDs[0]).Return([]*model.FormationAssignment{formationAssignments[1]}, nil)

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmIDs[0]).Return(nil, nil)

				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			TransactionerFn:    txGen.ThatSucceeds,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Returns error when listing all runtimes fails",
			LabelDefServiceFn: unusedLabelDefService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(nil).Once()
				mockRepo.On("ListAll", ctx, tenantID.String()).Return(make([]*model.AutomaticScenarioAssignment, 0), nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Once()

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Twice()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimes[0].ID, model.ScenariosKey).Return(nil).Once()
				return repo
			},
			NotificationSvcFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentSvcFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmIDs[0]).Return([]*model.FormationAssignment{formationAssignments[1]}, nil)

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmIDs[0]).Return(nil, nil)

				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			TransactionerFn:    txGen.ThatSucceeds,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:              "Returns error when unassigning runtime from formation fails",
			LabelDefServiceFn: unusedLabelDefService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(nil).Once()
				mockRepo.On("ListAll", ctx, tenantID.String()).Return(make([]*model.AutomaticScenarioAssignment, 0), nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Twice()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimes[0].ID, model.ScenariosKey).Return(testErr).Once()
				return repo
			},
			NotificationSvcFn:        unusedNotificationsService,
			TransactionerFn:          txGen.ThatDoesntStartTransaction,
			FormationAssignmentSvcFn: unusedFormationAssignmentService,
			InputASA:                 fixModel(testFormationName),
			ExpectedASA:              model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:       testErr.Error(),
		},
		{
			Name:              "Returns error when checking if runtime exists by id fails",
			LabelDefServiceFn: unusedLabelDefService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, testErr).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn:           unusedLabelService,
			LabelRepositoryFn:        unusedLabelRepo,
			NotificationSvcFn:        unusedNotificationsService,
			FormationAssignmentSvcFn: unusedFormationAssignmentService,
			TransactionerFn:          txGen.ThatDoesntStartTransaction,
			InputASA:                 fixModel(testFormationName),
			ExpectedASA:              model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:       testErr.Error(),
		},
		{
			Name:              "Returns error when listing owned runtimes fails",
			LabelDefServiceFn: unusedLabelDefService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn:           unusedLabelService,
			LabelRepositoryFn:        unusedLabelRepo,
			NotificationSvcFn:        unusedNotificationsService,
			FormationAssignmentSvcFn: unusedFormationAssignmentService,
			TransactionerFn:          txGen.ThatDoesntStartTransaction,
			InputASA:                 fixModel(testFormationName),
			ExpectedASA:              model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:       testErr.Error(),
		},
		{
			Name:              "Returns error when getting formation template by id fails",
			LabelDefServiceFn: unusedLabelDefService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			LabelServiceFn:           unusedLabelService,
			LabelRepositoryFn:        unusedLabelRepo,
			NotificationSvcFn:        unusedNotificationsService,
			FormationAssignmentSvcFn: unusedFormationAssignmentService,
			TransactionerFn:          txGen.ThatDoesntStartTransaction,
			InputASA:                 fixModel(testFormationName),
			ExpectedASA:              model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:       testErr.Error(),
		},
		{
			Name:              "Returns error when getting formation by name fails",
			LabelDefServiceFn: unusedLabelDefService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(nil, testErr).Once()
				return repo
			},
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			LabelServiceFn:                unusedLabelService,
			LabelRepositoryFn:             unusedLabelRepo,
			NotificationSvcFn:             unusedNotificationsService,
			FormationAssignmentSvcFn:      unusedFormationAssignmentService,
			TransactionerFn:               txGen.ThatDoesntStartTransaction,
			InputASA:                      fixModel(testFormationName),
			ExpectedASA:                   model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:            testErr.Error(),
		},
		{
			Name:              "Returns error when deleting ASA for scenario name fails",
			LabelDefServiceFn: unusedLabelDefService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), testFormationName).Return(testErr).Once()
				return mockRepo
			},
			RuntimeRepoFN:                 unusedRuntimeRepo,
			RuntimeContextRepoFn:          unusedRuntimeContextRepo,
			FormationRepositoryFn:         unusedFormationRepo,
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			LabelServiceFn:                unusedLabelService,
			LabelRepositoryFn:             unusedLabelRepo,
			NotificationSvcFn:             unusedNotificationsService,
			FormationAssignmentSvcFn:      unusedFormationAssignmentService,
			TransactionerFn:               txGen.ThatDoesntStartTransaction,
			InputASA:                      fixModel(testFormationName),
			ExpectedASA:                   model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:            testErr.Error(),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			asaRepo := testCase.AsaRepoFn()
			tenantSvc := &automock.TenantService{}
			labelDefService := testCase.LabelDefServiceFn()
			runtimeRepo := testCase.RuntimeRepoFN()
			runtimeContextRepo := testCase.RuntimeContextRepoFn()
			formationRepo := testCase.FormationRepositoryFn()
			formationTemplateRepo := testCase.FormationTemplateRepositoryFn()
			lblService := testCase.LabelServiceFn()
			lblRepo := testCase.LabelRepositoryFn()
			notificationSvc := testCase.NotificationSvcFn()
			formationAssignmentSvc := testCase.FormationAssignmentSvcFn()
			persist, transact := testCase.TransactionerFn()

			svc := formation.NewService(transact, nil, lblRepo, formationRepo, formationTemplateRepo, lblService, nil, labelDefService, asaRepo, nil, tenantSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, notificationSvc, runtimeType, applicationType)

			// WHEN
			err := svc.DeleteAutomaticScenarioAssignment(ctx, testCase.InputASA)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, asaRepo, labelDefService, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, lblService, lblRepo, notificationSvc, formationAssignmentSvc)
		})
	}

	t.Run("returns error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(context.TODO(), fixModel(ScenarioName))

		// THEN
		assert.EqualError(t, err, "cannot read tenant from context")
	})
}

func TestService_EnsureScenarioAssigned(t *testing.T) {
	ctx := fixCtxWithTenant()

	testErr := errors.New("test err")

	rtmIDs := []string{"123", "456", "789"}
	rtmNames := []string{"first", "second", "third"}

	runtimes := []*model.Runtime{
		{
			ID:   rtmIDs[0],
			Name: rtmNames[0],
		},
		{
			ID:   rtmIDs[1],
			Name: rtmNames[1],
		},
		{
			ID:   rtmIDs[2],
			Name: rtmNames[2],
		},
	}
	formationAssignments := []*model.FormationAssignment{
		{
			ID:          "fa1",
			FormationID: FormationID,
			Source:      "1",
		},
		{
			ID:          "fa2",
			FormationID: FormationID,
			Source:      "123",
		},
		{
			ID:          "fa3",
			FormationID: FormationID,
			Source:      "2",
		},
	}
	notifications := []*webhookclient.NotificationRequest{
		{
			Webhook: graphql.Webhook{
				ID: "wid1",
			},
		},
		{
			Webhook: graphql.Webhook{
				ID: "wid2",
			},
		},
		{
			Webhook: graphql.Webhook{
				ID: "wid3",
			},
		},
	}

	ownedRuntimes := []*model.Runtime{runtimes[0], runtimes[1]}

	runtimeLblInputs := []*model.LabelInput{
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmIDs[0],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmIDs[1],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
	}

	rtmContexts := []*model.RuntimeContext{
		{
			ID:        "1",
			RuntimeID: rtmIDs[0],
			Key:       "test",
			Value:     "test",
		},
		{
			ID:        "2",
			RuntimeID: rtmIDs[2],
			Key:       "test",
			Value:     "test",
		},
	}

	runtimeCtxLblInputs := []*model.LabelInput{
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmContexts[0].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmContexts[1].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
	}
	tnt := tenantID.String()

	expectedRtmCtxLabels := []*model.Label{
		{
			ID:         "1",
			Tenant:     &tnt,
			Key:        "scenarios",
			Value:      []interface{}{},
			ObjectID:   rtmContexts[0].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
		{
			ID:         "2",
			Tenant:     &tnt,
			Key:        "scenarios",
			Value:      []interface{}{},
			ObjectID:   rtmContexts[1].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
	}
	runtimeTypeLblInput := []*model.LabelInput{
		{
			Key:        runtimeType,
			ObjectID:   rtmIDs[0],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
		{
			Key:        runtimeType,
			ObjectID:   rtmIDs[1],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
		{
			Key:        runtimeType,
			ObjectID:   rtmIDs[2],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
	}
	runtimeTypeLbl := &model.Label{
		ID:         "123",
		Key:        runtimeType,
		Value:      runtimeType,
		Tenant:     str.Ptr(Tnt),
		ObjectID:   rtmIDs[0],
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}

	expectedRuntimeLabel := &model.Label{
		ID:         "1",
		Tenant:     &tnt,
		Key:        "scenarios",
		Value:      []interface{}{},
		ObjectID:   rtmIDs[0],
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}

	testCases := []struct {
		Name                          string
		RuntimeRepoFN                 func() *automock.RuntimeRepository
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		LabelServiceFn                func() *automock.LabelService
		LabelRepositoryFn             func() *automock.LabelRepository
		NotificationServiceFn         func() *automock.NotificationsService
		FormationAssignmentServiceFn  func() *automock.FormationAssignmentService
		InputASA                      model.AutomaticScenarioAssignment
		ExpectedErrMessage            string
	}{
		{
			Name: "Success",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(nil, apperrors.NewNotFoundError(resource.RuntimeContext, rtmContexts[0].ID)).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[2]).Return(rtmContexts[1], nil).Once()

				runtimeContextRepo.On("GetByID", ctx, tenantID.String(), rtmContexts[0].ID).Return(&model.RuntimeContext{
					RuntimeID: runtimes[0].ID,
				}, nil)
				runtimeContextRepo.On("GetByID", ctx, tenantID.String(), rtmContexts[1].ID).Return(&model.RuntimeContext{
					RuntimeID: runtimes[1].ID,
				}, nil)

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Times(3)

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(4)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Times(4)
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[0]).Return(runtimeTypeLbl, nil)
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[1]).Return(runtimeTypeLbl, nil)

				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[0]).Return(expectedRtmCtxLabels[0], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[0].ID, runtimeCtxLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[1]).Return(expectedRtmCtxLabels[1], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[1].ID, runtimeCtxLblInputs[1]).Return(nil)
				return svc
			},
			LabelRepositoryFn: unusedLabelRepo,
			NotificationServiceFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[0].ID, &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[0]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[1].ID, &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[2]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmContexts[0].ID, graphql.FormationObjectTypeRuntimeContext, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[0]}, nil).Once()
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmIDs[0], graphql.FormationObjectTypeRuntime, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[1]}, nil).Once()
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmContexts[1].ID, graphql.FormationObjectTypeRuntimeContext, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[2]}, nil).Once()

				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[0]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[0]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[2]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[2]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when assigning runtime context to formation fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}

				runtimeContextRepo.On("GetByID", ctx, tenantID.String(), rtmContexts[0].ID).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByID", ctx, tenantID.String(), rtmContexts[1].ID).Return(rtmContexts[1], nil).Once()

				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(nil, apperrors.NewNotFoundError(resource.RuntimeContext, rtmContexts[0].ID)).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[2]).Return(rtmContexts[1], nil).Once()

				runtimeContextRepo.On("GetByID", ctx, tenantID.String(), rtmContexts[0].ID).Return(&model.RuntimeContext{
					RuntimeID: runtimes[0].ID,
				}, nil)

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Times(2)

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(4)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Times(4)
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[0]).Return(runtimeTypeLbl, nil)
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[2]).Return(runtimeTypeLbl, nil)

				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[0]).Return(expectedRtmCtxLabels[0], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[0].ID, runtimeCtxLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[1]).Return(expectedRtmCtxLabels[1], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[1].ID, runtimeCtxLblInputs[1]).Return(testErr)
				return svc
			},
			NotificationServiceFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[0].ID, &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[0]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmContexts[0].ID, graphql.FormationObjectTypeRuntimeContext, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[0]}, nil).Once()
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmIDs[0], graphql.FormationObjectTypeRuntime, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[1]}, nil).Once()

				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[0]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[0]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			LabelRepositoryFn:  unusedLabelRepo,
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing runtime contexts for runtime fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(nil, testErr).Once()

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(2)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Times(2)
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[0]).Return(runtimeTypeLbl, nil)
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)
				return svc
			},
			NotificationServiceFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmIDs[0], graphql.FormationObjectTypeRuntime, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[1]}, nil).Once()

				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			LabelRepositoryFn:  unusedLabelRepo,
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing all runtimes fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(2)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Times(2)
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[0]).Return(runtimeTypeLbl, nil)
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)
				return svc
			},
			NotificationServiceFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctx, tnt, rtmIDs[0], graphql.FormationObjectTypeRuntime, &modelFormation).Return([]*model.FormationAssignment{formationAssignments[1]}, nil).Once()

				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			LabelRepositoryFn:  unusedLabelRepo,
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when assigning runtime to formation fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(2)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Times(2)
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeTypeLblInput[0]).Return(runtimeTypeLbl, nil)
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(nil, testErr)
				return svc
			},
			LabelRepositoryFn:            unusedLabelRepo,
			NotificationServiceFn:        unusedNotificationsService,
			FormationAssignmentServiceFn: unusedFormationAssignmentService,
			InputASA:                     fixModel(testFormationName),
			ExpectedErrMessage:           testErr.Error(),
		},
		{
			Name: "Returns error when checking if runtime exists by id fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, testErr).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(1)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelRepositoryFn:            unusedLabelRepo,
			NotificationServiceFn:        unusedNotificationsService,
			FormationAssignmentServiceFn: unusedFormationAssignmentService,
			LabelServiceFn:               unusedLabelService,
			InputASA:                     fixModel(testFormationName),
			ExpectedErrMessage:           testErr.Error(),
		},
		{
			Name: "Returns error when listing owned runtimes fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(1)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelRepositoryFn:            unusedLabelRepo,
			NotificationServiceFn:        unusedNotificationsService,
			FormationAssignmentServiceFn: unusedFormationAssignmentService,
			LabelServiceFn:               unusedLabelService,
			InputASA:                     fixModel(testFormationName),
			ExpectedErrMessage:           testErr.Error(),
		},
		{
			Name:                 "Returns error getting formation template by ID fails",
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(1)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			LabelRepositoryFn:            unusedLabelRepo,
			NotificationServiceFn:        unusedNotificationsService,
			FormationAssignmentServiceFn: unusedFormationAssignmentService,
			LabelServiceFn:               unusedLabelService,
			InputASA:                     fixModel(testFormationName),
			ExpectedErrMessage:           testErr.Error(),
		},
		{
			Name:                 "Returns error getting formation by name fails",
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(nil, testErr).Times(1)
				return repo
			},
			LabelRepositoryFn:             unusedLabelRepo,
			NotificationServiceFn:         unusedNotificationsService,
			FormationAssignmentServiceFn:  unusedFormationAssignmentService,
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			LabelServiceFn:                unusedLabelService,
			InputASA:                      fixModel(testFormationName),
			ExpectedErrMessage:            testErr.Error(),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			runtimeRepo := testCase.RuntimeRepoFN()
			runtimeContextRepo := testCase.RuntimeContextRepoFn()
			formationRepo := testCase.FormationRepositoryFn()
			formationTemplateRepo := testCase.FormationTemplateRepositoryFn()
			lblService := testCase.LabelServiceFn()
			labelRepo := testCase.LabelRepositoryFn()

			notificationSvc := testCase.NotificationServiceFn()
			formationAssignmentSvc := testCase.FormationAssignmentServiceFn()
			svc := formation.NewService(nil, nil, labelRepo, formationRepo, formationTemplateRepo, lblService, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, notificationSvc, runtimeType, applicationType)

			// WHEN
			err := svc.EnsureScenarioAssigned(ctx, testCase.InputASA)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, lblService, labelRepo, notificationSvc, formationAssignmentSvc)
		})
	}
}

func TestService_RemoveAssignedScenario(t *testing.T) {
	ctx := fixCtxWithTenant()

	testErr := errors.New("test err")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	rtmIDs := []string{"123", "456", "789"}
	rtmNames := []string{"first", "second", "third"}

	runtimes := []*model.Runtime{
		{
			ID:   rtmIDs[0],
			Name: rtmNames[0],
		},
		{
			ID:   rtmIDs[1],
			Name: rtmNames[1],
		},
		{
			ID:   rtmIDs[2],
			Name: rtmNames[2],
		},
	}
	formationAssignments := []*model.FormationAssignment{
		{
			ID:          "fa1",
			FormationID: FormationID,
			Source:      "1",
		},
		{
			ID:          "fa2",
			FormationID: FormationID,
			Source:      "123",
		},
		{
			ID:          "fa3",
			FormationID: FormationID,
			Source:      "2",
		},
	}
	notifications := []*webhookclient.NotificationRequest{
		{
			Webhook: graphql.Webhook{
				ID: "wid1",
			},
		},
		{
			Webhook: graphql.Webhook{
				ID: "wid2",
			},
		},
		{
			Webhook: graphql.Webhook{
				ID: "wid3",
			},
		},
	}
	ownedRuntimes := []*model.Runtime{runtimes[0], runtimes[1]}

	runtimeLblInputs := []*model.LabelInput{
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmIDs[0],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmIDs[1],
			ObjectType: model.RuntimeLabelableObject,
			Version:    0,
		},
	}

	rtmContexts := []*model.RuntimeContext{
		{
			ID:        "1",
			RuntimeID: rtmIDs[0],
			Key:       "test",
			Value:     "test",
		},
		{
			ID:        "2",
			RuntimeID: rtmIDs[2],
			Key:       "test",
			Value:     "test",
		},
	}

	runtimeCtxLblInputs := []*model.LabelInput{
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmContexts[0].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
		{
			Key:        "scenarios",
			Value:      []string{testFormationName},
			ObjectID:   rtmContexts[1].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
	}
	tnt := tenantID.String()

	expectedRtmCtxLabels := []*model.Label{
		{
			ID:         "1",
			Tenant:     &tnt,
			Key:        "scenarios",
			Value:      []interface{}{testFormationName},
			ObjectID:   rtmContexts[0].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
		{
			ID:         "2",
			Tenant:     &tnt,
			Key:        "scenarios",
			Value:      []interface{}{testFormationName},
			ObjectID:   rtmContexts[1].ID,
			ObjectType: model.RuntimeContextLabelableObject,
			Version:    0,
		},
	}

	expectedRuntimeLabel := &model.Label{
		ID:         "1",
		Tenant:     &tnt,
		Key:        "scenarios",
		Value:      []interface{}{testFormationName},
		ObjectID:   rtmIDs[0],
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}

	testCases := []struct {
		Name                          string
		AsaRepoFn                     func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFN                 func() *automock.RuntimeRepository
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		LabelServiceFn                func() *automock.LabelService
		LabelRepositoryFn             func() *automock.LabelRepository
		NotificationServiceFn         func() *automock.NotificationsService
		FormationAssignmentServiceFn  func() *automock.FormationAssignmentService
		TransactionerFn               func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		InputASA                      model.AutomaticScenarioAssignment
		ExpectedErrMessage            string
	}{
		{
			Name: "Success",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("ListAll", ctx, tenantID.String()).Return(make([]*model.AutomaticScenarioAssignment, 0), nil).Times(3)
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(nil, apperrors.NewNotFoundError(resource.RuntimeContext, rtmContexts[0].ID)).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[2]).Return(rtmContexts[1], nil).Once()

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Times(3)

				return runtimeContextRepo
			},
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(4)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil).Once()

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[0]).Return(expectedRtmCtxLabels[0], nil).Once()

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[1]).Return(expectedRtmCtxLabels[1], nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimes[0].ID, model.ScenariosKey).Return(nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeContextLabelableObject, rtmContexts[0].ID, model.ScenariosKey).Return(nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeContextLabelableObject, rtmContexts[1].ID, model.ScenariosKey).Return(nil).Once()

				return repo
			},
			NotificationServiceFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[0].ID, &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[0]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[1].ID, &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[2]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmContexts[0].ID).Return([]*model.FormationAssignment{formationAssignments[0]}, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmIDs[0]).Return([]*model.FormationAssignment{formationAssignments[1]}, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmContexts[1].ID).Return([]*model.FormationAssignment{formationAssignments[2]}, nil)

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmContexts[0].ID).Return(nil, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmIDs[0]).Return(nil, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmContexts[1].ID).Return(nil, nil)

				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[0]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[0]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[2]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[2]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when unassigning runtime context fails",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("ListAll", ctx, tenantID.String()).Return(make([]*model.AutomaticScenarioAssignment, 0), nil).Times(3)
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()

				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(nil, apperrors.NewNotFoundError(resource.RuntimeContext, rtmContexts[0].ID)).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[2]).Return(rtmContexts[1], nil).Once()

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Twice()

				return runtimeContextRepo
			},
			TransactionerFn: txGen.ThatSucceedsTwice,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(4)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil).Once()

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[0]).Return(expectedRtmCtxLabels[0], nil).Once()

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[1]).Return(expectedRtmCtxLabels[1], nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimes[0].ID, model.ScenariosKey).Return(nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeContextLabelableObject, rtmContexts[0].ID, model.ScenariosKey).Return(nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeContextLabelableObject, rtmContexts[1].ID, model.ScenariosKey).Return(testErr).Once()

				return repo
			},
			NotificationServiceFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmContexts[0].ID, &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntimeContext).Return([]*webhookclient.NotificationRequest{notifications[0]}, nil).Once()
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmContexts[0].ID).Return([]*model.FormationAssignment{formationAssignments[0]}, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmIDs[0]).Return([]*model.FormationAssignment{formationAssignments[1]}, nil)

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmContexts[0].ID).Return(nil, nil)
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmIDs[0]).Return(nil, nil)

				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[0]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[0]}, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing runtime contexts for runtime fails",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("ListAll", ctx, tenantID.String()).Return(make([]*model.AutomaticScenarioAssignment, 0), nil).Times(1)
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(nil, testErr).Once()

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Once()
				return runtimeContextRepo
			},
			TransactionerFn: txGen.ThatSucceeds,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Twice()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimes[0].ID, model.ScenariosKey).Return(nil).Once()

				return repo
			},
			NotificationServiceFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmIDs[0]).Return([]*model.FormationAssignment{formationAssignments[1]}, nil)

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmIDs[0]).Return(nil, nil)

				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing all runtimes fails",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("ListAll", ctx, tenantID.String()).Return(make([]*model.AutomaticScenarioAssignment, 0), nil).Times(1)
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("ListByIDs", mock.Anything, tenantID.String(), []string{}).Return(nil, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Twice()
				return repo
			},
			TransactionerFn: txGen.ThatSucceeds,
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimes[0].ID, model.ScenariosKey).Return(nil).Once()

				return repo
			},
			NotificationServiceFn: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, tnt, rtmIDs[0], &modelFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return([]*webhookclient.NotificationRequest{notifications[1]}, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, modelFormation.ID, rtmIDs[0]).Return([]*model.FormationAssignment{formationAssignments[1]}, nil)

				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), modelFormation.ID, rtmIDs[0]).Return(nil, nil)

				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), []*model.FormationAssignment{formationAssignments[1]}, map[string]string{}, []*webhookclient.NotificationRequest{notifications[1]}, mock.Anything).Return(nil).Once()
				return formationAssignmentSvc
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when unassigning runtime from formation fails",
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("ListAll", ctx, tenantID.String()).Return(make([]*model.AutomaticScenarioAssignment, 0), nil).Times(1)
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Twice()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimes[0].ID, model.ScenariosKey).Return(testErr).Once()
				return repo
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:      "Returns error when checking if runtime exists by id fails",
			AsaRepoFn: unusedASARepo,
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, testErr).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			TransactionerFn:    txGen.ThatDoesntStartTransaction,
			LabelServiceFn:     unusedLabelService,
			LabelRepositoryFn:  unusedLabelRepo,
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:      "Returns error when listing owned runtimes fails",
			AsaRepoFn: unusedASARepo,
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			TransactionerFn:    txGen.ThatDoesntStartTransaction,
			LabelServiceFn:     unusedLabelService,
			LabelRepositoryFn:  unusedLabelRepo,
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:                 "Returns error when getting formation template by id fails",
			AsaRepoFn:            unusedASARepo,
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			LabelServiceFn:     unusedLabelService,
			LabelRepositoryFn:  unusedLabelRepo,
			TransactionerFn:    txGen.ThatDoesntStartTransaction,
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:                 "Returns error when getting formation by name fails",
			AsaRepoFn:            unusedASARepo,
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(nil, testErr).Once()
				return repo
			},
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			LabelServiceFn:                unusedLabelService,
			LabelRepositoryFn:             unusedLabelRepo,
			TransactionerFn:               txGen.ThatDoesntStartTransaction,
			InputASA:                      fixModel(testFormationName),
			ExpectedErrMessage:            testErr.Error(),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			asaRepo := testCase.AsaRepoFn()
			runtimeRepo := testCase.RuntimeRepoFN()
			runtimeContextRepo := testCase.RuntimeContextRepoFn()
			formationRepo := testCase.FormationRepositoryFn()
			formationTemplateRepo := testCase.FormationTemplateRepositoryFn()
			lblService := testCase.LabelServiceFn()
			lblRepo := testCase.LabelRepositoryFn()
			persist, transact := testCase.TransactionerFn()

			notificationSvc := unusedNotificationsService()
			if testCase.NotificationServiceFn != nil {
				notificationSvc = testCase.NotificationServiceFn()
			}
			formationAssignmentSvc := unusedFormationAssignmentService()
			if testCase.FormationAssignmentServiceFn != nil {
				formationAssignmentSvc = testCase.FormationAssignmentServiceFn()
			}
			svc := formation.NewService(transact, nil, lblRepo, formationRepo, formationTemplateRepo, lblService, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, notificationSvc, runtimeType, applicationType)

			// WHEN
			err := svc.RemoveAssignedScenario(ctx, testCase.InputASA)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, lblService, lblRepo, asaRepo, notificationSvc, formationAssignmentSvc)
		})
	}
}

func TestService_MergeScenariosFromInputLabelsAndAssignments(t *testing.T) {
	// GIVEN
	ctx := fixCtxWithTenant()

	testErr := errors.New("Test error")

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

			svc := formation.NewService(nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, runtimeType, applicationType)

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

func TestService_GetScenariosFromMatchingASAs(t *testing.T) {
	ctx := fixCtxWithTenant()
	runtimeID := "runtimeID"
	runtimeID2 := "runtimeID2"

	testErr := errors.New(ErrMsg)
	notFoudErr := apperrors.NewNotFoundError(resource.Runtime, runtimeID2)

	testScenarios := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   ScenarioName,
			Tenant:         tenantID.String(),
			TargetTenantID: TargetTenantID,
		},
		{
			ScenarioName:   ScenarioName2,
			Tenant:         TenantID2,
			TargetTenantID: TargetTenantID2,
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

	rtmCtx := &model.RuntimeContext{
		ID:        RuntimeContextID,
		Key:       "subscription",
		Value:     "subscriptionValue",
		RuntimeID: runtimeID,
	}

	rtmCtx2 := &model.RuntimeContext{
		ID:        RuntimeContextID,
		Key:       "subscription",
		Value:     "subscriptionValue",
		RuntimeID: runtimeID2,
	}

	testCases := []struct {
		Name                     string
		ScenarioAssignmentRepoFn func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFn            func() *automock.RuntimeRepository
		RuntimeContextRepoFn     func() *automock.RuntimeContextRepository
		FormationRepoFn          func() *automock.FormationRepository
		FormationTemplateRepoFn  func() *automock.FormationTemplateRepository
		ObjectID                 string
		ObjectType               graphql.FormationObjectType
		ExpectedError            error
		ExpectedScenarios        []string
	}{
		{
			Name: "Success for runtime",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, RuntimeID).Return(false, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[0].TargetTenantID, RuntimeID, runtimeLblFilters).Return(true, nil).Once()
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[1].TargetTenantID, RuntimeID, runtimeLblFilters).Return(false, nil).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName2, testScenarios[1].Tenant).Return(formations[1], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:          RuntimeID,
			ObjectType:        graphql.FormationObjectTypeRuntime,
			ExpectedError:     nil,
			ExpectedScenarios: []string{ScenarioName},
		},
		{
			Name: "Success for runtime context",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("GetByID", ctx, testScenarios[0].TargetTenantID, RuntimeContextID).Return(rtmCtx, nil).Once()
				runtimeContextRepo.On("GetByID", ctx, testScenarios[1].TargetTenantID, RuntimeContextID).Return(rtmCtx2, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("GetByFiltersAndIDUsingUnion", ctx, testScenarios[0].TargetTenantID, rtmCtx.RuntimeID, runtimeLblFilters).Return(&model.Runtime{}, nil).Once()
				runtimeRepo.On("GetByFiltersAndIDUsingUnion", ctx, testScenarios[1].TargetTenantID, rtmCtx2.RuntimeID, runtimeLblFilters).Return(nil, notFoudErr).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName2, testScenarios[1].Tenant).Return(formations[1], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:          RuntimeContextID,
			ObjectType:        graphql.FormationObjectTypeRuntimeContext,
			ExpectedError:     nil,
			ExpectedScenarios: []string{ScenarioName},
		},
		{
			Name: "Returns an error when getting runtime contexts",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("GetByID", ctx, testScenarios[0].TargetTenantID, RuntimeContextID).Return(nil, testErr).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: unusedRuntimeRepo,
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:          RuntimeContextID,
			ObjectType:        graphql.FormationObjectTypeRuntimeContext,
			ExpectedError:     nil,
			ExpectedScenarios: nil,
		},
		{
			Name: "Returns an not found error when getting runtime contexts",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("GetByID", ctx, testScenarios[0].TargetTenantID, RuntimeContextID).Return(nil, notFoudErr).Once()
				runtimeContextRepo.On("GetByID", ctx, testScenarios[1].TargetTenantID, RuntimeContextID).Return(nil, notFoudErr).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: unusedRuntimeRepo,
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName2, testScenarios[1].Tenant).Return(formations[1], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:          RuntimeContextID,
			ObjectType:        graphql.FormationObjectTypeRuntimeContext,
			ExpectedError:     nil,
			ExpectedScenarios: nil,
		},
		{
			Name: "Returns an error when getting runtime",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("GetByID", ctx, testScenarios[0].TargetTenantID, RuntimeContextID).Return(rtmCtx, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("GetByFiltersAndIDUsingUnion", ctx, testScenarios[0].TargetTenantID, rtmCtx.RuntimeID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:          RuntimeContextID,
			ObjectType:        graphql.FormationObjectTypeRuntimeContext,
			ExpectedError:     nil,
			ExpectedScenarios: nil,
		},
		{
			Name: "Returns an error when getting formations",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			RuntimeRepoFn:        unusedRuntimeRepo,
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(nil, testErr).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: unusedFormationTemplateRepo,
			ObjectID:                RuntimeContextID,
			ObjectType:              graphql.FormationObjectTypeRuntimeContext,
			ExpectedError:           nil,
			ExpectedScenarios:       nil,
		},
		{
			Name: "Returns error for runtime when checking if the runtime has context fails",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, RuntimeID).Return(false, testErr).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[0].TargetTenantID, RuntimeID, runtimeLblFilters).Return(true, nil).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:          RuntimeID,
			ObjectType:        graphql.FormationObjectTypeRuntime,
			ExpectedError:     testErr,
			ExpectedScenarios: nil,
		},
		{
			Name: "Returns error for runtime when checking if runtime exists by filters and ID and has owner=true fails",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[0].TargetTenantID, RuntimeID, runtimeLblFilters).Return(false, testErr).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:          RuntimeID,
			ObjectType:        graphql.FormationObjectTypeRuntime,
			ExpectedError:     testErr,
			ExpectedScenarios: nil,
		},
		{
			Name: "Returns error for runtime when getting formation template runtime type fails",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			RuntimeRepoFn:        unusedRuntimeRepo,
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(nil, testErr).Once()
				return formationTemplateRepo
			},
			ObjectID:          RuntimeID,
			ObjectType:        graphql.FormationObjectTypeRuntime,
			ExpectedError:     testErr,
			ExpectedScenarios: nil,
		},
		{
			Name: "Returns error when listing ASAs fails",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(nil, testErr)
				return repo
			},
			RuntimeContextRepoFn:    unusedRuntimeContextRepo,
			RuntimeRepoFn:           unusedRuntimeRepo,
			FormationRepoFn:         unusedFormationRepo,
			FormationTemplateRepoFn: unusedFormationTemplateRepo,
			ObjectID:                RuntimeID,
			ObjectType:              graphql.FormationObjectTypeRuntime,
			ExpectedError:           testErr,
			ExpectedScenarios:       nil,
		},
		{
			Name:                     "Returns error when can't find matching func",
			ScenarioAssignmentRepoFn: unusedASARepo,
			RuntimeContextRepoFn:     unusedRuntimeContextRepo,
			RuntimeRepoFn:            unusedRuntimeRepo,
			FormationRepoFn:          unusedFormationRepo,
			FormationTemplateRepoFn:  unusedFormationTemplateRepo,
			ObjectID:                 "",
			ObjectType:               "test",
			ExpectedError:            errors.New("unexpected formation object type"),
			ExpectedScenarios:        nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			asaRepo := testCase.ScenarioAssignmentRepoFn()
			runtimeRepo := testCase.RuntimeRepoFn()
			runtimeContextRepo := testCase.RuntimeContextRepoFn()
			formationRepo := testCase.FormationRepoFn()
			formationTemplateRepo := testCase.FormationTemplateRepoFn()

			svc := formation.NewService(nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, runtimeType, applicationType)

			// WHEN
			scenarios, err := svc.GetScenariosFromMatchingASAs(ctx, testCase.ObjectID, testCase.ObjectType)

			// THEN
			if testCase.ExpectedError == nil {
				require.ElementsMatch(t, scenarios, testCase.ExpectedScenarios)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
				require.Nil(t, testCase.ExpectedScenarios)
			}

			mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo)
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

		svc := formation.NewService(nil, nil, nil, nil, nil, labelService, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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

		svc := formation.NewService(nil, nil, nil, nil, nil, labelService, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		formations, err := svc.GetFormationsForObject(ctx, tenantID.String(), model.RuntimeLabelableObject, id)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "while fetching scenario label for")
		require.Nil(t, formations)
		mock.AssertExpectationsForObjects(t, labelService)
	})
}
