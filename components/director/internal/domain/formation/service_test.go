package formation_test

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

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

			svc := formation.NewService(nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

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

			svc := formation.NewService(nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

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

	defaultSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
	assert.NoError(t, err)
	defaultSchemaLblDef := fixDefaultScenariosLabelDefinition(Tnt, defaultSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormationName})
	assert.NoError(t, err)
	newSchemaLblDef := fixDefaultScenariosLabelDefinition(Tnt, newSchema)

	emptySchemaLblDef := fixDefaultScenariosLabelDefinition(Tnt, defaultSchema)
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, defaultSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when validating automatic scenario assignment against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, defaultSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, defaultSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when update with version fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(testErr)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, defaultSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, defaultSchemaLblDef.Key).Return(nil)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when getting formation template by name fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
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

			svc := formation.NewService(lblDefRepo, nil, formationRepoMock, formationTemplateRepoMock, nil, uuidSvcMock, lblDefService, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

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

	defaultSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormationName})
	assert.NoError(t, err)
	defaultSchemaLblDef := fixDefaultScenariosLabelDefinition(Tnt, defaultSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
	assert.NoError(t, err)
	newSchemaLblDef := fixDefaultScenariosLabelDefinition(Tnt, newSchema)

	emptySchema, err := labeldef.NewSchemaForFormations([]string{})
	assert.NoError(t, err)
	emptySchemaLblDef := fixDefaultScenariosLabelDefinition(Tnt, emptySchema)

	nilSchemaLblDef := fixDefaultScenariosLabelDefinition(Tnt, defaultSchema)
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
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
			Name: "success when default scenario",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&newSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, emptySchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, emptySchema, Tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, emptySchema, Tnt, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			FormationRepoFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("DeleteByName", ctx, Tnt, model.DefaultScenario).Return(nil).Once()
				return repo
			},
			InputFormation:     model.Formation{Name: model.DefaultScenario},
			ExpectedFormation:  &defaultFormation,
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
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

			svc := formation.NewService(lblDefRepo, nil, formationRepoMock, nil, nil, nil, lblDefService, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

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

func TestServiceAssignFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("test error")

	inputFormation := model.Formation{
		Name: testFormationName,
	}
	expectedFormation := &model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            Tnt,
	}

	inputSecondFormation := model.Formation{
		Name: secondTestFormationName,
	}
	expectedSecondFormation := &model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            Tnt,
	}

	objectID := "123"

	applicationLblNoFormations := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	applicationLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	applicationLblInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	applicationLblInputInDefaultScenario := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{model.DefaultScenario},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	runtimeLblNoFormations := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}

	runtimeLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeLblInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeCtxLblInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   objectID,
		ObjectType: model.RuntimeContextLabelableObject,
		Version:    0,
	}

	asa := model.AutomaticScenarioAssignment{
		ScenarioName:   testFormationName,
		Tenant:         Tnt,
		TargetTenantID: TargetTenant,
	}

	testCases := []struct {
		Name                          string
		UIDServiceFn                  func() *automock.UuidService
		LabelServiceFn                func() *automock.LabelService
		LabelDefServiceFn             func() *automock.LabelDefService
		TenantServiceFn               func() *automock.TenantService
		AsaRepoFn                     func() *automock.AutomaticFormationAssignmentRepository
		AsaServiceFN                  func() *automock.AutomaticFormationAssignmentService
		RuntimeRepoFN                 func() *automock.RuntimeRepository
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		ApplicationRepoFN             func() *automock.ApplicationRepository
		WebhookRepoFN                 func() *automock.WebhookRepository
		WebhookConverterFN            func() *automock.WebhookConverter
		WebhookClientFN               func() *automock.WebhookClient
		ApplicationTemplateRepoFN     func() *automock.ApplicationTemplateRepository
		LabelRepoFN                   func() *automock.LabelRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		ObjectType                    graphql.FormationObjectType
		InputFormation                model.Formation
		ExpectedFormation             *model.Formation
		ExpectedErrMessage            string
	}{
		{
			Name: "success for application if label does not exist",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(&model.Application{}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application if formation is already added",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(applicationLblNoFormations, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(&model.Application{}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application with new formation",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(&model.Application{}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application with default formation",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInputInDefaultScenario).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInputInDefaultScenario).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(&model.Application{}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     defaultFormation,
			ExpectedFormation:  &defaultFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime if label does not exist",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, objectID))
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime if formation is already added",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(runtimeLblNoFormations, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, objectID))
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime with new formation",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, objectID))
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for tenant",
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, objectID).Return(TargetTenant, nil)
				return svc
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", ctx, Tnt).Return(nil)
				labelDefSvc.On("GetAvailableScenarios", ctx, Tnt).Return([]string{testFormationName}, nil)

				return labelDefSvc
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctx, asa).Return(nil)

				return asaRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenant, runtimeLblFilters).Return(make([]*model.Runtime, 0), nil).Once()
				runtimeRepo.On("ListAll", ctx, TargetTenant, runtimeLblFilters).Return(make([]*model.Runtime, 0), nil).Once()
				return runtimeRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Twice()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for application when label does not exist and can't create it",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(testErr)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application while getting label",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, testErr)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application while converting label values to string slice",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name: "error for application while converting label value to string",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name: "error for application when updating label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(applicationLblNoFormations, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &applicationLblInput).Return(testErr)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime when label does not exist and can't create it",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(testErr)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime while getting label",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, testErr)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime while converting label values to string slice",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name: "error for runtime while converting label value to string",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name: "error for runtime when updating label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(runtimeLblNoFormations, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &runtimeLblInput).Return(testErr)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant when tenant conversion fails",
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, objectID).Return("", testErr)
				return svc
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant when create fails",
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, objectID).Return(TargetTenant, nil)
				return svc
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctx, model.AutomaticScenarioAssignment{ScenarioName: testFormationName, Tenant: Tnt, TargetTenantID: TargetTenant}).Return(testErr)

				return asaRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", ctx, Tnt).Return(nil)
				labelDefSvc.On("GetAvailableScenarios", ctx, Tnt).Return([]string{testFormationName}, nil)

				return labelDefSvc
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when can't get formation by name",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, Tnt).Return(nil, testErr).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "error when object type is unknown",
			ObjectType:         "UNKNOWN",
			InputFormation:     inputFormation,
			ExpectedErrMessage: "unknown formation type",
		},
		{
			Name: "success for application if label does not exist with notifications",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeID, RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookID, RuntimeID), fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeID), nil)
				repo.On("ToGraphQL", fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID), nil)
				return repo
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, nil)
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, nil)

				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID), fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for application webhook client request fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID), nil)
				return repo
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, testErr)

				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when webhook conversion fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtime context labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtime contexts in scenario fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtimes in scenario fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching listening runtimes labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching listening runtimes fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching webhooks fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application template labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application template fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for runtime if label does not exist with notifications",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:           model.AssignFormation,
						FormationID:         expectedFormation.ID,
						ApplicationTemplate: nil,
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModelWithoutTemplate(Application2ID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime if webhook client call fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: FormationID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, testErr).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if webhook conversion fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(nil, testErr)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application template labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application templates fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for runtime if there are no applications to notify",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime if fetching applications fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching webhooks fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching runtime labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching runtime fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for runtime context if label does not exist with notifications",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:           model.AssignFormation,
						FormationID:         expectedFormation.ID,
						ApplicationTemplate: nil,
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModelWithoutTemplate(Application2ID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime context if webhook client call fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, testErr).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if webhook conversion fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeContextRuntimeID)).Return(nil, testErr)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application template labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application templates fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching applications fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return(nil, testErr)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching webhook fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(nil, testErr)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime context labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime context fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			uidService := unusedUUIDService()
			if testCase.UIDServiceFn != nil {
				uidService = testCase.UIDServiceFn()
			}
			labelService := unusedLabelService()
			if testCase.LabelServiceFn != nil {
				labelService = testCase.LabelServiceFn()
			}
			asaRepo := unusedASARepo()
			if testCase.AsaRepoFn != nil {
				asaRepo = testCase.AsaRepoFn()
			}
			asaService := unusedASAService()
			if testCase.AsaServiceFN != nil {
				asaService = testCase.AsaServiceFN()
			}
			tenantSvc := &automock.TenantService{}
			if testCase.TenantServiceFn != nil {
				tenantSvc = testCase.TenantServiceFn()
			}
			labelDefService := unusedLabelDefService()
			if testCase.LabelDefServiceFn != nil {
				labelDefService = testCase.LabelDefServiceFn()
			}
			runtimeRepo := unusedRuntimeRepo()
			if testCase.RuntimeRepoFN != nil {
				runtimeRepo = testCase.RuntimeRepoFN()
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
			applicationRepo := unusedApplicationRepo()
			if testCase.ApplicationRepoFN != nil {
				applicationRepo = testCase.ApplicationRepoFN()
			}
			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFN != nil {
				webhookRepo = testCase.WebhookRepoFN()
			}
			webhookConverter := unusedWebhookConverter()
			if testCase.WebhookConverterFN != nil {
				webhookConverter = testCase.WebhookConverterFN()
			}
			webhookClient := unusedWebhookClient()
			if testCase.WebhookClientFN != nil {
				webhookClient = testCase.WebhookClientFN()
			}
			appTemplateRepo := unusedAppTemplateRepository()
			if testCase.ApplicationTemplateRepoFN != nil {
				appTemplateRepo = testCase.ApplicationTemplateRepoFN()
			}
			labelRepo := unusedLabelRepo()
			if testCase.LabelRepoFN != nil {
				labelRepo = testCase.LabelRepoFN()
			}

			svc := formation.NewService(nil, labelRepo, formationRepo, formationTemplateRepo, labelService, uidService, labelDefService, asaRepo, asaService, tenantSvc, runtimeRepo, runtimeContextRepo, webhookRepo, webhookClient, applicationRepo, appTemplateRepo, webhookConverter)

			// WHEN
			actual, err := svc.AssignFormation(ctx, Tnt, objectID, testCase.ObjectType, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, uidService, labelService, asaService, tenantSvc, asaRepo, labelDefService, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, applicationRepo, webhookRepo, webhookConverter, webhookClient, appTemplateRepo, labelRepo)
		})
	}
}

func TestServiceUnassignFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("test error")

	in := model.Formation{
		Name: testFormationName,
	}
	secondIn := model.Formation{
		Name: secondTestFormationName,
	}

	expected := &model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            Tnt,
	}
	secondFormation := model.Formation{
		ID:                  fixUUID(),
		Name:                secondTestFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            Tnt,
	}

	objectID := "123"
	applicationLblSingleFormation := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	applicationLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName, secondTestFormationName},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	applicationLblInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	applicationLblInputWithDefaultScenario := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{model.DefaultScenario},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	applicationLblWithDefaultScenario := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName, model.DefaultScenario},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	runtimeLblSingleFormation := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName, secondTestFormationName},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeCtxLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName, secondTestFormationName},
		ObjectID:   objectID,
		ObjectType: model.RuntimeContextLabelableObject,
		Version:    0,
	}
	runtimeLblInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeCtxLblInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   objectID,
		ObjectType: model.RuntimeContextLabelableObject,
		Version:    0,
	}

	asa := model.AutomaticScenarioAssignment{
		ScenarioName:   testFormationName,
		Tenant:         Tnt,
		TargetTenantID: objectID,
	}

	testCases := []struct {
		Name                          string
		UIDServiceFn                  func() *automock.UuidService
		LabelServiceFn                func() *automock.LabelService
		LabelRepoFn                   func() *automock.LabelRepository
		AsaServiceFN                  func() *automock.AutomaticFormationAssignmentService
		AsaRepoFN                     func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFN                 func() *automock.RuntimeRepository
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		ApplicationRepoFN             func() *automock.ApplicationRepository
		WebhookRepoFN                 func() *automock.WebhookRepository
		WebhookConverterFN            func() *automock.WebhookConverter
		WebhookClientFN               func() *automock.WebhookClient
		ApplicationTemplateRepoFN     func() *automock.ApplicationTemplateRepository
		ObjectType                    graphql.FormationObjectType
		InputFormation                model.Formation
		ExpectedFormation             *model.Formation
		ExpectedErrMessage            string
	}{
		{
			Name:         "success for application",
			UIDServiceFn: unusedUUIDService,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN:                     unusedASARepo,
			AsaServiceFN:                  unusedASAService,
			RuntimeRepoFN:                 unusedRuntimeRepo,
			RuntimeContextRepoFn:          unusedRuntimeContextRepo,
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(&model.Application{}, nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application if formation do not exist",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(&model.Application{}, nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application when formation is last",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLblSingleFormation, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(&model.Application{}, nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, Tnt, model.ApplicationLabelableObject, objectID, model.ScenariosKey).Return(nil)
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application when the formation is default",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInputWithDefaultScenario).Return(applicationLblWithDefaultScenario, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(&model.Application{}, nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     defaultFormation,
			ExpectedFormation:  &defaultFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, objectID))
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime when formation is coming from ASA",
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return([]*model.AutomaticScenarioAssignment{{
					ScenarioName:   ScenarioName,
					Tenant:         Tnt,
					TargetTenantID: Tnt,
				}}, nil)
				return asaRepo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLblSingleFormation, nil)
				return labelService
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, Tnt, objectID, runtimeLblFilters).Return(true, nil).Once()
				runtimeRepo.On("GetByID", ctx, Tnt, objectID).Return(nil, nil)
				return runtimeRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, Tnt, model.RuntimeLabelableObject, "123", model.ScenariosKey).Return(nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, objectID))
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ExistsByRuntimeID", ctx, Tnt, objectID).Return(false, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime if formation do not exist",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLblSingleFormation, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, Tnt).Return(&secondFormation, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, objectID))
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     secondIn,
			ExpectedFormation:  &secondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime when formation is last",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLblSingleFormation, nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, Tnt, model.RuntimeLabelableObject, objectID, model.ScenariosKey).Return(nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, objectID))
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for tenant",
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}

				asaRepo.On("DeleteForScenarioName", ctx, Tnt, testFormationName).Return(nil)

				return asaRepo
			},
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormationName).Return(asa, nil)
				return asaService
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, asa.TargetTenantID, runtimeLblFilters).Return(make([]*model.Runtime, 0), nil).Once()
				runtimeRepo.On("ListAll", ctx, asa.TargetTenantID, runtimeLblFilters).Return(make([]*model.Runtime, 0), nil).Once()
				return runtimeRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Twice()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for application while getting label",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(nil, testErr)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application while converting label values to string slice",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name: "error for application while converting label value to string",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name: "error for application when formation is last and delete fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLblSingleFormation, nil)
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, Tnt, model.ApplicationLabelableObject, objectID, model.ScenariosKey).Return(testErr)
				return labelRepo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when updating label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(testErr)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime when can't get formations that are coming from ASAs",
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, testErr)
				return asaRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime while getting label",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(nil, testErr)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime while converting label values to string slice",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name: "error for runtime while converting label value to string",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name: "error for runtime when formation is last and delete fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLblSingleFormation, nil)
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, Tnt, model.RuntimeLabelableObject, objectID, model.ScenariosKey).Return(testErr)
				return labelRepo
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime when updating label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(testErr)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant when delete fails",
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}

				asaRepo.On("DeleteForScenarioName", ctx, Tnt, testFormationName).Return(testErr)

				return asaRepo
			},
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormationName).Return(asa, nil)
				return asaService
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant when delete fails",
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormationName).Return(model.AutomaticScenarioAssignment{}, testErr)
				return asaService
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when fetching formation fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(nil, testErr).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "error when object type is unknown",
			ObjectType:         "UNKNOWN",
			InputFormation:     in,
			ExpectedErrMessage: "unknown formation type",
		},
		{
			Name: "success for application with notifications",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeID, RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookID, RuntimeID), fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeID), nil)
				repo.On("ToGraphQL", fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID), nil)
				return repo
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.UnassignFormation,
						FormationID: expected.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, nil)
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.UnassignFormation,
						FormationID: expected.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, nil)

				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID), fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{in.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{in.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for application when webhook client request fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID), nil)
				return repo
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.UnassignFormation,
						FormationID: FormationID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, testErr)

				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{in.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{in.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when webhook conversion fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{in.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{in.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtime context labels fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{in.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{in.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for application when fetching runtime contexts in scenario fails",
			UIDServiceFn: unusedUUIDService,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{in.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{in.Name}, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtimes in scenario fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{in.Name}, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching listening runtimes labels fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching listening runtimes fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching webhooks fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, testErr)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application template labels fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(nil, testErr)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application template fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application labels fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(nil, testErr)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, testErr)
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for runtime with notifications",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.UnassignFormation,
						FormationID: expected.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:           model.UnassignFormation,
						FormationID:         expected.ID,
						ApplicationTemplate: nil,
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModelWithoutTemplate(Application2ID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime if webhook client call fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.UnassignFormation,
						FormationID: expected.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, testErr).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if webhook conversion fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(nil, testErr)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application template labels fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application template fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application labels fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching applications fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching webhook fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, objectID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching runtime labels fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, objectID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching runtime fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for runtime context with notifications",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.UnassignFormation,
						FormationID: expected.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:           model.UnassignFormation,
						FormationID:         expected.ID,
						ApplicationTemplate: nil,
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModelWithoutTemplate(Application2ID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime context if webhook client call fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.UnassignFormation,
						FormationID: FormationID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, testErr).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if webhook conversion fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeContextRuntimeID)).Return(nil, testErr)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application template labels fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application templates fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application labels fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching applications fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expected.Name}).Return(nil, testErr)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching webhook fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime labels fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for runtime context if fetching runtime fails",
			UIDServiceFn: unusedUUIDService,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(nil, testErr)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(fixRuntimeContextLabels(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime context labels fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, objectID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime context fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(runtimeCtxLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeCtxLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   objectID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			uidService := unusedUUIDService()
			if testCase.UIDServiceFn != nil {
				uidService = testCase.UIDServiceFn()
			}
			labelService := unusedLabelService()
			if testCase.LabelServiceFn != nil {
				labelService = testCase.LabelServiceFn()
			}
			asaRepo := unusedASARepo()
			if testCase.AsaRepoFN != nil {
				asaRepo = testCase.AsaRepoFN()
			}
			asaService := unusedASAService()
			if testCase.AsaServiceFN != nil {
				asaService = testCase.AsaServiceFN()
			}
			runtimeRepo := unusedRuntimeRepo()
			if testCase.RuntimeRepoFN != nil {
				runtimeRepo = testCase.RuntimeRepoFN()
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
			applicationRepo := unusedApplicationRepo()
			if testCase.ApplicationRepoFN != nil {
				applicationRepo = testCase.ApplicationRepoFN()
			}
			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFN != nil {
				webhookRepo = testCase.WebhookRepoFN()
			}
			webhookConverter := unusedWebhookConverter()
			if testCase.WebhookConverterFN != nil {
				webhookConverter = testCase.WebhookConverterFN()
			}
			webhookClient := unusedWebhookClient()
			if testCase.WebhookClientFN != nil {
				webhookClient = testCase.WebhookClientFN()
			}
			appTemplateRepo := unusedAppTemplateRepository()
			if testCase.ApplicationTemplateRepoFN != nil {
				appTemplateRepo = testCase.ApplicationTemplateRepoFN()
			}
			labelRepo := unusedLabelRepo()
			if testCase.LabelRepoFn != nil {
				labelRepo = testCase.LabelRepoFn()
			}

			svc := formation.NewService(nil, labelRepo, formationRepo, formationTemplateRepo, labelService, uidService, nil, asaRepo, asaService, nil, runtimeRepo, runtimeContextRepo, webhookRepo, webhookClient, applicationRepo, appTemplateRepo, webhookConverter)

			// WHEN
			actual, err := svc.UnassignFormation(ctx, Tnt, objectID, testCase.ObjectType, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}
			mock.AssertExpectationsForObjects(t, uidService, labelService, asaRepo, asaService, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, applicationRepo, webhookRepo, webhookConverter, webhookClient, appTemplateRepo)
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

	testCases := []struct {
		Name                          string
		LabelDefServiceFn             func() *automock.LabelDefService
		AsaRepoFn                     func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFN                 func() *automock.RuntimeRepository
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		WebhookRepoFN                 func() *automock.WebhookRepository
		LabelRepoFN                   func() *automock.LabelRepository
		LabelServiceFn                func() *automock.LabelService
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[1].ID).Return(nil, nil)
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

				return runtimeContextRepo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[1].ID).Return(nil, nil)
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[0].ID).Return(nil, nil)
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[1].ID).Return(nil, nil)
				return labelRepo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[1].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[1].ID))
				return webhookRepo
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
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
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(3)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[0]).Return(expectedRtmCtxLabels[0], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[0].ID, runtimeCtxLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[1]).Return(expectedRtmCtxLabels[1], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[1].ID, runtimeCtxLblInputs[1]).Return(testErr)
				return svc
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[0].ID).Return(nil, nil)
				return labelRepo
			},
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(nil, testErr).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(2)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)
				return svc
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				return labelRepo
			},
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(2)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)
				return svc
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				return labelRepo
			},
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
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(nil, testErr)
				return svc
			},
			WebhookRepoFN:      unusedWebhookRepository,
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
			WebhookRepoFN:      unusedWebhookRepository,
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
			WebhookRepoFN:      unusedWebhookRepository,
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
			WebhookRepoFN:      unusedWebhookRepository,
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
			WebhookRepoFN:                 unusedWebhookRepository,
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
			WebhookRepoFN:                 unusedWebhookRepository,
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
			WebhookRepoFN:                 unusedWebhookRepository,
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
			WebhookRepoFN:                 unusedWebhookRepository,
			LabelRepoFN:                   unusedLabelRepo,
			InputASA:                      fixModel(ScenarioName),
			ExpectedASA:                   model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:            "while persisting Assignment: some error",
		},
		{
			Name: "returns error on ensuring that scenarios label definition exist",
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}
				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, mock.Anything).Return(fixError()).Once()
				return labelDefSvc
			},
			AsaRepoFn:                     unusedASARepo,
			RuntimeRepoFN:                 unusedRuntimeRepo,
			RuntimeContextRepoFn:          unusedRuntimeContextRepo,
			LabelServiceFn:                unusedLabelService,
			FormationRepositoryFn:         unusedFormationRepo,
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			WebhookRepoFN:                 unusedWebhookRepository,
			LabelRepoFN:                   unusedLabelRepo,
			InputASA:                      fixModel(ScenarioName),
			ExpectedASA:                   model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:            "while ensuring that `scenarios` label definition exist: some error",
		},
		{
			Name: "returns error on getting available scenarios from label definition",
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}
				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, mock.Anything).Return(nil).Once()
				labelDefSvc.On("GetAvailableScenarios", mock.Anything, tenantID.String()).Return(nil, fixError()).Once()
				return labelDefSvc
			},
			AsaRepoFn:                     unusedASARepo,
			RuntimeRepoFN:                 unusedRuntimeRepo,
			RuntimeContextRepoFn:          unusedRuntimeContextRepo,
			LabelServiceFn:                unusedLabelService,
			FormationRepositoryFn:         unusedFormationRepo,
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			WebhookRepoFN:                 unusedWebhookRepository,
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
			webhookRepo := testCase.WebhookRepoFN()
			labelRepo := testCase.LabelRepoFN()

			svc := formation.NewService(nil, labelRepo, formationRepo, formationTemplateRepo, lblService, nil, labelDefService, asaRepo, nil, tenantSvc, runtimeRepo, runtimeContextRepo, webhookRepo, nil, nil, nil, nil)

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

			mock.AssertExpectationsForObjects(t, tenantSvc, asaRepo, labelDefService, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, lblService, webhookRepo, labelRepo)
		})
	}

	t.Run("returns error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

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

		runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(make([]*model.Runtime, 0), nil).Twice()

		formationRepo := &automock.FormationRepository{}
		formationRepo.On("GetByName", ctx, scenarioNameA, "").Return(formations[0], nil).Once()
		formationRepo.On("GetByName", ctx, scenarioNameB, "").Return(formations[1], nil).Once()

		formationTemplateRepo := &automock.FormationTemplateRepository{}
		formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
		formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()

		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, formationRepo, formationTemplateRepo)

		svc := formation.NewService(nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

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

		svc := formation.NewService(nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("return error when input slice is empty", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

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

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignments: %s", ErrMsg))
	})

	t.Run("returns error when empty tenant", func(t *testing.T) {
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		err := svc.DeleteManyASAForSameTargetTenant(context.TODO(), models)
		require.EqualError(t, err, "cannot read tenant from context")
	})
}

func TestService_DeleteAutomaticScenarioAssignment(t *testing.T) {
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
		WebhookRepoFN                 func() *automock.WebhookRepository
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[1].ID).Return(nil, nil)

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

				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[1].ID).Return(nil, nil)
				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[0].ID).Return(nil, nil)
				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[1].ID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[1].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
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
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(3)
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

				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[0].ID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(nil, testErr).Once()
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
				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
			},
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()
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
				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
			},
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
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
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
			WebhookRepoFN:      unusedWebhookRepository,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
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
			LabelServiceFn:     unusedLabelService,
			LabelRepositoryFn:  unusedLabelRepo,
			WebhookRepoFN:      unusedWebhookRepository,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
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
			LabelServiceFn:     unusedLabelService,
			LabelRepositoryFn:  unusedLabelRepo,
			WebhookRepoFN:      unusedWebhookRepository,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
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
			LabelServiceFn:     unusedLabelService,
			LabelRepositoryFn:  unusedLabelRepo,
			WebhookRepoFN:      unusedWebhookRepository,
			InputASA:           fixModel(testFormationName),
			ExpectedASA:        model.AutomaticScenarioAssignment{},
			ExpectedErrMessage: testErr.Error(),
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
			WebhookRepoFN:                 unusedWebhookRepository,
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
			WebhookRepoFN:                 unusedWebhookRepository,
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
			webhookRepo := testCase.WebhookRepoFN()

			svc := formation.NewService(nil, lblRepo, formationRepo, formationTemplateRepo, lblService, nil, labelDefService, asaRepo, nil, tenantSvc, runtimeRepo, runtimeContextRepo, webhookRepo, nil, nil, nil, nil)

			// WHEN
			err := svc.DeleteAutomaticScenarioAssignment(ctx, testCase.InputASA)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, tenantSvc, asaRepo, labelDefService, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, lblService, lblRepo, webhookRepo)
		})
	}

	t.Run("returns error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

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
		WebhookRepositoryFn           func() *automock.WebhookRepository
		InputASA                      model.AutomaticScenarioAssignment
		ExpectedErrMessage            string
	}{
		{
			Name: "Success",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[1].ID).Return(nil, nil)
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
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[0]).Return(expectedRtmCtxLabels[0], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[0].ID, runtimeCtxLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[1]).Return(expectedRtmCtxLabels[1], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[1].ID, runtimeCtxLblInputs[1]).Return(nil)
				return svc
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), rtmIDs[0], model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), rtmIDs[1], model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, rtmIDs[0]).Return(nil, nil)
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, rtmIDs[1]).Return(nil, nil)
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[0].ID).Return(nil, nil)
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[1].ID).Return(nil, nil)
				return labelRepo
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when assigning runtime context to formation fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()

				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
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
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(3)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[0]).Return(expectedRtmCtxLabels[0], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[0].ID, runtimeCtxLblInputs[0]).Return(nil)

				svc.On("GetLabel", ctx, tnt, runtimeCtxLblInputs[1]).Return(expectedRtmCtxLabels[1], nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRtmCtxLabels[1].ID, runtimeCtxLblInputs[1]).Return(testErr)
				return svc
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), rtmIDs[0], model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, rtmIDs[0]).Return(nil, nil)
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[0].ID).Return(nil, nil)
				return labelRepo
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing runtime contexts for runtime fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()

				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(nil, testErr).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(2)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)
				return svc
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), rtmIDs[0], model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, rtmIDs[0]).Return(nil, nil)
				return labelRepo
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing all runtimes fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()

				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(2)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(expectedRuntimeLabel, nil)
				svc.On("UpdateLabel", ctx, tnt, expectedRuntimeLabel.ID, runtimeLblInputs[0]).Return(nil)
				return svc
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), rtmIDs[0], model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, rtmIDs[0]).Return(nil, nil)
				return labelRepo
			},
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
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetLabel", ctx, tnt, runtimeLblInputs[0]).Return(nil, testErr)
				return svc
			},
			LabelRepositoryFn:   unusedLabelRepo,
			WebhookRepositoryFn: unusedWebhookRepository,
			InputASA:            fixModel(testFormationName),
			ExpectedErrMessage:  testErr.Error(),
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
			LabelRepositoryFn:   unusedLabelRepo,
			WebhookRepositoryFn: unusedWebhookRepository,
			LabelServiceFn:      unusedLabelService,
			InputASA:            fixModel(testFormationName),
			ExpectedErrMessage:  testErr.Error(),
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
			LabelRepositoryFn:   unusedLabelRepo,
			WebhookRepositoryFn: unusedWebhookRepository,
			LabelServiceFn:      unusedLabelService,
			InputASA:            fixModel(testFormationName),
			ExpectedErrMessage:  testErr.Error(),
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
			LabelRepositoryFn:   unusedLabelRepo,
			WebhookRepositoryFn: unusedWebhookRepository,
			LabelServiceFn:      unusedLabelService,
			InputASA:            fixModel(testFormationName),
			ExpectedErrMessage:  testErr.Error(),
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
			WebhookRepositoryFn:           unusedWebhookRepository,
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
			webhookRepo := testCase.WebhookRepositoryFn()

			svc := formation.NewService(nil, labelRepo, formationRepo, formationTemplateRepo, lblService, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo, webhookRepo, nil, nil, nil, nil)

			// WHEN
			err := svc.EnsureScenarioAssigned(ctx, testCase.InputASA)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, lblService, labelRepo, webhookRepo)
		})
	}
}

func TestService_RemoveAssignedScenario(t *testing.T) {
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
		WebhookRepositoryFn           func() *automock.WebhookRepository
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()

				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[1].ID).Return(nil, nil)
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

				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[1].ID).Return(nil, nil)
				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[0].ID).Return(nil, nil)
				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[1].ID).Return(nil, nil)
				return repo
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[1].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()

				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
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
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(3)
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

				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[0].ID).Return(nil, nil)
				return repo
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()

				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(nil, testErr).Once()
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

				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				return repo
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
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
				runtimeRepo.On("ListAll", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()
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

				repo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
				return repo
			},
			WebhookRepositoryFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
				return webhookRepo
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
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
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
			WebhookRepositoryFn: unusedWebhookRepository,
			InputASA:            fixModel(testFormationName),
			ExpectedErrMessage:  testErr.Error(),
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
			WebhookRepositoryFn: unusedWebhookRepository,
			LabelServiceFn:      unusedLabelService,
			LabelRepositoryFn:   unusedLabelRepo,
			InputASA:            fixModel(testFormationName),
			ExpectedErrMessage:  testErr.Error(),
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
			WebhookRepositoryFn: unusedWebhookRepository,
			LabelServiceFn:      unusedLabelService,
			LabelRepositoryFn:   unusedLabelRepo,
			InputASA:            fixModel(testFormationName),
			ExpectedErrMessage:  testErr.Error(),
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
			LabelServiceFn:      unusedLabelService,
			LabelRepositoryFn:   unusedLabelRepo,
			WebhookRepositoryFn: unusedWebhookRepository,
			InputASA:            fixModel(testFormationName),
			ExpectedErrMessage:  testErr.Error(),
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
			WebhookRepositoryFn:           unusedWebhookRepository,
			LabelServiceFn:                unusedLabelService,
			LabelRepositoryFn:             unusedLabelRepo,
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
			webhookRepo := testCase.WebhookRepositoryFn()

			svc := formation.NewService(nil, lblRepo, formationRepo, formationTemplateRepo, lblService, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, webhookRepo, nil, nil, nil, nil)

			// WHEN
			err := svc.RemoveAssignedScenario(ctx, testCase.InputASA)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, lblService, lblRepo, asaRepo, webhookRepo)
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

			svc := formation.NewService(nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

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
	testErr := errors.New(ErrMsg)
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
				runtimeContextRepo.On("Exists", ctx, testScenarios[0].TargetTenantID, RuntimeContextID).Return(true, nil).Once()

				runtimeContextRepo.On("Exists", ctx, testScenarios[1].TargetTenantID, RuntimeContextID).Return(false, nil).Once()
				runtimeContextRepo.On("Exists", ctx, testScenarios[1].TargetTenantID, RuntimeContextID).Return(false, nil).Once()
				runtimeContextRepo.On("Exists", ctx, testScenarios[1].TargetTenantID, RuntimeContextID).Return(false, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListAll", ctx, testScenarios[0].TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				runtimeRepo.On("ListAll", ctx, testScenarios[1].TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
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
			Name: "Returns error for runtime context when checking if runtime exists for a runtime ctx",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("Exists", ctx, testScenarios[0].TargetTenantID, RuntimeContextID).Return(false, testErr).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListAll", ctx, testScenarios[0].TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
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
			ExpectedError:     testErr,
			ExpectedScenarios: nil,
		},
		{
			Name: "Returns error for runtime context when listing all runtimes fails",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListAll", ctx, testScenarios[0].TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
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
			ExpectedError:     testErr,
			ExpectedScenarios: nil,
		},
		{
			Name: "Returns error for runtime context getting formation template runtime type fails",
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
			ObjectID:          RuntimeContextID,
			ObjectType:        graphql.FormationObjectTypeRuntimeContext,
			ExpectedError:     testErr,
			ExpectedScenarios: nil,
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

			svc := formation.NewService(nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

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

		svc := formation.NewService(nil, nil, nil, nil, labelService, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

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

		svc := formation.NewService(nil, nil, nil, nil, labelService, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		formations, err := svc.GetFormationsForObject(ctx, tenantID.String(), model.RuntimeLabelableObject, id)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "while fetching scenario label for")
		require.Nil(t, formations)
		mock.AssertExpectationsForObjects(t, labelService)
	})
}
