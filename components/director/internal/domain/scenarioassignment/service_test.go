package scenarioassignment_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("Create", fixCtxWithTenant(), fixModel()).Return(nil).Once()
		sut := scenarioassignment.NewService(mockRepo)
		// WHEN
		actual, err := sut.Create(fixCtxWithTenant(), fixModel())
		// THEN
		require.NoError(t, err)
		assert.Equal(t, fixModel(), actual)

	})

	t.Run("error on missing tenant in context", func(t *testing.T) {
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		sut := scenarioassignment.NewService(mockRepo)
		// WHEN
		_, err := sut.Create(context.TODO(), fixModel())
		// THEN
		assert.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("Create", fixCtxWithTenant(), fixModel()).Return(fixError()).Once()
		sut := scenarioassignment.NewService(mockRepo)
		// WHEN
		_, err := sut.Create(fixCtxWithTenant(), fixModel())
		// THEN
		require.EqualError(t, err, fmt.Sprintf("while persisting Assignment: %s", fixError().Error()))
	})
}

func TestService_GetByScenarioName(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("GetForScenarioName", fixCtxWithTenant(), mock.Anything, scenarioName).Return(fixModel(), nil).Once()
		sut := scenarioassignment.NewService(mockRepo)
		// WHEN
		actual, err := sut.GetForScenarioName(fixCtxWithTenant(), scenarioName)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, fixModel(), actual)

	})

	t.Run("error on missing tenant in context", func(t *testing.T) {
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		sut := scenarioassignment.NewService(mockRepo)
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
		sut := scenarioassignment.NewService(mockRepo)
		// WHEN
		_, err := sut.GetForScenarioName(fixCtxWithTenant(), scenarioName)
		// THEN
		require.EqualError(t, err, fmt.Sprintf("while getting Assignment: %s", errMsg))
	})
}

func TestService_GetForSelector(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		selector := fixLabelSelector()
		assignment := fixModel()
		result := []*model.AutomaticScenarioAssignment{&assignment}
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("GetForSelector", mock.Anything, selector, tenantID).Return(result, nil).Once()
		sut := scenarioassignment.NewService(mockRepo)
		// WHEN
		actual, err := sut.GetForSelector(fixCtxWithTenant(), selector)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, result, actual)

	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		selector := fixLabelSelector()
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("GetForSelector", mock.Anything, selector, tenantID).Return(nil, fixError()).Once()
		sut := scenarioassignment.NewService(mockRepo)
		// WHEN
		actual, err := sut.GetForSelector(fixCtxWithTenant(), selector)
		// THEN
		require.EqualError(t, err, "while getting the assignments: some error")
		require.Nil(t, actual)
	})

	t.Run("returns error when no tenant in context", func(t *testing.T) {
		sut := scenarioassignment.NewService(nil)
		_, err := sut.GetForSelector(context.TODO(), fixLabelSelector())
		require.EqualError(t, err, "cannot read tenant from context")
	})
}

func TestService_DeleteForSelector(t *testing.T) {
	ctx := fixCtxWithTenant()
	selector := fixLabelSelector()

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("DeleteForSelector", ctx, tenantID, selector).Return(nil).Once()
		sut := scenarioassignment.NewService(mockRepo)
		// WHEN
		err := sut.DeleteForSelector(ctx, selector)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("DeleteForSelector", ctx, tenantID, selector).Return(fixError()).Once()
		sut := scenarioassignment.NewService(mockRepo)
		// WHEN
		err := sut.DeleteForSelector(ctx, selector)
		// THEN
		require.EqualError(t, err, "while deleting the Assignments: some error")
	})

	t.Run("returns error when empty tenant", func(t *testing.T) {
		sut := scenarioassignment.NewService(nil)
		err := sut.DeleteForSelector(context.TODO(), selector)
		require.EqualError(t, err, "cannot read tenant from context")
	})
}
