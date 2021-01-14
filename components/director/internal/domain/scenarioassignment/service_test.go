package scenarioassignment_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const validPageSize = 2

func TestService_Create(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.Repository{}
		mockRepo.On("Create", ctx, fixModel()).Return(nil)
		mockScenarioDefSvc := mockScenarioDefServiceThatReturns([]string{scenarioName})
		mockEngine := &automock.AssignmentEngine{}
		mockEngine.On("EnsureScenarioAssigned", ctx, fixModel()).Return(nil).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo, mockScenarioDefSvc, mockEngine)

		sut := scenarioassignment.NewService(mockRepo, mockScenarioDefSvc, mockEngine)

		// WHEN
		actual, err := sut.Create(fixCtxWithTenant(), fixModel())

		// THEN
		require.NoError(t, err)
		assert.Equal(t, fixModel(), actual)

	})

	t.Run("return error when ensuring scenarios for runtimes fails", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.Repository{}
		mockRepo.On("Create", ctx, fixModel()).Return(nil)
		mockScenarioDefSvc := mockScenarioDefServiceThatReturns([]string{scenarioName})
		mockEngine := &automock.AssignmentEngine{}
		mockEngine.On("EnsureScenarioAssigned", ctx, fixModel()).Return(fixError()).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo, mockScenarioDefSvc, mockEngine)

		sut := scenarioassignment.NewService(mockRepo, mockScenarioDefSvc, mockEngine)

		// WHEN
		_, err := sut.Create(fixCtxWithTenant(), fixModel())

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("returns error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		sut := scenarioassignment.NewService(nil, nil, nil)

		// WHEN
		_, err := sut.Create(context.TODO(), fixModel())

		// THEN
		assert.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("returns error when scenario already has an assignment", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}

		mockRepo.On("Create", mock.Anything, fixModel()).Return(apperrors.NewNotUniqueError(""))
		mockScenarioDefSvc := mockScenarioDefServiceThatReturns([]string{scenarioName})

		defer mock.AssertExpectationsForObjects(t, mockRepo, mockScenarioDefSvc)
		sut := scenarioassignment.NewService(mockRepo, mockScenarioDefSvc, nil)
		// WHEN
		_, err := sut.Create(fixCtxWithTenant(), fixModel())
		// THEN
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "a given scenario already has an assignment")
	})

	t.Run("returns error when given scenario does not exist", func(t *testing.T) {
		// GIVEN
		mockScenarioDefSvc := mockScenarioDefServiceThatReturns([]string{"completely-different-scenario"})
		defer mock.AssertExpectationsForObjects(t, mockScenarioDefSvc)
		sut := scenarioassignment.NewService(nil, mockScenarioDefSvc, nil)

		// WHEN
		_, err := sut.Create(fixCtxWithTenant(), fixModel())

		// THEN
		require.EqualError(t, err, "scenario `scenario-A` does not exist")
	})

	t.Run("returns error on persisting in DB", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		mockRepo.On("Create", mock.Anything, fixModel()).Return(fixError())
		mockScenarioDefSvc := mockScenarioDefServiceThatReturns([]string{scenarioName})

		defer mock.AssertExpectationsForObjects(t, mockRepo, mockScenarioDefSvc)
		sut := scenarioassignment.NewService(mockRepo, mockScenarioDefSvc, nil)

		// WHEN
		_, err := sut.Create(fixCtxWithTenant(), fixModel())

		// THEN
		require.EqualError(t, err, "while persisting Assignment: some error")
	})

	t.Run("returns error on ensuring that scenarios label definition exist", func(t *testing.T) {
		// GIVEN
		mockScenarioDefSvc := &automock.ScenariosDefService{}
		defer mock.AssertExpectationsForObjects(t, mockScenarioDefSvc)
		mockScenarioDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, mock.Anything).Return(fixError())
		sut := scenarioassignment.NewService(nil, mockScenarioDefSvc, nil)
		// WHEN
		_, err := sut.Create(fixCtxWithTenant(), fixModel())
		// THEN
		require.EqualError(t, err, "while ensuring that `scenarios` label definition exist: some error")
	})

	t.Run("returns error on getting available scenarios from label definition", func(t *testing.T) {
		// GIVEN
		mockScenarioDefSvc := &automock.ScenariosDefService{}
		defer mock.AssertExpectationsForObjects(t, mockScenarioDefSvc)
		mockScenarioDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, mock.Anything).Return(nil)
		mockScenarioDefSvc.On("GetAvailableScenarios", mock.Anything, tenantID).Return(nil, fixError())
		sut := scenarioassignment.NewService(nil, mockScenarioDefSvc, nil)
		// WHEN
		_, err := sut.Create(fixCtxWithTenant(), fixModel())
		// THEN
		require.EqualError(t, err, "while getting available scenarios: some error")
	})

}

func TestService_GetByScenarioName(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("GetForScenarioName", fixCtxWithTenant(), mock.Anything, scenarioName).Return(fixModel(), nil).Once()
		sut := scenarioassignment.NewService(mockRepo, nil, nil)

		// WHEN
		actual, err := sut.GetForScenarioName(fixCtxWithTenant(), scenarioName)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, fixModel(), actual)

	})

	t.Run("error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		sut := scenarioassignment.NewService(mockRepo, nil, nil)

		// WHEN
		_, err := sut.GetForScenarioName(context.TODO(), scenarioName)

		// THEN
		assert.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("GetForScenarioName", fixCtxWithTenant(), mock.Anything, scenarioName).Return(model.AutomaticScenarioAssignment{}, fixError()).Once()
		sut := scenarioassignment.NewService(mockRepo, nil, nil)

		// WHEN
		_, err := sut.GetForScenarioName(fixCtxWithTenant(), scenarioName)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while getting Assignment: %s", errMsg))
	})
}

func TestService_ListForSelector(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		selector := fixLabelSelector()
		assignment := fixModel()
		result := []*model.AutomaticScenarioAssignment{&assignment}
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("ListForSelector", mock.Anything, selector, tenantID).Return(result, nil).Once()
		sut := scenarioassignment.NewService(mockRepo, nil, nil)

		// WHEN
		actual, err := sut.ListForSelector(fixCtxWithTenant(), selector)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, result, actual)
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		selector := fixLabelSelector()
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("ListForSelector", mock.Anything, selector, tenantID).Return(nil, fixError()).Once()
		sut := scenarioassignment.NewService(mockRepo, nil, nil)

		// WHEN
		actual, err := sut.ListForSelector(fixCtxWithTenant(), selector)

		// THEN
		require.EqualError(t, err, "while getting the assignments: some error")
		require.Nil(t, actual)
	})

	t.Run("returns error when no tenant in context", func(t *testing.T) {
		sut := scenarioassignment.NewService(nil, nil, nil)
		_, err := sut.ListForSelector(context.TODO(), fixLabelSelector())

		require.EqualError(t, err, "cannot read tenant from context")
	})
}

func TestService_List(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	mod1 := fixModelWithScenarioName("foo")
	mod2 := fixModelWithScenarioName("bar")
	modItems := []*model.AutomaticScenarioAssignment{
		&mod1, &mod2,
	}

	modelPage := fixModelPageWithItems(modItems)

	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.Repository
		ExpectedResult     *model.AutomaticScenarioAssignmentPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("List", ctx, tenantID, validPageSize, after).Return(&modelPage, nil).Once()
				return repo
			},
			PageSize:           validPageSize,
			ExpectedResult:     &modelPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Return error when page size is less than 1",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				return repo
			},
			PageSize:           0,
			ExpectedResult:     &modelPage,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Return error when page size is bigger than 200",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				return repo
			},
			PageSize:           201,
			ExpectedResult:     &modelPage,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when Assignments listing failed",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("List", ctx, tenantID, 2, after).Return(nil, testErr).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := scenarioassignment.NewService(repo, nil, nil)

			// WHEN
			items, err := svc.List(ctx, testCase.PageSize, after)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, items)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := scenarioassignment.NewService(nil, nil, nil)
		// WHEN
		_, err := svc.List(context.TODO(), 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_DeleteManyForSameSelector(t *testing.T) {
	ctx := fixCtxWithTenant()
	selector := fixLabelSelector()

	scenarioNameA := "scenario-A"
	scenarioNameB := "scenario-B"
	models := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName: scenarioNameA,
			Selector: model.LabelSelector{
				Key:   "key",
				Value: "value",
			},
		},
		{
			ScenarioName: scenarioNameB,
			Selector: model.LabelSelector{
				Key:   "key",
				Value: "value",
			},
		},
	}

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		mockRepo.On("DeleteForSelector", ctx, tenantID, selector).Return(nil).Once()
		mockEngine := &automock.AssignmentEngine{}
		mockEngine.On("RemoveAssignedScenarios", ctx, models).Return(nil).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo, mockEngine)

		sut := scenarioassignment.NewService(mockRepo, nil, mockEngine)
		// WHEN
		err := sut.DeleteManyForSameSelector(ctx, models)
		// THEN
		require.NoError(t, err)
	})

	t.Run("return error when unassigning scenarios from runtimes fails", func(t *testing.T) {
		// GIVEN
		mockEngine := &automock.AssignmentEngine{}
		mockEngine.On("RemoveAssignedScenarios", ctx, models).Return(fixError()).Once()
		defer mock.AssertExpectationsForObjects(t, mockEngine)

		sut := scenarioassignment.NewService(nil, nil, mockEngine)
		// WHEN
		err := sut.DeleteManyForSameSelector(ctx, models)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("return error when input slice is empty", func(t *testing.T) {
		// GIVEN
		mockEngine := &automock.AssignmentEngine{}
		defer mock.AssertExpectationsForObjects(t, mockEngine)

		sut := scenarioassignment.NewService(nil, nil, mockEngine)
		// WHEN
		err := sut.DeleteManyForSameSelector(ctx, []*model.AutomaticScenarioAssignment{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected at least one item in Assignments slice")
	})

	t.Run("return error when input slice contains assignments with different selectors", func(t *testing.T) {
		// GIVEN
		modelsWithDifferentSelectors := []*model.AutomaticScenarioAssignment{
			{
				ScenarioName: scenarioNameA,
				Selector: model.LabelSelector{
					Key:   "key",
					Value: "value",
				},
			},
			{
				ScenarioName: scenarioNameB,
				Selector: model.LabelSelector{
					Key:   "key",
					Value: "bar",
				},
			},
		}

		mockEngine := &automock.AssignmentEngine{}
		defer mock.AssertExpectationsForObjects(t, mockEngine)

		sut := scenarioassignment.NewService(nil, nil, mockEngine)
		// WHEN
		err := sut.DeleteManyForSameSelector(ctx, modelsWithDifferentSelectors)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "all input items have to have the same selector")
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		mockRepo := &automock.Repository{}
		mockEngine := &automock.AssignmentEngine{}
		mockEngine.On("RemoveAssignedScenarios", ctx, models).Return(nil).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo, mockEngine)

		mockRepo.On("DeleteForSelector", ctx, tenantID, selector).Return(fixError()).Once()
		sut := scenarioassignment.NewService(mockRepo, nil, mockEngine)
		// WHEN
		err := sut.DeleteManyForSameSelector(ctx, models)
		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignments: %s", errMsg))
	})

	t.Run("returns error when empty tenant", func(t *testing.T) {
		sut := scenarioassignment.NewService(nil, nil, nil)
		err := sut.DeleteManyForSameSelector(context.TODO(), models)
		require.EqualError(t, err, "cannot read tenant from context")
	})
}

func TestService_DeleteForScenarioName(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.Repository{}
		mockRepo.On("DeleteForScenarioName", ctx, tenantID, scenarioName).Return(nil)
		mockEngine := &automock.AssignmentEngine{}
		mockEngine.On("RemoveAssignedScenario", ctx, fixModel()).Return(nil).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo, mockEngine)

		svc := scenarioassignment.NewService(mockRepo, nil, mockEngine)

		// WHEN
		err := svc.Delete(fixCtxWithTenant(), fixModel())

		// THEN
		require.NoError(t, err)
	})

	t.Run("return error when unassigning scenarios from runtimes fails", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.Repository{}
		mockEngine := &automock.AssignmentEngine{}
		mockEngine.On("RemoveAssignedScenario", ctx, fixModel()).Return(fixError()).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo, mockEngine)

		svc := scenarioassignment.NewService(mockRepo, nil, mockEngine)

		// WHEN
		err := svc.Delete(fixCtxWithTenant(), fixModel())

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), fixError().Error())
	})

	t.Run("error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		svc := scenarioassignment.NewService(mockRepo, nil, nil)

		// WHEN
		err := svc.Delete(context.TODO(), fixModel())

		// THEN
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		ctx := fixCtxWithTenant()
		mockRepo := &automock.Repository{}
		mockRepo.On("DeleteForScenarioName", ctx, tenantID, scenarioName).Return(fixError())
		mockEngine := &automock.AssignmentEngine{}
		mockEngine.On("RemoveAssignedScenario", ctx, fixModel()).Return(nil).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepo, mockEngine)

		svc := scenarioassignment.NewService(mockRepo, nil, mockEngine)

		// WHEN
		err := svc.Delete(fixCtxWithTenant(), fixModel())

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignment: %s", errMsg))
	})
}

func mockScenarioDefServiceThatReturns(scenarios []string) *automock.ScenariosDefService {
	mockScenarioDefSvc := &automock.ScenariosDefService{}
	mockScenarioDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, tenantID).Return(nil)
	mockScenarioDefSvc.On("GetAvailableScenarios", mock.Anything, tenantID).Return(scenarios, nil)
	return mockScenarioDefSvc
}
