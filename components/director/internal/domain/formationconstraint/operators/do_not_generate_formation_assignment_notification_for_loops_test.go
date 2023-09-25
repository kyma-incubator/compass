package operators_test

import (
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstraintOperators_DoNotGenerateFormationAssignmentNotificationForLoops(t *testing.T) {
	testCases := []struct {
		Name                  string
		Input                 operators.OperatorInput
		LabelSvc              func() *automock.LabelService
		RuntimeContextRepo    func() *automock.RuntimeContextRepo
		FormationTemplateRepo func() *automock.FormationTemplateRepo
		ExpectedResult        bool
		ExpectedErrorMsg      string
	}{
		{
			Name:           "Success for a system when notifications should NOT be skipped when notification is not about a loop",
			Input:          in,
			ExpectedResult: true,
		},
		{
			Name:  "Success for a system when notifications should be skipped",
			Input: inLoop,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, inputAppID, applicationTypeLabel).Return(&model.Label{Value: inputAppType}, nil).Once()
				return svc
			},
			ExpectedResult: false,
		},
		{
			Name: "Success when all notifications should be skipped",
			Input: &formationconstraintpkg.DoNotGenerateFormationAssignmentNotificationInput{
				ResourceType:       model.ApplicationResourceType,
				ResourceSubtype:    inputAppType,
				ResourceID:         inputAppID,
				SourceResourceType: model.ApplicationResourceType,
				SourceResourceID:   inputAppID,
				Tenant:             testTenantID,
			},
			ExpectedResult: false,
		},
		{
			Name:  "Success for a formation type that is excepted and notifications should NOT be skipped",
			Input: inWithFormationTypeExceptionLoop,
			LabelSvc: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			FormationTemplateRepo: func() *automock.FormationTemplateRepo {
				repo := &automock.FormationTemplateRepo{}
				repo.On("Get", ctx, formationTemplateID).Return(&model.FormationTemplate{Name: formationType}, nil).Once()
				return repo
			},
			ExpectedResult: true,
		},
		{
			Name:     "Error when get formation type fails",
			Input:    inWithFormationTypeExceptionLoop,
			LabelSvc: unusedLabelService,
			FormationTemplateRepo: func() *automock.FormationTemplateRepo {
				repo := &automock.FormationTemplateRepo{}
				repo.On("Get", ctx, formationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:  "Success for a system that is excepted and notifications should NOT be skipped",
			Input: inLoop,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, inputAppID, applicationTypeLabel).Return(&model.Label{Value: exceptType}, nil).Once()
				return svc
			},
			ExpectedResult: true,
		},
		{
			Name:  "Error for a system if get label fail",
			Input: inLoop,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, inputAppID, applicationTypeLabel).Return(nil, testErr).Once()
				return svc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "Error when input is with wrong type",
			Input:            inputTenantResourceType,
			ExpectedResult:   false,
			ExpectedErrorMsg: "Incompatible input for operator",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelSvc := unusedLabelService()
			if testCase.LabelSvc != nil {
				labelSvc = testCase.LabelSvc()
			}
			runtimeContextRepo := unusedRuntimeContextRepo()
			if testCase.RuntimeContextRepo != nil {
				runtimeContextRepo = testCase.RuntimeContextRepo()
			}

			formationTemplateRepo := unusedFormationTemplateRepo()
			if testCase.FormationTemplateRepo != nil {
				formationTemplateRepo = testCase.FormationTemplateRepo()
			}
			engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, labelSvc, nil, runtimeContextRepo, formationTemplateRepo, nil, runtimeType, applicationType)

			result, err := engine.DoNotGenerateFormationAssignmentNotificationForLoops(ctx, testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.Equal(t, testCase.ExpectedResult, result)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, formationTemplateRepo, runtimeContextRepo, labelSvc)
		})
	}
}
