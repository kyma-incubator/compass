package formation_test

import (
	"context"

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

			svc := formation.NewService(nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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

			svc := formation.NewService(nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("Test error")

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
				repo.On("GetByName", ctx, testFormationName, Tnt).Return(&modelFormation, nil).Once()
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
				repo.On("GetByName", ctx, testFormationName, Tnt).Return(nil, testErr).Once()
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

			svc := formation.NewService(nil, nil, nil, formationRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			actual, err := svc.GetFormationByName(ctx, testCase.Input, Tnt)

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

	testCases := []struct {
		Name                    string
		UUIDServiceFn           func() *automock.UuidService
		LabelDefRepositoryFn    func() *automock.LabelDefRepository
		LabelDefServiceFn       func() *automock.LabelDefService
		FormationTemplateRepoFn func() *automock.FormationTemplateRepository
		FormationRepoFn         func() *automock.FormationRepository
		ConstraintEngineFn      func() *automock.ConstraintEngine
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
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, fixFormationModel()).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			TemplateName:       testFormationTemplateName,
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
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, fixFormationModel()).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			TemplateName:       testFormationTemplateName,
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
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, Tnt, []string{testFormationName}).Return(nil)
				return labelDefService
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, fixFormationModel()).Return(testErr).Once()
				return formationRepoMock
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "An error occurred while creating formation with name",
		},
		{
			Name: "error when can not get labeldef",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(nil, testErr)
				return labelDefRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&emptySchemaLblDef, nil)
				return labelDefRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, testSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
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
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&testSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, Tnt, testSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, Tnt, testSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
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
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
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
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(nil, testErr).Once()
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
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, fixFormationModel()).Return(testErr).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "An error occurred while creating formation with name: \"test-formation\"",
		},
		{
			Name: "error while enforcing constraint pre operation",
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepoMock := &automock.FormationTemplateRepository{}
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
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
			Name: "error while enforcing constraints post operation",
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
				formationTemplateRepoMock.On("GetByNameAndTenant", ctx, testFormationTemplateName, Tnt).Return(fixFormationTemplateModel(), nil).Once()
				return formationTemplateRepoMock
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("Create", ctx, fixFormationModel()).Return(nil).Once()
				return formationRepoMock
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preCreateLocation, createFormationDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postCreateLocation, createFormationDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			TemplateName:       testFormationTemplateName,
			ExpectedErrMessage: "While enforcing constraints for target operation \"CREATE_FORMATION\" and constraint type \"POST\": Test error",
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

			svc := formation.NewService(nil, labelDefRepo, nil, formationRepo, formationTemplateRepo, nil, uidService, labelDefService, nil, nil, nil, nil, nil, nil, nil, constraintEngine, runtimeType, applicationType)

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

			mock.AssertExpectationsForObjects(t, uidService, labelDefRepo, labelDefService, formationRepo, formationTemplateRepo, constraintEngine)
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
		Name                    string
		LabelDefRepositoryFn    func() *automock.LabelDefRepository
		LabelDefServiceFn       func() *automock.LabelDefService
		FormationRepoFn         func() *automock.FormationRepository
		FormationTemplateRepoFn func() *automock.FormationTemplateRepository
		ConstraintEngineFn      func() *automock.ConstraintEngine
		InputFormation          model.Formation
		ExpectedFormation       *model.Formation
		ExpectedErrMessage      string
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
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error when can not get labeldef",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(nil, testErr).Once()
				return labelDefRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			InputFormation:     in,
			ExpectedErrMessage: "while getting `scenarios` label definition: Test error",
		},
		{
			Name: "error when labeldef's schema is missing",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, Tnt, model.ScenariosKey).Return(&nilSchemaLblDef, nil)
				return labelDefRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			InputFormation:     in,
			ExpectedErrMessage: "while validating Scenario Assignments against a new schema: Test error",
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
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when can't get formation by name",
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, Tnt).Return(nil, testErr).Once()
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
			InputFormation:     in,
			ExpectedErrMessage: "An error occurred while deleting formation with name: \"test-formation\": Test error",
		},
		{
			Name: "error while enforcing constraints pre operation",
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepoMock := &automock.FormationRepository{}
				formationRepoMock.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			InputFormation:     in,
			ExpectedErrMessage: "While enforcing constraints for target operation \"DELETE_FORMATION\" and constraint type \"POST\": Test error",
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

			svc := formation.NewService(nil, labelDefRepo, nil, formationRepo, formationTemplateRepo, nil, nil, labelDefService, nil, nil, nil, nil, nil, nil, nil, constraintEngine, runtimeType, applicationType)

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

			mock.AssertExpectationsForObjects(t, labelDefRepo, labelDefService, formationRepo, formationTemplateRepo, constraintEngine)
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

		svc := formation.NewService(nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, runtimeType, applicationType)

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

		svc := formation.NewService(nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("return error when input slice is empty", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
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

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, mockRepo, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignments: %s", ErrMsg))
	})

	t.Run("returns error when empty tenant", func(t *testing.T) {
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
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

			svc := formation.NewServiceWithAsaEngine(nil, nil, nil, nil, nil, nil, nil, labelDefService, asaRepo, nil, tenantSvc, nil, nil, nil, nil, nil, runtimeType, applicationType, asaEngine)

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
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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

			svc := formation.NewServiceWithAsaEngine(nil, nil, nil, nil, nil, nil, nil, nil, asaRepo, nil, tenantSvc, nil, nil, nil, nil, nil, runtimeType, applicationType, asaEngine)

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
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(context.TODO(), fixModel(ScenarioName))

		// THEN
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
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

			svc := formation.NewService(nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo, nil, nil, nil, runtimeType, applicationType)

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

		svc := formation.NewService(nil, nil, nil, nil, nil, labelService, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

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

		svc := formation.NewService(nil, nil, nil, nil, nil, labelService, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		formations, err := svc.GetFormationsForObject(ctx, tenantID.String(), model.RuntimeLabelableObject, id)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "while fetching scenario label for")
		require.Nil(t, formations)
		mock.AssertExpectationsForObjects(t, labelService)
	})
}
