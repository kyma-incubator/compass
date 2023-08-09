package operators_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConstraintOperators_IsNotAssignedToAnyFormationOfType(t *testing.T) {
	testCases := []struct {
		Name                  string
		Input                 operators.OperatorInput
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

			engine := operators.NewConstraintEngine(nil, nil, tenantSvc, asaSvc, nil, nil, formationRepo, labelRepo, nil, nil, nil, nil, nil, runtimeType, applicationType)
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
