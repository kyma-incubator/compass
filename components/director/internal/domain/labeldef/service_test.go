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
		defWithID.ID = fixUUID()
		mockUID.On("Generate").Return(fixUUID())
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

		mockUID.On("Generate").Return(fixUUID())
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

		mockUID.On("Generate").Return(fixUUID())
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

	t.Run("success when schema is not nil", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		newSchema := fixBasicSchema(t)

		ld := model.LabelDefinition{
			ID:     fixUUID(),
			Tenant: tenant,
			Key:    key,
			Schema: fixBasicSchema(t),
		}

		in := model.LabelDefinition{
			ID:     fixUUID(),
			Key:    key,
			Tenant: tenant,
			Schema: newSchema,
		}

		existingLabels := []*model.Label{
			{
				ID:     "b9566e9d-83a2-4091-8c65-7a512b88f89e",
				Tenant: tenant,
				Key:    key,
				Value: map[string]interface{}{
					"firstName": "val",
					"lastName":  "val2",
					"age":       1235,
				},
				ObjectID:   "foo",
				ObjectType: model.RuntimeLabelableObject,
			},
			{
				ID:     "2037fc3d-be6c-4489-94cf-05518bac709f",
				Tenant: tenant,
				Key:    key,
				Value: map[string]interface{}{
					"firstName": "val3",
					"lastName":  "val4",
				},
				ObjectID:   "bar",
				ObjectType: model.ApplicationLabelableObject,
			},
		}

		defWithID := in
		defWithID.ID = fixUUID()

		mockRepository.On("GetByKey", mock.Anything, tenant, key).Return(&ld, nil).Once()
		mockRepository.On("Update", mock.Anything, defWithID).Return(nil)

		mockLabelRepository.On("ListByKey", context.TODO(), tenant, key).Return(existingLabels, nil).Once()

		ctx := context.TODO()
		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil)
		// WHEN
		err := sut.Update(ctx, in)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when existing label doesn't match new schema", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		oldProperty := "oldProperty"
		nonExistingProperty := "nonExistingProp"

		ld := model.LabelDefinition{
			ID:     fixUUID(),
			Tenant: tenant,
			Key:    key,
			Schema: fixSchema(t, oldProperty, "string", "desc", "oldProperty"),
		}

		in := model.LabelDefinition{
			ID:     fixUUID(),
			Key:    key,
			Tenant: tenant,
			Schema: fixSchema(t, nonExistingProperty, "integer", "desc", "nonExistingProp"),
		}

		existingLabels := []*model.Label{
			{
				ID:     "b9566e9d-83a2-4091-8c65-7a512b88f89e",
				Tenant: tenant,
				Key:    oldProperty,
				Value: map[string]interface{}{
					"key": "val",
				},
				ObjectID:   "foo",
				ObjectType: model.RuntimeLabelableObject,
			},
			{
				ID:     "2037fc3d-be6c-4489-94cf-05518bac709f",
				Tenant: tenant,
				Key:    oldProperty,
				Value: map[string]interface{}{
					"key": "val2",
				},
				ObjectID:   "bar",
				ObjectType: model.ApplicationLabelableObject,
			},
		}

		defWithID := in
		defWithID.ID = fixUUID()

		mockRepository.On("GetByKey", mock.Anything, tenant, key).Return(&ld, nil).Once()

		mockLabelRepository.On("ListByKey", context.TODO(), tenant, key).Return(existingLabels, nil).Once()

		ctx := context.TODO()
		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil)
		// WHEN
		err := sut.Update(ctx, in)
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, `label with key "oldProperty" is not valid against new schema for Runtime with ID "foo": (root): nonExistingProp is required`)
	})

	t.Run("returns error when validation of Label Definition failed", func(t *testing.T) {
		// GIVEN

		sut := labeldef.NewService(nil, nil, nil)
		// WHEN
		err := sut.Update(context.TODO(), model.LabelDefinition{})
		// THEN
		require.EqualError(t, err, "while validating Label Definition: missing Tenant field")
	})

	t.Run("returns error when error occured during receiving Label Definition", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(nil, errors.New("some error"))
		sut := labeldef.NewService(mockRepository, nil, nil)
		// WHEN
		err := sut.Update(context.TODO(), model.LabelDefinition{Key: key, Tenant: tenant, Schema: fixBasicSchema(t)})
		// THEN
		require.EqualError(t, err, "while receiving Label Definition: some error")
	})

	t.Run("returns error if Label Definition was not found", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(nil, nil)
		sut := labeldef.NewService(mockRepository, nil, nil)
		// WHEN
		err := sut.Update(context.TODO(), model.LabelDefinition{Key: key, Tenant: tenant, Schema: fixBasicSchema(t)})
		// THEN
		require.EqualError(t, err, "definition with firstName key doesn't exist")
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
				ID:     "b9566e9d-83a2-4091-8c65-7a512b88f89e",
				Tenant: tenant,
				Key:    key,
				Value: map[string]interface{}{
					"firstName": "val",
					"lastName":  "val2",
				},
				ObjectID:   "foo",
				ObjectType: model.RuntimeLabelableObject,
			},
			{
				ID:     "2037fc3d-be6c-4489-94cf-05518bac709f",
				Tenant: tenant,
				Key:    key,
				Value: map[string]interface{}{
					"firstName": "val3",
					"lastName":  "val4",
					"age":       22,
				},
				ObjectID:   "bar",
				ObjectType: model.ApplicationLabelableObject,
			},
		}

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(ld, nil)
		mockRepository.On("Update", context.TODO(), *ld).Return(errors.New("some error"))

		mockLabelRepository.On("ListByKey", context.TODO(), "tenant", "firstName").Return(existingLabels, nil).Once()

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil)
		// WHEN
		err := sut.Update(context.TODO(), *ld)
		// THEN
		require.EqualError(t, err, "while updating Label Definition: some error")
	})

	t.Run("label definition schema for update is nil", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		ld := &model.LabelDefinition{
			ID:     "8b131225-f09d-4035-8091-1f12933863b3",
			Tenant: tenant,
			Key:    key,
		}

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(ld, nil).Once()
		mockRepository.On("Update", context.TODO(), *ld).Return(nil).Once()

		sut := labeldef.NewService(mockRepository, nil, nil)
		// WHEN
		err := sut.Update(context.TODO(), *ld)
		// THEN
		require.NoError(t, err)
	})
}

func TestServiceDelete(t *testing.T) {
	t.Run("success when no labels use labeldef", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		tnt := "tenant"
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    "key",
			Tenant: tnt,
		}
		deleteRelatedResources := false
		mockRepository.On("GetByKey", ctx, tnt, given.Key).Return(&given, nil).Once()
		mockRepository.On("DeleteByKey", ctx, tnt, given.Key).Return(nil).Once()
		mockLabelRepository.On("ListByKey", ctx, tnt, given.Key).Return([]*model.Label{}, nil)

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil)
		// WHEN
		err := sut.Delete(ctx, tnt, given.Key, deleteRelatedResources)
		// THEN
		require.NoError(t, err)
	})

	t.Run("error when deleting scenarios key", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		tnt := "tenant"
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    model.ScenariosKey,
			Tenant: tnt,
		}
		deleteRelatedResources := false

		sut := labeldef.NewService(mockRepository, nil, nil)
		// WHEN
		err := sut.Delete(ctx, tnt, given.Key, deleteRelatedResources)
		// THEN
		require.Error(t, err)
	})

	t.Run("error when some labels use labeldef", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		tnt := "tenant"
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    "key",
			Tenant: "tenant",
		}
		existingLabels := []*model.Label{
			fixLabel("test", tnt, given.Key, nil, "object1", model.ApplicationLabelableObject),
			fixLabel("test2", tnt, given.Key, nil, "object2", model.RuntimeLabelableObject),
			fixLabel("test3", tnt, given.Key, nil, "object3", model.ApplicationLabelableObject),
		}
		deleteRelatedResources := false
		mockRepository.On("GetByKey", ctx, tnt, given.Key).Return(&given, nil).Once()
		mockLabelRepository.On("ListByKey", ctx, tnt, given.Key).Return(existingLabels, nil)

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil)
		// WHEN
		err := sut.Delete(ctx, "tenant", given.Key, deleteRelatedResources)
		// THEN
		require.Error(t, err)
	})

	t.Run("error when listing existing labels failed", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		tnt := "tenant"
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    "key",
			Tenant: "tenant",
		}
		deleteRelatedResources := false
		mockRepository.On("GetByKey", ctx, tnt, given.Key).Return(&given, nil).Once()
		mockLabelRepository.On("ListByKey", ctx, tnt, given.Key).Return([]*model.Label{}, errors.New("test"))

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil)
		// WHEN
		err := sut.Delete(ctx, "tenant", given.Key, deleteRelatedResources)
		// THEN
		require.Error(t, err)
	})

	t.Run("error when label definition not found", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		tnt := "tenant"
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    "key",
			Tenant: tnt,
		}
		deleteRelatedResources := false
		mockRepository.On("GetByKey", ctx, tnt, given.Key).Return(nil, nil).Once()

		sut := labeldef.NewService(mockRepository, nil, nil)
		// WHEN
		err := sut.Delete(ctx, tnt, given.Key, deleteRelatedResources)
		// THEN
		require.Error(t, err)
	})

	t.Run("error when getting label definition failed", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		tnt := "tenant"
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    "key",
			Tenant: tnt,
		}
		deleteRelatedResources := false
		mockRepository.On("GetByKey", ctx, tnt, given.Key).Return(nil, errors.New("")).Once()

		sut := labeldef.NewService(mockRepository, nil, nil)
		// WHEN
		err := sut.Delete(ctx, tnt, given.Key, deleteRelatedResources)
		// THEN
		require.Error(t, err)
	})

	t.Run("error when trying to delete related resources", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		tnt := "tenant"
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    "key",
			Tenant: tnt,
		}
		deleteRelatedResources := true

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil)
		// WHEN
		err := sut.Delete(ctx, tnt, given.Key, deleteRelatedResources)
		// THEN
		require.Error(t, err)
	})
}

func fixUUID() string {
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

func fixSchema(t *testing.T, propertyName, propertyType, propertyDescription, requiredProperty string) *interface{} {
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
		"required": ["%s"]
	  }`, propertyName, propertyType, propertyDescription, requiredProperty)
	var obj map[string]interface{}

	err := json.Unmarshal([]byte(sch), &obj)
	require.NoError(t, err)
	var objTemp interface{}
	objTemp = obj
	return &objTemp
}

func fixLabel(id, tenant, key string, value interface{}, objectID string, objectType model.LabelableObject) *model.Label {
	return &model.Label{
		ID:         id,
		Tenant:     tenant,
		Key:        key,
		Value:      value,
		ObjectID:   objectID,
		ObjectType: objectType,
	}
}
