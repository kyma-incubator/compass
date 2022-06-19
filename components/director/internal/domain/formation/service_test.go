package formation_test

import (
	"context"

	"github.com/pkg/errors"

	"fmt"
	"testing"

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

func TestServiceCreateFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, formation.Tnt, formation.ExternalTnt)

	testErr := errors.New("Test error")

	in := model.Formation{
		Name: testFormation,
	}
	expected := &model.Formation{
		Name: testFormation,
	}

	defaultSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
	assert.NoError(t, err)
	defaultSchemaLblDef := formation.FixDefaultScenariosLabelDefinition(formation.Tnt, defaultSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
	assert.NoError(t, err)
	newSchemaLblDef := formation.FixDefaultScenariosLabelDefinition(formation.Tnt, newSchema)

	emptySchemaLblDef := formation.FixDefaultScenariosLabelDefinition(formation.Tnt, defaultSchema)
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
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, formation.Tnt, []string{testFormation}).Return(nil)
				return labelDefService
			},
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success when labeldef exists",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, formation.Tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, formation.Tnt, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error when labeldef is missing and can not create it",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("CreateWithFormations", ctx, formation.Tnt, []string{testFormation}).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when can not get labeldef",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(nil, testErr)
				return labelDefRepo
			},
			LabelDefServiceFn:  formation.UnusedLabelDefService,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when labeldef's schema is missing",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(&emptySchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn:  formation.UnusedLabelDefService,
			ExpectedErrMessage: "missing schema",
		},
		{
			Name: "error when validating existing labels against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, formation.Tnt, defaultSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when validating automatic scenario assignment against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, formation.Tnt, defaultSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, formation.Tnt, defaultSchemaLblDef.Key).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when update with version fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(testErr)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, formation.Tnt, defaultSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, formation.Tnt, defaultSchemaLblDef.Key).Return(nil)
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
			actual, err := svc.CreateFormation(ctx, formation.Tnt, in)

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
	ctx = tenant.SaveToContext(ctx, formation.Tnt, formation.ExternalTnt)

	testErr := errors.New("Test error")

	in := model.Formation{
		Name: testFormation,
	}

	expected := &model.Formation{
		Name: testFormation,
	}

	defaultSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
	assert.NoError(t, err)
	defaultSchemaLblDef := formation.FixDefaultScenariosLabelDefinition(formation.Tnt, defaultSchema)

	newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
	assert.NoError(t, err)
	newSchemaLblDef := formation.FixDefaultScenariosLabelDefinition(formation.Tnt, newSchema)

	emptySchemaLblDef := formation.FixDefaultScenariosLabelDefinition(formation.Tnt, defaultSchema)
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
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, formation.Tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, formation.Tnt, model.ScenariosKey).Return(nil)
				return labelDefService
			},
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error when can not get labeldef",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(nil, testErr)
				return labelDefRepo
			},
			LabelDefServiceFn:  formation.UnusedLabelDefService,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when labeldef's schema is missing",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(&emptySchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn:  formation.UnusedLabelDefService,
			ExpectedErrMessage: "missing schema",
		},
		{
			Name: "error when validating existing labels against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, formation.Tnt, model.ScenariosKey).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when validating automatic scenario assignment against the schema",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(&defaultSchemaLblDef, nil)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, formation.Tnt, model.ScenariosKey).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, formation.Tnt, model.ScenariosKey).Return(testErr)
				return labelDefService
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when update with version fails",
			LabelDefRepositoryFn: func() *automock.LabelDefRepository {
				labelDefRepo := &automock.LabelDefRepository{}
				labelDefRepo.On("GetByKey", ctx, formation.Tnt, model.ScenariosKey).Return(&newSchemaLblDef, nil)
				labelDefRepo.On("UpdateWithVersion", ctx, newSchemaLblDef).Return(testErr)
				return labelDefRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefService := &automock.LabelDefService{}
				labelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, formation.Tnt, newSchemaLblDef.Key).Return(nil)
				labelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, formation.Tnt, newSchemaLblDef.Key).Return(nil)
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
			actual, err := svc.DeleteFormation(ctx, formation.Tnt, in)

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
	ctx = tenant.SaveToContext(ctx, formation.Tnt, formation.ExternalTnt)

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
		Tenant:     str.Ptr(formation.Tnt),
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
		Tenant:     str.Ptr(formation.Tnt),
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
		Tenant:         formation.Tnt,
		TargetTenantID: formation.TargetTenant,
	}

	testCases := []struct {
		Name                 string
		UIDServiceFn         func() *automock.UidService
		LabelServiceFn       func() *automock.LabelService
		LabelDefServiceFn    func() *automock.LabelDefService
		TenantServiceFn      func() *automock.TenantService
		AsaRepoFn            func() *automock.AutomaticFormationAssignmentRepository
		AsaServiceFN         func() *automock.AutomaticFormationAssignmentService
		RuntimeRepoFN        func() *automock.RuntimeRepository
		RuntimeContextRepoFn func() *automock.RuntimeContextRepository
		ObjectType           graphql.FormationObjectType
		InputFormation       model.Formation
		ExpectedFormation    *model.Formation
		ExpectedErrMessage   string
	}{
		{
			Name: "success for application if label does not exist",
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
				uidService.On("Generate").Return(formation.FixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, formation.Tnt, formation.FixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedFormation:    expectedFormation,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for application if formation is already added",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, applicationLbl.ID, &applicationLblInput).Return(nil)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedFormation:    expectedFormation,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for application with new formation",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{"test-formation-2"},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation, "test-formation-2"},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputSecondFormation,
			ExpectedFormation:    expectedSecondFormation,
			ExpectedErrMessage:   "",
		},
		{
			Name: "success for runtime if label does not exist",
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
				uidService.On("Generate").Return(formation.FixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, formation.Tnt, formation.FixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       inputFormation,
			ExpectedFormation:    expectedFormation,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for runtime if formation is already added",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, runtimeLbl.ID, &runtimeLblInput).Return(nil)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       inputFormation,
			ExpectedFormation:    expectedFormation,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for runtime with new formation",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{"test-formation-2"},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation, "test-formation-2"},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       inputSecondFormation,
			ExpectedFormation:    expectedSecondFormation,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for tenant",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, objectID).Return(formation.TargetTenant, nil)
				return svc
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", ctx, formation.Tnt).Return(nil)
				labelDefSvc.On("GetAvailableScenarios", ctx, formation.Tnt).Return([]string{testFormation}, nil)

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
				runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenant, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ListAll", ctx, formation.TargetTenant).Return(make([]*model.RuntimeContext, 0), nil).Once()
				return runtimeContextRepo
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
				uidService.On("Generate").Return(formation.FixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, formation.Tnt, formation.FixUUID(), &applicationLblInput).Return(testErr)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:         "error for application while getting label",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &applicationLblInput).Return(nil, testErr)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			AsaServiceFN:         formation.UnusedASAService,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:         "error for application while converting label values to string slice",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(formation.Tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for application while converting label value to string",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &applicationLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(formation.Tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   "cannot cast label value as a string",
		},
		{
			Name:         "error for application when updating label fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, applicationLbl.ID, &applicationLblInput).Return(testErr)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name: "error for runtime when label does not exist and can't create it",
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
				uidService.On("Generate").Return(formation.FixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, formation.Tnt, formation.FixUUID(), &runtimeLblInput).Return(testErr)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:         "error for runtime while getting label",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &runtimeLblInput).Return(nil, testErr)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:         "error for runtime while converting label values to string slice",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(formation.Tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for runtime while converting label value to string",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &runtimeLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(formation.Tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   "cannot cast label value as a string",
		},
		{
			Name:         "error for runtime when updating label fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, runtimeLbl.ID, &runtimeLblInput).Return(testErr)
				return labelService
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:         "error for tenant when tenant conversion fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, objectID).Return("", testErr)
				return svc
			},
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			LabelServiceFn:       formation.UnusedLabelService,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeTenant,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:         "error for tenant when create fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, objectID).Return(formation.TargetTenant, nil)
				return svc
			},
			LabelServiceFn: formation.UnusedLabelService,
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctx, model.AutomaticScenarioAssignment{ScenarioName: testFormation, Tenant: formation.Tnt, TargetTenantID: formation.TargetTenant}).Return(testErr)

				return asaRepo
			},
			AsaServiceFN: formation.UnusedASAService,
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", ctx, formation.Tnt).Return(nil)
				labelDefSvc.On("GetAvailableScenarios", ctx, formation.Tnt).Return([]string{testFormation}, nil)

				return labelDefSvc
			},
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeTenant,
			InputFormation:       inputFormation,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:                 "error when object type is unknown",
			UIDServiceFn:         frmtest.UnusedUUIDService(),
			LabelServiceFn:       formation.UnusedLabelService,
			LabelDefServiceFn:    formation.UnusedLabelDefServiceFn,
			AsaRepoFn:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           "UNKNOWN",
			InputFormation:       inputFormation,
			ExpectedErrMessage:   "unknown formation type",
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

			if testCase.TenantServiceFn != nil {
				tenantSvc = testCase.TenantServiceFn()
			}

			svc := formation.NewService(nil, nil, labelService, uidService, labelDefService, asaRepo, asaService, tenantSvc, runtimeRepo, runtimeContextRepo)

			// WHEN
			actual, err := svc.AssignFormation(ctx, formation.Tnt, objectID, testCase.ObjectType, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, uidService, labelService, asaService, tenantSvc, asaRepo, labelDefService, runtimeRepo, runtimeContextRepo)
		})
	}
}

func TestServiceUnassignFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, formation.Tnt, formation.ExternalTnt)

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
		Tenant:     str.Ptr(formation.Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormation},
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	applicationLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(formation.Tnt),
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
		Tenant:     str.Ptr(formation.Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormation},
		ObjectID:   objectID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(formation.Tnt),
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
		Tenant:         formation.Tnt,
		TargetTenantID: objectID,
	}

	testCases := []struct {
		Name                 string
		UIDServiceFn         func() *automock.UidService
		LabelServiceFn       func() *automock.LabelService
		LabelRepoFn          func() *automock.LabelRepository
		AsaServiceFN         func() *automock.AutomaticFormationAssignmentService
		AsaRepoFN            func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFN        func() *automock.RuntimeRepository
		RuntimeContextRepoFn func() *automock.RuntimeContextRepository
		ObjectType           graphql.FormationObjectType
		InputFormation       model.Formation
		ExpectedFormation    *model.Formation
		ExpectedErrMessage   string
	}{
		{
			Name:         "success for application",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelRepoFn:          formation.UnusedLabelRepo,
			AsaRepoFN:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       in,
			ExpectedFormation:    expected,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for application if formation do not exist",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelRepoFn:          formation.UnusedLabelRepo,
			AsaRepoFN:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       in,
			ExpectedFormation:    expected,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for application when formation is last",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, applicationLblInput).Return(applicationLblSingleFormation, nil)
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, formation.Tnt, model.ApplicationLabelableObject, objectID, model.ScenariosKey).Return(nil)
				return labelRepo
			},
			AsaRepoFN:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       in,
			ExpectedFormation:    expected,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for runtime",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelRepoFn: formation.UnusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, formation.Tnt).Return(nil, nil)
				return asaRepo
			},
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       in,
			ExpectedFormation:    expected,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for runtime when formation is coming from ASA",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, formation.Tnt, model.RuntimeLabelableObject, "123", model.ScenariosKey).Return(nil)

				return labelRepo
			},
			AsaServiceFN: formation.UnusedASAService,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, formation.Tnt).Return([]*model.AutomaticScenarioAssignment{{
					ScenarioName:   formation.ScenarioName,
					Tenant:         formation.Tnt,
					TargetTenantID: formation.Tnt,
				}}, nil)
				return asaRepo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLblSingleFormation, nil)
				return labelService
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("Exists", ctx, formation.Tnt, "123").Return(true, nil)
				return runtimeRepo
			},
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       in,
			ExpectedFormation:    expected,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for runtime if formation do not exist",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLblSingleFormation, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			LabelRepoFn: formation.UnusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, formation.Tnt).Return(nil, nil)
				return asaRepo
			},
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       secondFormation,
			ExpectedFormation:    &secondFormation,
			ExpectedErrMessage:   "",
		},
		{
			Name:         "success for runtime when formation is last",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, runtimeLblInput).Return(runtimeLblSingleFormation, nil)
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, formation.Tnt, model.RuntimeLabelableObject, objectID, model.ScenariosKey).Return(nil)
				return labelRepo
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, formation.Tnt).Return(nil, nil)
				return asaRepo
			},
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       in,
			ExpectedFormation:    expected,
			ExpectedErrMessage:   "",
		},
		{
			Name:           "success for tenant",
			UIDServiceFn:   frmtest.UnusedUUIDService(),
			LabelServiceFn: formation.UnusedLabelService,
			LabelRepoFn:    formation.UnusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}

				asaRepo.On("DeleteForScenarioName", ctx, formation.Tnt, testFormation).Return(nil)

				return asaRepo
			},
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormation).Return(asa, nil)
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
				labelService.On("GetLabel", ctx, formation.Tnt, applicationLblInput).Return(nil, testErr)
				return labelService
			},
			LabelRepoFn:          formation.UnusedLabelRepo,
			AsaRepoFN:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       in,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:         "error for application while converting label values to string slice",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(formation.Tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelRepoFn:          formation.UnusedLabelRepo,
			AsaRepoFN:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       in,
			ExpectedErrMessage:   "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for application while converting label value to string",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, applicationLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(formation.Tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelRepoFn:          formation.UnusedLabelRepo,
			AsaRepoFN:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       in,
			ExpectedErrMessage:   "cannot cast label value as a string",
		},
		{
			Name:         "error for application when formation is last and delete fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, applicationLblInput).Return(applicationLblSingleFormation, nil)
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, formation.Tnt, model.ApplicationLabelableObject, objectID, model.ScenariosKey).Return(testErr)
				return labelRepo
			},
			AsaRepoFN:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       in,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:         "error for application when updating label fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(testErr)
				return labelService
			},
			LabelRepoFn:          formation.UnusedLabelRepo,
			AsaRepoFN:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeApplication,
			InputFormation:       in,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:           "error for runtime when can't get formations that are coming from ASAs",
			UIDServiceFn:   frmtest.UnusedUUIDService(),
			LabelServiceFn: formation.UnusedLabelService,
			LabelRepoFn:    formation.UnusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, formation.Tnt).Return(nil, testErr)
				return asaRepo
			},
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       in,
			ExpectedFormation:    expected,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:         "error for runtime while getting label",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, runtimeLblInput).Return(nil, testErr)
				return labelService
			},
			LabelRepoFn: formation.UnusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, formation.Tnt).Return(nil, nil)
				return asaRepo
			},
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       in,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:         "error for runtime while converting label values to string slice",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(formation.Tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelRepoFn: formation.UnusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, formation.Tnt).Return(nil, nil)
				return asaRepo
			},
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       in,
			ExpectedErrMessage:   "cannot convert label value to slice of strings",
		},
		{
			Name:         "error for runtime while converting label value to string",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, runtimeLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(formation.Tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			LabelRepoFn: formation.UnusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, formation.Tnt).Return(nil, nil)
				return asaRepo
			},
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       in,
			ExpectedErrMessage:   "cannot cast label value as a string",
		},
		{
			Name:         "error for runtime when formation is last and delete fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, runtimeLblInput).Return(runtimeLblSingleFormation, nil)
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, formation.Tnt, model.RuntimeLabelableObject, objectID, model.ScenariosKey).Return(testErr)
				return labelRepo
			},
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, formation.Tnt).Return(nil, nil)
				return asaRepo
			},
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       in,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:         "error for runtime when updating label fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, formation.Tnt, runtimeLblInput).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, formation.Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormation},
					ObjectID:   objectID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(testErr)
				return labelService
			},
			LabelRepoFn: formation.UnusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("ListAll", ctx, formation.Tnt).Return(nil, nil)
				return asaRepo
			},
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			InputFormation:       in,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:           "error for tenant when delete fails",
			UIDServiceFn:   frmtest.UnusedUUIDService(),
			LabelServiceFn: formation.UnusedLabelService,
			LabelRepoFn:    formation.UnusedLabelRepo,
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}

				asaRepo.On("DeleteForScenarioName", ctx, formation.Tnt, testFormation).Return(testErr)

				return asaRepo
			},
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormation).Return(asa, nil)
				return asaService
			},
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           graphql.FormationObjectTypeTenant,
			InputFormation:       in,
			ExpectedErrMessage:   testErr.Error(),
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
			ObjectType:           graphql.FormationObjectTypeTenant,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			InputFormation:       in,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name:                 "error when object type is unknown",
			UIDServiceFn:         frmtest.UnusedUUIDService(),
			LabelServiceFn:       formation.UnusedLabelService,
			LabelRepoFn:          formation.UnusedLabelRepo,
			AsaRepoFN:            formation.UnusedASARepo,
			AsaServiceFN:         formation.UnusedASAService,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectType:           "UNKNOWN",
			InputFormation:       in,
			ExpectedErrMessage:   "unknown formation type",
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
			svc := formation.NewService(nil, labelRepo, labelService, uidService, nil, asaRepo, asaService, nil, runtimeRepo, runtimeContextRepo)

			// WHEN
			actual, err := svc.UnassignFormation(ctx, formation.Tnt, objectID, testCase.ObjectType, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}
			mock.AssertExpectationsForObjects(t, uidService, labelService, asaRepo, asaService, runtimeRepo, runtimeContextRepo)
		})
	}
}

func TestService_CreateAutomaticScenarioAssignment(t *testing.T) {
	ctx := formation.FixCtxWithTenant()
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
				return formation.MockScenarioDefServiceThatReturns([]string{formation.ScenarioName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, formation.FixModel()).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListAll", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(make([]*model.RuntimeContext, 0), nil).Once()
				return runtimeContextRepo
			},
			InputASA:           formation.FixModel(),
			ExpectedASA:        formation.FixModel(),
			ExpectedErrMessage: "",
		},
		{
			Name: "return error when ensuring scenarios for runtimes fails",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return formation.MockScenarioDefServiceThatReturns([]string{formation.ScenarioName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", ctx, formation.FixModel()).Return(nil).Once()
				return mockRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListAll", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, formation.FixError()).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			InputASA:             formation.FixModel(),
			ExpectedASA:          model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:   formation.FixError().Error(),
		},
		{
			Name: "returns error when scenario already has an assignment",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return formation.MockScenarioDefServiceThatReturns([]string{formation.ScenarioName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", mock.Anything, formation.FixModel()).Return(apperrors.NewNotUniqueError("")).Once()
				return mockRepo
			},
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			InputASA:             formation.FixModel(),
			ExpectedASA:          model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:   "a given scenario already has an assignment",
		},
		{
			Name: "returns error when given scenario does not exist",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return formation.MockScenarioDefServiceThatReturns([]string{"completely-different-scenario"})
			},
			AsaRepoFn:            formation.UnusedASARepo,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			InputASA:             formation.FixModel(),
			ExpectedASA:          model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:   apperrors.NewNotFoundError(resource.AutomaticScenarioAssigment, formation.FixModel().ScenarioName).Error(),
		},
		{
			Name: "returns error on persisting in DB",
			LabelDefServiceFn: func() *automock.LabelDefService {
				return formation.MockScenarioDefServiceThatReturns([]string{formation.ScenarioName})
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				mockRepo := &automock.AutomaticFormationAssignmentRepository{}
				mockRepo.On("Create", mock.Anything, formation.FixModel()).Return(formation.FixError()).Once()
				return mockRepo
			},
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			InputASA:             formation.FixModel(),
			ExpectedASA:          model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:   "while persisting Assignment: some error",
		},
		{
			Name: "returns error on ensuring that scenarios label definition exist",
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}
				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, mock.Anything).Return(formation.FixError()).Once()
				return labelDefSvc
			},
			AsaRepoFn:            formation.UnusedASARepo,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			InputASA:             formation.FixModel(),
			ExpectedASA:          model.AutomaticScenarioAssignment{},
			ExpectedErrMessage:   "while ensuring that `scenarios` label definition exist: some error",
		},
		{
			Name: "returns error on getting available scenarios from label definition",
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}
				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, mock.Anything).Return(nil).Once()
				labelDefSvc.On("GetAvailableScenarios", mock.Anything, formation.TenantID.String()).Return(nil, formation.FixError()).Once()
				return labelDefSvc
			},
			AsaRepoFn:            formation.UnusedASARepo,
			RuntimeRepoFN:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			InputASA:             formation.FixModel(),
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

			svc := formation.NewService(nil, nil, nil, nil, labelDefService, asaRepo, nil, tenantSvc, runtimeRepo, runtimeContextRepo)

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
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		_, err := svc.CreateAutomaticScenarioAssignment(context.TODO(), formation.FixModel())

		// THEN
		assert.EqualError(t, err, "cannot read tenant from context")
	})
}

func TestService_DeleteManyASAForSameTargetTenant(t *testing.T) {
	ctx := formation.FixCtxWithTenant()

	scenarioNameA := "scenario-A"
	scenarioNameB := "scenario-B"
	models := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   scenarioNameA,
			TargetTenantID: formation.TargetTenantID,
		},
		{
			ScenarioName:   scenarioNameB,
			TargetTenantID: formation.TargetTenantID,
		},
	}

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForTargetTenant", ctx, formation.TenantID.String(), formation.TargetTenantID).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil)

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(make([]*model.RuntimeContext, 0), nil)
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, runtimeContextRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, runtimeContextRepo)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.NoError(t, err)
	})

	t.Run("return error when listing runtimes fails", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForTargetTenant", ctx, formation.TenantID.String(), formation.TargetTenantID).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, formation.FixError())
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), formation.FixError().Error())
	})

	t.Run("return error when listing runtimes contexts fails", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForTargetTenant", ctx, formation.TenantID.String(), formation.TargetTenantID).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil)

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(nil, formation.FixError())
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, runtimeContextRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, runtimeContextRepo)

		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), formation.FixError().Error())
	})

	t.Run("return error when input slice is empty", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

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
				TargetTenantID: formation.TargetTenantID,
			},
			{
				ScenarioName:   scenarioNameB,
				TargetTenantID: "differentTargetTenantID",
			},
		}

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, modelsWithDifferentSelectors)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "all input items have to have the same target tenant")
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}

		mockRepo.On("DeleteForTargetTenant", ctx, formation.TenantID.String(), formation.TargetTenantID).Return(formation.FixError()).Once()

		defer mock.AssertExpectationsForObjects(t, mockRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, nil, nil)
		// WHEN
		err := svc.DeleteManyASAForSameTargetTenant(ctx, models)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignments: %s", formation.ErrMsg))
	})

	t.Run("returns error when empty tenant", func(t *testing.T) {
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		err := svc.DeleteManyASAForSameTargetTenant(context.TODO(), models)
		require.EqualError(t, err, "cannot read tenant from context")
	})
}

func TestService_DeleteForScenarioName(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		ctx := formation.FixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForScenarioName", ctx, formation.TenantID.String(), formation.ScenarioName).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(make([]*model.RuntimeContext, 0), nil).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, runtimeContextRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, runtimeContextRepo)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(formation.FixCtxWithTenant(), formation.FixModel())

		// THEN
		require.NoError(t, err)
	})

	t.Run("return error when listing runtimes fails", func(t *testing.T) {
		// GIVEN
		ctx := formation.FixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForScenarioName", ctx, formation.TenantID.String(), formation.ScenarioName).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, formation.FixError()).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(ctx, formation.FixModel())

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), formation.FixError().Error())
	})

	t.Run("return error when listing runtimes contexts fails", func(t *testing.T) {
		// GIVEN
		ctx := formation.FixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForScenarioName", ctx, formation.TenantID.String(), formation.ScenarioName).Return(nil).Once()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(nil, formation.FixError())
		defer mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo, runtimeContextRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, runtimeRepo, runtimeContextRepo)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(ctx, formation.FixModel())

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), formation.FixError().Error())
	})

	t.Run("error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(context.TODO(), formation.FixModel())

		// THEN
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		ctx := formation.FixCtxWithTenant()
		mockRepo := &automock.AutomaticFormationAssignmentRepository{}
		mockRepo.On("DeleteForScenarioName", ctx, formation.TenantID.String(), formation.ScenarioName).Return(formation.FixError()).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo)

		svc := formation.NewService(nil, nil, nil, nil, nil, mockRepo, nil, nil, nil, nil)

		// WHEN
		err := svc.DeleteAutomaticScenarioAssignment(formation.FixCtxWithTenant(), formation.FixModel())

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignment: %s", formation.ErrMsg))
	})
}

func TestService_EnsureScenarioAssigned(t *testing.T) {
	selectorScenario := "SELECTOR_SCENARIO"
	in := formation.FixAutomaticScenarioAssigment(selectorScenario)
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

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(runtimeContexts, nil).Once()

		upsertSvc := &automock.LabelService{}
		upsertSvc.On("GetLabel", ctx, formation.TenantID.String(), &labelInputWithoutScenario).Return(&labelWithoutScenario, nil).Once()
		upsertSvc.On("GetLabel", ctx, formation.TenantID.String(), &labelInputWithScenario).Return(&labelWithScenario, nil).Once()
		upsertSvc.On("GetLabel", ctx, formation.TenantID.String(), &rtmCtxLabelInputWithoutScenario).Return(&rtmCtxLabelWithoutScenario, nil).Once()
		upsertSvc.On("GetLabel", ctx, formation.TenantID.String(), &rtmCtxLabelInputWithScenario).Return(&rtmCtxLabelWithScenario, nil).Once()

		upsertSvc.On("UpdateLabel", ctx, formation.TenantID.String(), rtmIDWithoutScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{selectorScenario},
			ObjectID:   rtmIDWithoutScenario,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(nil).Once()
		upsertSvc.On("UpdateLabel", ctx, formation.TenantID.String(), rtmIDWithScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{otherScenario, basicScenario, selectorScenario},
			ObjectID:   rtmIDWithScenario,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(nil).Once()
		upsertSvc.On("UpdateLabel", ctx, formation.TenantID.String(), rtmIDWithoutScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{selectorScenario},
			ObjectID:   rtmCtxIDWithoutScenario,
			ObjectType: model.RuntimeContextLabelableObject,
		}).Return(nil).Once()
		upsertSvc.On("UpdateLabel", ctx, formation.TenantID.String(), rtmIDWithScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{otherScenario, basicScenario, selectorScenario},
			ObjectID:   rtmCtxIDWithScenario,
			ObjectType: model.RuntimeContextLabelableObject,
		}).Return(nil).Once()

		svc := formation.NewService(nil, nil, upsertSvc, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo)

		// WHEN
		err := svc.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo, upsertSvc)
	})

	t.Run("Failed when insert new Label on upsert failed ", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return([]*model.Runtime{{ID: rtmIDWithoutScenario}}, nil).Once()

		upsertSvc := &automock.LabelService{}
		upsertSvc.On("GetLabel", ctx, formation.TenantID.String(), &labelInputWithoutScenario).Return(&labelWithoutScenario, nil).Once()
		upsertSvc.On("UpdateLabel", ctx, formation.TenantID.String(), rtmIDWithoutScenario, &model.LabelInput{
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
		// GIVEN
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, formation.TenantID.String(), &labelInputWithoutScenario).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, nil, nil, nil, runtimeRepo, nil)

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
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, nil)

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
		runtimeRepo.On("ListAll", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(runtimeContexts, nil).Once()

		upsertSvc := &automock.LabelService{}
		upsertSvc.On("GetLabel", ctx, formation.TenantID.String(), &rtmCtxLabelInputWithoutScenario).Return(&rtmCtxLabelWithoutScenario, nil).Once()

		upsertSvc.On("UpdateLabel", ctx, formation.TenantID.String(), rtmIDWithoutScenario, &model.LabelInput{
			Key:        "scenarios",
			Value:      []string{selectorScenario},
			ObjectID:   rtmCtxIDWithoutScenario,
			ObjectType: model.RuntimeContextLabelableObject,
		}).Return(testErr).Once()

		svc := formation.NewService(nil, nil, upsertSvc, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo)

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
		runtimeRepo.On("ListAll", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(runtimeContexts, nil).Once()

		upsertSvc := &automock.LabelService{}
		upsertSvc.On("GetLabel", ctx, formation.TenantID.String(), &rtmCtxLabelInputWithoutScenario).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, upsertSvc, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo)

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
		runtimeRepo.On("ListAll", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo)

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
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(make([]*model.RuntimeContext, 0), nil).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo)

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
	in := formation.FixAutomaticScenarioAssigment(selectorScenario)
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

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		ctx := formation.FixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(rtmContexts, nil).Once()

		asaRepo := &automock.AutomaticFormationAssignmentRepository{}
		asaRepo.On("ListAll", ctx, formation.TenantID.String()).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, formation.TenantID.String(), &labelInput).Return(&label, nil).Once()
		labelService.On("UpdateLabel", ctx, formation.TenantID.String(), rtmID, &model.LabelInput{
			Key:        "scenarios",
			Value:      stringScenarios,
			ObjectID:   rtmID,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(nil).Once()
		labelService.On("GetLabel", ctx, formation.TenantID.String(), &rtmCtxLabelInput).Return(&rtmCtxLabel, nil).Once()
		labelService.On("UpdateLabel", ctx, formation.TenantID.String(), rtmCtxID, &model.LabelInput{
			Key:        "scenarios",
			Value:      stringScenarios,
			ObjectID:   rtmCtxID,
			ObjectType: model.RuntimeContextLabelableObject,
		}).Return(nil).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo, runtimeContextRepo)
	})

	t.Run("Failed when Label Upsert failed ", func(t *testing.T) {
		// GIVEN
		ctx := formation.FixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		asaRepo := &automock.AutomaticFormationAssignmentRepository{}
		asaRepo.On("ListAll", ctx, formation.TenantID.String()).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, formation.TenantID.String(), &labelInput).Return(&label, nil).Once()
		labelService.On("UpdateLabel", ctx, formation.TenantID.String(), rtmID, &model.LabelInput{
			Key:        "scenarios",
			Value:      stringScenarios,
			ObjectID:   rtmID,
			ObjectType: model.RuntimeLabelableObject,
		}).Return(testErr).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, asaRepo, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo, asaRepo)
	})

	t.Run("Failed when ListAll returns error", func(t *testing.T) {
		// GIVEN
		ctx := formation.FixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, runtimeRepo)
	})

	t.Run("Failed when GetScenarioLabelsForRuntimes failed", func(t *testing.T) {
		// GIVEN
		ctx := formation.FixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		asaRepo := &automock.AutomaticFormationAssignmentRepository{}
		asaRepo.On("ListAll", ctx, formation.TenantID.String()).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, formation.TenantID.String(), &labelInput).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, asaRepo, nil, nil, runtimeRepo, nil)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo, asaRepo)
	})

	t.Run("Failed when Label Upsert for runtime context failed ", func(t *testing.T) {
		// GIVEN
		ctx := formation.FixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(rtmContexts, nil).Once()

		asaRepo := &automock.AutomaticFormationAssignmentRepository{}
		asaRepo.On("ListAll", ctx, formation.TenantID.String()).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, formation.TenantID.String(), &rtmCtxLabelInput).Return(&rtmCtxLabel, nil).Once()
		labelService.On("UpdateLabel", ctx, formation.TenantID.String(), rtmCtxID, &model.LabelInput{
			Key:        "scenarios",
			Value:      stringScenarios,
			ObjectID:   rtmCtxID,
			ObjectType: model.RuntimeContextLabelableObject,
		}).Return(testErr).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo, asaRepo, runtimeContextRepo)
	})

	t.Run("Failed when ListAll for runtime context returns error", func(t *testing.T) {
		// GIVEN
		ctx := formation.FixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, runtimeRepo, runtimeContextRepo)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo)
	})

	t.Run("Failed when GetScenarioLabelsForRuntimes for runtime context failed", func(t *testing.T) {
		// GIVEN
		ctx := formation.FixCtxWithTenant()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListOwnedRuntimes", ctx, formation.TargetTenantID, []*labelfilter.LabelFilter(nil)).Return(make([]*model.Runtime, 0), nil).Once()

		runtimeContextRepo := &automock.RuntimeContextRepository{}
		runtimeContextRepo.On("ListAll", ctx, formation.TargetTenantID).Return(rtmContexts, nil).Once()

		asaRepo := &automock.AutomaticFormationAssignmentRepository{}
		asaRepo.On("ListAll", ctx, formation.TenantID.String()).Return(nil, nil)

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, formation.TenantID.String(), &rtmCtxLabelInput).Return(nil, testErr).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo)

		// WHEN
		err := svc.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, labelService, runtimeRepo, asaRepo, runtimeContextRepo)
	})
}

// todo:: retrigger the tests and adapt ListAll to ListOwnedRuntimes where needed
func TestService_MergeScenariosFromInputLabelsAndAssignments_Success(t *testing.T) {
	// GIVEN
	ctx := formation.FixCtxWithTenant()
	differentTargetTenant := "differentTargetTenant"
	runtimeID := "runtimeID"
	labelKey := "key"
	labelValue := "val"

	inputLabels := map[string]interface{}{
		labelKey: labelValue,
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   formation.ScenarioName,
			Tenant:         formation.TenantID.String(),
			TargetTenantID: formation.TargetTenantID,
		},
		{
			ScenarioName:   formation.ScenarioName2,
			Tenant:         formation.TenantID.String(),
			TargetTenantID: differentTargetTenant,
		},
	}

	expectedScenarios := []interface{}{formation.ScenarioName}

	asaRepo := &automock.AutomaticFormationAssignmentRepository{}
	asaRepo.On("ListAll", ctx, formation.TenantID.String()).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", ctx, formation.TargetTenantID, runtimeID).Return(true, nil).Once()
	runtimeRepo.On("Exists", ctx, differentTargetTenant, runtimeID).Return(false, nil).Once()

	svc := formation.NewService(nil, nil, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, nil)

	// WHEN
	actualScenarios, err := svc.MergeScenariosFromInputLabelsAndAssignments(ctx, inputLabels, runtimeID)

	// THEN
	require.NoError(t, err)
	require.ElementsMatch(t, expectedScenarios, actualScenarios)

	mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo)
}

func TestService_MergeScenariosFromInputLabelsAndAssignments_SuccessIfScenariosLabelIsInInput(t *testing.T) {
	// GIVEN
	ctx := formation.FixCtxWithTenant()
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
			ScenarioName:   formation.ScenarioName,
			Tenant:         formation.TenantID.String(),
			TargetTenantID: formation.TargetTenantID,
		},
	}

	expectedScenarios := []interface{}{formation.ScenarioName, scenario}

	asaRepo := &automock.AutomaticFormationAssignmentRepository{}
	asaRepo.On("ListAll", ctx, formation.TenantID.String()).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", ctx, formation.TargetTenantID, runtimeID).Return(true, nil).Once()

	svc := formation.NewService(nil, nil, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, nil)

	// WHEN
	actualScenarios, err := svc.MergeScenariosFromInputLabelsAndAssignments(ctx, inputLabels, runtimeID)

	// THEN
	require.NoError(t, err)
	require.ElementsMatch(t, expectedScenarios, actualScenarios)

	mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo)
}

func TestService_MergeScenariosFromInputLabelsAndAssignments_ReturnsErrorIfListAllFailed(t *testing.T) {
	// GIVEN
	ctx := formation.FixCtxWithTenant()
	testErr := errors.New("testErr")
	labelKey := "key"
	labelValue := "val"

	inputLabels := map[string]interface{}{
		labelKey: labelValue,
	}

	asaRepo := &automock.AutomaticFormationAssignmentRepository{}
	asaRepo.On("ListAll", ctx, formation.TenantID.String()).Return(nil, testErr)

	svc := formation.NewService(nil, nil, nil, nil, nil, asaRepo, nil, nil, nil, nil)

	// WHEN
	_, err := svc.MergeScenariosFromInputLabelsAndAssignments(ctx, inputLabels, "runtimeID")

	// THEN
	require.Error(t, err)

	asaRepo.AssertExpectations(t)
}

func TestService_MergeScenariosFromInputLabelsAndAssignments_ReturnsErrorIfExistsFailed(t *testing.T) {
	// GIVEN
	ctx := formation.FixCtxWithTenant()
	runtimeID := "runtimeID"
	testErr := errors.New("testErr")
	labelKey := "key"
	labelValue := "val"

	inputLabels := map[string]interface{}{
		labelKey: labelValue,
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   formation.ScenarioName,
			Tenant:         formation.TenantID.String(),
			TargetTenantID: formation.TargetTenantID,
		},
	}

	asaRepo := &automock.AutomaticFormationAssignmentRepository{}
	asaRepo.On("ListAll", ctx, formation.TenantID.String()).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", ctx, formation.TargetTenantID, runtimeID).Return(false, testErr).Once()

	svc := formation.NewService(nil, nil, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, nil)

	// WHEN
	_, err := svc.MergeScenariosFromInputLabelsAndAssignments(ctx, inputLabels, runtimeID)

	// THEN
	require.Error(t, err)

	mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo)
}

func TestService_MergeScenariosFromInputLabelsAndAssignments_ReturnsErrorIfScenariosFromInputWereNotInterfaceSlice(t *testing.T) {
	//  GIVEN
	ctx := formation.FixCtxWithTenant()
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
			ScenarioName:   formation.ScenarioName,
			Tenant:         formation.TenantID.String(),
			TargetTenantID: formation.TargetTenantID,
		},
	}

	asaRepo := &automock.AutomaticFormationAssignmentRepository{}
	asaRepo.On("ListAll", ctx, formation.TenantID.String()).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", ctx, formation.TargetTenantID, runtimeID).Return(true, nil).Once()

	svc := formation.NewService(nil, nil, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, nil)

	// WHEN
	_, err := svc.MergeScenariosFromInputLabelsAndAssignments(ctx, inputLabels, runtimeID)

	// THEN
	require.Error(t, err)

	mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo)
}

func TestService_GetScenariosFromMatchingASAs(t *testing.T) {
	ctx := formation.FixCtxWithTenant()
	testErr := errors.New(formation.ErrMsg)
	testScenarios := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   formation.ScenarioName,
			Tenant:         formation.TenantID.String(),
			TargetTenantID: formation.TargetTenantID,
		},
		{
			ScenarioName:   formation.ScenarioName2,
			Tenant:         formation.TenantID2,
			TargetTenantID: formation.TargetTenantID2,
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
				repo.On("ListAll", ctx, formation.TenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, formation.TargetTenantID, formation.RuntimeID).Return(true, nil)
				repo.On("Exists", ctx, formation.TargetTenantID2, formation.RuntimeID).Return(false, nil)
				return repo
			},
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectID:             formation.RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedError:        nil,
			ExpectedScenarios:    []string{formation.ScenarioName},
		},
		{
			Name: "Success for runtime context",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, formation.TenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeRepoFn: formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, formation.TargetTenantID, formation.RuntimeContextID).Return(true, nil)
				repo.On("Exists", ctx, formation.TargetTenantID2, formation.RuntimeContextID).Return(false, nil)
				return repo
			},
			ObjectID:          formation.RuntimeContextID,
			ObjectType:        graphql.FormationObjectTypeRuntimeContext,
			ExpectedError:     nil,
			ExpectedScenarios: []string{formation.ScenarioName},
		},
		{
			Name:                     "Returns error when formation object type is not expected",
			ScenarioAssignmentRepoFn: formation.UnusedASARepo,
			RuntimeRepoFn:            formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn:     formation.UnusedRuntimeContextRepo,
			ObjectID:                 formation.RuntimeID,
			ObjectType:               graphql.FormationObjectTypeTenant,
			ExpectedError:            errors.Errorf("unexpected formation object type %q", graphql.FormationObjectTypeTenant),
			ExpectedScenarios:        nil,
		},
		{
			Name: "Returns error when can't list ASAs",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, formation.TenantID.String()).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFn:        formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectID:             formation.RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedError:        testErr,
			ExpectedScenarios:    nil,
		},
		{
			Name: "Returns error when checking if asa matches runtime",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, formation.TenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, formation.TargetTenantID, formation.RuntimeID).Return(false, testErr)
				return repo
			},
			RuntimeContextRepoFn: formation.UnusedRuntimeContextRepo,
			ObjectID:             formation.RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedError:        testErr,
			ExpectedScenarios:    nil,
		},
		{
			Name: "Returns error when checking if asa matches runtime context",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, formation.TenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeRepoFn: formation.UnusedRuntimeRepo,
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, formation.TargetTenantID, formation.RuntimeContextID).Return(false, testErr)
				return repo
			},
			ObjectID:          formation.RuntimeContextID,
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

			svc := formation.NewService(nil, nil, nil, nil, nil, asaRepo, nil, nil, runtimeRepo, runtimeContextRepo)

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
		labelService.On("GetLabel", ctx, formation.TenantID.String(), labelInput).Return(label, nil).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		formations, err := svc.GetFormationsForObject(ctx, formation.TenantID.String(), model.RuntimeLabelableObject, id)

		// THEN
		require.NoError(t, err)
		require.ElementsMatch(t, formations, scenarios)
		mock.AssertExpectationsForObjects(t, labelService)
	})

	t.Run("Returns error while getting label", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		labelService := &automock.LabelService{}
		labelService.On("GetLabel", ctx, formation.TenantID.String(), labelInput).Return(nil, errors.New(testErr)).Once()

		svc := formation.NewService(nil, nil, labelService, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		formations, err := svc.GetFormationsForObject(ctx, formation.TenantID.String(), model.RuntimeLabelableObject, id)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "while fetching scenario label for")
		require.Nil(t, formations)
		mock.AssertExpectationsForObjects(t, labelService)
	})
}
