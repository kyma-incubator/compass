package labeldef_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

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
		sut := labeldef.NewService(mockRepository, nil, nil, nil, mockUID)
		// WHEN
		actual, err := sut.Create(ctx, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, defWithID, actual)
	})

	t.Run("returns error if cannot persist Label Definition", func(t *testing.T) {
		// GIVEN
		mockUID := &automock.UIDService{}
		defer mockUID.AssertExpectations(t)

		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockUID.On("Generate").Return(fixUUID())
		mockRepository.On("Create", mock.Anything, mock.Anything).Return(errors.New("some error"))
		sut := labeldef.NewService(mockRepository, nil, nil, nil, mockUID)
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
		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
		// WHEN
		actual, err := sut.Get(ctx, "tenant", "key")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, &given, actual)
	})

	t.Run("success when getting scenarios LD if it doesn't exist", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		testTenant := "tenant"
		testKey := model.ScenariosKey
		given := model.LabelDefinition{
			ID:     "foo",
			Key:    testKey,
			Tenant: testTenant,
			Schema: fixBasicSchema(t),
		}

		mockRepository := &automock.Repository{}
		mockRepository.On("GetByKey", ctx, testTenant, testKey).Return(&given, nil).Once()
		mockScenariosSvc := &automock.ScenariosService{}
		mockScenariosSvc.On("EnsureScenariosLabelDefinitionExists", ctx, testTenant).Return(nil).Once()
		defer mock.AssertExpectationsForObjects(t, mockRepository, mockScenariosSvc)

		sut := labeldef.NewService(mockRepository, nil, nil, mockScenariosSvc, nil)

		// WHEN
		actual, err := sut.Get(ctx, testTenant, testKey)

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

	t.Run("error when getting scenarios LD if it doesn't exist and ensuring fails", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		testTenant := "tenant"
		testKey := model.ScenariosKey

		testError := errors.New("some error")

		mockScenariosSvc := &automock.ScenariosService{}
		mockScenariosSvc.On("EnsureScenariosLabelDefinitionExists", ctx, testTenant).Return(testError).Once()
		defer mock.AssertExpectationsForObjects(t, mockScenariosSvc)

		sut := labeldef.NewService(nil, nil, nil, mockScenariosSvc, nil)

		// WHEN
		actual, err := sut.Get(ctx, testTenant, testKey)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		assert.Nil(t, actual)
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
		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil, nil, nil)
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
		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil, nil, nil)
		// WHEN
		err := sut.Update(ctx, in)
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, `Invalid data [reason=label with key="oldProperty" is not valid against new schema for Runtime with ID="foo": (root): nonExistingProp is required]`)
	})

	t.Run("returns error when error occurred during receiving Label Definition", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(nil, errors.New("some error"))
		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
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
		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
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

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil, nil, nil)
		// WHEN
		err := sut.Update(context.TODO(), *ld)
		// THEN
		require.EqualError(t, err, "while updating Label Definition: some error")
	})

	t.Run("label definition schema for update is nil", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)

		ld := &model.LabelDefinition{
			ID:     "8b131225-f09d-4035-8091-1f12933863b3",
			Tenant: tenant,
			Key:    key,
		}

		mockRepository.On("GetByKey", context.TODO(), tenant, key).Return(ld, nil).Once()
		mockRepository.On("Update", context.TODO(), *ld).Return(nil).Once()

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
		// WHEN
		err := sut.Update(context.TODO(), *ld)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success when updating scenarios label definition and no automatic assignment exists", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		mockLabelRepo := &automock.LabelRepository{}
		mockScenarioAssignmentLister := &automock.ScenarioAssignmentLister{}
		defer mock.AssertExpectationsForObjects(t, mockRepository, mockLabelRepo, mockScenarioAssignmentLister)
		sut := labeldef.NewService(mockRepository, mockLabelRepo, mockScenarioAssignmentLister, nil, nil)

		defaultLD := fixDefaultScenariosLabelDefinition(tenant)
		ld := fixModifiedScenariosLabelDefinition(tenant)
		mockRepository.On("GetByKey", mock.Anything, tenant, model.ScenariosKey).Return(&defaultLD, nil)
		mockRepository.On("Update", context.TODO(), ld).Return(nil).Once()

		mockLabelRepo.On("ListByKey", mock.Anything, tenant, model.ScenariosKey).Return(nil, nil)
		mockScenarioAssignmentLister.On("List", mock.Anything, tenant, 100, "").Return(&model.AutomaticScenarioAssignmentPage{
			TotalCount: 0,
			PageInfo: &pagination.Page{
				HasNextPage: false,
			},
		}, nil)
		// WHEN
		err := sut.Update(context.TODO(), fixModifiedScenariosLabelDefinition(tenant))
		// THEN
		require.NoError(t, err)
	})

	t.Run("success when updating scenarios label definition and all automatic assignments are valid", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		mockLabelRepo := &automock.LabelRepository{}
		mockScenarioAssignmentLister := &automock.ScenarioAssignmentLister{}
		defer mock.AssertExpectationsForObjects(t, mockRepository, mockLabelRepo, mockScenarioAssignmentLister)
		sut := labeldef.NewService(mockRepository, mockLabelRepo, mockScenarioAssignmentLister, nil, nil)

		defaultLD := fixDefaultScenariosLabelDefinition(tenant)
		ld := fixModifiedScenariosLabelDefinition(tenant)
		mockRepository.On("GetByKey", mock.Anything, tenant, model.ScenariosKey).Return(&defaultLD, nil)
		mockRepository.On("Update", context.TODO(), ld).Return(nil).Once()

		mockLabelRepo.On("ListByKey", mock.Anything, tenant, model.ScenariosKey).Return(nil, nil)
		mockScenarioAssignmentLister.On("List", mock.Anything, tenant, 100, "").Return(&model.AutomaticScenarioAssignmentPage{
			TotalCount: 200,
			Data:       []*model.AutomaticScenarioAssignment{{ScenarioName: "scenario-A"}},
			PageInfo: &pagination.Page{
				HasNextPage: true,
				EndCursor:   "secondPage",
			},
		}, nil).Once()
		mockScenarioAssignmentLister.On("List", mock.Anything, tenant, 100, "secondPage").Return(&model.AutomaticScenarioAssignmentPage{
			TotalCount: 200,
			Data:       []*model.AutomaticScenarioAssignment{{ScenarioName: "scenario-B"}},
			PageInfo: &pagination.Page{
				HasNextPage: false,
			},
		}, nil).Once()
		// WHEN
		err := sut.Update(context.TODO(), ld)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error if cannot fetch automatic assignments", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		mockLabelRepo := &automock.LabelRepository{}
		mockScenarioAssignmentLister := &automock.ScenarioAssignmentLister{}
		defer mock.AssertExpectationsForObjects(t, mockRepository, mockLabelRepo, mockScenarioAssignmentLister)
		sut := labeldef.NewService(mockRepository, mockLabelRepo, mockScenarioAssignmentLister, nil, nil)

		defaultLD := fixDefaultScenariosLabelDefinition(tenant)
		ld := fixModifiedScenariosLabelDefinition(tenant)
		mockRepository.On("GetByKey", mock.Anything, tenant, model.ScenariosKey).Return(&defaultLD, nil)

		mockLabelRepo.On("ListByKey", mock.Anything, tenant, model.ScenariosKey).Return(nil, nil)
		mockScenarioAssignmentLister.On("List", mock.Anything, tenant, 100, "").Return(nil, errors.New("some error")).Once()

		// WHEN
		err := sut.Update(context.TODO(), ld)
		// THEN
		require.EqualError(t, err, "while validating Scenario Assignments against a new schema: while getting page of Automatic Scenario Assignments: some error")
	})

	t.Run("returns error if automatic assignment is not valid against a new schema", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		mockLabelRepo := &automock.LabelRepository{}
		mockScenarioAssignmentLister := &automock.ScenarioAssignmentLister{}
		defer mock.AssertExpectationsForObjects(t, mockRepository, mockLabelRepo, mockScenarioAssignmentLister)
		sut := labeldef.NewService(mockRepository, mockLabelRepo, mockScenarioAssignmentLister, nil, nil)

		defaultLD := fixDefaultScenariosLabelDefinition(tenant)
		ld := fixModifiedScenariosLabelDefinition(tenant)
		mockRepository.On("GetByKey", mock.Anything, tenant, model.ScenariosKey).Return(&defaultLD, nil)

		mockLabelRepo.On("ListByKey", mock.Anything, tenant, model.ScenariosKey).Return(nil, nil)
		mockScenarioAssignmentLister.On("List", mock.Anything, tenant, 100, "").Return(&model.AutomaticScenarioAssignmentPage{
			TotalCount: 1,
			Data:       []*model.AutomaticScenarioAssignment{{ScenarioName: "scenario-that-is-invalid"}},
			PageInfo: &pagination.Page{
				HasNextPage: false,
			},
		}, nil).Once()
		// WHEN
		err := sut.Update(context.TODO(), ld)
		// THEN
		require.EqualError(t, err, "while validating Scenario Assignments against a new schema: Scenario Assignment [scenario=scenario-that-is-invalid] is not valid against a new schema: 0: 0 must be one of the following: \"DEFAULT\", \"scenario-A\", \"scenario-B\"")
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

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil, nil, nil)
		// WHEN
		err := sut.Delete(ctx, tnt, given.Key, deleteRelatedResources)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success when some labels use labeldef", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		tnt := "tenant"
		key := "key"
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    key,
			Tenant: tnt,
		}

		deleteRelatedResources := true
		mockRepository.On("GetByKey", ctx, tnt, given.Key).Return(&given, nil).Once()
		mockRepository.On("DeleteByKey", ctx, tnt, given.Key).Return(nil).Once()
		mockLabelRepository.On("DeleteByKey", ctx, tnt, given.Key).Return(nil).Once()
		mockLabelRepository.On("ListByKey", ctx, tnt, given.Key).Return([]*model.Label{}, nil).Once()

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil, nil, nil)
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

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
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

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil, nil, nil)
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

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil, nil, nil)
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

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
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

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil)
		// WHEN
		err := sut.Delete(ctx, tnt, given.Key, deleteRelatedResources)
		// THEN
		require.Error(t, err)
	})

	t.Run("error during listing labels when trying to delete related resources", func(t *testing.T) {
		// GIVEN
		testErr := errors.New("testErr")

		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockLabelRepository := &automock.LabelRepository{}
		defer mockLabelRepository.AssertExpectations(t)

		tnt := "tenant"
		key := "key"
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    key,
			Tenant: tnt,
		}
		deleteRelatedResources := true
		mockRepository.On("GetByKey", ctx, tnt, given.Key).Return(&given, nil).Once()
		mockLabelRepository.On("DeleteByKey", ctx, tnt, given.Key).Return(testErr).Once()

		sut := labeldef.NewService(mockRepository, mockLabelRepository, nil, nil, nil)
		// WHEN
		err := sut.Delete(ctx, tnt, given.Key, deleteRelatedResources)
		// THEN
		require.Error(t, err)
		errMsg := fmt.Sprintf("while deleting labels with key \"%s\": %s", key, testErr)
		assert.EqualError(t, err, errMsg)
	})
}

func fixUUID() string {
	return "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
}

func fixBasicInputSchema() *graphql.JSONSchema {
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
	jsonSchema := graphql.JSONSchema(sch)
	return &jsonSchema
}

func fixBasicSchema(t *testing.T) *interface{} {
	sch := fixBasicInputSchema()
	require.NotNil(t, sch)
	var obj map[string]interface{}

	err := json.Unmarshal([]byte(*sch), &obj)
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

func fixDefaultScenariosLabelDefinition(tenantID string) model.LabelDefinition {
	var schema interface{} = model.ScenariosSchema
	ld := model.LabelDefinition{
		Key:    model.ScenariosKey,
		Tenant: tenantID,
		Schema: &schema,
	}
	return ld
}

func fixModifiedScenariosLabelDefinition(tenantID string) model.LabelDefinition {
	m := map[string]interface{}{
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
		"items": map[string]interface{}{
			"type":      "string",
			"pattern":   "^[A-Za-z0-9]([-_A-Za-z0-9\\s]*[A-Za-z0-9])$",
			"enum":      []string{"DEFAULT", "scenario-A", "scenario-B"},
			"maxLength": 128,
		},
	}

	var schema interface{} = m
	ld := model.LabelDefinition{
		Key:    model.ScenariosKey,
		Tenant: tenantID,
		Schema: &schema,
	}
	return ld
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
