package formationconstraint_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConstraintOperators_IsNotAssignedToAnyFormationOfType(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	testCases := []struct {
		Name                  string
		Input                 formationconstraint.OperatorInput
		TenantServiceFn       func() *automock.TenantService
		AsaServiceFn          func() *automock.AutomaticScenarioAssignmentService
		LabelRepositoryFn     func() *automock.LabelRepository
		FormationRepositoryFn func() *automock.FormationRepository
		ExpectedResult        bool
		ExpectedErrorMsg      string
	}{
		{
			Name:  "Success for tenant when participating in formation",
			Input: inputTenantResourceType,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, testID).Return(testInternalTenantID, nil).Once()
				return svc
			},
			AsaServiceFn: func() *automock.AutomaticScenarioAssignmentService {
				svc := &automock.AutomaticScenarioAssignmentService{}
				svc.On("ListForTargetTenant", ctx, testInternalTenantID).Return(assignments, nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListByFormationNames", ctx, []string{scenario}, testTenantID).Return(formations, nil).Once()
				return repo
			},
			ExpectedResult: true,
		},
		{
			Name:  "Success for tenant when not participating in any formations",
			Input: inputTenantResourceType,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, testID).Return(testInternalTenantID, nil).Once()
				return svc
			},
			AsaServiceFn: func() *automock.AutomaticScenarioAssignmentService {
				svc := &automock.AutomaticScenarioAssignmentService{}
				svc.On("ListForTargetTenant", ctx, testInternalTenantID).Return(emptyAssignments, nil).Once()
				return svc
			},
			ExpectedResult: true,
		},
		{
			Name:  "Success for tenant when already participating in formations of given type",
			Input: inputTenantResourceType,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, testID).Return(testInternalTenantID, nil).Once()
				return svc
			},
			AsaServiceFn: func() *automock.AutomaticScenarioAssignmentService {
				svc := &automock.AutomaticScenarioAssignmentService{}
				svc.On("ListForTargetTenant", ctx, testInternalTenantID).Return(assignments, nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListByFormationNames", ctx, []string{scenario}, testTenantID).Return(formations2, nil).Once()
				return repo
			},
			ExpectedResult: false,
		},
		{
			Name:  "Error when listing formations by names",
			Input: inputTenantResourceType,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, testID).Return(testInternalTenantID, nil).Once()
				return svc
			},
			AsaServiceFn: func() *automock.AutomaticScenarioAssignmentService {
				svc := &automock.AutomaticScenarioAssignmentService{}
				svc.On("ListForTargetTenant", ctx, testInternalTenantID).Return(assignments, nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListByFormationNames", ctx, []string{scenario}, testTenantID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:  "Error when listing ASAs",
			Input: inputTenantResourceType,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, testID).Return(testInternalTenantID, nil).Once()
				return svc
			},
			AsaServiceFn: func() *automock.AutomaticScenarioAssignmentService {
				svc := &automock.AutomaticScenarioAssignmentService{}
				svc.On("ListForTargetTenant", ctx, testInternalTenantID).Return(nil, testErr).Once()
				return svc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:  "Error when getting internal tenant",
			Input: inputTenantResourceType,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, testID).Return("", testErr).Once()
				return svc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:  "Success for Application when participating in formation",
			Input: inputApplicationResourceType,
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, testID, model.ScenariosKey).Return(scenariosLabel, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListByFormationNames", ctx, []string{scenario}, testTenantID).Return(formations, nil).Once()
				return repo
			},
			ExpectedResult: true,
		},
		{
			Name:  "Success for Application when not participating in formation",
			Input: inputApplicationResourceType,
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, testID, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError("", testID)).Once()
				return repo
			},
			ExpectedResult: true,
		},
		{
			Name:  "Success for Application when already participating in formation of the given type",
			Input: inputApplicationResourceType,
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, testID, model.ScenariosKey).Return(scenariosLabel, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListByFormationNames", ctx, []string{scenario}, testTenantID).Return(formations2, nil).Once()
				return repo
			},
			ExpectedResult: false,
		},
		{
			Name:  "Success for Application when already participating in formation of the given type but the subtype is part of the exception",
			Input: inputApplicationResourceTypeWithSubtypeThatIsException,
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, testID, model.ScenariosKey).Return(scenariosLabel, nil).Once()
				return repo
			},
			ExpectedResult: true,
		},
		{
			Name:  "Error when converting label value",
			Input: inputApplicationResourceType,
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, testID, model.ScenariosKey).Return(scenariosLabelInvalidValue, nil).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "cannot convert label value to slice of strings",
		},
		{
			Name:  "Error when getting label",
			Input: inputApplicationResourceType,
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, testID, model.ScenariosKey).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "Error when input is incompatible",
			Input:            "incompatible",
			ExpectedResult:   false,
			ExpectedErrorMsg: "Incompatible input",
		},
		{
			Name:             "Error when type is unsupported",
			Input:            inputRuntimeResourceType,
			ExpectedResult:   false,
			ExpectedErrorMsg: "Unsupported resource type \"RUNTIME\"",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := UnusedTenantService()
			if testCase.TenantServiceFn != nil {
				tenantSvc = testCase.TenantServiceFn()
			}
			asaSvc := UnusedASAService()
			if testCase.AsaServiceFn != nil {
				asaSvc = testCase.AsaServiceFn()
			}
			labelRepo := UnusedLabelRepo()
			if testCase.LabelRepositoryFn != nil {
				labelRepo = testCase.LabelRepositoryFn()
			}
			formationRepo := UnusedFormationRepo()
			if testCase.FormationRepositoryFn != nil {
				formationRepo = testCase.FormationRepositoryFn()
			}

			engine := formationconstraint.NewConstraintEngine(nil, tenantSvc, asaSvc, formationRepo, labelRepo, nil, nil, nil, runtimeType, applicationType)
			// WHEN
			result, err := engine.IsNotAssignedToAnyFormationOfType(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.Equal(t, testCase.ExpectedResult, result)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tenantSvc, asaSvc, formationRepo, labelRepo)
		})
	}
}

func TestConstraintOperators_DoesNotContainResourceOfSubtype(t *testing.T) {
	ctx := context.TODO()
	testErr := errors.New("test error")

	applicationTypeLabel := "applicationType"
	inputAppType := "input-type"
	inputAppID := "eb2d5110-ca3a-11ed-afa1-0242ac120002"
	appID := "b55131c4-ca3a-11ed-afa1-0242ac120002"

	in := &formationconstraintpkg.DoesNotContainResourceOfSubtypeInput{
		FormationName:   scenario,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: inputAppType,
		ResourceID:      inputAppID,
		Tenant:          testTenantID,
	}

	testCases := []struct {
		Name             string
		Input            formationconstraint.OperatorInput
		LabelSvc         func() *automock.LabelService
		ApplicationRepo  func() *automock.ApplicationRepository
		ExpectedResult   bool
		ExpectedErrorMsg string
	}{
		{
			Name:  "Success for a system when there is NO such system type in that formation",
			Input: in,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, appID, applicationTypeLabel).Return(&model.Label{Value: "different-type"}, nil).Once()
				return svc
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, testTenantID, []string{scenario}).Return([]*model.Application{{BaseEntity: &model.BaseEntity{ID: appID}}}, nil).Once()
				return repo
			},
			ExpectedResult:   true,
			ExpectedErrorMsg: "",
		},
		{
			Name:  "Success for a system when there is such system type in that formation",
			Input: in,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, appID, applicationTypeLabel).Return(&model.Label{Value: inputAppType}, nil).Once()
				return svc
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, testTenantID, []string{scenario}).Return([]*model.Application{{BaseEntity: &model.BaseEntity{ID: appID}}}, nil).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "",
		},
		{
			Name:  "Returns error when can't get the label of another system in the formation",
			Input: in,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, appID, applicationTypeLabel).Return(nil, testErr).Once()
				return svc
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, testTenantID, []string{scenario}).Return([]*model.Application{{BaseEntity: &model.BaseEntity{ID: appID}}}, nil).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:     "Returns error when can't get the applications in the formation",
			Input:    in,
			LabelSvc: UnusedLabelService,
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, testTenantID, []string{scenario}).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "Returns error when the operator input is incompatible",
			Input:            "incompatible",
			LabelSvc:         UnusedLabelService,
			ApplicationRepo:  UnusedApplicationRepo,
			ExpectedResult:   false,
			ExpectedErrorMsg: "Incompatible input",
		},
		{
			Name: "Returns error when the resource type is unknown",
			Input: &formationconstraintpkg.DoesNotContainResourceOfSubtypeInput{
				FormationName:   scenario,
				ResourceType:    "Unknown",
				ResourceSubtype: inputAppType,
				ResourceID:      inputAppID,
				Tenant:          testTenantID,
			},
			LabelSvc:         UnusedLabelService,
			ApplicationRepo:  UnusedApplicationRepo,
			ExpectedResult:   false,
			ExpectedErrorMsg: "Unsupported resource type",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelSvc := testCase.LabelSvc()
			appRepo := testCase.ApplicationRepo()
			engine := formationconstraint.NewConstraintEngine(nil, nil, nil, nil, nil, labelSvc, appRepo, nil, runtimeType, applicationType)

			result, err := engine.DoesNotContainResourceOfSubtype(ctx, testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.Equal(t, testCase.ExpectedResult, result)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, labelSvc, appRepo)
		})
	}
}

func TestConstraintOperators_DoNotSendNotification(t *testing.T) {
	ctx := context.TODO()
	testErr := errors.New("test error")

	applicationTypeLabel := "applicationType"
	runtimeTypeLabel := "runtimeType"
	inputAppType := "input-type"
	inputAppID := "eb2d5110-ca3a-11ed-afa1-0242ac120002"
	appID := "b55131c4-ca3a-11ed-afa1-0242ac120002"
	runtimeID := "c66341c4-ca3a-11ed-afa1-0242ac120564"
	runtimeCtxID := "f7156h4-ca3a-11ed-afa1-0242ac121237"

	exceptType := "except-type"

	in := &formationconstraintpkg.DoNotSendNotificationInput{
		ResourceType:       model.ApplicationResourceType,
		ResourceSubtype:    inputAppType,
		ResourceID:         inputAppID,
		SourceResourceType: model.ApplicationResourceType,
		SourceResourceID:   appID,
		Tenant:             testTenantID,
		ExceptSubtypes:     []string{exceptType},
	}

	runtimeIn := &formationconstraintpkg.DoNotSendNotificationInput{
		ResourceType:       model.ApplicationResourceType,
		ResourceSubtype:    inputAppType,
		ResourceID:         inputAppID,
		SourceResourceType: model.RuntimeResourceType,
		SourceResourceID:   runtimeID,
		Tenant:             testTenantID,
		ExceptSubtypes:     []string{exceptType},
	}

	runtimeContextIn := &formationconstraintpkg.DoNotSendNotificationInput{
		ResourceType:       model.ApplicationResourceType,
		ResourceSubtype:    inputAppType,
		ResourceID:         inputAppID,
		SourceResourceType: model.RuntimeContextResourceType,
		SourceResourceID:   runtimeCtxID,
		Tenant:             testTenantID,
		ExceptSubtypes:     []string{exceptType},
	}

	testCases := []struct {
		Name               string
		Input              formationconstraint.OperatorInput
		LabelSvc           func() *automock.LabelService
		RuntimeContextRepo func() *automock.RuntimeContextRepo
		ExpectedResult     bool
		ExpectedErrorMsg   string
	}{
		{
			Name:  "Success for a system when notifications should be skipped",
			Input: in,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, appID, applicationTypeLabel).Return(&model.Label{Value: inputAppType}, nil).Once()
				return svc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "",
		},
		{
			Name: "Success when all notifications should be skipped",
			Input: &formationconstraintpkg.DoNotSendNotificationInput{
				ResourceType:       model.ApplicationResourceType,
				ResourceSubtype:    inputAppType,
				ResourceID:         inputAppID,
				SourceResourceType: model.ApplicationResourceType,
				SourceResourceID:   appID,
				Tenant:             testTenantID,
			},
			LabelSvc: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "",
		},
		{
			Name:  "Success for a system that is excepted and notifications should NOT be skipped",
			Input: in,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, appID, applicationTypeLabel).Return(&model.Label{Value: exceptType}, nil).Once()
				return svc
			},
			ExpectedResult:   true,
			ExpectedErrorMsg: "",
		},
		{
			Name:  "Error for a system if get label fail",
			Input: in,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, appID, applicationTypeLabel).Return(nil, testErr).Once()
				return svc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:  "Success for runtime when notifications should be skipped",
			Input: runtimeIn,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.RuntimeLabelableObject, runtimeID, runtimeTypeLabel).Return(&model.Label{Value: inputAppType}, nil).Once()
				return svc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "",
		},
		{
			Name:  "Error for runtime if get label fail",
			Input: runtimeIn,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.RuntimeLabelableObject, runtimeID, runtimeTypeLabel).Return(nil, testErr).Once()
				return svc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:  "Success for runtime context when notifications should be skipped",
			Input: runtimeContextIn,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.RuntimeLabelableObject, runtimeID, runtimeTypeLabel).Return(&model.Label{Value: inputAppType}, nil).Once()
				return svc
			},
			RuntimeContextRepo: func() *automock.RuntimeContextRepo {
				repo := &automock.RuntimeContextRepo{}
				repo.On("GetByID", ctx, testTenantID, runtimeCtxID).Return(&model.RuntimeContext{RuntimeID: runtimeID}, nil).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "",
		},
		{
			Name:  "Error for runtime context when get rt ctx fails",
			Input: runtimeContextIn,
			LabelSvc: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			RuntimeContextRepo: func() *automock.RuntimeContextRepo {
				repo := &automock.RuntimeContextRepo{}
				repo.On("GetByID", ctx, testTenantID, runtimeCtxID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:  "Error for runtime context if runtime get label fail",
			Input: runtimeContextIn,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.RuntimeLabelableObject, runtimeID, runtimeTypeLabel).Return(nil, testErr).Once()
				return svc
			},
			RuntimeContextRepo: func() *automock.RuntimeContextRepo {
				repo := &automock.RuntimeContextRepo{}
				repo.On("GetByID", ctx, testTenantID, runtimeCtxID).Return(&model.RuntimeContext{RuntimeID: runtimeID}, nil).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelSvc := testCase.LabelSvc()
			var runtimeContextRepo *automock.RuntimeContextRepo
			if testCase.RuntimeContextRepo != nil {
				runtimeContextRepo = testCase.RuntimeContextRepo()
			}
			engine := formationconstraint.NewConstraintEngine(nil, nil, nil, nil, nil, labelSvc, nil, runtimeContextRepo, runtimeType, applicationType)

			result, err := engine.DoNotSendNotification(ctx, testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.Equal(t, testCase.ExpectedResult, result)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, labelSvc)
			if runtimeContextRepo != nil {
				runtimeContextRepo.AssertExpectations(t)
			}
		})
	}
}
