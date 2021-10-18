package formation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateFormation(t *testing.T) {
	formationInput := graphql.FormationInput{
		Name: testFormation,
	}
	tnt := "tenant"
	externalTnt := "external-tenant"
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully created formation", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockConverter := &automock.Converter{}
		mockService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormation}).Return(&model.Formation{Name: testFormation}, nil)

		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})
		mockConverter.On("ToGraphQL", &model.Formation{Name: testFormation}).Return(&graphql.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		actual, err := sut.CreateFormation(ctx, formationInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, testFormation, actual.Name)
	})
	t.Run("returns error when can not load tenant from context", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		sut := formation.NewResolver(nil, nil, nil)

		// WHEN
		_, err := sut.CreateFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewCannotReadTenantError().Error())
	})
	t.Run("returns error when can not start db transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, nil, nil)

		// WHEN
		_, err := sut.CreateFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("returns error when commit fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnCommit()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormation}).Return(&model.Formation{Name: testFormation}, nil)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		_, err := sut.CreateFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("returns error when create formation fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormation}).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		actual, err := sut.CreateFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, actual)
	})
}

func TestDeleteFormation(t *testing.T) {
	testFormation := testFormation
	formationInput := graphql.FormationInput{
		Name: testFormation,
	}
	tnt := "tenant"
	externalTnt := "external-tenant"
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully delete formation", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("DeleteFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormation}).Return(&model.Formation{Name: testFormation}, nil)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})
		mockConverter.On("ToGraphQL", &model.Formation{Name: testFormation}).Return(&graphql.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		actual, err := sut.DeleteFormation(ctx, formationInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, testFormation, actual.Name)
	})
	t.Run("returns error when can not load tenant from context", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		sut := formation.NewResolver(nil, nil, nil)

		// WHEN
		_, err := sut.DeleteFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
	})
	t.Run("returns error when can not start db transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, nil, nil)

		// WHEN
		_, err := sut.DeleteFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("returns error when commit fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnCommit()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("DeleteFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormation}).Return(&model.Formation{Name: testFormation}, nil)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		_, err := sut.DeleteFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("returns error when create formation fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("DeleteFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormation}).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		actual, err := sut.DeleteFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, actual)
	})
}

func TestAssignFormation(t *testing.T) {
	formationInput := graphql.FormationInput{
		Name: testFormation,
	}
	tnt := "tenant"
	externalTnt := "external-tenant"
	testObjectType := graphql.FormationObjectType("Application")
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully assigned formation", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockConverter := &automock.Converter{}
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormation}).Return(&model.Formation{Name: testFormation}, nil)

		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})
		mockConverter.On("ToGraphQL", &model.Formation{Name: testFormation}).Return(&graphql.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		actual, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, testFormation, actual.Name)
	})
	t.Run("returns error when can not load tenant from context", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		sut := formation.NewResolver(nil, nil, nil)

		// WHEN
		_, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewCannotReadTenantError().Error())
	})
	t.Run("returns error when can not start db transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, nil, nil)

		// WHEN
		_, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("returns error when commit fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnCommit()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormation}).Return(&model.Formation{Name: testFormation}, nil)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		_, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("returns error when assign formation fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormation}).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		actual, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, actual)
	})
}

func TestUnassignFormation(t *testing.T) {
	formationInput := graphql.FormationInput{
		Name: testFormation,
	}
	tnt := "tenant"
	externalTnt := "external-tenant"
	testObjectType := graphql.FormationObjectType("Application")
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully unassigned formation", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockConverter := &automock.Converter{}
		mockService.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormation}).Return(&model.Formation{Name: testFormation}, nil)

		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})
		mockConverter.On("ToGraphQL", &model.Formation{Name: testFormation}).Return(&graphql.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		actual, err := sut.UnassignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, testFormation, actual.Name)
	})
	t.Run("returns error when can not load tenant from context", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		sut := formation.NewResolver(nil, nil, nil)

		// WHEN
		_, err := sut.UnassignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewCannotReadTenantError().Error())
	})
	t.Run("returns error when can not start db transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, nil, nil)

		// WHEN
		_, err := sut.UnassignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("returns error when commit fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnCommit()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormation}).Return(&model.Formation{Name: testFormation}, nil)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		_, err := sut.UnassignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("returns error when assign formation fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()
		defer persist.AssertExpectations(t)
		defer transact.AssertExpectations(t)

		mockService := &automock.Service{}
		defer mockService.AssertExpectations(t)
		mockService.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormation}).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter)

		// WHEN
		actual, err := sut.UnassignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, actual)
	})
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
