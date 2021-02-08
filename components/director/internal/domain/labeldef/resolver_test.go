package labeldef_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	pautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
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
	externalTnt := "external-tenant"
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully created Label Definition", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Create", contextThatHasTenant(tnt), model.LabelDefinition{Key: "scenarios", Tenant: tnt}).
			Return(model.LabelDefinition{Key: "scenarios", Tenant: tnt, ID: "id"}, nil)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    "scenarios",
			Tenant: tnt,
		}, nil)
		mockConverter.On("ToGraphQL", model.LabelDefinition{Key: "scenarios", Tenant: tnt, ID: "id"}).Return(graphql.LabelDefinition{
			Key: "scenarios",
		}, nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockConverter)
		// WHEN
		actual, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.NoError(t, err)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		assert.Equal(t, "scenarios", actual.Key)
	})

	t.Run("got error on converting to graphql", func(t *testing.T) {
		// GIVEN
		ldModel := model.LabelDefinition{Key: "scenarios", Tenant: tnt, ID: "id"}
		persist, transact := txGen.ThatDoesntExpectCommit()
		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Create", contextThatHasTenant(tnt), model.LabelDefinition{Key: "scenarios", Tenant: tnt}).
			Return(ldModel, nil)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{Key: "scenarios", Tenant: tnt}, nil)
		mockConverter.On("ToGraphQL", ldModel).Return(graphql.LabelDefinition{}, testErr)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockConverter)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, testErr.Error())
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
	})

	t.Run("got error when missing tenant in context", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewResolver(nil, nil, nil)
		// WHEN
		_, err := sut.CreateLabelDefinition(context.TODO(), graphql.LabelDefinitionInput{})
		// THEN
		require.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()
		ctx := tenant.SaveToContext(context.TODO(), "tenant", "external-tenant")

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		sut := labeldef.NewResolver(transact, nil, mockConverter)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, graphql.LabelDefinitionInput{Key: "scenarios"})
		// THEN
		require.EqualError(t, err, "while starting transaction: test error")
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("got error on creating Label Definition", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()
		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Create", mock.Anything, model.LabelDefinition{Key: "scenarios", Tenant: tnt}).
			Return(model.LabelDefinition{}, errors.New("some error"))

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    "scenarios",
			Tenant: tnt,
		}, nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockConverter)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
		require.EqualError(t, err, "while creating label definition: some error")
	})

	t.Run("got error on converting to model", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()
		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{}, errors.New("json schema is not valid"))

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, nil, mockConverter)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.EqualError(t, err, "json schema is not valid")
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("got error on committing transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnCommit()
		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Create", contextThatHasTenant(tnt), model.LabelDefinition{Key: "scenarios", Tenant: tnt}).
			Return(model.LabelDefinition{Key: "scenarios", Tenant: tnt, ID: "id"}, nil)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    "scenarios",
			Tenant: tnt,
		}, nil)
		mockConverter.On("ToGraphQL", model.LabelDefinition{Key: "scenarios", Tenant: tnt, ID: "id"}).Return(graphql.LabelDefinition{
			Key: "scenarios",
		}, nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockConverter)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.EqualError(t, err, "while committing transaction: test error")
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})
}

func TestQueryLabelDefinitions(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tenant"
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully returns definitions", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
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

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", givenModels[0]).Return(graphql.LabelDefinition{
			Key: "key1",
		}, nil)
		mockConverter.On("ToGraphQL", givenModels[1]).Return(graphql.LabelDefinition{
			Key: "key2",
		}, nil)

		sut := labeldef.NewResolver(transact, mockService, mockConverter)
		// WHEN
		actual, err := sut.LabelDefinitions(ctx)
		// THEN
		require.NoError(t, err)
		require.Len(t, actual, 2)
		assert.Equal(t, actual[0].Key, "key1")
		assert.Equal(t, actual[1].Key, "key2")
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("got error when convert to graphql failed", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		givenModels := []model.LabelDefinition{{ID: "id1", Key: "key1", Tenant: tnt}, {ID: "id2", Key: "key2", Tenant: tnt}}

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("List",
			contextThatHasTenant(tnt),
			tnt).Return(givenModels, nil)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", givenModels[0]).Return(graphql.LabelDefinition{Key: "key1"}, nil)
		mockConverter.On("ToGraphQL", givenModels[1]).Return(graphql.LabelDefinition{}, testErr)

		sut := labeldef.NewResolver(transact, mockService, mockConverter)
		// WHEN
		_, err := sut.LabelDefinitions(ctx)
		// THEN
		require.Error(t, err)
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("successfully returns empty slice if no definitions", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("List",
			contextThatHasTenant(tnt),
			tnt).Return(nil, nil)

		sut := labeldef.NewResolver(transact, mockService, nil)
		// WHEN
		actual, err := sut.LabelDefinitions(ctx)
		// THEN
		require.NoError(t, err)
		require.Empty(t, actual)
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("got error when missing tenant in context", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewResolver(nil, nil, nil)
		// WHEN
		_, err := sut.LabelDefinitions(context.TODO())
		// THEN
		require.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()
		ctx := tenant.SaveToContext(context.TODO(), "tenant", "external-tenant")
		sut := labeldef.NewResolver(transact, nil, nil)
		// WHEN
		_, err := sut.LabelDefinitions(ctx)
		// THEN
		require.EqualError(t, err, "while starting transaction: test error")
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("got error on getting definitions from service", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()
		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("List", contextThatHasTenant(tnt), tnt).
			Return(nil, testErr)

		sut := labeldef.NewResolver(transact, mockService, nil)
		// WHEN
		_, err := sut.LabelDefinitions(ctx)
		// THEN
		require.EqualError(t, err, "while listing Label Definitions: test error")
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("got error on committing transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnCommit()
		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("List", contextThatHasTenant(tnt), tnt).Return(nil, nil)

		sut := labeldef.NewResolver(transact, mockService, nil)
		// WHEN
		_, err := sut.LabelDefinitions(ctx)
		// THEN
		require.EqualError(t, err, "while committing transaction: test error")
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})
}

func TestQueryGivenLabelDefinition(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tenant"
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	t.Run("successfully returns single definition", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(nil)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
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

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", *givenModel).Return(graphql.LabelDefinition{
			Key: "key",
		}, nil)

		sut := labeldef.NewResolver(mockTransactioner, mockService, mockConverter)
		// WHEN
		actual, err := sut.LabelDefinition(ctx, "key")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "key", actual.Key)
		assert.Nil(t, actual.Schema)
	})

	t.Run("returns error when convert to graphql failed", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		givenModel := &model.LabelDefinition{ID: "id", Key: "key", Tenant: tnt}

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, "key").Return(givenModel, nil)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", *givenModel).Return(graphql.LabelDefinition{}, testErr)

		sut := labeldef.NewResolver(transact, mockService, mockConverter)
		// WHEN
		_, err := sut.LabelDefinition(ctx, "key")
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, testErr.Error())
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("returns nil if definition does not exist", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockPersistanceCtx.On("Commit").Return(nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, "key").Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))

		sut := labeldef.NewResolver(mockTransactioner, mockService, nil)
		// WHEN
		_, err := sut.LabelDefinition(ctx, "key")
		// THEN
		require.Nil(t, err)
	})

	t.Run("got error on getting label definition from service", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("error from service"))

		sut := labeldef.NewResolver(mockTransactioner, mockService, nil)
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
		require.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		mockTransactioner := getInvalidMockTransactioner()
		defer mockTransactioner.AssertExpectations(t)
		ctx := tenant.SaveToContext(context.TODO(), "tenant", "external-tenant")
		sut := labeldef.NewResolver(mockTransactioner, nil, nil)
		// WHEN
		_, err := sut.LabelDefinition(ctx, "anything")
		// THEN
		require.EqualError(t, err, "while starting transaction: some error")
	})

	t.Run("got error on committing transaction", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(errors.New("commit errror"))

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, "key").Return(nil, nil)

		sut := labeldef.NewResolver(mockTransactioner, mockService, nil)
		// WHEN
		_, err := sut.LabelDefinition(ctx, "key")
		// THEN
		require.EqualError(t, err, "while committing transaction: commit errror")
	})
}

func TestResolver_DeleteLabelDefinition(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tenant"
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("success", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
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

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", *givenModel).Return(graphql.LabelDefinition{
			Key: givenModel.Key,
		}, nil)

		sut := labeldef.NewResolver(transact, mockService, mockConverter)
		// WHEN
		actual, err := sut.DeleteLabelDefinition(ctx, givenModel.Key, &deleteRelatedLabels)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "key", actual.Key)
		assert.Nil(t, actual.Schema)
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("error when convertion to graphql failed", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
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

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", *givenModel).Return(graphql.LabelDefinition{}, testErr)

		sut := labeldef.NewResolver(transact, mockService, mockConverter)
		// WHEN
		_, err := sut.DeleteLabelDefinition(ctx, givenModel.Key, &deleteRelatedLabels)
		// THEN
		require.Error(t, err)
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("error when get failed", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
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

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)

		sut := labeldef.NewResolver(mockTransactioner, mockService, mockConverter)
		// WHEN
		_, err := sut.DeleteLabelDefinition(ctx, givenModel.Key, &deleteRelatedLabels)
		// THEN
		require.Error(t, err)
	})

	t.Run("error when label definition not found", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
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

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)

		sut := labeldef.NewResolver(mockTransactioner, mockService, mockConverter)
		// WHEN
		_, err := sut.DeleteLabelDefinition(ctx, givenModel.Key, &deleteRelatedLabels)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("error when label definition delete failed", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
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

		sut := labeldef.NewResolver(mockTransactioner, mockService, nil)
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
		require.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		mockTransactioner := getInvalidMockTransactioner()
		defer mockTransactioner.AssertExpectations(t)
		ctx := tenant.SaveToContext(context.TODO(), "tenant", "external-tenant")
		deleteRelatedLabels := true
		sut := labeldef.NewResolver(mockTransactioner, nil, nil)
		// WHEN
		_, err := sut.DeleteLabelDefinition(ctx, "test", &deleteRelatedLabels)
		// THEN
		require.EqualError(t, err, "while starting transaction: some error")
	})

	t.Run("got error on committing transaction", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(errors.New("commit errror"))

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
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

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToGraphQL", *givenModel).Return(graphql.LabelDefinition{
			Key: givenModel.Key,
		}, nil)

		sut := labeldef.NewResolver(mockTransactioner, mockService, mockConverter)
		// WHEN
		_, err := sut.DeleteLabelDefinition(ctx, givenModel.Key, &deleteRelatedLabels)
		// THEN
		require.EqualError(t, err, "while committing transaction: commit errror")
	})
}

func TestUpdateLabelDefinition(t *testing.T) {

	tnt := "tenant"
	externalTnt := "external-tenant"
	gqlLabelDefinitionInput := graphql.LabelDefinitionInput{
		Key:    "key",
		Schema: fixBasicInputSchema(),
	}
	modelLabelDefinition := model.LabelDefinition{
		Key:    "key",
		Schema: fixBasicSchema(t),
	}
	updatedGQLLabelDefinition := graphql.LabelDefinition{
		Key:    "key",
		Schema: fixBasicInputSchema(),
	}
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully updated Label Definition", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(nil)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(modelLabelDefinition, nil)
		mockConverter.On("ToGraphQL", modelLabelDefinition).Return(updatedGQLLabelDefinition, nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Update", contextThatHasTenant(tnt), modelLabelDefinition).Return(nil)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&modelLabelDefinition, nil).Once()

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(mockTransactioner, mockService, mockConverter)
		// WHEN
		actual, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "key", actual.Key)
	})

	t.Run("got error when convert to graphql failed", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(modelLabelDefinition, nil)
		mockConverter.On("ToGraphQL", modelLabelDefinition).Return(graphql.LabelDefinition{}, testErr)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Update", contextThatHasTenant(tnt), modelLabelDefinition).Return(nil)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&modelLabelDefinition, nil).Once()

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockConverter)
		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, testErr.Error())
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("missing tenant in context", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewResolver(nil, nil, nil)
		// WHEN
		_, err := sut.UpdateLabelDefinition(context.TODO(), graphql.LabelDefinitionInput{})
		// THEN
		require.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(nil, errors.New("some error"))
		defer mockTransactioner.AssertExpectations(t)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(mockTransactioner, nil, mockConverter)
		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.EqualError(t, err, "while starting transaction: some error")
	})

	t.Run("got error when convert to model failed", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(model.LabelDefinition{}, testErr)

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(transact, nil, mockConverter)
		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, testErr.Error())
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("got error on updating Label Definition", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(modelLabelDefinition, nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Update", contextThatHasTenant(tnt), modelLabelDefinition).Return(errors.New("some error"))

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(mockTransactioner, mockService, mockConverter)
		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.EqualError(t, err, "while updating label definition: some error")
	})

	t.Run("got error on getting Label Definition", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(modelLabelDefinition, nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Update", contextThatHasTenant(tnt), modelLabelDefinition).Return(nil)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(nil, testErr)

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockConverter)
		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "while receiving updated label definition: test error")
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("got error on committing transaction", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockPersistanceCtx.On("Commit").Return(errors.New("error on commit"))

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(nil)
		defer mockTransactioner.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Update", contextThatHasTenant(tnt), modelLabelDefinition).Return(nil)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&modelLabelDefinition, nil)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(modelLabelDefinition, nil)
		mockConverter.On("ToGraphQL", modelLabelDefinition).Return(updatedGQLLabelDefinition, nil)

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(mockTransactioner, mockService, mockConverter)
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
