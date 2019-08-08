package labeldef_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	pautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateLabelDefinition(t *testing.T) {
	labelDefInput := graphql.LabelDefinitionInput{
		Key: "scenarios",
	}
	tnt := "tenant"
	t.Run("successfully created Label Definition", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(nil)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Create", contextThatHasTenant(tnt), model.LabelDefinition{Key: "scenarios", Tenant: tnt}).
			Return(model.LabelDefinition{Key: "scenarios", Tenant: tnt, ID: "id"}, nil)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    "scenarios",
			Tenant: tnt,
		})
		mockConverter.On("ToGraphQL", model.LabelDefinition{Key: "scenarios", Tenant: tnt, ID: "id"}).Return(graphql.LabelDefinition{
			Key: "scenarios",
		})

		ctx := tenant.SaveToContext(context.TODO(), tnt)
		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		actual, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "scenarios", actual.Key)
	})
	t.Run("got error when missing tenant in context", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewResolver(nil, nil, nil)
		// WHEN
		_, err := sut.CreateLabelDefinition(context.TODO(), graphql.LabelDefinitionInput{})
		// THEN
		require.EqualError(t, err, "Cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		mockTransactioner := getInvalidMockTransactioner()
		defer mockTransactioner.AssertExpectations(t)
		ctx := tenant.SaveToContext(context.TODO(), "tenant")
		sut := labeldef.NewResolver(nil, nil, mockTransactioner)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, graphql.LabelDefinitionInput{})
		// THEN
		require.EqualError(t, err, "while starting transaction: some error")
	})

	t.Run("got error on creating Label Definition", func(t *testing.T) {
		// GIVEN
		mockPersistenceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistenceCtx.AssertExpectations(t)
		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistenceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mockPersistenceCtx).Return(nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Create", mock.Anything, model.LabelDefinition{Key: "scenarios", Tenant: tnt}).
			Return(model.LabelDefinition{}, errors.New("some error"))

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    "scenarios",
			Tenant: tnt,
		})

		defer mockTransactioner.AssertExpectations(t)
		ctx := tenant.SaveToContext(context.TODO(), tnt)
		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.EqualError(t, err, "while creating label definition: some error")
	})

	t.Run("got error on committing transaction", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(errors.New("error on commit"))

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Create", contextThatHasTenant(tnt), model.LabelDefinition{Key: "scenarios", Tenant: tnt}).
			Return(model.LabelDefinition{Key: "scenarios", Tenant: tnt, ID: "id"}, nil)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    "scenarios",
			Tenant: tnt,
		})

		ctx := tenant.SaveToContext(context.TODO(), tnt)
		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.EqualError(t, err, "while committing transaction: error on commit")
	})
}

func TestQueryLabelDefinitions(t *testing.T) {
	tnt := "tenant"
	t.Run("successfully returns definitions", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(nil)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)
		givenModels := []model.LabelDefinition{{
			ID:     "id1",
			Key:    "key1",
			Tenant: tnt,
		}, {ID: "id2", Key: "key2", Tenant: tnt}}

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("List",
			contextThatHasTenant(tnt),
			tnt).Return(givenModels, nil)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", givenModels[0]).Return(graphql.LabelDefinition{
			Key: "key1",
		})
		mockConverter.On("ToGraphQL", givenModels[1]).Return(graphql.LabelDefinition{
			Key: "key2",
		})

		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		actual, err := sut.LabelDefinitions(ctx)
		// THEN
		require.NoError(t, err)
		require.Len(t, actual, 2)
		assert.Equal(t, actual[0].Key, "key1")
		assert.Equal(t, actual[1].Key, "key2")

	})

	t.Run("successfully returns empty slice if no definitions", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(nil)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("List",
			contextThatHasTenant(tnt),
			tnt).Return(nil, nil)

		sut := labeldef.NewResolver(mockService, nil, mockTransactioner)
		// WHEN
		actual, err := sut.LabelDefinitions(ctx)
		// THEN
		require.NoError(t, err)
		require.Empty(t, actual)
	})

	t.Run("got error when missing tenant in context", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewResolver(nil, nil, nil)
		// WHEN
		_, err := sut.LabelDefinitions(context.TODO())
		// THEN
		require.EqualError(t, err, "Cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		mockTransactioner := getInvalidMockTransactioner()
		defer mockTransactioner.AssertExpectations(t)
		ctx := tenant.SaveToContext(context.TODO(), "tenant")
		sut := labeldef.NewResolver(nil, nil, mockTransactioner)
		// WHEN
		_, err := sut.LabelDefinitions(ctx)
		// THEN
		require.EqualError(t, err, "while starting transaction: some error")
	})

	t.Run("got error on getting definitions from service", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("List",
			contextThatHasTenant(tnt),
			tnt).Return(nil, errors.New("error on list"))

		sut := labeldef.NewResolver(mockService, nil, mockTransactioner)
		// WHEN
		_, err := sut.LabelDefinitions(ctx)
		// THEN
		require.EqualError(t, err, "while listing Label Definitions: error on list")

	})

	t.Run("got error on committing transaction", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(errors.New("commit error"))

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("List",
			contextThatHasTenant(tnt),
			tnt).Return(nil, nil)

		sut := labeldef.NewResolver(mockService, nil, mockTransactioner)
		// WHEN
		_, err := sut.LabelDefinitions(ctx)
		// THEN
		require.EqualError(t, err, "while committing transaction: commit error")

	})
}

func TestQueryGivenLabelDefinition(t *testing.T) {
	tnt := "tenant"
	t.Run("successfully returns single definition", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(nil)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)
		givenModel := &model.LabelDefinition{
			ID:     "id",
			Key:    "key",
			Tenant: tnt,
		}

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, "key").Return(givenModel, nil)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", *givenModel).Return(graphql.LabelDefinition{
			Key: "key",
		})

		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		actual, err := sut.LabelDefinition(ctx, "key")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "key", actual.Key)
		assert.Nil(t, actual.Schema)
	})

	t.Run("returns error if definition does not exist", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(nil)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, "key").Return(nil, nil)

		sut := labeldef.NewResolver(mockService, nil, mockTransactioner)
		// WHEN
		_, err := sut.LabelDefinition(ctx, "key")
		// THEN
		require.EqualError(t, err, "label definition with key 'key' does not exist")
	})

	t.Run("got error on getting label definition from service", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("error from service"))

		sut := labeldef.NewResolver(mockService, nil, mockTransactioner)
		// WHEN
		_, err := sut.LabelDefinition(ctx, "key")
		// THEN
		require.EqualError(t, err, "while getting Label Definition: error from service")
	})

	t.Run("got error when missing tenant in context", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewResolver(nil, nil, nil)
		// WHEN
		_, err := sut.LabelDefinition(context.TODO(), "anything")
		// THEN
		require.EqualError(t, err, "Cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		mockTransactioner := getInvalidMockTransactioner()
		defer mockTransactioner.AssertExpectations(t)
		ctx := tenant.SaveToContext(context.TODO(), "tenant")
		sut := labeldef.NewResolver(nil, nil, mockTransactioner)
		// WHEN
		_, err := sut.LabelDefinition(ctx, "anything")
		// THEN
		require.EqualError(t, err, "while starting transaction: some error")
	})

	t.Run("got error on committing transaction", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(errors.New("commit errror"))

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, "key").Return(nil, nil)

		sut := labeldef.NewResolver(mockService, nil, mockTransactioner)
		// WHEN
		_, err := sut.LabelDefinition(ctx, "key")
		// THEN
		require.EqualError(t, err, "while committing transaction: commit errror")

	})
}

func TestResolver_DeleteLabelDefinition(t *testing.T) {
	tnt := "tenant"

	t.Run("success", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(nil)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)
		givenModel := &model.LabelDefinition{
			ID:     "id",
			Key:    "key",
			Tenant: tnt,
		}
		deleteRelatedLabels := true

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, givenModel.Key).Return(givenModel, nil)
		mockService.On("Delete", contextThatHasTenant(tnt), tnt, givenModel.Key, deleteRelatedLabels).Return(nil).Once()

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", *givenModel).Return(graphql.LabelDefinition{
			Key: givenModel.Key,
		})

		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		actual, err := sut.DeleteLabelDefinition(ctx, givenModel.Key, &deleteRelatedLabels)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "key", actual.Key)
		assert.Nil(t, actual.Schema)
	})

	t.Run("error when get failed", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)
		givenModel := &model.LabelDefinition{
			ID:     "id",
			Key:    "key",
			Tenant: tnt,
		}
		deleteRelatedLabels := true

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, givenModel.Key).Return(nil, errors.New("test"))

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		_, err := sut.DeleteLabelDefinition(ctx, givenModel.Key, &deleteRelatedLabels)
		// THEN
		require.Error(t, err)
	})

	t.Run("error when label definition not found", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)
		givenModel := &model.LabelDefinition{
			ID:     "id",
			Key:    "key",
			Tenant: tnt,
		}
		deleteRelatedLabels := true

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, givenModel.Key).Return(nil, nil)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		_, err := sut.DeleteLabelDefinition(ctx, givenModel.Key, &deleteRelatedLabels)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("error when label definition delete failed", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)
		givenModel := &model.LabelDefinition{
			ID:     "id",
			Key:    "key",
			Tenant: tnt,
		}
		deleteRelatedLabels := true

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, givenModel.Key).Return(givenModel, nil)
		mockService.On("Delete", contextThatHasTenant(tnt), tnt, givenModel.Key, deleteRelatedLabels).Return(errors.New("test")).Once()

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", *givenModel).Return(graphql.LabelDefinition{
			Key: givenModel.Key,
		})

		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		_, err := sut.DeleteLabelDefinition(ctx, givenModel.Key, &deleteRelatedLabels)
		// THEN
		require.Error(t, err)
	})

	t.Run("got error when missing tenant in context", func(t *testing.T) {
		// GIVEN
		deleteRelatedLabels := true
		sut := labeldef.NewResolver(nil, nil, nil)
		// WHEN
		_, err := sut.DeleteLabelDefinition(context.TODO(), "test", &deleteRelatedLabels)
		// THEN
		require.EqualError(t, err, "Cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		mockTransactioner := getInvalidMockTransactioner()
		defer mockTransactioner.AssertExpectations(t)
		ctx := tenant.SaveToContext(context.TODO(), "tenant")
		deleteRelatedLabels := true
		sut := labeldef.NewResolver(nil, nil, mockTransactioner)
		// WHEN
		_, err := sut.DeleteLabelDefinition(ctx, "test", &deleteRelatedLabels)
		// THEN
		require.EqualError(t, err, "while starting transaction: some error")
	})

	t.Run("got error on committing transaction", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(errors.New("commit errror"))

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt)
		givenModel := &model.LabelDefinition{
			ID:     "id",
			Key:    "key",
			Tenant: tnt,
		}
		deleteRelatedLabels := true

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, givenModel.Key).Return(givenModel, nil)
		mockService.On("Delete", contextThatHasTenant(tnt), tnt, givenModel.Key, deleteRelatedLabels).Return(nil).Once()

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", *givenModel).Return(graphql.LabelDefinition{
			Key: givenModel.Key,
		})

		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		_, err := sut.DeleteLabelDefinition(ctx, givenModel.Key, &deleteRelatedLabels)
		// THEN
		require.EqualError(t, err, "while committing transaction: commit errror")
	})
}

func TestUpdateLabelDefinition(t *testing.T) {

	tnt := "tenant"
	gqlLabelDefinitionInput := graphql.LabelDefinitionInput{
		Key:    "key",
		Schema: fixBasicSchema(t),
	}
	modelLabelDefinition := model.LabelDefinition{
		Key:    "key",
		Schema: fixBasicSchema(t),
	}
	updatedGQLLabelDefinition := graphql.LabelDefinition{
		Key:    "key",
		Schema: fixBasicSchema(t),
	}

	t.Run("successfully updated Label Definition", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(nil)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(modelLabelDefinition)
		mockConverter.On("ToGraphQL", modelLabelDefinition).Return(updatedGQLLabelDefinition)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Update", contextThatHasTenant(tnt), modelLabelDefinition).Return(nil)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&modelLabelDefinition, nil).Once()

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt)
		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		actual, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "key", actual.Key)
	})
	t.Run("missing tenant in context", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewResolver(nil, nil, nil)
		// WHEN
		_, err := sut.UpdateLabelDefinition(context.TODO(), graphql.LabelDefinitionInput{})
		// THEN
		require.EqualError(t, err, "Cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(nil, errors.New("some error"))
		defer mockTransactioner.AssertExpectations(t)
		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt)
		sut := labeldef.NewResolver(nil, nil, mockTransactioner)
		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, graphql.LabelDefinitionInput{})
		// THEN
		require.EqualError(t, err, "while starting transaction: some error")
	})

	t.Run("got error on updating Label Definition", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(modelLabelDefinition)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Update", contextThatHasTenant(tnt), modelLabelDefinition).Return(errors.New("some error"))

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt)
		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.EqualError(t, err, "while updating label definition: some error")

	})

	t.Run("got error on committing transaction", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(errors.New("error on commit"))

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Update", contextThatHasTenant(tnt), modelLabelDefinition).Return(nil)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&modelLabelDefinition, nil)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(modelLabelDefinition)

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt)
		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.EqualError(t, err, "while committing transaction: error on commit")
	})
}

func getInvalidMockTransactioner() *pautomock.Transactioner {
	mockTransactioner := &pautomock.Transactioner{}
	mockTransactioner.On("Begin").Return(nil, errors.New("some error"))
	return mockTransactioner
}

func contextThatHasTenant(expectedTenant string) interface{} {
	return mock.MatchedBy(func(actual context.Context) bool {
		actualTenant, err := tenant.LoadFromContext(actual)
		if err != nil {
			return false
		}
		return actualTenant == expectedTenant
	})
}
