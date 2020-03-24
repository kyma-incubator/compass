package scenarioassignment_test

import (
	"context"
	"testing"

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
		mockRepo.On("Create", mock.Anything, fixModel()).Return(nil)
		sut := scenarioassignment.NewService(mockRepo)
		// WHEN
		actual, err := sut.Create(context.TODO(), fixModel())
		// THEN
		require.NoError(t, err)
		assert.Equal(t, fixModel(), actual)

	})

	t.Run("returns error on error from repository", func(t *testing.T) {
		// GIVEN
		mockRepo := &automock.Repository{}
		defer mockRepo.AssertExpectations(t)
		mockRepo.On("Create", mock.Anything, fixModel()).Return(fixError())
		sut := scenarioassignment.NewService(mockRepo)
		// WHEN
		_, err := sut.Create(context.TODO(), fixModel())
		// THEN
		require.EqualError(t, err, "while persisting Assignment: some error")
	})
}
