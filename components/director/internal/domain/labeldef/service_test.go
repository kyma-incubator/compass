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
			Schema: fixBasicSchema(t),
		}

		defWithID := in
		defWithID.ID = fixUID()
		mockUID.On("Generate").Return(fixUID())
		mockRepository.On("Create", mock.Anything, defWithID).Return(nil)

		ctx := context.TODO()
		sut := labeldef.NewService(mockRepository, nil, mockUID)
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
		sut := labeldef.NewService(nil, nil, mockUID)
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
		sut := labeldef.NewService(mockRepository, nil, mockUID)
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
		sut := labeldef.NewService(mockRepository, nil, nil)
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

		sut := labeldef.NewService(mockRepository, nil, nil)
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

		sut := labeldef.NewService(mockRepository, nil, nil)
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
		sut := labeldef.NewService(mockRepository, nil, nil)
		// WHEN
		_, err := sut.List(ctx, "tenant")
		// THEN
		require.EqualError(t, err, "while fetching Label Definitions: some error")
	})
}

func TestServiceUpdate(t *testing.T) {
	tenant := "tenant"
	key := "firstName"

	t.Run("success", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		newSchema := fixBasicSchema(t)

		ld := model.LabelDefinition{
			ID:     fixUID(),
			Tenant: tenant,
			Key:    key,
			Schema: fixBasicSchema(t),
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

		existingLabels := []*model.Label{
			{
				ID:         "b9566e9d-83a2-4091-8c65-7a512b88f89e",
				Tenant:     tenant,
				Key:        key,
				Value:      "val",
				ObjectID:   "foo",
				ObjectType: model.RuntimeLabelableObject,
			},
			{
				ID:         "2037fc3d-be6c-4489-94cf-05518bac709f",
				Tenant:     tenant,
				Key:        key,
				Value:      "val2",
				ObjectID:   "bar",
				ObjectType: model.ApplicationLabelableObject,
			},
		}

		defWithID := in
		defWithID.ID = fixUID()

		mockRepository.On("GetByKey", mock.Anything, tenant, key).Return(&ld, nil).Once()
		mockRepository.On("Update", mock.Anything, defWithID).Return(nil)
		mockRepository.On("GetByKey", mock.Anything, tenant, key).Return(&in, nil).Once()

		mockLabelRepository.On("ListByKey", context.TODO(), tenant, key).Return(existingLabels, nil).Once()

		ctx := context.TODO()
		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil)
		// WHEN
		actual, err := sut.Update(ctx, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedLD, actual)
	})

	t.Run("returns error when existing label doesn't match new schema", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		newSchema := fixSchema(t, "nonExistingProp", "integer", "desc")

		ld := model.LabelDefinition{
			ID:     fixUID(),
			Tenant: tenant,
			Key:    key,
			Schema: fixBasicSchema(t),
		}

		in := model.LabelDefinition{
			ID:     fixUID(),
			Key:    key,
			Tenant: tenant,
			Schema: newSchema,
		}

		existingLabels := []*model.Label{
			{
				ID:         "b9566e9d-83a2-4091-8c65-7a512b88f89e",
				Tenant:     tenant,
				Key:        key,
				Value:      "val",
				ObjectID:   "foo",
				ObjectType: model.RuntimeLabelableObject,
			},
			{
				ID:         "2037fc3d-be6c-4489-94cf-05518bac709f",
				Tenant:     tenant,
				Key:        key,
				Value:      "val2",
				ObjectID:   "bar",
				ObjectType: model.ApplicationLabelableObject,
			},
		}

		defWithID := in
		defWithID.ID = fixUID()

		mockRepository.On("GetByKey", mock.Anything, tenant, key).Return(&ld, nil).Once()

		mockLabelRepository.On("ListByKey", context.TODO(), tenant, key).Return(existingLabels, nil).Once()

		ctx := context.TODO()
		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil)
		// WHEN
		_, err := sut.Update(ctx, in)
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "label with key firstName is not valid against new schema")
	})

	t.Run("returns error when validation of Label Definition failed", func(t *testing.T) {
		// GIVEN

		sut := labeldef.NewService(nil, nil, nil)
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
		sut := labeldef.NewService(mockRepository, nil, nil)
		// WHEN
		_, err := sut.Update(context.TODO(), model.LabelDefinition{Key: key, Tenant: tenant})
		// THEN
		require.EqualError(t, err, "while receiving Label Definition: some error")
	})

	t.Run("returns error if Label Definition update failed", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		ld := &model.LabelDefinition{
			ID:     "8b131225-f09d-4035-8091-1f12933863b3",
			Tenant: tenant,
			Key:    key,
			Schema: fixBasicSchema(t),
		}

		existingLabels := []*model.Label{
			{
				ID:         "b9566e9d-83a2-4091-8c65-7a512b88f89e",
				Tenant:     tenant,
				Key:        key,
				Value:      "val",
				ObjectID:   "foo",
				ObjectType: model.RuntimeLabelableObject,
			},
			{
				ID:         "2037fc3d-be6c-4489-94cf-05518bac709f",
				Tenant:     tenant,
				Key:        key,
				Value:      "val2",
				ObjectID:   "bar",
				ObjectType: model.ApplicationLabelableObject,
			},
		}

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(ld, nil)
		mockRepository.On("Update", context.TODO(), *ld).Return(errors.New("some error"))

		mockLabelRepository.On("ListByKey", context.TODO(), "tenant", "firstName").Return(existingLabels, nil).Once()

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil)
		// WHEN
		_, err := sut.Update(context.TODO(), *ld)
		// THEN
		require.EqualError(t, err, "while updating Label Definition: some error")
	})

	t.Run("returns error if receive of updated Label Definition failed", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		ld := &model.LabelDefinition{
			ID:     "8b131225-f09d-4035-8091-1f12933863b3",
			Tenant: tenant,
			Key:    key,
			Schema: fixBasicSchema(t),
		}

		existingLabels := []*model.Label{
			{
				ID:         "b9566e9d-83a2-4091-8c65-7a512b88f89e",
				Tenant:     tenant,
				Key:        key,
				Value:      "val",
				ObjectID:   "foo",
				ObjectType: model.RuntimeLabelableObject,
			},
			{
				ID:         "2037fc3d-be6c-4489-94cf-05518bac709f",
				Tenant:     tenant,
				Key:        key,
				Value:      "val2",
				ObjectID:   "bar",
				ObjectType: model.ApplicationLabelableObject,
			},
		}

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(ld, nil).Once()
		mockRepository.On("Update", context.TODO(), *ld).Return(nil)
		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(ld, errors.New("some error")).Once()

		mockLabelRepository.On("ListByKey", context.TODO(), "tenant", "firstName").Return(existingLabels, nil).Once()

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil)
		// WHEN
		_, err := sut.Update(context.TODO(), *ld)
		// THEN
		require.EqualError(t, err, "while receiving updated Label Definition: some error")
	})

	t.Run("returns error if received Label Definition has empty ID", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		ld := &model.LabelDefinition{
			ID:     "",
			Tenant: tenant,
			Key:    key,
			Schema: fixBasicSchema(t),
		}

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(ld, nil).Once()

		sut := labeldef.NewService(mockRepository, nil, nil)
		// WHEN
		_, err := sut.Update(context.TODO(), *ld)
		// THEN
		require.EqualError(t, err, "id of received label definition is empty")
	})
}

func fixUID() string {
	return "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
}

func fixBasicSchema(t *testing.T) *interface{} {
	sch := `{
		"$id": "https://example.com/person.schema.json",
  		"$schema": "http://json-schema.org/draft-07/schema#",
  		"title": "Person",
  		"type": "object",
  		"properties": {
  		  "firstName": {
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
	  }`
	var obj map[string]interface{}

	err := json.Unmarshal([]byte(sch), &obj)
	require.NoError(t, err)
	var objTemp interface{}
	objTemp = obj
	return &objTemp
}

func fixSchema(t *testing.T, propertyName, propertyType, propertyDescription string) *interface{} {
	sch := fmt.Sprintf(`{
		"$id": "https://example.com/person.schema.json",
		"$schema": "http://json-schema.org/draft-07/schema#",
		"title": "Person",
		"type": "object",
		"properties": {
		  "%s": {
		    "type": "%s",
		    "description": "%s"
		  }
		},
		"required": ["nonExistingProp"]
	  }`, propertyName, propertyType, propertyDescription)
	var obj map[string]interface{}

	err := json.Unmarshal([]byte(sch), &obj)
	require.NoError(t, err)
	var objTemp interface{}
	objTemp = obj
	return &objTemp
}
