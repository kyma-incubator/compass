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
		mockService.On("Create", mock.Anything /*ctx TODO*/, model.LabelDefinition{Key: "scenarios", Tenant: tnt}).
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

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt)
		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		actual, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "scenarios", actual.Key)
	})
	t.Run("missing tenant in context", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewResolver(nil, nil, nil)
		// WHEN
		_, err := sut.CreateLabelDefinition(context.TODO(), graphql.LabelDefinitionInput{})
		// THEN
		require.EqualError(t, err, "Cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(nil, errors.New("some error"))
		defer mockTransactioner.AssertExpectations(t)
		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, "tenant")
		sut := labeldef.NewResolver(nil, nil, mockTransactioner)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, graphql.LabelDefinitionInput{})
		// THEN
		require.EqualError(t, err, "while starting transaction: some error")
	})

	t.Run("got error on creating Label Definition", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTxOp{}
		defer mockPersistanceCtx.AssertExpectations(t)
		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommited", mockPersistanceCtx).Return(nil)

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
		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt)
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

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt)
		sut := labeldef.NewResolver(mockService, mockConverter, mockTransactioner)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.EqualError(t, err, "while committing transaction: error on commit")
	})
}
