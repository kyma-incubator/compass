package formation_test

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/pkg/errors"

	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

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
			FormationRepoFn:       unusedFormationRepo,
			InputID:               FormationID,
			InputPageSize:         300,
			ExpectedFormationPage: nil,
			ExpectedErrMessage:    "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			formationRepo := testCase.FormationRepoFn()

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

	asa := model.AutomaticScenarioAssignment{
		ScenarioName:   testFormationName,
		Tenant:         Tnt,
		TargetTenantID: TargetTenant,
	}

	testCases := []struct {
		Name                      string
		UIDServiceFn              func() *automock.UuidService
		LabelServiceFn            func() *automock.LabelService
		LabelDefServiceFn         func() *automock.LabelDefService
		TenantServiceFn           func() *automock.TenantService
		AsaRepoFn                 func() *automock.AutomaticFormationAssignmentRepository
		AsaServiceFN              func() *automock.AutomaticFormationAssignmentService
		RuntimeRepoFN             func() *automock.RuntimeRepository
		ApplicationRepoFN         func() *automock.ApplicationRepository
		WebhookRepoFN             func() *automock.WebhookRepository
		WebhookConverterFN        func() *automock.WebhookConverter
		WebhookClientFN           func() *automock.WebhookClient
		ApplicationTemplateRepoFN func() *automock.ApplicationTemplateRepository
		LabelRepoFN               func() *automock.LabelRepository
		RuntimeContextRepoFn      func() *automock.RuntimeContextRepository
		FormationRepositoryFn     func() *automock.FormationRepository
		ObjectType                graphql.FormationObjectType
		InputFormation            model.Formation
		ExpectedFormation         *model.Formation
		ExpectedErrMessage        string
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
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
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            inputFormation,
			ExpectedFormation:         expectedFormation,
			ExpectedErrMessage:        "",
		},
		{
			Name:         "success for application if formation is already added",
			UIDServiceFn: unusedUUIDService,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(applicationLbl, nil)
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
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            inputFormation,
			ExpectedFormation:         expectedFormation,
			ExpectedErrMessage:        "",
		},
		{
			Name:         "success for application with new formation",
			UIDServiceFn: unusedUUIDService,
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
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
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            inputSecondFormation,
			ExpectedFormation:         expectedSecondFormation,
			ExpectedErrMessage:        "",
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
			FormationRepositoryFn: unusedFormationRepo,
			LabelDefServiceFn:     unusedLabelDefServiceFn,
			AsaRepoFn:             unusedASARepo,
			AsaServiceFN:          unusedASAService,
			WebhookConverterFN:    unusedWebhookConverter,
			WebhookClientFN:       unusedWebhookClient,
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
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            defaultFormation,
			ExpectedFormation:         &defaultFormation,
			ExpectedErrMessage:        "",
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
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
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
			Name:         "success for runtime if formation is already added",
			UIDServiceFn: unusedUUIDService,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
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
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            inputFormation,
			ExpectedFormation:         expectedFormation,
			ExpectedErrMessage:        "",
		},
		{
			Name:         "success for runtime with new formation",
			UIDServiceFn: unusedUUIDService,
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
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
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
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       inputSecondFormation,
			ExpectedFormation:    expectedSecondFormation,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for tenant",
			UIDServiceFn: unusedUUIDService,
			LabelRepoFN:  unusedLabelRepo,
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
			LabelServiceFn: unusedLabelService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctx, asa).Return(nil)

				return asaRepo
			},
			AsaServiceFN:              unusedASAService,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenant, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ListAll", ctx, TargetTenant).Return(make([]*model.RuntimeContext, 0), nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name:        "error for application when label does not exist and can't create it",
			LabelRepoFN: unusedLabelRepo,
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
			FormationRepositoryFn:     unusedFormationRepo,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error for application while getting label",
			UIDServiceFn: unusedUUIDService,
			LabelRepoFN:  unusedLabelRepo,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, testErr)
				return labelService
			},
			FormationRepositoryFn:     unusedFormationRepo,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			AsaServiceFN:              unusedASAService,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error for application while converting label values to string slice",
			UIDServiceFn: unusedUUIDService,
			LabelRepoFN:  unusedLabelRepo,
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
			FormationRepositoryFn:     unusedFormationRepo,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for application while converting label value to string",
			UIDServiceFn: unusedUUIDService,
			LabelRepoFN:  unusedLabelRepo,
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
			FormationRepositoryFn:     unusedFormationRepo,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        "cannot cast label value as a string",
		},
		{
			Name:         "error for application when updating label fails",
			UIDServiceFn: unusedUUIDService,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &applicationLblInput).Return(testErr)
				return labelService
			},
			LabelRepoFN:               unusedLabelRepo,
			FormationRepositoryFn:     unusedFormationRepo,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name: "error for runtime when label does not exist and can't create it",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelRepoFN: unusedLabelRepo,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(testErr)
				return labelService
			},
			FormationRepositoryFn:     unusedFormationRepo,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error for runtime while getting label",
			UIDServiceFn: unusedUUIDService,
			LabelRepoFN:  unusedLabelRepo,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, testErr)
				return labelService
			},
			FormationRepositoryFn:     unusedFormationRepo,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error for runtime while converting label values to string slice",
			UIDServiceFn: unusedUUIDService,
			LabelRepoFN:  unusedLabelRepo,
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
			FormationRepositoryFn:     unusedFormationRepo,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for runtime while converting label value to string",
			UIDServiceFn: unusedUUIDService,
			LabelRepoFN:  unusedLabelRepo,
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
			FormationRepositoryFn:     unusedFormationRepo,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        "cannot cast label value as a string",
		},
		{
			Name:         "error for runtime when updating label fails",
			UIDServiceFn: unusedUUIDService,
			LabelRepoFN:  unusedLabelRepo,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &runtimeLblInput).Return(testErr)
				return labelService
			},
			FormationRepositoryFn:     unusedFormationRepo,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error for tenant when tenant conversion fails",
			UIDServiceFn: unusedUUIDService,
			LabelRepoFN:  unusedLabelRepo,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, objectID).Return("", testErr)
				return svc
			},
			FormationRepositoryFn:     unusedFormationRepo,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			LabelServiceFn:            unusedLabelService,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeTenant,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error for tenant when create fails",
			UIDServiceFn: unusedUUIDService,
			LabelRepoFN:  unusedLabelRepo,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, objectID).Return(TargetTenant, nil)
				return svc
			},
			LabelServiceFn: unusedLabelService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctx, model.AutomaticScenarioAssignment{ScenarioName: testFormationName, Tenant: Tnt, TargetTenantID: TargetTenant}).Return(testErr)

				return asaRepo
			},
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: unusedApplicationRepo,
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", ctx, Tnt).Return(nil)
				labelDefSvc.On("GetAvailableScenarios", ctx, Tnt).Return([]string{testFormationName}, nil)

				return labelDefSvc
			},
			FormationRepositoryFn:     unusedFormationRepo,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeTenant,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error when can't get formation by name",
			UIDServiceFn: unusedUUIDService,
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
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			LabelRepoFN:               unusedLabelRepo,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            inputSecondFormation,
			ExpectedFormation:         expectedSecondFormation,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:                      "error when object type is unknown",
			FormationRepositoryFn:     unusedFormationRepo,
			LabelRepoFN:               unusedLabelRepo,
			UIDServiceFn:              unusedUUIDService,
			LabelServiceFn:            unusedLabelService,
			LabelDefServiceFn:         unusedLabelDefServiceFn,
			AsaRepoFn:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                "UNKNOWN",
			InputFormation:            inputFormation,
			ExpectedErrMessage:        "unknown formation type",
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
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
							Application: fixApplicationModel(),
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
							Application: fixApplicationModel(),
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
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
							Application: fixApplicationModel(),
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
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
			WebhookClientFN: unusedWebhookClient,
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
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
			WebhookConverterFN: unusedWebhookConverter,
			WebhookClientFN:    unusedWebhookClient,
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
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
			WebhookConverterFN: unusedWebhookConverter,
			WebhookClientFN:    unusedWebhookClient,
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
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
			WebhookConverterFN: unusedWebhookConverter,
			WebhookClientFN:    unusedWebhookClient,
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
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
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
			WebhookConverterFN: unusedWebhookConverter,
			WebhookClientFN:    unusedWebhookClient,
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
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
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
			WebhookConverterFN: unusedWebhookConverter,
			WebhookClientFN:    unusedWebhookClient,
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
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
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
			WebhookConverterFN: unusedWebhookConverter,
			WebhookClientFN:    unusedWebhookClient,
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN:      unusedWebhookRepository,
			WebhookConverterFN: unusedWebhookConverter,
			WebhookClientFN:    unusedWebhookClient,
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(fixApplicationLabels(), nil)
				return repo
			},
			WebhookRepoFN:      unusedWebhookRepository,
			WebhookConverterFN: unusedWebhookConverter,
			WebhookClientFN:    unusedWebhookClient,
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(fixApplicationModel(), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, objectID).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        testErr.Error(),
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
			LabelDefServiceFn: unusedLabelDefServiceFn,
			AsaRepoFn:         unusedASARepo,
			AsaServiceFN:      unusedASAService,
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, testErr)
				return repo
			},
			LabelRepoFN:               unusedLabelRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            inputFormation,
			ExpectedErrMessage:        testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			uidService := testCase.UIDServiceFn()
			labelService := testCase.LabelServiceFn()
			asaRepo := testCase.AsaRepoFn()
			asaService := testCase.AsaServiceFN()
			tenantSvc := &automock.TenantService{}
			labelDefService := testCase.LabelDefServiceFn()
			runtimeRepo := testCase.RuntimeRepoFN()
			runtimeContextRepo := testCase.RuntimeContextRepoFn()
			formationRepo := testCase.FormationRepositoryFn()
			applicationRepo := testCase.ApplicationRepoFN()
			webhookRepo := testCase.WebhookRepoFN()
			webhookConverter := testCase.WebhookConverterFN()
			webhookClient := testCase.WebhookClientFN()
			appTemplateRepo := testCase.ApplicationTemplateRepoFN()
			labelRepo := testCase.LabelRepoFN()

			if testCase.TenantServiceFn != nil {
				tenantSvc = testCase.TenantServiceFn()
			}

			svc := formation.NewService(nil, labelRepo, formationRepo, nil, labelService, uidService, labelDefService, asaRepo, asaService, tenantSvc, runtimeRepo, runtimeContextRepo, webhookRepo, webhookClient, applicationRepo, appTemplateRepo, webhookConverter)

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

			mock.AssertExpectationsForObjects(t, uidService, labelService, asaService, tenantSvc, asaRepo, labelDefService, runtimeRepo, runtimeContextRepo, formationRepo, applicationRepo, webhookRepo, webhookConverter, webhookClient, appTemplateRepo, labelRepo)
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
	runtimeLblInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}

	asa := model.AutomaticScenarioAssignment{
		ScenarioName:   testFormationName,
		Tenant:         Tnt,
		TargetTenantID: objectID,
	}

	testCases := []struct {
		Name                      string
		UIDServiceFn              func() *automock.UuidService
		LabelServiceFn            func() *automock.LabelService
		LabelRepoFn               func() *automock.LabelRepository
		AsaServiceFN              func() *automock.AutomaticFormationAssignmentService
		AsaRepoFN                 func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFN             func() *automock.RuntimeRepository
		RuntimeContextRepoFn      func() *automock.RuntimeContextRepository
		FormationRepositoryFn     func() *automock.FormationRepository
		ApplicationRepoFN         func() *automock.ApplicationRepository
		WebhookRepoFN             func() *automock.WebhookRepository
		WebhookConverterFN        func() *automock.WebhookConverter
		WebhookClientFN           func() *automock.WebhookClient
		ApplicationTemplateRepoFN func() *automock.ApplicationTemplateRepository
		ObjectType                graphql.FormationObjectType
		InputFormation            model.Formation
		ExpectedFormation         *model.Formation
		ExpectedErrMessage        string
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
			AsaRepoFN:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
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
			Name:         "success for application if formation do not exist",
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
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			AsaRepoFN:            unusedASARepo,
			AsaServiceFN:         unusedASAService,
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
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
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            in,
			ExpectedFormation:         expected,
			ExpectedErrMessage:        "",
		},
		{
			Name:         "success for application when formation is last",
			UIDServiceFn: unusedUUIDService,
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
			AsaRepoFN:            unusedASARepo,
			AsaServiceFN:         unusedASAService,
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
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
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            in,
			ExpectedFormation:         expected,
			ExpectedErrMessage:        "",
		},
		{
			Name:         "success for application when the formation is default",
			UIDServiceFn: unusedUUIDService,
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
			AsaRepoFN:            unusedASARepo,
			AsaServiceFN:         unusedASAService,
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
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
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			FormationRepositoryFn:     unusedFormationRepo,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            defaultFormation,
			ExpectedFormation:         &defaultFormation,
			ExpectedErrMessage:        "",
		},
		{
			Name:         "success for runtime",
			UIDServiceFn: unusedUUIDService,
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
			AsaServiceFN:              unusedASAService,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            in,
			ExpectedFormation:         expected,
			ExpectedErrMessage:        "",
		},
		{
			Name:         "success for runtime when formation is coming from ASA",
			UIDServiceFn: unusedUUIDService,
			AsaServiceFN: unusedASAService,
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
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, Tnt, "123").Return(true, nil)
				repo.On("GetByID", ctx, Tnt, objectID).Return(nil, nil)
				return repo
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
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookClientFN:           unusedWebhookClient,
			WebhookConverterFN:        unusedWebhookConverter,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            in,
			ExpectedFormation:         expected,
			ExpectedErrMessage:        "",
		},
		{
			Name:         "success for runtime if formation do not exist",
			UIDServiceFn: unusedUUIDService,
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
			AsaServiceFN:         unusedASAService,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
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
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            secondIn,
			ExpectedFormation:         &secondFormation,
			ExpectedErrMessage:        "",
		},
		{
			Name:         "success for runtime when formation is last",
			UIDServiceFn: unusedUUIDService,
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
			AsaServiceFN:              unusedASAService,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            in,
			ExpectedFormation:         expected,
			ExpectedErrMessage:        "",
		},
		{
			Name:                      "success for tenant",
			UIDServiceFn:              unusedUUIDService,
			LabelServiceFn:            unusedLabelService,
			LabelRepoFn:               unusedLabelRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
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
				runtimeRepo.On("ListOwnedRuntimes", ctx, "123", []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil)

				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ListAll", ctx, "123").Return(make([]*model.RuntimeContext, 0), nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name:         "error for application while getting label",
			UIDServiceFn: unusedUUIDService,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(nil, testErr)
				return labelService
			},
			FormationRepositoryFn:     unusedFormationRepo,
			LabelRepoFn:               unusedLabelRepo,
			AsaRepoFN:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            in,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error for application while converting label values to string slice",
			UIDServiceFn: unusedUUIDService,
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
			FormationRepositoryFn:     unusedFormationRepo,
			LabelRepoFn:               unusedLabelRepo,
			AsaRepoFN:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            in,
			ExpectedErrMessage:        "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for application while converting label value to string",
			UIDServiceFn: unusedUUIDService,
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
			FormationRepositoryFn:     unusedFormationRepo,
			LabelRepoFn:               unusedLabelRepo,
			AsaRepoFN:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            in,
			ExpectedErrMessage:        "cannot cast label value as a string",
		},
		{
			Name:         "error for application when formation is last and delete fails",
			UIDServiceFn: unusedUUIDService,
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
			FormationRepositoryFn:     unusedFormationRepo,
			AsaRepoFN:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            in,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error for application when updating label fails",
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
				}).Return(testErr)
				return labelService
			},
			FormationRepositoryFn:     unusedFormationRepo,
			LabelRepoFn:               unusedLabelRepo,
			AsaRepoFN:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeApplication,
			InputFormation:            in,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:           "error for runtime when can't get formations that are coming from ASAs",
			UIDServiceFn:   unusedUUIDService,
			LabelServiceFn: unusedLabelService,
			LabelRepoFn:    unusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, testErr)
				return asaRepo
			},
			FormationRepositoryFn:     unusedFormationRepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            in,
			ExpectedFormation:         expected,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error for runtime while getting label",
			UIDServiceFn: unusedUUIDService,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(nil, testErr)
				return labelService
			},
			LabelRepoFn: unusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn:     unusedFormationRepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            in,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error for runtime while converting label values to string slice",
			UIDServiceFn: unusedUUIDService,
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
			LabelRepoFn: unusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn:     unusedFormationRepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            in,
			ExpectedErrMessage:        "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for runtime while converting label value to string",
			UIDServiceFn: unusedUUIDService,
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
			FormationRepositoryFn: unusedFormationRepo,
			LabelRepoFn:           unusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            in,
			ExpectedErrMessage:        "cannot cast label value as a string",
		},
		{
			Name:         "error for runtime when formation is last and delete fails",
			UIDServiceFn: unusedUUIDService,
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
			FormationRepositoryFn:     unusedFormationRepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            in,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:         "error for runtime when updating label fails",
			UIDServiceFn: unusedUUIDService,
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
			LabelRepoFn: unusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, Tnt).Return(nil, nil)
				return asaRepo
			},
			FormationRepositoryFn:     unusedFormationRepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeRuntime,
			InputFormation:            in,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:           "error for tenant when delete fails",
			UIDServiceFn:   unusedUUIDService,
			LabelServiceFn: unusedLabelService,
			LabelRepoFn:    unusedLabelRepo,
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
			FormationRepositoryFn:     unusedFormationRepo,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                graphql.FormationObjectTypeTenant,
			InputFormation:            in,
			ExpectedErrMessage:        testErr.Error(),
		},
		{
			Name:                      "error for tenant when delete fails",
			UIDServiceFn:              unusedUUIDService,
			LabelServiceFn:            unusedLabelService,
			LabelRepoFn:               unusedLabelRepo,
			AsaRepoFN:                 unusedASARepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormationName).Return(model.AutomaticScenarioAssignment{}, testErr)
				return asaService
			},
			FormationRepositoryFn: unusedFormationRepo,
			ObjectType:            graphql.FormationObjectTypeTenant,
			RuntimeRepoFN:         unusedRuntimeRepo,
			RuntimeContextRepoFn:  unusedRuntimeContextRepo,
			InputFormation:        in,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name:         "success for runtime",
			UIDServiceFn: unusedUUIDService,
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
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			LabelRepoFn:               unusedLabelRepo,
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
			AsaServiceFN:         unusedASAService,
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       in,
			ExpectedFormation:    expected,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:                      "error when object type is unknown",
			UIDServiceFn:              unusedUUIDService,
			LabelServiceFn:            unusedLabelService,
			LabelRepoFn:               unusedLabelRepo,
			AsaRepoFN:                 unusedASARepo,
			AsaServiceFN:              unusedASAService,
			RuntimeRepoFN:             unusedRuntimeRepo,
			RuntimeContextRepoFn:      unusedRuntimeContextRepo,
			FormationRepositoryFn:     unusedFormationRepo,
			ApplicationRepoFN:         unusedApplicationRepo,
			WebhookRepoFN:             unusedWebhookRepository,
			WebhookConverterFN:        unusedWebhookConverter,
			WebhookClientFN:           unusedWebhookClient,
			ApplicationTemplateRepoFN: unusedAppTemplateRepository,
			ObjectType:                "UNKNOWN",
			InputFormation:            in,
			ExpectedErrMessage:        "unknown formation type",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			uidService := testCase.UIDServiceFn()
			labelService := testCase.LabelServiceFn()
			labelRepo := testCase.LabelRepoFn()
			asaRepo := testCase.AsaRepoFN()
			asaService := testCase.AsaServiceFN()
			runtimeRepo := testCase.RuntimeRepoFN()
			runtimeContextRepo := testCase.RuntimeContextRepoFn()
			formationRepo := testCase.FormationRepositoryFn()
			applicationRepo := testCase.ApplicationRepoFN()
			webhookRepo := testCase.WebhookRepoFN()
			webhookConverter := testCase.WebhookConverterFN()
			webhookClient := testCase.WebhookClientFN()
			appTemplateRepo := testCase.ApplicationTemplateRepoFN()

			svc := formation.NewService(nil, labelRepo, formationRepo, nil, labelService, uidService, nil, asaRepo, asaService, nil, runtimeRepo, runtimeContextRepo, webhookRepo, webhookClient, applicationRepo, appTemplateRepo, webhookConverter)

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
			mock.AssertExpectationsForObjects(t, uidService, labelService, asaRepo, asaService, runtimeRepo, runtimeContextRepo, formationRepo)
		})
	}
}

func TestService_CreateAutomaticScenarioAssignment(t *testing.T) {
	ctx := fixCtxWithTenant()
	testCases := []struct {
		Name                 string
		LabelDefServiceFn    func() *automock.LabelDefService
		AsaRepoFn            func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFN        func() *automock.RuntimeRepository
		RuntimeContextRepoFn func() *automock.RuntimeContextRepository
		InputASA             model.AutomaticScenarioAssignment
		ExpectedASA          model.AutomaticScenarioAssignment
		ExpectedErrMessage   string
	}{
		{
			Name: "happy path",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{ScenarioName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel()).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(make([]*model.RuntimeContext, 0), nil).Once()
				return runtimeContextRepo
			},
			InputASA:           fixModel(),
			ExpectedASA:        fixModel(),
			ExpectedErrMessage: "",
		},
		{
			Name: "return error when ensuring scenarios for runtimes fails",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{ScenarioName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, fixModel()).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, fixError()).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			InputASA:             fixModel(),
			ExpectedASA:          model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:   fixError().Error(),
		},
		{
			Name: "returns error when scenario already has an assignment",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{ScenarioName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", mock.Anything, fixModel()).Return(apperrors.NewNotUniqueError("")).Once()
				return mockRepo
			},
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			InputASA:             fixModel(),
			ExpectedASA:          model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:   "a given scenario already has an assignment",
		},
		{
			Name: "returns error when given scenario does not exist",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{"completely-different-scenario"})
			},
			AsaRepoFn:            unusedASARepo,
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			InputASA:             fixModel(),
			ExpectedASA:          model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:   apperrors.NewNotFoundError(resource.AutomaticScenarioAssigment, fixModel().ScenarioName).Error(),
		},
		{
			Name: "returns error on persisting in DB",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return mockScenarioDefServiceThatReturns([]string{ScenarioName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", mock.Anything, fixModel()).Return(fixError()).Once()
				return mockRepo
			},
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			InputASA:             fixModel(),
			ExpectedASA:          model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:   "while persisting Assignment: some error",
		},
		{
			Name: "returns error on ensuring that scenarios label definition exist",
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}
				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, mock.Anything).Return(fixError()).Once()
				return labelDefSvc
			},
			AsaRepoFn:            unusedASARepo,
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			InputASA:             fixModel(),
			ExpectedASA:          model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:   "while ensuring that `scenarios` label definition exist: some error",
		},
		{
			Name: "returns error on getting available scenarios from label definition",
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}
				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, mock.Anything).Return(nil).Once()
				labelDefSvc.On("GetAvailableScenarios", mock.Anything, tenantID.String()).Return(nil, fixError()).Once()
				return labelDefSvc
			},
			AsaRepoFn:            unusedASARepo,
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			InputASA:             fixModel(),
			ExpectedASA:          model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:   "while getting available scenarios: some error",
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

			svc := formation.NewService(nil, nil, nil, nil, nil, nil, labelDefService, asaRepo, nil, tenantSvc, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

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

			mock.AssertExpectationsForObjects(t, tenantSvc, asaRepo, labelDefService, runtimeRepo, runtimeContextRepo)
		})
	}

	t.Run("returns error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(context.TODO(), fixModel())

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

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForTargetTenant", ctx, tenantID.String(), TargetTenantID).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil)

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(make([]*model.RuntimeContext, 0), nil)
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, runtimeContextRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.NoError(t, err)
	})

	t.Run("return error when listing runtimes fails", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForTargetTenant", ctx, tenantID.String(), TargetTenantID).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, fixError())
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("return error when listing runtimes contexts fails", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForTargetTenant", ctx, tenantID.String(), TargetTenantID).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil)

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(nil, fixError())
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, runtimeContextRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

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

func TestService_DeleteForScenarioName(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), ScenarioName).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(make([]*model.RuntimeContext, 0), nil).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, runtimeContextRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(fixCtxWithTenant(), fixModel())

		// THEN
		require.NoError(t, err)
	})

	t.Run("return error when listing runtimes fails", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), ScenarioName).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, fixError()).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(ctx, fixModel())

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("return error when listing runtimes contexts fails", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), ScenarioName).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(nil, fixError())
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, runtimeContextRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(ctx, fixModel())

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(context.TODO(), fixModel())

		// THEN
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), ScenarioName).Return(fixError()).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(fixCtxWithTenant(), fixModel())

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignment: %s", ErrMsg))
	})
}

func TestService_EnsureScenarioAssigned(t *testing.T) {
	selectorScenario := "SELECTOR_SCENARIO"
	in := fixAutomaticScenarioAssigment(selectorScenario)
	testErr := errors.New("test err")
	otherScenario := "OTHER"
	basicScenario := "SCENARIO"
	scenarios := []interface{}{otherScenario, basicScenario}

	rtmIDWithScenario := "rtm1_scenario"
	rtmIDWithoutScenario := "rtm1_no_scenario"

	runtimes := []*model.Runtime{{ID: rtmIDWithoutScenario}, {ID: rtmIDWithScenario}}
	labelWithoutScenario := model.Label{
		ID:         rtmIDWithoutScenario,
		Key:        "scenarios",
		Value:      []interface{}{selectorScenario},
		ObjectID:   rtmIDWithoutScenario,
		ObjectType: model.RuntimeLabelableObject,
	}
	labelWithScenario := model.Label{
		ID:         rtmIDWithScenario,
		Key:        "scenarios",
		Value:      scenarios,
		ObjectID:   rtmIDWithScenario,
		ObjectType: model.RuntimeLabelableObject,
	}
	labelInputWithoutScenario := model.LabelInput{
		Key:        "scenarios",
		Value:      []string{selectorScenario},
		ObjectID:   rtmIDWithoutScenario,
		ObjectType: model.RuntimeLabelableObject,
	}
	labelInputWithScenario := model.LabelInput{
		Key:        "scenarios",
		Value:      []string{selectorScenario},
		ObjectID:   rtmIDWithScenario,
		ObjectType: model.RuntimeLabelableObject,
	}

	rtmCtxIDWithScenario := "rtmCtx_scenario"
	rtmCtxIDWithoutScenario := "rtmCtx_no_scenario"

	runtimeContexts := []*model.RuntimeContext{{ID: rtmCtxIDWithoutScenario, RuntimeID: rtmIDWithoutScenario}, {ID: rtmCtxIDWithScenario, RuntimeID: rtmIDWithScenario}}
	rtmCtxLabelWithoutScenario := model.Label{
		ID:         rtmIDWithoutScenario,
		Key:        "scenarios",
		Value:      []interface{}{selectorScenario},
		ObjectID:   rtmCtxIDWithoutScenario,
		ObjectType: model.RuntimeContextLabelableObject,
	}
	rtmCtxLabelWithScenario := model.Label{
		ID:         rtmIDWithScenario,
		Key:        "scenarios",
		Value:      scenarios,
		ObjectID:   rtmCtxIDWithScenario,
		ObjectType: model.RuntimeContextLabelableObject,
	}
	rtmCtxLabelInputWithoutScenario := model.LabelInput{
		Key:        "scenarios",
		Value:      []string{selectorScenario},
		ObjectID:   rtmCtxIDWithoutScenario,
		ObjectType: model.RuntimeContextLabelableObject,
	}
	rtmCtxLabelInputWithScenario := model.LabelInput{
		Key:        "scenarios",
		Value:      []string{selectorScenario},
		ObjectID:   rtmCtxIDWithScenario,
		ObjectType: model.RuntimeContextLabelableObject,
	}

	expectedFormation := &model.Formation{
		ID:                  fixUUID(),
		Name:                in.ScenarioName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            in.Tenant,
	}

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[1].ID).Return(nil, nil)
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
		labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[1].ID).Return(nil, nil)
		labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, runtimeContexts[0].ID).Return(nil, nil)
		labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, runtimeContexts[1].ID).Return(nil, nil)
		webhookRepo := &automock.WebhookRepository{}
		webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))
		webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[1].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[1].ID))

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(runtimeContexts, nil).Once()
		runtimeContextRepo.On("GetByID", ctx, tenantID.String(), runtimeContexts[0].ID).Return(&model.RuntimeContext{
			RuntimeID: runtimes[0].ID,
		}, nil)
		runtimeContextRepo.On("GetByID", ctx, tenantID.String(), runtimeContexts[1].ID).Return(&model.RuntimeContext{
			RuntimeID: runtimes[1].ID,
		}, nil)

		upsertSvc := &automock.LabelService{}
		upsertSvc.On("GetLabel", ctx, tenantID.String(), &labelInputWithoutScenario).Return(&labelWithoutScenario, nil).Once()
		upsertSvc.On("GetLabel", ctx, tenantID.String(), &labelInputWithScenario).Return(&labelWithScenario, nil).Once()
		upsertSvc.On("GetLabel", ctx, tenantID.String(), &rtmCtxLabelInputWithoutScenario).Return(&rtmCtxLabelWithoutScenario, nil).Once()
		upsertSvc.On("GetLabel", ctx, tenantID.String(), &rtmCtxLabelInputWithScenario).Return(&rtmCtxLabelWithScenario, nil).Once()

		upsertSvc.On("UpdateLabel", ctx, tenantID.String(), rtmIDWithoutScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{selectorScenario},
			ObjectID:   rtmIDWithoutScenario,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(nil).Once()
		upsertSvc.On("UpdateLabel", ctx, tenantID.String(), rtmIDWithScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{otherScenario, basicScenario, selectorScenario},
			ObjectID:   rtmIDWithScenario,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(nil).Once()
		upsertSvc.On("UpdateLabel", ctx, tenantID.String(), rtmIDWithoutScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{selectorScenario},
			ObjectID:   rtmCtxIDWithoutScenario,
			ObjectType: model.RuntimeContextLabelableObject,
		}).Return(nil).Once()
		upsertSvc.On("UpdateLabel", ctx, tenantID.String(), rtmIDWithScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{otherScenario, basicScenario, selectorScenario},
			ObjectID:   rtmCtxIDWithScenario,
			ObjectType: model.RuntimeContextLabelableObject,
		}).Return(nil).Once()

		formationRepo := &automock.FormationRepository{}
		formationRepo.On("GetByName", ctx, selectorScenario, in.Tenant).Return(expectedFormation, nil).Times(4)

		svc := formation.NewService(nil, labelRepo, formationRepo, nil, upsertSvc, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo, webhookRepo, nil, nil, nil, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo, upsertSvc, formationRepo)
	})

	t.Run("Failed when insert new Label on upsert failed ", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return([]*model.Runtime{{ID: rtmIDWithoutScenario}}, nil).Once()

		upsertSvc := &automock.LabelService{}
		upsertSvc.On("GetLabel", ctx, tenantID.String(), &labelInputWithoutScenario).Return(&labelWithoutScenario, nil).Once()
		upsertSvc.On("UpdateLabel", ctx, tenantID.String(), rtmIDWithoutScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{selectorScenario},
			ObjectID:   rtmIDWithoutScenario,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, upsertSvc, nil, nil, nil, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, upsertSvc, runtimeRepo)
	})

	t.Run("Failed when GetScenarioLabelsForRuntimes returns error", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &labelInputWithoutScenario).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, labelService, nil, nil, nil, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo)
	})

	t.Run("Failed when listing runtimes returns error", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Failed when insert new Label for runtime context on upsert failed ", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(runtimeContexts, nil).Once()

		upsertSvc := &automock.LabelService{}
		upsertSvc.On("GetLabel", ctx, tenantID.String(), &rtmCtxLabelInputWithoutScenario).Return(&rtmCtxLabelWithoutScenario, nil).Once()

		upsertSvc.On("UpdateLabel", ctx, tenantID.String(), rtmIDWithoutScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{selectorScenario},
			ObjectID:   rtmCtxIDWithoutScenario,
			ObjectType: model.RuntimeContextLabelableObject,
		}).Return(testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, upsertSvc, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, upsertSvc, runtimeRepo, runtimeContextRepo)
	})

	t.Run("Failed when GetScenarioLabelsForRuntimes returns error", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(runtimeContexts, nil).Once()

		upsertSvc := &automock.LabelService{}
		upsertSvc.On("GetLabel", ctx, tenantID.String(), &rtmCtxLabelInputWithoutScenario).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, upsertSvc, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, upsertSvc, runtimeRepo, runtimeContextRepo)
	})

	t.Run("Failed when listing runtime contexts returns error", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo)
	})

	t.Run("Success, no runtimes found", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(make([]*model.RuntimeContext, 0), nil).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo)
	})
}

func TestService_RemoveAssignedScenario(t *testing.T) {
	selectorScenario := "SELECTOR_SCENARIO"
	rtmID := "8c4de4d8-dcfa-47a9-95c9-3c8b1f5b907c"
	in := fixAutomaticScenarioAssigment(selectorScenario)
	runtimes := []*model.Runtime{{ID: rtmID}}
	testErr := errors.New("test err")
	otherScenario := "OTHER"
	basicScenario := "SCENARIO"
	scenarios := []interface{}{otherScenario, basicScenario}
	stringScenarios := []string{otherScenario, basicScenario}
	labelInput := model.LabelInput{
		Key:        "scenarios",
		Value:      []string{selectorScenario},
		ObjectID:   rtmID,
		ObjectType: model.RuntimeLabelableObject,
	}
	label := model.Label{
		ID:         rtmID,
		Key:        "scenarios",
		Value:      append(scenarios, selectorScenario),
		ObjectID:   rtmID,
		ObjectType: model.RuntimeLabelableObject,
	}

	rtmCtxID := "7c4de4d8-dcfa-47a9-95c9-3c8b1f5b907d"
	rtmContexts := []*model.RuntimeContext{{ID: rtmCtxID, RuntimeID: rtmID}}
	rtmCtxLabelInput := model.LabelInput{
		Key:        "scenarios",
		Value:      []string{selectorScenario},
		ObjectID:   rtmCtxID,
		ObjectType: model.RuntimeContextLabelableObject,
	}
	rtmCtxLabel := model.Label{
		ID:         rtmCtxID,
		Key:        "scenarios",
		Value:      append(scenarios, selectorScenario),
		ObjectID:   rtmCtxID,
		ObjectType: model.RuntimeContextLabelableObject,
	}

	expectedFormation := &model.Formation{
		ID:                  fixUUID(),
		Name:                in.ScenarioName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            in.Tenant,
	}

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		webhookRepo := &automock.WebhookRepository{}
		webhookRepo.On("GetByIDAndWebhookType", ctx, tenantID.String(), runtimes[0].ID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, runtimes[0].ID))

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeLabelableObject, runtimes[0].ID).Return(nil, nil)
		labelRepo.On("ListForObject", ctx, tenantID.String(), model.RuntimeContextLabelableObject, rtmContexts[0].ID).Return(nil, nil)

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
		runtimeRepo.On("GetByID", ctx, tenantID.String(), runtimes[0].ID).Return(nil, nil)

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(rtmContexts, nil).Once()
		runtimeContextRepo.On("GetByID", ctx, tenantID.String(), rtmContexts[0].ID).Return(&model.RuntimeContext{
			RuntimeID: runtimes[0].ID,
		}, nil)

		asaRepo := &automock.AutomaticFormationAssignmentRepository{}
		asaRepo.On("ListAll", ctx, tenantID.String()).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &labelInput).Return(&label, nil).Once()
		labelService.On("UpdateLabel", ctx, tenantID.String(), rtmID, &model.LabelInput{
			Key:        "scenarios",
			Value:      stringScenarios,
			ObjectID:   rtmID,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(nil).Once()
		labelService.On("GetLabel", ctx, tenantID.String(), &rtmCtxLabelInput).Return(&rtmCtxLabel, nil).Once()
		labelService.On("UpdateLabel", ctx, tenantID.String(), rtmCtxID, &model.LabelInput{
			Key:        "scenarios",
			Value:      stringScenarios,
			ObjectID:   rtmCtxID,
			ObjectType: model.RuntimeContextLabelableObject,
		}).Return(nil).Once()

		formationRepo := &automock.FormationRepository{}
		formationRepo.On("GetByName", ctx, selectorScenario, in.Tenant).Return(expectedFormation, nil).Times(2)

		svc := formation.NewService(nil, labelRepo, formationRepo, nil, labelService, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, webhookRepo, nil, nil, nil, nil)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo, runtimeContextRepo, formationRepo)
	})

	t.Run("Failed when Label Upsert failed ", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		asaRepo := &automock.AutomaticFormationAssignmentRepository{}
		asaRepo.On("ListAll", ctx, tenantID.String()).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &labelInput).Return(&label, nil).Once()
		labelService.On("UpdateLabel", ctx, tenantID.String(), rtmID, &model.LabelInput{
			Key:        "scenarios",
			Value:      stringScenarios,
			ObjectID:   rtmID,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, labelService, nil, nil, asaRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo, asaRepo)
	})

	t.Run("Failed when ListAll returns error", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Failed when GetScenarioLabelsForRuntimes failed", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		asaRepo := &automock.AutomaticFormationAssignmentRepository{}
		asaRepo.On("ListAll", ctx, tenantID.String()).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &labelInput).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, labelService, nil, nil, asaRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo, asaRepo)
	})

	t.Run("Failed when Label Upsert for runtime context failed ", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(rtmContexts, nil).Once()

		asaRepo := &automock.AutomaticFormationAssignmentRepository{}
		asaRepo.On("ListAll", ctx, tenantID.String()).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &rtmCtxLabelInput).Return(&rtmCtxLabel, nil).Once()
		labelService.On("UpdateLabel", ctx, tenantID.String(), rtmCtxID, &model.LabelInput{
			Key:        "scenarios",
			Value:      stringScenarios,
			ObjectID:   rtmCtxID,
			ObjectType: model.RuntimeContextLabelableObject,
		}).Return(testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, labelService, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo, asaRepo, runtimeContextRepo)
	})

	t.Run("Failed when ListAll for runtime context returns error", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo)
	})

	t.Run("Failed when GetScenarioLabelsForRuntimes for runtime context failed", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, TargetTenantID).Return(rtmContexts, nil).Once()

		asaRepo := &automock.AutomaticFormationAssignmentRepository{}
		asaRepo.On("ListAll", ctx, tenantID.String()).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &rtmCtxLabelInput).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, labelService, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo, asaRepo, runtimeContextRepo)
	})
}

func TestService_MergeScenariosFromInputLabelsAndAssignments_Success(t *testing.T) {
	// GIVEN
	ctx := fixCtxWithTenant()
	differentTargetTenant := "differentTargetTenant"
	runtimeID := "runtimeID"
	labelKey := "key"
	labelValue := "val"

	inputLabels := map[string]interface{}{
		labelKey: labelValue,
	}

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

	expectedScenarios := []interface{}{ScenarioName}

	asaRepo := &automock.AutomaticFormationAssignmentRepository{}
	asaRepo.On("ListAll", ctx, tenantID.String()).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", ctx, TargetTenantID, runtimeID).Return(true, nil).Once()
	runtimeRepo.On("Exists", ctx, differentTargetTenant, runtimeID).Return(false, nil).Once()

	svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

	// WHEN
	actualScenarios, err := svc.MergeScenariosFromInputLabelsAndAssignments(ctx, inputLabels, runtimeID)

	// THEN
	require.NoError(t, err)
	require.ElementsMatch(t, expectedScenarios, actualScenarios)

	mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo)
}

func TestService_MergeScenariosFromInputLabelsAndAssignments_SuccessIfScenariosLabelIsInInput(t *testing.T) {
	// GIVEN
	ctx := fixCtxWithTenant()
	runtimeID := "runtimeID"
	labelKey := "key"
	labelValue := "val"

	scenario := "SCENARIO"
	inputLabels := map[string]interface{}{
		labelKey:           labelValue,
		model.ScenariosKey: []interface{}{scenario},
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   ScenarioName,
			Tenant:         tenantID.String(),
			TargetTenantID: TargetTenantID,
		},
	}

	expectedScenarios := []interface{}{ScenarioName, scenario}

	asaRepo := &automock.AutomaticFormationAssignmentRepository{}
	asaRepo.On("ListAll", ctx, tenantID.String()).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", ctx, TargetTenantID, runtimeID).Return(true, nil).Once()

	svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

	// WHEN
	actualScenarios, err := svc.MergeScenariosFromInputLabelsAndAssignments(ctx, inputLabels, runtimeID)

	// THEN
	require.NoError(t, err)
	require.ElementsMatch(t, expectedScenarios, actualScenarios)

	mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo)
}

func TestService_MergeScenariosFromInputLabelsAndAssignments_ReturnsErrorIfListAllFailed(t *testing.T) {
	// GIVEN
	ctx := fixCtxWithTenant()
	testErr := errors.New("testErr")
	labelKey := "key"
	labelValue := "val"

	inputLabels := map[string]interface{}{
		labelKey: labelValue,
	}

	asaRepo := &automock.AutomaticFormationAssignmentRepository{}
	asaRepo.On("ListAll", ctx, tenantID.String()).Return(nil, testErr)

	svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, asaRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	// WHEN
	_, err := svc.MergeScenariosFromInputLabelsAndAssignments(ctx, inputLabels, "runtimeID")

	// THEN
	require.Error(t, err)

	asaRepo.AssertExpectations(t)
}

func TestService_MergeScenariosFromInputLabelsAndAssignments_ReturnsErrorIfExistsFailed(t *testing.T) {
	// GIVEN
	ctx := fixCtxWithTenant()
	runtimeID := "runtimeID"
	testErr := errors.New("testErr")
	labelKey := "key"
	labelValue := "val"

	inputLabels := map[string]interface{}{
		labelKey: labelValue,
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   ScenarioName,
			Tenant:         tenantID.String(),
			TargetTenantID: TargetTenantID,
		},
	}

	asaRepo := &automock.AutomaticFormationAssignmentRepository{}
	asaRepo.On("ListAll", ctx, tenantID.String()).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", ctx, TargetTenantID, runtimeID).Return(false, testErr).Once()

	svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

	// WHEN
	_, err := svc.MergeScenariosFromInputLabelsAndAssignments(ctx, inputLabels, runtimeID)

	// THEN
	require.Error(t, err)

	mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo)
}

func TestService_MergeScenariosFromInputLabelsAndAssignments_ReturnsErrorIfScenariosFromInputWereNotInterfaceSlice(t *testing.T) {
	//  GIVEN
	ctx := fixCtxWithTenant()
	runtimeID := "runtimeID"
	labelKey := "key"
	labelValue := "val"

	scenario := "SCENARIO"
	inputLabels := map[string]interface{}{
		labelKey:           labelValue,
		model.ScenariosKey: []string{scenario},
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   ScenarioName,
			Tenant:         tenantID.String(),
			TargetTenantID: TargetTenantID,
		},
	}

	asaRepo := &automock.AutomaticFormationAssignmentRepository{}
	asaRepo.On("ListAll", ctx, tenantID.String()).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", ctx, TargetTenantID, runtimeID).Return(true, nil).Once()

	svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, nil, nil)

	// WHEN
	_, err := svc.MergeScenariosFromInputLabelsAndAssignments(ctx, inputLabels, runtimeID)

	// THEN
	require.Error(t, err)

	mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo)
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

	testCases := []struct {
		Name                     string
		ScenarioAssignmentRepoFn func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFn            func() *automock.RuntimeRepository
		RuntimeContextRepoFn     func() *automock.RuntimeContextRepository
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
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, TargetTenantID, RuntimeID).Return(true, nil)
				repo.On("Exists", ctx, TargetTenantID2, RuntimeID).Return(false, nil)
				return repo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			ObjectID:             RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedError:        nil,
			ExpectedScenarios:    []string{ScenarioName},
		},
		{
			Name: "Success for runtime context",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeRepoFn: unusedRuntimeRepo,
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, TargetTenantID, RuntimeContextID).Return(true, nil)
				repo.On("Exists", ctx, TargetTenantID2, RuntimeContextID).Return(false, nil)
				return repo
			},
			ObjectID:          RuntimeContextID,
			ObjectType:        graphql.FormationObjectTypeRuntimeContext,
			ExpectedError:     nil,
			ExpectedScenarios: []string{ScenarioName},
		},
		{
			Name:                     "Returns error when formation object type is not expected",
			ScenarioAssignmentRepoFn: unusedASARepo,
			RuntimeRepoFn:            unusedRuntimeRepo,
			RuntimeContextRepoFn:     unusedRuntimeContextRepo,
			ObjectID:                 RuntimeID,
			ObjectType:               graphql.FormationObjectTypeTenant,
			ExpectedError:            errors.Errorf("unexpected formation object type %q", graphql.FormationObjectTypeTenant),
			ExpectedScenarios:        nil,
		},
		{
			Name: "Returns error when can't list ASAs",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFn:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			ObjectID:             RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedError:        testErr,
			ExpectedScenarios:    nil,
		},
		{
			Name: "Returns error when checking if asa matches runtime",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, TargetTenantID, RuntimeID).Return(false, testErr)
				return repo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			ObjectID:             RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedError:        testErr,
			ExpectedScenarios:    nil,
		},
		{
			Name: "Returns error when checking if asa matches runtime context",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeRepoFn: unusedRuntimeRepo,
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, TargetTenantID, RuntimeContextID).Return(false, testErr)
				return repo
			},
			ObjectID:          RuntimeContextID,
			ObjectType:        graphql.FormationObjectTypeRuntimeContext,
			ExpectedError:     testErr,
			ExpectedScenarios: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			asaRepo := testCase.ScenarioAssignmentRepoFn()
			runtimeRepo := testCase.RuntimeRepoFn()
			runtimeContextRepo := testCase.RuntimeContextRepoFn()

			svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil)

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

			mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo, runtimeContextRepo)
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
