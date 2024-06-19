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

func TestConstraintOperators_DoesNotContainResourceOfSubtype(t *testing.T) {
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

	appIDs := []string{appID}
	apps := []*model.Application{{BaseEntity: &model.BaseEntity{ID: appID}}}

	testCases := []struct {
		Name             string
		Input            operators.OperatorInput
		LabelSvc         func() *automock.LabelService
		FormationRepo    func() *automock.FormationRepository
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
			FormationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListObjectIDsOfTypeForFormations", ctx, testTenantID, []string{scenario}, model.FormationAssignmentTypeApplication).Return(appIDs, nil).Once()
				return repo
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllByIDs", ctx, testTenantID, appIDs).Return(apps, nil).Once()
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
			FormationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListObjectIDsOfTypeForFormations", ctx, testTenantID, []string{scenario}, model.FormationAssignmentTypeApplication).Return(appIDs, nil).Once()
				return repo
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllByIDs", ctx, testTenantID, appIDs).Return(apps, nil).Once()
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
			FormationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListObjectIDsOfTypeForFormations", ctx, testTenantID, []string{scenario}, model.FormationAssignmentTypeApplication).Return(appIDs, nil).Once()
				return repo
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllByIDs", ctx, testTenantID, appIDs).Return(apps, nil).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:  "Returns error when can't list applications by IDs",
			Input: in,
			FormationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListObjectIDsOfTypeForFormations", ctx, testTenantID, []string{scenario}, model.FormationAssignmentTypeApplication).Return(appIDs, nil).Once()
				return repo
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllByIDs", ctx, testTenantID, appIDs).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:  "Returns error when can't list object IDs for the formation",
			Input: in,
			FormationRepo: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("ListObjectIDsOfTypeForFormations", ctx, testTenantID, []string{scenario}, model.FormationAssignmentTypeApplication).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "Returns error when the operator input is incompatible",
			Input:            "incompatible",
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
			ExpectedResult:   false,
			ExpectedErrorMsg: "Unsupported resource type",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelSvc := &automock.LabelService{}
			if testCase.LabelSvc != nil {
				labelSvc = testCase.LabelSvc()
			}
			appRepo := &automock.ApplicationRepository{}
			if testCase.ApplicationRepo != nil {
				appRepo = testCase.ApplicationRepo()
			}
			formationRepo := &automock.FormationRepository{}
			if testCase.FormationRepo != nil {
				formationRepo = testCase.FormationRepo()
			}
			engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, formationRepo, nil, labelSvc, appRepo, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			result, err := engine.DoesNotContainResourceOfSubtype(ctx, testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.Equal(t, testCase.ExpectedResult, result)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, labelSvc, appRepo, formationRepo)
		})
	}
}
