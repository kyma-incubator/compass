package formation_test

import (
	"context"
	"errors"
	"testing"

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
	tnt          = "tenant"
	targetTenant = "targetTenant"
	externalTnt  = "external-tnt"
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

			svc := formation.NewService(lblDefRepo, nil, nil, nil, lblDefService, nil, nil, nil)

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

			svc := formation.NewService(lblDefRepo, nil, nil, nil, lblDefService, nil, nil, nil)

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

	subaccountInput := func() model.BusinessTenantMappingInput {
		return model.BusinessTenantMappingInput{
			ExternalTenant: objectID,
			Parent:         tnt,
			Type:           "subaccount",
			Provider:       "lazilyWhileFormationCreation",
		}
	}

	testCases := []struct {
		Name               string
		UIDServiceFn       func() *automock.UidService
		LabelServiceFn     func() *automock.LabelService
		TenantServiceFn    func() *automock.TenantService
		AsaServiceFN       func() *automock.AutomaticFormationAssignmentService
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
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
				svc.On("CreateManyIfNotExists", ctx, subaccountInput()).Return(nil).Once()
				svc.On("GetInternalTenant", ctx, objectID).Return(targetTenant, nil)
				return svc
			},
			LabelServiceFn: formation.UnusedLabelService,
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("Create", ctx, asa).Return(asa, nil)
				return asaService
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
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
			AsaServiceFN:       formation.UnusedASAService,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for tenant when tenant creation fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("CreateManyIfNotExists", ctx, subaccountInput()).Return(testErr).Once()
				return svc
			},
			LabelServiceFn:     formation.UnusedLabelService,
			AsaServiceFN:       formation.UnusedASAService,
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for tenant when tenant conversion fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("CreateManyIfNotExists", ctx, subaccountInput()).Return(nil).Once()
				svc.On("GetInternalTenant", ctx, objectID).Return("", testErr)
				return svc
			},
			LabelServiceFn:     formation.UnusedLabelService,
			AsaServiceFN:       formation.UnusedASAService,
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "error for tenant when create fails",
			UIDServiceFn: frmtest.UnusedUUIDService(),
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("CreateManyIfNotExists", ctx, subaccountInput()).Return(nil).Once()
				svc.On("GetInternalTenant", ctx, objectID).Return(targetTenant, nil)
				return svc
			},
			LabelServiceFn: formation.UnusedLabelService,
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("Create", ctx, asa).Return(asa, testErr)
				return asaService
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "error when object type is unknown",
			UIDServiceFn:       frmtest.UnusedUUIDService(),
			LabelServiceFn:     formation.UnusedLabelService,
			AsaServiceFN:       formation.UnusedASAService,
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
			asaService := testCase.AsaServiceFN()
			tenantSvc := &automock.TenantService{}
			if testCase.TenantServiceFn != nil {
				tenantSvc = testCase.TenantServiceFn()
			}

			svc := formation.NewService(nil, nil, labelService, uidService, nil, asaService, tenantSvc, nil)

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

			mock.AssertExpectationsForObjects(t, uidService, labelService, asaService, tenantSvc)
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
		EngineFn           func() *automock.ScenarioAssignmentEngine
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
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
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
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
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
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
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
			AsaServiceFN: formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
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
			AsaServiceFN:   formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return([]string{testFormation}, nil)
				return engine
			},
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
			AsaServiceFN: formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
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
			AsaServiceFN: formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
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
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormation).Return(asa, nil)
				asaService.On("Delete", ctx, asa).Return(nil)
				return asaService
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
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
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
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
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
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
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
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
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
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
			ObjectType:         graphql.FormationObjectTypeApplication,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:           "error for runtime when can't get formations that are coming from ASAs",
			UIDServiceFn:   frmtest.UnusedUUIDService(),
			LabelServiceFn: formation.UnusedLabelService,
			LabelRepoFn:    formation.UnusedLabelRepo,
			AsaServiceFN:   formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, testErr)
				return engine
			},
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
			AsaServiceFN: formation.UnusedASAService,
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
			AsaServiceFN: formation.UnusedASAService,
			EngineFn: func() *automock.ScenarioAssignmentEngine {
				engine := &automock.ScenarioAssignmentEngine{}
				engine.On("GetScenariosFromMatchingASAs", ctx, objectID).Return(nil, nil)
				return engine
			},
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
			LabelRepoFn:  formation.UnusedLabelRepo,
			AsaServiceFN: formation.UnusedASAService,
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
			AsaServiceFN: formation.UnusedASAService,
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
			LabelRepoFn:  formation.UnusedLabelRepo,
			AsaServiceFN: formation.UnusedASAService,
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
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormation).Return(asa, nil)
				asaService.On("Delete", ctx, asa).Return(testErr)
				return asaService
			},
			EngineFn:           formation.UnusedEngine,
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:           "error for tenant when delete fails",
			UIDServiceFn:   frmtest.UnusedUUIDService(),
			LabelServiceFn: formation.UnusedLabelService,
			LabelRepoFn:    formation.UnusedLabelRepo,
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormation).Return(model.AutomaticScenarioAssignment{}, testErr)
				return asaService
			},
			EngineFn:           formation.UnusedEngine,
			ObjectType:         graphql.FormationObjectTypeTenant,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "error when object type is unknown",
			UIDServiceFn:       frmtest.UnusedUUIDService(),
			LabelServiceFn:     formation.UnusedLabelService,
			LabelRepoFn:        formation.UnusedLabelRepo,
			AsaServiceFN:       formation.UnusedASAService,
			EngineFn:           formation.UnusedEngine,
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
			asaService := testCase.AsaServiceFN()
			engine := testCase.EngineFn()

			svc := formation.NewService(nil, labelRepo, labelService, uidService, nil, asaService, nil, engine)

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
			mock.AssertExpectationsForObjects(t, uidService, labelService, asaService)
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
