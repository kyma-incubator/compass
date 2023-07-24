package operators_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConstraintOperators_DoNotGenerateFormationAssignmentNotification(t *testing.T) {

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
			Name:  "Success for a system when notifications should be skipped",
			Input: in,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, appID, applicationTypeLabel).Return(&model.Label{Value: inputAppType}, nil).Once()
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
				SourceResourceID:   appID,
				Tenant:             testTenantID,
			},
			LabelSvc:       UnusedLabelService,
			ExpectedResult: false,
		},
		{
			Name:  "Success for a formation type that is excepted and notifications should NOT be skipped",
			Input: inWithFormationTypeException,
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
			Input:    inWithFormationTypeException,
			LabelSvc: UnusedLabelService,
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
			Input: in,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, appID, applicationTypeLabel).Return(&model.Label{Value: exceptType}, nil).Once()
				return svc
			},
			ExpectedResult: true,
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
			ExpectedResult: false,
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
			ExpectedResult: false,
		},
		{
			Name:     "Error for runtime context when get rt ctx fails",
			Input:    runtimeContextIn,
			LabelSvc: UnusedLabelService,
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
			var formationTemplateRepo *automock.FormationTemplateRepo
			if testCase.FormationTemplateRepo != nil {
				formationTemplateRepo = testCase.FormationTemplateRepo()
			}
			engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, labelSvc, nil, runtimeContextRepo, formationTemplateRepo, nil, runtimeType, applicationType)

			result, err := engine.DoNotGenerateFormationAssignmentNotification(ctx, testCase.Input)

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
			if formationTemplateRepo != nil {
				formationTemplateRepo.AssertExpectations(t)
			}
		})
	}
}
