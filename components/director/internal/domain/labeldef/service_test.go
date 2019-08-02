package labeldef_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		mockUID := &automock.UIDService{}
		defer mockRepository.AssertExpectations(t)
		defer mockUID.AssertExpectations(t)

		in := model.LabelDefinition{
			Key:    "some-key",
			Tenant: "tenant",
			Schema: fixSchema(t, "firstName"),
		}

		defWithID := in
		defWithID.ID = fixUID()
		mockUID.On("Generate").Return(fixUID())
		mockRepository.On("Create", mock.Anything, defWithID).Return(nil)

		ctx := context.TODO()
		sut := labeldef.NewService(mockRepository, mockUID)
		// WHEN
		actual, err := sut.Create(ctx, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, defWithID, actual)
	})

	t.Run("returns error if Label Definition is invalid", func(t *testing.T) {
		// GIVEN
		mockUID := &automock.UIDService{}
		defer mockUID.AssertExpectations(t)

		mockUID.On("Generate").Return(fixUID())
		sut := labeldef.NewService(nil, mockUID)
		// WHEN
		_, err := sut.Create(context.TODO(), model.LabelDefinition{})
		// THEN
		require.EqualError(t, err, "while validation Label Definition: missing Tenant field")
	})

	t.Run("returns error if cannot persist Label Definition", func(t *testing.T) {
		// GIVEN
		mockUID := &automock.UIDService{}
		defer mockUID.AssertExpectations(t)

		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockUID.On("Generate").Return(fixUID())
		mockRepository.On("Create", mock.Anything, mock.Anything).Return(errors.New("some error"))
		sut := labeldef.NewService(mockRepository, mockUID)
		// WHEN
		_, err := sut.Create(context.TODO(), model.LabelDefinition{Key: "key", Tenant: "tenant"})
		// THEN
		require.EqualError(t, err, "while storing Label Definition: some error")
	})

}

func TestServiceGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    "key",
			Tenant: "tenant",
		}
		mockRepository.On("GetByKey", ctx, "tenant", "key").Return(&given, nil)
		sut := labeldef.NewService(mockRepository, nil)
		// WHEN
		actual, err := sut.Get(ctx, "tenant", "key")
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

		sut := labeldef.NewService(mockRepository, nil)
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

		sut := labeldef.NewService(mockRepository, nil)
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
		sut := labeldef.NewService(mockRepository, nil)
		// WHEN
		_, err := sut.List(ctx, "tenant")
		// THEN
		require.EqualError(t, err, "while fetching Label Definitions: some error")
	})
}

func TestServiceUpdate(t *testing.T) {
	tenant := "tenant"
	key := "key"

	t.Run("success", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		newSchema := fixSchema(t, "newFirstName")

		ld := model.LabelDefinition{
			ID:     fixUID(),
			Tenant: tenant,
			Key:    key,
			Schema: fixSchema(t, "FirstName"),
		}

		in := model.LabelDefinition{
			ID:     fixUID(),
			Key:    key,
			Tenant: tenant,
			Schema: newSchema,
		}

		expectedLD := model.LabelDefinition{
			ID:     fixUID(),
			Tenant: tenant,
			Key:    key,
			Schema: newSchema,
		}

		defWithID := in
		defWithID.ID = fixUID()
		mockRepository.On("GetByKey", mock.Anything, tenant, key).Return(&ld, nil).Once()
		mockRepository.On("Update", mock.Anything, defWithID).Return(nil)
		mockRepository.On("GetByKey", mock.Anything, tenant, key).Return(&in, nil).Once()

		ctx := context.TODO()
		sut := labeldef.NewService(mockRepository, nil)
		// WHEN
		actual, err := sut.Update(ctx, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedLD, actual)
	})

	t.Run("returns error when validation of Label Definition failed", func(t *testing.T) {
		// GIVEN

		sut := labeldef.NewService(nil, nil)
		// WHEN
		_, err := sut.Update(context.TODO(), model.LabelDefinition{})
		// THEN
		require.EqualError(t, err, "while validating Label Definition: missing Tenant field")
	})

	t.Run("returns error if Label Definition was not found", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(nil, errors.New("some error"))
		sut := labeldef.NewService(mockRepository, nil)
		// WHEN
		_, err := sut.Update(context.TODO(), model.LabelDefinition{Key: key, Tenant: tenant})
		// THEN
		require.EqualError(t, err, "while receiving Label Definition: some error")
	})

	t.Run("returns error if Label Definition update failed", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		ld := &model.LabelDefinition{
			Tenant: tenant,
			Key:    key,
		}

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(ld, nil)
		mockRepository.On("Update", context.TODO(), *ld).Return(errors.New("some error"))
		sut := labeldef.NewService(mockRepository, nil)
		// WHEN
		_, err := sut.Update(context.TODO(), *ld)
		// THEN
		require.EqualError(t, err, "while updating Label Definition: some error")
	})

	t.Run("returns error if received of updated Label Definition failed", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		ld := &model.LabelDefinition{
			Tenant: tenant,
			Key:    key,
		}

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(ld, nil).Once()
		mockRepository.On("Update", context.TODO(), *ld).Return(nil)
		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(ld, errors.New("some error")).Once()
		sut := labeldef.NewService(mockRepository, nil)
		// WHEN
		_, err := sut.Update(context.TODO(), *ld)
		// THEN
		require.EqualError(t, err, "while receiving updated Label Definition: some error")
	})
}

func fixUID() string {
	return "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
}

func fixSchema(t *testing.T, firstPropertyName string) *interface{} {
	sch := fmt.Sprintf(`{
		"$id": "https://example.com/person.schema.json",
  		"$schema": "http://json-schema.org/draft-07/schema#",
  		"title": "Person",
  		"type": "object",
  		"properties": {
  		  "%s": {
  		    "type": "string",
  		    "description": "The person's first name."
  		  },
  		  "lastName": {
  		    "type": "string",
  		    "description": "The person's last name."
  		  },
  		  "age": {
  		    "description": "Age in years which must be equal to or greater than zero.",
  		    "type": "integer",
  		    "minimum": 0
  		  }
  		}
	  }`, firstPropertyName)
	var obj map[string]interface{}

	err := json.Unmarshal([]byte(sch), &obj)
	require.NoError(t, err)
	var objTemp interface{}
	objTemp = obj
	return &objTemp
}
