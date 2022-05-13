package labeldef_test

import (
	"context"
	"encoding/json"
	"fmt"
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
	scenariosSchema := model.ScenariosSchema
	var scenarioSchemaInt interface{} = scenariosSchema

	labelDefInput := graphql.LabelDefinitionInput{
		Key:    "scenarios",
		Schema: getJSONSchemaFromSchema(t, &scenarioSchemaInt),
	}
	tnt := "tenant"
	storedModel := model.LabelDefinition{Key: model.ScenariosKey, Tenant: tnt, ID: fixUUID(), Schema: &scenarioSchemaInt}
	externalTnt := "external-tenant"
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully created Label Definition", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockFormationsService := &automock.FormationService{}
		defer mockFormationsService.AssertExpectations(t)

		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    model.ScenariosKey,
			Tenant: tnt,
			Schema: &scenarioSchemaInt,
		}, nil)

		defer mockService.AssertExpectations(t)
		mockService.On("GetWithoutCreating",
			contextThatHasTenant(tnt),
			tnt, model.ScenariosKey).Return(nil, errors.New("Object not found")).Once()

		mockFormationsService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "DEFAULT"}).Return(nil, nil)

		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, model.ScenariosKey).Return(&storedModel, nil).Once()

		mockConverter.On("ToGraphQL", storedModel).Return(graphql.LabelDefinition{
			Key:    model.ScenariosKey,
			Schema: getJSONSchemaFromSchema(t, &scenarioSchemaInt),
		}, nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockFormationsService, mockConverter)
		// WHEN
		actual, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.NoError(t, err)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		assert.Equal(t, model.ScenariosKey, actual.Key)
	})

	t.Run("got error on converting to graphql", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()
		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)

		mockFormationsService := &automock.FormationService{}
		defer mockFormationsService.AssertExpectations(t)
		mockService.On("GetWithoutCreating",
			contextThatHasTenant(tnt),
			tnt, model.ScenariosKey).Return(nil, errors.New("Object not found")).Once()

		mockFormationsService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "DEFAULT"}).Return(nil, nil)

		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, model.ScenariosKey).Return(&storedModel, nil).Once()

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    model.ScenariosKey,
			Tenant: tnt,
			Schema: &scenarioSchemaInt,
		}, nil)
		mockConverter.On("ToGraphQL", storedModel).Return(graphql.LabelDefinition{}, testErr)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockFormationsService, mockConverter)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, testErr.Error())
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
	})

	t.Run("got error different from not found while getting label definition", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    model.ScenariosKey,
			Tenant: tnt,
			Schema: &scenarioSchemaInt,
		}, nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("GetWithoutCreating",
			contextThatHasTenant(tnt),
			tnt, model.ScenariosKey).Return(nil, errors.New("db error")).Once()

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, nil, mockConverter)

		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "while getting label definition: db error")
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
	})

	t.Run("got error during formation creating", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockFormationsService := &automock.FormationService{}
		defer mockFormationsService.AssertExpectations(t)

		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    model.ScenariosKey,
			Tenant: tnt,
			Schema: &scenarioSchemaInt,
		}, nil)

		defer mockService.AssertExpectations(t)
		mockService.On("GetWithoutCreating",
			contextThatHasTenant(tnt),
			tnt, model.ScenariosKey).Return(nil, errors.New("Object not found")).Once()

		mockFormationsService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "DEFAULT"}).Return(nil, testErr)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockFormationsService, mockConverter)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		require.Error(t, err)
		require.EqualError(t, err, fmt.Sprintf("while creating formation: %s", testErr.Error()))
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
	})

	t.Run("got error during getting updated label definition", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockFormationsService := &automock.FormationService{}
		defer mockFormationsService.AssertExpectations(t)

		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    model.ScenariosKey,
			Tenant: tnt,
			Schema: &scenarioSchemaInt,
		}, nil)

		defer mockService.AssertExpectations(t)
		mockService.On("GetWithoutCreating",
			contextThatHasTenant(tnt),
			tnt, model.ScenariosKey).Return(nil, errors.New("Object not found")).Once()

		mockFormationsService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "DEFAULT"}).Return(nil, nil)

		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, model.ScenariosKey).Return(nil, testErr).Once()

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockFormationsService, mockConverter)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		require.Error(t, err)
		require.EqualError(t, err, fmt.Sprintf("while getting label definition: %s", testErr.Error()))
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
	})

	t.Run("got error when missing tenant in context", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewResolver(nil, nil, nil, nil)
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
		sut := labeldef.NewResolver(transact, nil, nil, mockConverter)
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
		mockService.On("GetWithoutCreating",
			contextThatHasTenant(tnt),
			tnt, model.ScenariosKey).Return(&storedModel, nil).Once()

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    model.ScenariosKey,
			Tenant: tnt,
			Schema: &scenarioSchemaInt,
		}, nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, nil, mockConverter)
		// WHEN
		_, err := sut.CreateLabelDefinition(ctx, labelDefInput)
		// THEN
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
		require.EqualError(t, err, "Object is not unique [object=labelDefinition]")
	})

	t.Run("got error on converting to model", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()
		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{}, errors.New("json schema is not valid"))

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, nil, nil, mockConverter)
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
		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockFormationsService := &automock.FormationService{}
		defer mockFormationsService.AssertExpectations(t)

		mockConverter.On("FromGraphQL", labelDefInput, tnt).Return(model.LabelDefinition{
			Key:    model.ScenariosKey,
			Tenant: tnt,
			Schema: &scenarioSchemaInt,
		}, nil)

		defer mockService.AssertExpectations(t)
		mockService.On("GetWithoutCreating",
			contextThatHasTenant(tnt),
			tnt, model.ScenariosKey).Return(nil, errors.New("Object not found")).Once()

		mockFormationsService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "DEFAULT"}).Return(nil, nil)

		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, model.ScenariosKey).Return(&storedModel, nil).Once()

		mockConverter.On("ToGraphQL", storedModel).Return(graphql.LabelDefinition{
			Key:    model.ScenariosKey,
			Schema: getJSONSchemaFromSchema(t, &scenarioSchemaInt),
		}, nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockFormationsService, mockConverter)
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

		sut := labeldef.NewResolver(transact, mockService, nil, mockConverter)
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

		sut := labeldef.NewResolver(transact, mockService, nil, mockConverter)
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

		sut := labeldef.NewResolver(transact, mockService, nil, nil)
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
		sut := labeldef.NewResolver(nil, nil, nil, nil)
		// WHEN
		_, err := sut.LabelDefinitions(context.TODO())
		// THEN
		require.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("got error on starting transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()
		ctx := tenant.SaveToContext(context.TODO(), "tenant", "external-tenant")
		sut := labeldef.NewResolver(transact, nil, nil, nil)
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

		sut := labeldef.NewResolver(transact, mockService, nil, nil)
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

		sut := labeldef.NewResolver(transact, mockService, nil, nil)
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
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(false)
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

		sut := labeldef.NewResolver(mockTransactioner, mockService, nil, mockConverter)
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

		sut := labeldef.NewResolver(transact, mockService, nil, mockConverter)
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
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(true)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, "key").Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))

		sut := labeldef.NewResolver(mockTransactioner, mockService, nil, nil)
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
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(true)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("error from service"))

		sut := labeldef.NewResolver(mockTransactioner, mockService, nil, nil)
		// WHEN
		_, err := sut.LabelDefinition(ctx, "key")
		// THEN
		require.EqualError(t, err, "while getting Label Definition: error from service")
	})

	t.Run("got error when missing tenant in context", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewResolver(nil, nil, nil, nil)
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
		sut := labeldef.NewResolver(mockTransactioner, nil, nil, nil)
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
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(true)
		defer mockTransactioner.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get",
			contextThatHasTenant(tnt),
			tnt, "key").Return(nil, nil)

		sut := labeldef.NewResolver(mockTransactioner, mockService, nil, nil)
		// WHEN
		_, err := sut.LabelDefinition(ctx, "key")
		// THEN
		require.EqualError(t, err, "while committing transaction: commit errror")
	})
}

func TestUpdateLabelDefinition(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tenant"
	desiredSchema := getScenarioSchemaWithFormations([]string{"DEFAULT", "additional-1"})
	var desiredSchemaInt interface{} = desiredSchema
	initialSchema := getScenarioSchemaWithFormations([]string{"DEFAULT", "initial-1"})
	var initialSchemaInt interface{} = initialSchema

	gqlLabelDefinitionInput := graphql.LabelDefinitionInput{
		Key:    model.ScenariosKey,
		Schema: getJSONSchemaFromSchema(t, &desiredSchemaInt),
	}
	gqlLabelDefinition := model.LabelDefinition{
		ID:     fixUUID(),
		Tenant: tnt,
		Key:    model.ScenariosKey,
		Schema: &desiredSchemaInt,
	}
	modelLabelDefinition := model.LabelDefinition{
		ID:     fixUUID(),
		Tenant: tnt,
		Key:    model.ScenariosKey,
		Schema: &initialSchemaInt,
	}
	updatedGQLLabelDefinition := graphql.LabelDefinition{
		Key:    model.ScenariosKey,
		Schema: getJSONSchemaFromSchema(t, &desiredSchemaInt),
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
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(false)
		defer mockTransactioner.AssertExpectations(t)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(gqlLabelDefinition, nil)
		mockConverter.On("ToGraphQL", gqlLabelDefinition).Return(updatedGQLLabelDefinition, nil)

		mockFormationsService := &automock.FormationService{}
		defer mockFormationsService.AssertExpectations(t)
		mockFormationsService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "additional-1"}).Return(nil, nil)
		mockFormationsService.On("DeleteFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "initial-1"}).Return(nil, nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&modelLabelDefinition, nil).Once()
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&gqlLabelDefinition, nil).Once()

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(mockTransactioner, mockService, mockFormationsService, mockConverter)
		// WHEN
		actual, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, model.ScenariosKey, actual.Key)
	})

	t.Run("got error when convert to graphql failed", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(gqlLabelDefinition, nil)
		mockConverter.On("ToGraphQL", gqlLabelDefinition).Return(graphql.LabelDefinition{}, testErr)

		mockFormationsService := &automock.FormationService{}
		defer mockFormationsService.AssertExpectations(t)
		mockFormationsService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "additional-1"}).Return(nil, nil)
		mockFormationsService.On("DeleteFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "initial-1"}).Return(nil, nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&modelLabelDefinition, nil).Once()
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&gqlLabelDefinition, nil).Once()

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockFormationsService, mockConverter)
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
		sut := labeldef.NewResolver(nil, nil, nil, nil)
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
		sut := labeldef.NewResolver(mockTransactioner, nil, nil, mockConverter)
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
		sut := labeldef.NewResolver(transact, nil, nil, mockConverter)
		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, testErr.Error())
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("got error on creating formation", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(false)
		defer mockTransactioner.AssertExpectations(t)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(gqlLabelDefinition, nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&modelLabelDefinition, nil).Once()

		mockFormationsService := &automock.FormationService{}
		defer mockFormationsService.AssertExpectations(t)
		mockFormationsService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "additional-1"}).Return(nil, testErr)

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(mockTransactioner, mockService, mockFormationsService, mockConverter)
		defer mockService.AssertExpectations(t)

		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.EqualError(t, err, "while creating formation: test error")
	})
	t.Run("got error on deleting formation", func(t *testing.T) {
		// GIVEN
		mockPersistanceCtx := &pautomock.PersistenceTx{}
		defer mockPersistanceCtx.AssertExpectations(t)

		mockTransactioner := &pautomock.Transactioner{}
		mockTransactioner.On("Begin").Return(mockPersistanceCtx, nil)
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(false)
		defer mockTransactioner.AssertExpectations(t)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(gqlLabelDefinition, nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&modelLabelDefinition, nil).Once()

		mockFormationsService := &automock.FormationService{}
		defer mockFormationsService.AssertExpectations(t)
		mockFormationsService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "additional-1"}).Return(nil, nil)
		mockFormationsService.On("DeleteFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "initial-1"}).Return(nil, testErr)

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(mockTransactioner, mockService, mockFormationsService, mockConverter)
		defer mockService.AssertExpectations(t)

		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.EqualError(t, err, "while deleting formation: test error")
	})

	t.Run("got error on getting Label Definition", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(modelLabelDefinition, nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(nil, testErr)

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, nil, mockConverter)
		// WHEN
		_, err := sut.UpdateLabelDefinition(ctx, gqlLabelDefinitionInput)
		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "while receiving stored label definition: test error")
		transact.AssertExpectations(t)
		persist.AssertExpectations(t)
	})

	t.Run("get error on getting updated label definition", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(gqlLabelDefinition, nil)

		mockFormationsService := &automock.FormationService{}
		defer mockFormationsService.AssertExpectations(t)
		mockFormationsService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "additional-1"}).Return(nil, nil)
		mockFormationsService.On("DeleteFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "initial-1"}).Return(nil, nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&modelLabelDefinition, nil).Once()
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(nil, testErr).Once()

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(transact, mockService, mockFormationsService, mockConverter)
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
		mockTransactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return(false)
		defer mockTransactioner.AssertExpectations(t)

		mockConverter := &automock.ModelConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", gqlLabelDefinitionInput, tnt).Return(gqlLabelDefinition, nil)
		mockConverter.On("ToGraphQL", gqlLabelDefinition).Return(updatedGQLLabelDefinition, nil)

		mockFormationsService := &automock.FormationService{}
		defer mockFormationsService.AssertExpectations(t)
		mockFormationsService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "additional-1"}).Return(nil, nil)
		mockFormationsService.On("DeleteFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: "initial-1"}).Return(nil, nil)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&modelLabelDefinition, nil).Once()
		mockService.On("Get", contextThatHasTenant(tnt), tnt, modelLabelDefinition.Key).Return(&gqlLabelDefinition, nil).Once()

		ctx := persistence.SaveToContext(context.TODO(), nil)
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
		sut := labeldef.NewResolver(mockTransactioner, mockService, mockFormationsService, mockConverter)

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

func getJSONSchemaFromSchema(t *testing.T, schema *interface{}) *graphql.JSONSchema {
	data, err := json.Marshal(schema)
	require.NoError(t, err)
	jsonSchema := graphql.JSONSchema(data)
	return &jsonSchema
}

func getScenarioSchemaWithFormations(formations []string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
		"items": map[string]interface{}{
			"type":      "string",
			"pattern":   "^[A-Za-z0-9]([-_A-Za-z0-9\\s]*[A-Za-z0-9])$",
			"enum":      formations,
			"maxLength": 128,
		},
	}
}
