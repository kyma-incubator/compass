package labeldef_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
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
		mockService.On("Create", mock.Anything /*ctx TODO*/, model.LabelDefinition{Key: "scenarios", Tenant: tnt}).
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
