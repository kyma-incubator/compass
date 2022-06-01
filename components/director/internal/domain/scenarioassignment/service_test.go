package scenarioassignment_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const validPageSize = 2

func TestService_GetByScenarioName(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("GetForScenarioName", fixCtxWithTenant(), mock.Anything, scenarioName).Return(fixModel(), nil).Once()
		sut := scenarioassignment.NewService(mockRepo, nil)

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
		sut := scenarioassignment.NewService(mockRepo, nil)

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
		sut := scenarioassignment.NewService(mockRepo, nil)

		// WHEN
		_, err := sut.GetForScenarioName(fixCtxWithTenant(), scenarioName)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while getting Assignment: %s", errMsg))
	})
}

func TestService_ListForTargetTenant(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		assignment := fixModel()
		result := []*model.AutomaticScenarioAssignment{&assignment}
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("ListForTargetTenant", mock.Anything, tenantID, targetTenantID).Return(result, nil).Once()
		sut := scenarioassignment.NewService(mockRepo, nil)

		// WHEN
		actual, err := sut.ListForTargetTenant(fixCtxWithTenant(), targetTenantID)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, result, actual)
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("ListForTargetTenant", mock.Anything, tenantID, targetTenantID).Return(nil, fixError()).Once()
		sut := scenarioassignment.NewService(mockRepo, nil)

		// WHEN
		actual, err := sut.ListForTargetTenant(fixCtxWithTenant(), targetTenantID)

		// THEN
		require.EqualError(t, err, "while getting the assignments: some error")
		require.Nil(t, actual)
	})

	t.Run("returns error when no tenant in context", func(t *testing.T) {
		sut := scenarioassignment.NewService(nil, nil)
		_, err := sut.ListForTargetTenant(context.TODO(), targetTenantID)

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

			svc := scenarioassignment.NewService(repo, nil)

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
		svc := scenarioassignment.NewService(nil, nil)
		// WHEN
		_, err := svc.List(context.TODO(), 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
