package operators_test

import (
	"testing"

	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConstraintOperators_ContainsScenarioGroups(t *testing.T) {
	in := &formationconstraintpkg.ContainsScenarioGroupsInput{
		ResourceType:           model.ApplicationResourceType,
		ResourceSubtype:        inputAppType,
		ResourceID:             inputAppID,
		Tenant:                 testTenantID,
		RequiredScenarioGroups: []string{testScenarioGroup},
	}

	testCases := []struct {
		Name              string
		Input             operators.OperatorInput
		ApplicationRepo   func() *automock.ApplicationRepository
		SystemAuthService func() *automock.SystemAuthService
		ExpectedResult    bool
		ExpectedErrorMsg  string
	}{
		{
			Name:  "Success for an application which has requested scenario group and is connected",
			Input: in,
			SystemAuthService: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", ctx, pkgmodel.ApplicationReference, inputAppID).Return([]pkgmodel.SystemAuth{{
					Value: &model.Auth{
						OneTimeToken: &model.OneTimeToken{
							ScenarioGroups: []string{`{"key": "scenarioGroup","description": "bar1"}`, `{"key": "scenarioGroup2","description": "bar2"}`},
						},
					},
				}}, nil)
				return svc
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, testTenantID, inputAppID).Return(&model.Application{BaseEntity: &model.BaseEntity{ID: appID}, Status: &model.ApplicationStatus{Condition: model.ApplicationStatusConditionConnected}}, nil).Once()
				return repo
			},
			ExpectedResult:   true,
			ExpectedErrorMsg: "",
		},
		{
			Name:  "Success for an application which has requested scenario group and is connected and scenario groups are in legacy string array format",
			Input: in,
			SystemAuthService: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", ctx, pkgmodel.ApplicationReference, inputAppID).Return([]pkgmodel.SystemAuth{{
					Value: &model.Auth{
						OneTimeToken: &model.OneTimeToken{
							ScenarioGroups: []string{"scenarioGroup"},
						},
					},
				}}, nil)
				return svc
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, testTenantID, inputAppID).Return(&model.Application{BaseEntity: &model.BaseEntity{ID: appID}, Status: &model.ApplicationStatus{Condition: model.ApplicationStatusConditionConnected}}, nil).Once()
				return repo
			},
			ExpectedResult:   true,
			ExpectedErrorMsg: "",
		},
		{
			Name:  "Error for an application which has scenario groups, but not the requested ones",
			Input: in,
			SystemAuthService: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", ctx, pkgmodel.ApplicationReference, inputAppID).Return([]pkgmodel.SystemAuth{{
					Value: &model.Auth{
						OneTimeToken: &model.OneTimeToken{
							ScenarioGroups: []string{"someOtherGroup"},
						},
					},
				}}, nil)
				return svc
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, testTenantID, inputAppID).Return(&model.Application{BaseEntity: &model.BaseEntity{ID: appID}, Status: &model.ApplicationStatus{Condition: model.ApplicationStatusConditionConnected}}, nil).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "",
		},
		{
			Name:  "Error for an application which has scenario groups, but the one time token is already used",
			Input: in,
			SystemAuthService: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", ctx, pkgmodel.ApplicationReference, inputAppID).Return([]pkgmodel.SystemAuth{{
					Value: &model.Auth{
						OneTimeToken: &model.OneTimeToken{
							Used:           true,
							ScenarioGroups: []string{"someOtherGroup"},
						},
					},
				}}, nil)
				return svc
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, testTenantID, inputAppID).Return(&model.Application{BaseEntity: &model.BaseEntity{ID: appID}, Status: &model.ApplicationStatus{Condition: model.ApplicationStatusConditionConnected}}, nil).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "",
		},
		{
			Name:  "Error for an application which has no system auths",
			Input: in,
			SystemAuthService: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", ctx, pkgmodel.ApplicationReference, inputAppID).Return([]pkgmodel.SystemAuth{}, nil)
				return svc
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, testTenantID, inputAppID).Return(&model.Application{BaseEntity: &model.BaseEntity{ID: appID}, Status: &model.ApplicationStatus{Condition: model.ApplicationStatusConditionConnected}}, nil).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "",
		},
		{
			Name:  "Error fetching system auths fails",
			Input: in,
			SystemAuthService: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", ctx, pkgmodel.ApplicationReference, inputAppID).Return(nil, testErr)
				return svc
			},
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, testTenantID, inputAppID).Return(&model.Application{BaseEntity: &model.BaseEntity{ID: appID}, Status: &model.ApplicationStatus{Condition: model.ApplicationStatusConditionConnected}}, nil).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:              "Error for application that is still in initial state",
			Input:             in,
			SystemAuthService: unusedSystemAuthService,
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, testTenantID, inputAppID).Return(&model.Application{BaseEntity: &model.BaseEntity{ID: appID}, Status: &model.ApplicationStatus{Condition: model.ApplicationStatusConditionInitial}}, nil).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: fmt.Sprintf("Application with ID %q is not in status %s", inputAppID, graphql.ApplicationStatusConditionConnected),
		},
		{
			Name:              "Error when getting application",
			Input:             in,
			SystemAuthService: unusedSystemAuthService,
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, testTenantID, inputAppID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name: "Success when the operator input is has no required scenario groups",
			Input: &formationconstraintpkg.ContainsScenarioGroupsInput{
				ResourceType:           model.ApplicationResourceType,
				ResourceSubtype:        inputAppType,
				ResourceID:             inputAppID,
				Tenant:                 testTenantID,
				RequiredScenarioGroups: nil,
			},
			SystemAuthService: unusedSystemAuthService,
			ApplicationRepo:   unusedApplicationRepo,
			ExpectedResult:    true,
			ExpectedErrorMsg:  "",
		},
		{
			Name:              "Returns error when the operator input is incompatible",
			Input:             "incompatible",
			SystemAuthService: unusedSystemAuthService,
			ApplicationRepo:   unusedApplicationRepo,
			ExpectedResult:    false,
			ExpectedErrorMsg:  "Incompatible input",
		},
		{
			Name: "Returns error when the resource type is unknown",
			Input: &formationconstraintpkg.ContainsScenarioGroupsInput{
				ResourceType:    "Unknown",
				ResourceSubtype: inputAppType,
				ResourceID:      inputAppID,
				Tenant:          testTenantID,
			},
			SystemAuthService: unusedSystemAuthService,
			ApplicationRepo:   unusedApplicationRepo,
			ExpectedResult:    false,
			ExpectedErrorMsg:  "Unsupported resource type",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			systemAuthSvc := testCase.SystemAuthService()
			appRepo := testCase.ApplicationRepo()
			engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, systemAuthSvc, nil, nil, nil, appRepo, nil, nil, nil, runtimeType, applicationType)

			result, err := engine.ContainsScenarioGroups(ctx, testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.Equal(t, testCase.ExpectedResult, result)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, systemAuthSvc, appRepo)
		})
	}
}
