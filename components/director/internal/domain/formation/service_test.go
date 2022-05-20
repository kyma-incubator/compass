package formation_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/frmtest"

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

const (
	targetTenantID = "targetTenantID"
	scenarioName   = "scenario-A"
	errMsg         = "some error"
	tnt            = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	targetTenant   = "targetTenant"
	externalTnt    = "external-tnt"
)

func TestServiceCreateFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	in := model.Formation{
		Name: testFormation,
	}
	expected := &model.Formation{
		Name: testFormation,
	}

	defaultSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
	assert.NoError(t, err)
	defaultSchemaLblDef := fixDefaultScenariosLabelDefinition(tnt, defaultSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
	assert.NoError(t, err)
	newSchemaLblDef := fixDefaultScenariosLabelDefinition(tnt, newSchema)

	emptySchemaLblDef := fixDefaultScenariosLabelDefinition(tnt, defaultSchema)
	emptySchemaLblDef.Schema = nil

	testCases := []struct {
		Name                 string
		LabelDefRepositoryFn func() *automock.LabelDefRepository
		LabelDefServiceFn    func() *automock.LabelDefService
		ExpectedFormation    *model.Formation
		ExpectedErrMessage   string
	}{
		{
			Name: "success when no labeldef exists",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, tnt, []string{testFormation}).Return(nil)
				return labelDefService
			},
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success when labeldef exists",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error when labeldef is missing and can not create it",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, tnt, []string{testFormation}).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when can not get labeldef",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(nil, testErr)
				return labelDefRepo
			},
			LabelDefServiceFn:  formation.UnusedLabelDefService,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when labeldef's schema is missing",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&emptySchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn:  formation.UnusedLabelDefService,
			ExpectedErrMessage: "missing schema",
		},
		{
			Name: "error when validating existing labels against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, defaultSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when validating automatic scenario assignment against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, defaultSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, tnt, defaultSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when update with version fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(testErr)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, defaultSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, tnt, defaultSchemaLblDef.Key).Return(nil)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			lblDefRepo := testCase.LabelDefRepositoryFn()
			lblDefService := testCase.LabelDefServiceFn()

			svc := formation.NewService(lblDefRepo, nil, nil, nil, lblDefService, nil, nil, nil, nil, nil)

			// WHEN
			actual, err := svc.CreateFormation(ctx, tnt, in)

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

func TestServiceDeleteFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	in := model.Formation{
		Name: testFormation,
	}

	expected := &model.Formation{
		Name: testFormation,
	}

	defaultSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
	assert.NoError(t, err)
	defaultSchemaLblDef := fixDefaultScenariosLabelDefinition(tnt, defaultSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
	assert.NoError(t, err)
	newSchemaLblDef := fixDefaultScenariosLabelDefinition(tnt, newSchema)

	emptySchemaLblDef := fixDefaultScenariosLabelDefinition(tnt, defaultSchema)
	emptySchemaLblDef.Schema = nil

	testCases := []struct {
		Name                 string
		LabelDefRepositoryFn func() *automock.LabelDefRepository
		LabelDefServiceFn    func() *automock.LabelDefService
		ExpectedFormation    *model.Formation
		ExpectedErrMessage   string
	}{
		{
			Name: "success",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error when can not get labeldef",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(nil, testErr)
				return labelDefRepo
			},
			LabelDefServiceFn:  formation.UnusedLabelDefService,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when labeldef's schema is missing",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&emptySchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn:  formation.UnusedLabelDefService,
			ExpectedErrMessage: "missing schema",
		},
		{
			Name: "error when validating existing labels against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when validating automatic scenario assignment against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when update with version fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&newSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(testErr)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, newSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, tnt, newSchemaLblDef.Key).Return(nil)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			lblDefRepo := testCase.LabelDefRepositoryFn()
			lblDefService := testCase.LabelDefServiceFn()

			svc := formation.NewService(lblDefRepo, nil, nil, nil, lblDefService, nil, nil, nil, nil, nil)

			// WHEN
			actual, err := svc.DeleteFormation(ctx, tnt, in)

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
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("test error")

	inputFormation := model.Formation{
		Name: testFormation,
	}
	expectedFormation := &model.Formation{
		Name: testFormation,
	}

	inputSecondFormation := model.Formation{
		Name: "test-formation-2",
	}
	expectedSecondFormation := &model.Formation{
		Name: "test-formation-2",
	}

	objectID := "123"
	applicationLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormation},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	applicationLblInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormation},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	runtimeLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormation},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeLblInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormation},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}

	asa := model.AutomaticScenarioAssignment{
		ScenarioName:   testFormation,
		Tenant:         tnt,
		TargetTenantID: targetTenant,
	}

	testCases := []struct {
		Name               string
		UIDServiceFn       func() *automock.UidService
		LabelServiceFn     func() *automock.LabelService
		LabelDefServiceFn  func() *automock.LabelDefService
		TenantServiceFn    func() *automock.TenantService
		AsaRepoFn          func() *automock.AutomaticFormationAssignmentRepository
		AsaServiceFN       func() *automock.AutomaticFormationAssignmentService
		RuntimeRepoFN      func() *automock.RuntimeRepository
		ObjectType         graphql.FormationObjectType
		InputFormation     model.Formation
		ExpectedFormation  *model.Formation
		ExpectedErrMessage string
	}{
		{
			Name: "success for application if label does not exist",
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name:         "success for application if formation is already added",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, tnt, applicationLbl.ID, &applicationLblInput).Return(nil)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name:         "success for application with new formation",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{"test-formation-2"},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation, "test-formation-2"},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime if label does not exist",
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name:         "success for runtime if formation is already added",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, tnt, runtimeLbl.ID, &runtimeLblInput).Return(nil)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name:         "success for runtime with new formation",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{"test-formation-2"},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation, "test-formation-2"},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name:         "success for tenant",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, objectID).Return(targetTenant, nil)
				return svc
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", ctx, tnt).Return(nil)
				labelDefSvc.On("GetAvailableScenarios", ctx, tnt).Return([]string{testFormation}, nil)

				return labelDefSvc
			},
			LabelServiceFn: formation.UnusedLabelService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctx, asa).Return(nil)

				return asaRepo
			},
			AsaServiceFN: formation.UnusedASAService,
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListAll", ctx, targetTenant, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil)

				return runtimeRepo
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for application when label does not exist and can't create it",
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, tnt, fixUUID(), &applicationLblInput).Return(testErr)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for application while getting label",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &applicationLblInput).Return(nil, testErr)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			AsaServiceFN:       formation.UnusedASAService,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for application while converting label values to string slice",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for application while converting label value to string",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &applicationLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name:         "error for application when updating label fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, tnt, applicationLbl.ID, &applicationLblInput).Return(testErr)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime when label does not exist and can't create it",
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, tnt, fixUUID(), &runtimeLblInput).Return(testErr)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for runtime while getting label",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &runtimeLblInput).Return(nil, testErr)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for runtime while converting label values to string slice",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for runtime while converting label value to string",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &runtimeLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name:         "error for runtime when updating label fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, tnt, runtimeLbl.ID, &runtimeLblInput).Return(testErr)
				return labelService
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for tenant when tenant conversion fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, objectID).Return("", testErr)
				return svc
			},
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			LabelServiceFn:     formation.UnusedLabelService,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for tenant when create fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, objectID).Return(targetTenant, nil)
				return svc
			},
			LabelServiceFn: formation.UnusedLabelService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctx, model.AutomaticScenarioAssignment{ScenarioName: testFormation, Tenant: tnt, TargetTenantID: targetTenant}).Return(testErr)

				return asaRepo
			},
			AsaServiceFN: formation.UnusedASAService,
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", ctx, tnt).Return(nil)
				labelDefSvc.On("GetAvailableScenarios", ctx, tnt).Return([]string{testFormation}, nil)

				return labelDefSvc
			},
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "error when object type is unknown",
			UIDServiceFn:       frmtest.UnusedUUIDService(),
			LabelServiceFn:     formation.UnusedLabelService,
			LabelDefServiceFn:  formation.UnusedLabelDefServiceFn,
			AsaRepoFn:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         "UNKNOWN",
			InputFormation:     inputFormation,
			ExpectedErrMessage: "unknown formation type",
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

			if testCase.TenantServiceFn != nil {
				tenantSvc = testCase.TenantServiceFn()
			}

			svc := formation.NewService(nil, nil, labelService, uidService, labelDefService, asaRepo, asaService, tenantSvc, runtimeRepo, nil)

			// WHEN
			actual, err := svc.AssignFormation(ctx, tnt, objectID, testCase.ObjectType, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, uidService, labelService, asaService, tenantSvc, asaRepo, labelDefService, runtimeRepo)
		})
	}
}

func TestServiceUnassignFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("test error")

	in := model.Formation{
		Name: testFormation,
	}
	expected := &model.Formation{
		Name: testFormation,
	}
	secondFormation := model.Formation{
		Name: secondTestFormation,
	}

	objectID := "123"
	applicationLblSingleFormation := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormation},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	applicationLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormation, secondTestFormation},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	applicationLblInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormation},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	runtimeLblSingleFormation := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormation},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormation, secondTestFormation},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeLblInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormation},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}

	asa := model.AutomaticScenarioAssignment{
		ScenarioName:   testFormation,
		Tenant:         tnt,
		TargetTenantID: objectID,
	}

	testCases := []struct {
		Name               string
		UIDServiceFn       func() *automock.UidService
		LabelServiceFn     func() *automock.LabelService
		LabelRepoFn        func() *automock.LabelRepository
		AsaServiceFN       func() *automock.AutomaticFormationAssignmentService
		AsaRepoFN          func() *automock.AutomaticFormationAssignmentRepository
		EngineFn           func() *automock.ScenarioAssignmentEngine
		RuntimeRepoFN      func() *automock.RuntimeRepository
		ObjectType         graphql.FormationObjectType
		InputFormation     model.Formation
		ExpectedFormation  *model.Formation
		ExpectedErrMessage string
	}{
		{
			Name:         "success for application",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelRepoFn:        formation.UnusedLabelRepo,
			AsaRepoFN:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name:         "success for application if formation do not exist",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelRepoFn:        formation.UnusedLabelRepo,
			AsaRepoFN:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name:         "success for application when formation is last",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, applicationLblInput).Return(applicationLblSingleFormation, nil)
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, tnt, model.ApplicationLabelableObject, objectID, model.ScenariosKey).Return(nil)
				return labelRepo
			},
			AsaRepoFN:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name:         "success for runtime",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelRepoFn:  formation.UnusedLabelRepo,
			AsaRepoFN:    formation.UnusedASARepo,
			AsaServiceFN: formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name:           "success for runtime when formation is coming from ASA",
			UIDServiceFn:   frmtest.UnusedUUIDService(),
			LabelServiceFn: formation.UnusedLabelService,
			LabelRepoFn:    formation.UnusedLabelRepo,
			AsaRepoFN:      formation.UnusedASARepo,
			AsaServiceFN:   formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return([]string{testFormation}, nil)
				return engine
			},
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name:         "success for runtime if formation do not exist",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLblSingleFormation, nil)
				labelService.On("UpdateLabel", ctx, tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelRepoFn:  formation.UnusedLabelRepo,
			AsaRepoFN:    formation.UnusedASARepo,
			AsaServiceFN: formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     secondFormation,
			ExpectedFormation:  &secondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name:         "success for runtime when formation is last",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, runtimeLblInput).Return(runtimeLblSingleFormation, nil)
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, objectID, model.ScenariosKey).Return(nil)
				return labelRepo
			},
			AsaRepoFN:    formation.UnusedASARepo,
			AsaServiceFN: formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name:           "success for tenant",
			UIDServiceFn:   frmtest.UnusedUUIDService(),
			LabelServiceFn: formation.UnusedLabelService,
			LabelRepoFn:    formation.UnusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}

				asaRepo.On("DeleteForScenarioName", ctx, tnt, testFormation).Return(nil)

				return asaRepo
			},
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormation).Return(asa, nil)
				return asaService
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListAll", ctx, "123", []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil)

				return runtimeRepo
			},
			EngineFn:           formation.UnusedEngine,
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name:         "error for application while getting label",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, applicationLblInput).Return(nil, testErr)
				return labelService
			},
			LabelRepoFn:        formation.UnusedLabelRepo,
			AsaRepoFN:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for application while converting label values to string slice",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelRepoFn:        formation.UnusedLabelRepo,
			AsaRepoFN:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for application while converting label value to string",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, applicationLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelRepoFn:        formation.UnusedLabelRepo,
			AsaRepoFN:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name:         "error for application when formation is last and delete fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, applicationLblInput).Return(applicationLblSingleFormation, nil)
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, tnt, model.ApplicationLabelableObject, objectID, model.ScenariosKey).Return(testErr)
				return labelRepo
			},
			AsaRepoFN:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for application when updating label fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(testErr)
				return labelService
			},
			LabelRepoFn:        formation.UnusedLabelRepo,
			AsaRepoFN:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:           "error for runtime when can't get formations that are coming from ASAs",
			UIDServiceFn:   frmtest.UnusedUUIDService(),
			LabelServiceFn: formation.UnusedLabelService,
			LabelRepoFn:    formation.UnusedLabelRepo,
			AsaRepoFN:      formation.UnusedASARepo,
			AsaServiceFN:   formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, testErr)
				return engine
			},
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for runtime while getting label",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, runtimeLblInput).Return(nil, testErr)
				return labelService
			},
			LabelRepoFn:  formation.UnusedLabelRepo,
			AsaRepoFN:    formation.UnusedASARepo,
			AsaServiceFN: formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for runtime while converting label values to string slice",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelRepoFn:  formation.UnusedLabelRepo,
			AsaRepoFN:    formation.UnusedASARepo,
			AsaServiceFN: formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for runtime while converting label value to string",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, runtimeLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelRepoFn:   formation.UnusedLabelRepo,
			AsaRepoFN:     formation.UnusedASARepo,
			AsaServiceFN:  formation.UnusedASAService,
			RuntimeRepoFN: formation.UnusedRuntimeRepo,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name:         "error for runtime when formation is last and delete fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, runtimeLblInput).Return(runtimeLblSingleFormation, nil)
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, objectID, model.ScenariosKey).Return(testErr)
				return labelRepo
			},
			AsaRepoFN:    formation.UnusedASARepo,
			AsaServiceFN: formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for runtime when updating label fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(testErr)
				return labelService
			},
			LabelRepoFn:   formation.UnusedLabelRepo,
			AsaRepoFN:     formation.UnusedASARepo,
			AsaServiceFN:  formation.UnusedASAService,
			RuntimeRepoFN: formation.UnusedRuntimeRepo,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:           "error for tenant when delete fails",
			UIDServiceFn:   frmtest.UnusedUUIDService(),
			LabelServiceFn: formation.UnusedLabelService,
			LabelRepoFn:    formation.UnusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}

				asaRepo.On("DeleteForScenarioName", ctx, tnt, testFormation).Return(testErr)

				return asaRepo
			},
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormation).Return(asa, nil)
				return asaService
			},
			EngineFn:           formation.UnusedEngine,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:           "error for tenant when delete fails",
			UIDServiceFn:   frmtest.UnusedUUIDService(),
			LabelServiceFn: formation.UnusedLabelService,
			LabelRepoFn:    formation.UnusedLabelRepo,
			AsaRepoFN:      formation.UnusedASARepo,
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormation).Return(model.AutomaticScenarioAssignment{}, testErr)
				return asaService
			},
			EngineFn:           formation.UnusedEngine,
			ObjectType:         graphql.FormationObjectTypeTenant,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "error when object type is unknown",
			UIDServiceFn:       frmtest.UnusedUUIDService(),
			LabelServiceFn:     formation.UnusedLabelService,
			LabelRepoFn:        formation.UnusedLabelRepo,
			AsaRepoFN:          formation.UnusedASARepo,
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
			RuntimeRepoFN:      formation.UnusedRuntimeRepo,
			ObjectType:         "UNKNOWN",
			InputFormation:     in,
			ExpectedErrMessage: "unknown formation type",
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
			engine := testCase.EngineFn()
			runtimeRepo := testCase.RuntimeRepoFN()

			svc := formation.NewService(nil, labelRepo, labelService, uidService, nil, asaRepo, asaService, nil, runtimeRepo, engine)

			// WHEN
			actual, err := svc.UnassignFormation(ctx, tnt, objectID, testCase.ObjectType, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}
			mock.AssertExpectationsForObjects(t, uidService, labelService, asaService, asaService, runtimeRepo, engine)
		})
	}
}
func fixUUID() string {
	return "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
}

func fixDefaultScenariosLabelDefinition(tenantID string, schema interface{}) model.LabelDefinition {
	return model.LabelDefinition{
		Key:     model.ScenariosKey,
		Tenant:  tenantID,
		Schema:  &schema,
		Version: 1,
	}
}

func fixAutomaticScenarioAssigment(selectorScenario string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   selectorScenario,
		Tenant:         tenantID.String(),
		TargetTenantID: targetTenantID,
	}
}

func TestService_CreateAutomaticScenarioAssignment(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("Create", ctx, fixModel()).Return(nil)
		mockScenarioDefSvc := mockScenarioDefServiceThatReturns([]string{scenarioName})
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil)
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, mockScenarioDefSvc)

		svc := formation.NewService(nil, nil, nil, nil, mockScenarioDefSvc, mockRepo, nil, nil, runtimeRepo, nil)

		// WHEN
		actual, err := svc.CreateAutomaticScenarioAssignment(fixCtxWithTenant(), fixModel())

		// THEN
		require.NoError(t, err)
		assert.Equal(t, fixModel(), actual)
	})

	t.Run("return error when ensuring scenarios for runtimes fails", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("Create", ctx, fixModel()).Return(nil)
		mockScenarioDefSvc := mockScenarioDefServiceThatReturns([]string{scenarioName})
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, fixError())
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, mockScenarioDefSvc)

		svc := formation.NewService(nil, nil, nil, nil, mockScenarioDefSvc, mockRepo, nil, nil, runtimeRepo, nil)

		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(fixCtxWithTenant(), fixModel())

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("returns error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(context.TODO(), fixModel())

		// THEN
		assert.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("returns error when scenario already has an assignment", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}

		mockRepo.On("Create", mock.Anything, fixModel()).Return(apperrors.NewNotUniqueError(""))
		mockScenarioDefSvc := mockScenarioDefServiceThatReturns([]string{scenarioName})
		runtimeRepo := &automock.RuntimeRepository{}
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, mockScenarioDefSvc)

		svc := formation.NewService(nil, nil, nil, nil, mockScenarioDefSvc, mockRepo, nil, nil, runtimeRepo, nil)

		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(fixCtxWithTenant(), fixModel())
		// THEN
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "a given scenario already has an assignment")
	})

	t.Run("returns error when given scenario does not exist", func(t *testing.T) {
		// GIVEN
		mockScenarioDefSvc := mockScenarioDefServiceThatReturns([]string{"completely-different-scenario"})
		runtimeRepo := &automock.RuntimeRepository{}
		defer mock.AssertExpectationsForObjects(t, runtimeRepo, mockScenarioDefSvc)

		svc := formation.NewService(nil, nil, nil, nil, mockScenarioDefSvc, nil, nil, nil, runtimeRepo, nil)
		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(fixCtxWithTenant(), fixModel())

		// THEN
		require.EqualError(t, err, apperrors.NewNotFoundError(resource.AutomaticScenarioAssigment, fixModel().ScenarioName).Error())
	})

	t.Run("returns error on persisting in DB", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("Create", mock.Anything, fixModel()).Return(fixError())
		mockScenarioDefSvc := mockScenarioDefServiceThatReturns([]string{scenarioName})

		runtimeRepo := &automock.RuntimeRepository{}
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, mockScenarioDefSvc)

		svc := formation.NewService(nil, nil, nil, nil, mockScenarioDefSvc, mockRepo, nil, nil, runtimeRepo, nil)

		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(fixCtxWithTenant(), fixModel())

		// THEN
		require.EqualError(t, err, "while persisting Assignment: some error")
	})

	t.Run("returns error on ensuring that scenarios label definition exist", func(t *testing.T) {
		// GIVEN
		mockScenarioDefSvc := &automock.LabelDefService{}
		mockScenarioDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, mock.Anything).Return(fixError())
		runtimeRepo := &automock.RuntimeRepository{}
		defer mock.AssertExpectationsForObjects(t, runtimeRepo, mockScenarioDefSvc)

		svc := formation.NewService(nil, nil, nil, nil, mockScenarioDefSvc, nil, nil, nil, runtimeRepo, nil)
		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(fixCtxWithTenant(), fixModel())
		// THEN
		require.EqualError(t, err, "while ensuring that `scenarios` label definition exist: some error")
	})

	t.Run("returns error on getting available scenarios from label definition", func(t *testing.T) {
		// GIVEN
		mockScenarioDefSvc := &automock.LabelDefService{}
		defer mock.AssertExpectationsForObjects(t, mockScenarioDefSvc)
		mockScenarioDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, mock.Anything).Return(nil)
		mockScenarioDefSvc.On("GetAvailableScenarios", mock.Anything, tenantID.String()).Return(nil, fixError())
		runtimeRepo := &automock.RuntimeRepository{}
		defer mock.AssertExpectationsForObjects(t, runtimeRepo, mockScenarioDefSvc)

		svc := formation.NewService(nil, nil, nil, nil, mockScenarioDefSvc, nil, nil, nil, runtimeRepo, nil)
		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(fixCtxWithTenant(), fixModel())
		// THEN
		require.EqualError(t, err, "while getting available scenarios: some error")
	})
}

func TestService_DeleteManyForSameTargetTenant(t *testing.T) {
	ctx := fixCtxWithTenant()

	scenarioNameA := "scenario-A"
	scenarioNameB := "scenario-B"
	models := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   scenarioNameA,
			TargetTenantID: targetTenantID,
		},
		{
			ScenarioName:   scenarioNameB,
			TargetTenantID: targetTenantID,
		},
	}

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForTargetTenant", ctx, tenantID.String(), targetTenantID).Return(nil).Once()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil)
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil)
		// WHEN
		err := svc.DeleteManyForSameTargetTenant(ctx, models)
		// THEN
		require.NoError(t, err)
	})

	t.Run("return error when unassigning scenarios from runtimes fails", func(t *testing.T) {
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForTargetTenant", ctx, tenantID.String(), targetTenantID).Return(nil).Once()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, fixError())
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil)
		// WHEN
		err := svc.DeleteManyForSameTargetTenant(ctx, models)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("return error when input slice is empty", func(t *testing.T) {
		// GIVEN

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		err := svc.DeleteManyForSameTargetTenant(ctx, []*model.AutomaticScenarioAssignment{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected at least one item in Assignments slice")
	})

	t.Run("return error when input slice contains assignments with different selectors", func(t *testing.T) {
		// GIVEN
		modelsWithDifferentSelectors := []*model.AutomaticScenarioAssignment{
			{
				ScenarioName:   scenarioNameA,
				TargetTenantID: targetTenantID,
			},
			{
				ScenarioName:   scenarioNameB,
				TargetTenantID: "differentTargetTenantID",
			},
		}

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		err := svc.DeleteManyForSameTargetTenant(ctx, modelsWithDifferentSelectors)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "all input items have to have the same target tenant")
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}

		mockRepo.On("DeleteForTargetTenant", ctx, tenantID.String(), targetTenantID).Return(fixError()).Once()

		defer mock.AssertExpectationsForObjects(t, mockRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, nil, nil)
		// WHEN
		err := svc.DeleteManyForSameTargetTenant(ctx, models)
		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignments: %s", errMsg))
	})

	t.Run("returns error when empty tenant", func(t *testing.T) {
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		err := svc.DeleteManyForSameTargetTenant(context.TODO(), models)
		require.EqualError(t, err, "cannot read tenant from context")
	})
}

func TestService_DeleteForScenarioName(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), scenarioName).Return(nil)
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil)
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(fixCtxWithTenant(), fixModel())

		// THEN
		require.NoError(t, err)
	})

	t.Run("return error when unassigning scenarios from runtimes fails", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), scenarioName).Return(nil)
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, fixError())
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(ctx, fixModel())

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(context.TODO(), fixModel())

		// THEN
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForScenarioName", ctx, tenantID.String(), scenarioName).Return(fixError())
		defer mock.AssertExpectationsForObjects(t, mockRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(fixCtxWithTenant(), fixModel())

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignment: %s", errMsg))
	})
}

func TestEngine_EnsureScenarioAssigned(t *testing.T) {
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

	t.Run("Success", func(t *testing.T) {
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
		upsertSvc := &automock.LabelService{}
		upsertSvc.On("GetLabel", ctx, tenantID.String(), &labelInputWithoutScenario).Return(&labelWithoutScenario, nil).Once()
		upsertSvc.On("GetLabel", ctx, tenantID.String(), &labelInputWithScenario).Return(&labelWithScenario, nil).Once()

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

		tenantSvc := &automock.TenantService{}

		svc := formation.NewService(nil, nil, upsertSvc, nil, nil, nil, nil, tenantSvc, runtimeRepo, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, runtimeRepo, tenantSvc, upsertSvc)
	})

	t.Run("Failed when insert new Label on upsert failed ", func(t *testing.T) {
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return([]*model.Runtime{{ID: rtmIDWithoutScenario}}, nil).Once()
		upsertSvc := &automock.LabelService{}
		upsertSvc.On("GetLabel", ctx, tenantID.String(), &labelInputWithoutScenario).Return(&labelWithoutScenario, nil).Once()
		upsertSvc.On("UpdateLabel", ctx, tenantID.String(), rtmIDWithoutScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{selectorScenario},
			ObjectID:   rtmIDWithoutScenario,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(testErr).Once()

		svc := formation.NewService(nil, nil, upsertSvc, nil, nil, nil, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, upsertSvc, runtimeRepo)
	})

	t.Run("Failed when GetScenarioLabelsForRuntimes returns error", func(t *testing.T) {
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &labelInputWithoutScenario).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, nil, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo)
	})

	t.Run("Failed when ListAll returns error", func(t *testing.T) {
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Success, no runtimes found", func(t *testing.T) {
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, nil).Once()

		labelService := &automock.LabelService{}

		svc := formation.NewService(nil, labelRepo, labelService, nil, nil, nil, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo, runtimeRepo)
	})
}

func TestEngine_RemoveAssignedScenario(t *testing.T) {
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

	t.Run("Success", func(t *testing.T) {
		ctx := context.TODO()

		engine := &automock.ScenarioAssignmentEngine{}
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
		engine.On("GetScenariosFromMatchingASAs", ctx, runtimes[0].ID).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &labelInput).Return(&label, nil).Once()
		labelService.On("UpdateLabel", ctx, tenantID.String(), rtmID, &model.LabelInput{
			Key:        "scenarios",
			Value:      stringScenarios,
			ObjectID:   rtmID,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(nil).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, nil, nil, nil, runtimeRepo, engine)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelService, engine, runtimeRepo)
	})

	t.Run("Failed when Label Upsert failed ", func(t *testing.T) {
		ctx := context.TODO()

		engine := &automock.ScenarioAssignmentEngine{}
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
		engine.On("GetScenariosFromMatchingASAs", ctx, runtimes[0].ID).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &labelInput).Return(&label, nil).Once()
		labelService.On("UpdateLabel", ctx, tenantID.String(), rtmID, &model.LabelInput{
			Key:        "scenarios",
			Value:      stringScenarios,
			ObjectID:   rtmID,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(testErr).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, nil, nil, nil, runtimeRepo, engine)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, engine, labelService, runtimeRepo)
	})

	t.Run("Failed when ListAll returns error", func(t *testing.T) {
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Failed when GetScenarioLabelsForRuntimes failed", func(t *testing.T) {
		ctx := context.TODO()

		engine := &automock.ScenarioAssignmentEngine{}
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
		engine.On("GetScenariosFromMatchingASAs", ctx, runtimes[0].ID).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &labelInput).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, nil, nil, nil, runtimeRepo, engine)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo)
	})
}

func TestEngine_RemoveAssignedScenarios(t *testing.T) {
	selectorScenario := "SCENARIO1"
	in := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   selectorScenario,
			Tenant:         tenantID.String(),
			TargetTenantID: targetTenantID,
		},
	}
	rtmID := "651038e0-e4b6-4036-a32f-f6e9846003f4"
	runtimes := []*model.Runtime{{ID: rtmID}}
	otherScenario := "SCENARIO1"
	basicScenario := "SCENARIO2"
	scenarios := []interface{}{otherScenario, basicScenario}

	labelInput := model.LabelInput{
		Key:        "scenarios",
		Value:      []string{selectorScenario},
		ObjectID:   rtmID,
		ObjectType: model.RuntimeLabelableObject,
	}
	label := model.Label{
		ID:         rtmID,
		Key:        "scenarios",
		Value:      scenarios,
		ObjectID:   rtmID,
		ObjectType: model.RuntimeLabelableObject,
	}

	t.Run("Success", func(t *testing.T) {
		ctx := context.TODO()

		engine := &automock.ScenarioAssignmentEngine{}
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
		engine.On("GetScenariosFromMatchingASAs", ctx, runtimes[0].ID).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &labelInput).Return(&label, nil).Once()
		labelService.On("UpdateLabel", ctx, tenantID.String(), rtmID, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{basicScenario},
			ObjectID:   rtmID,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(nil).Once()
		// GIVEN

		svc := formation.NewService(nil, nil, labelService, nil, nil, nil, nil, nil, runtimeRepo, engine)

		// WHEN
		err := svc.RemoveAssignedScenarios(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, engine, labelService, runtimeRepo)
	})

	t.Run("Error, while removing scenario - ListAll fail", func(t *testing.T) {
		// GIVEN
		testErr := errors.New("test error")
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, testErr).Once()
		labelRepo := &automock.LabelRepository{}

		labelService := &automock.LabelService{}

		svc := formation.NewService(nil, labelRepo, labelService, nil, nil, nil, nil, nil, runtimeRepo, nil)
		// WHEN
		err := svc.RemoveAssignedScenarios(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, runtimeRepo)
	})

	t.Run("Error, while removing scenario - GetScenarioLabelsForRuntimes fail", func(t *testing.T) {
		ctx := context.TODO()
		testErr := errors.New("test error")

		engine := &automock.ScenarioAssignmentEngine{}
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
		engine.On("GetScenariosFromMatchingASAs", ctx, runtimes[0].ID).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, tenantID.String(), &labelInput).Return(nil, testErr).Once()
		// GIVEN

		svc := formation.NewService(nil, nil, labelService, nil, nil, nil, nil, nil, runtimeRepo, engine)

		// WHEN
		err := svc.RemoveAssignedScenarios(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelService, engine, runtimeRepo)
	})
}

var tenantID = uuid.New()
var externalTenantID = uuid.New()

func fixCtxWithTenant() context.Context {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID.String(), externalTenantID.String())

	return ctx
}

func fixModel() model.AutomaticScenarioAssignment {
	return fixModelWithScenarioName(scenarioName)
}

func fixModelWithScenarioName(scenario string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   scenario,
		Tenant:         tenantID.String(),
		TargetTenantID: targetTenantID,
	}
}

func fixError() error {
	return errors.New(errMsg)
}

func mockScenarioDefServiceThatReturns(scenarios []string) *automock.LabelDefService {
	mockScenarioDefSvc := &automock.LabelDefService{}
	mockScenarioDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, tenantID.String()).Return(nil)
	mockScenarioDefSvc.On("GetAvailableScenarios", mock.Anything, tenantID.String()).Return(scenarios, nil)
	return mockScenarioDefSvc
}
