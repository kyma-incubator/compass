package labeldef_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceCreateWithFormations(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		testFormations := []string{"test-formation-one", "test-formation-two"}
		expectedFormations := testFormations
		ctx := context.TODO()

		mockUID := &automock.UIDService{}
		defer mockUID.AssertExpectations(t)
		mockUID.On("Generate").Return(fixUUID())

		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockRepository.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			if schemaArgs, ok := args.Get(1).(model.LabelDefinition); ok {
				formations, err := labeldef.ParseFormationsFromSchema(schemaArgs.Schema)
				require.NoError(t, err)
				require.ElementsMatch(t, formations, expectedFormations)
				return
			}
			t.Fatal("schema should contain desired formations")
		})

		sut := labeldef.NewService(mockRepository, nil, nil, nil, mockUID)
		// WHEN
		err := sut.CreateWithFormations(ctx, "tenant", testFormations)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error if cannot create Label Definition", func(t *testing.T) {
		// GIVEN
		testFormations := []string{"test-formation-one", "test-formation-two"}
		expectedFormations := testFormations
		ctx := context.TODO()
		testError := errors.New("test error")

		mockUID := &automock.UIDService{}
		defer mockUID.AssertExpectations(t)
		mockUID.On("Generate").Return(fixUUID())

		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockRepository.On("Create", mock.Anything, mock.Anything).Return(testError).Run(func(args mock.Arguments) {
			if schemaArgs, ok := args.Get(1).(model.LabelDefinition); ok {
				formations, err := labeldef.ParseFormationsFromSchema(schemaArgs.Schema)
				require.NoError(t, err)
				require.ElementsMatch(t, formations, expectedFormations)
				return
			}
			t.Fatal("schema should contain desired formations")
		})

		sut := labeldef.NewService(mockRepository, nil, nil, nil, mockUID)
		// WHEN
		err := sut.CreateWithFormations(ctx, "tenant", testFormations)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}

func TestServiceGet(t *testing.T) {
	t.Run("success when key is not scenarios key", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    "key",
			Tenant: "tenant",
		}
		mockRepository.On("GetByKey", ctx, "tenant", "key").Return(&given, nil)
		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
		// WHEN
		actual, err := sut.Get(ctx, "tenant", "key")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, &given, actual)
	})

	t.Run("success when LD exists", func(t *testing.T) {
		// GIVEN
		testKey := model.ScenariosKey
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    testKey,
			Tenant: "tenant",
		}

		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockRepository.On("GetByKey", ctx, "tenant", testKey).Return(&given, nil)

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
		// WHEN
		actual, err := sut.Get(ctx, "tenant", testKey)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, &given, actual)
	})

	t.Run("on error from repository", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockRepository.On("GetByKey", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("some error"))

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
		// WHEN
		_, err := sut.Get(context.TODO(), "tenant", "key")
		// THEN
		require.EqualError(t, err, "while fetching Label Definition: some error")
	})
}

func TestServiceList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		ctx := context.TODO()
		givenDefs := []model.LabelDefinition{
			{
				Tenant: "tenant",
				Key:    "key1",
			},
			{
				Tenant: "tenant",
				Key:    "key2",
			},
		}
		mockRepository.On("List", ctx, "tenant").Return(givenDefs, nil)

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
		// WHEN
		actual, err := sut.List(ctx, "tenant")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenDefs, actual)
	})

	t.Run("on error from repository", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		ctx := context.TODO()
		mockRepository.On("List", ctx, "tenant").Return(nil, errors.New("some error"))
		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
		// WHEN
		_, err := sut.List(ctx, "tenant")
		// THEN
		require.EqualError(t, err, "while fetching Label Definitions: some error")
	})
}

func TestGetAvailableScenarios(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		mockService := &automock.Repository{}
		defer mockService.AssertExpectations(t)
		var givenSchema interface{} = model.NewScenariosSchema(testScenarios)
		givenDef := model.LabelDefinition{
			Tenant: fixTenant(),
			Key:    model.ScenariosKey,
			Schema: &givenSchema,
		}
		mockService.On("GetByKey", mock.Anything, fixTenant(), model.ScenariosKey).Return(&givenDef, nil)
		sut := labeldef.NewService(mockService, nil, nil, nil, nil)
		// WHEN
		actualScenarios, err := sut.GetAvailableScenarios(context.TODO(), fixTenant())
		// THEN
		require.NoError(t, err)
		assert.Equal(t, testScenarios, actualScenarios)
	})

	t.Run("returns error from repository", func(t *testing.T) {
		// GIVEN
		mockService := &automock.Repository{}
		defer mockService.AssertExpectations(t)
		mockService.On("GetByKey", mock.Anything, fixTenant(), model.ScenariosKey).Return(nil, fixError())
		sut := labeldef.NewService(mockService, nil, nil, nil, nil)
		// WHEN
		_, err := sut.GetAvailableScenarios(context.TODO(), fixTenant())
		// THEN
		require.EqualError(t, err, "while getting `scenarios` label definition: some error")
	})

	t.Run("returns error when missing schema in label def", func(t *testing.T) {
		// GIVEN
		mockService := &automock.Repository{}
		defer mockService.AssertExpectations(t)
		mockService.On("GetByKey", mock.Anything, fixTenant(), model.ScenariosKey).Return(&model.LabelDefinition{}, nil)
		sut := labeldef.NewService(mockService, nil, nil, nil, nil)
		// WHEN
		_, err := sut.GetAvailableScenarios(context.TODO(), fixTenant())
		// THEN
		require.EqualError(t, err, "missing schema for `scenarios` label definition")
	})
}

func fixUUID() string {
	return "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
}
