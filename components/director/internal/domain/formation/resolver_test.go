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

var defaultTemplateName = "Side-by-side extensibility with Kyma"

func TestCreateFormation(t *testing.T) {
	formationInput := graphql.FormationInput{
		Name: testFormationName,
	}
	tnt := "tenant"
	externalTnt := "external-tenant"
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully created formation", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()

		mockService := &automock.Service{}
		mockConverter := &automock.Converter{}
		mockService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormationName}, &defaultTemplateName).Return(&model.Formation{Name: testFormationName}, nil)

		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormationName})
		mockConverter.On("ToGraphQL", &model.Formation{Name: testFormationName}).Return(&graphql.Formation{Name: testFormationName})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

		// WHEN
		actual, err := sut.CreateFormation(ctx, formationInput, &defaultTemplateName)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, testFormationName, actual.Name)
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
	})

	t.Run("successfully created formation when no template is provided", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()

		mockService := &automock.Service{}
		mockConverter := &automock.Converter{}
		mockService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormationName}, &defaultTemplateName).Return(&model.Formation{Name: testFormationName}, nil)

		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormationName})
		mockConverter.On("ToGraphQL", &model.Formation{Name: testFormationName}).Return(&graphql.Formation{Name: testFormationName})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

		// WHEN
		actual, err := sut.CreateFormation(ctx, formationInput, nil)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, testFormationName, actual.Name)
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
	})

	t.Run("returns error when can not load tenant from context", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		sut := formation.NewResolver(nil, nil, nil, nil)

		// WHEN
		_, err := sut.CreateFormation(ctx, formationInput, &defaultTemplateName)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewCannotReadTenantError().Error())
	})

	t.Run("returns error when can not start db transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, nil, nil, nil)

		// WHEN
		_, err := sut.CreateFormation(ctx, formationInput, &defaultTemplateName)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, persist, transact)
	})

	t.Run("returns error when commit fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnCommit()

		mockService := &automock.Service{}
		mockService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormationName}, &defaultTemplateName).Return(&model.Formation{Name: testFormationName}, nil)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormationName})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

		// WHEN
		_, err := sut.CreateFormation(ctx, formationInput, &defaultTemplateName)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
	})

	t.Run("returns error when create formation fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockService := &automock.Service{}
		mockService.On("CreateFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormationName}, &defaultTemplateName).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormationName})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

		// WHEN
		actual, err := sut.CreateFormation(ctx, formationInput, &defaultTemplateName)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, actual)
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
	})
}

func TestDeleteFormation(t *testing.T) {
	testFormation := testFormationName
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

		mockService := &automock.Service{}
		mockService.On("DeleteFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormation}).Return(&model.Formation{Name: testFormation}, nil)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})
		mockConverter.On("ToGraphQL", &model.Formation{Name: testFormation}).Return(&graphql.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

		// WHEN
		actual, err := sut.DeleteFormation(ctx, formationInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, testFormation, actual.Name)
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
	})
	t.Run("returns error when can not load tenant from context", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		sut := formation.NewResolver(nil, nil, nil, nil)

		// WHEN
		_, err := sut.DeleteFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
	})
	t.Run("returns error when can not start db transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, nil, nil, nil)

		// WHEN
		_, err := sut.DeleteFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, persist, transact)
	})
	t.Run("returns error when commit fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnCommit()

		mockService := &automock.Service{}
		mockService.On("DeleteFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormation}).Return(&model.Formation{Name: testFormation}, nil)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

		// WHEN
		_, err := sut.DeleteFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
	})
	t.Run("returns error when create formation fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockService := &automock.Service{}
		mockService.On("DeleteFormation", contextThatHasTenant(tnt), tnt, model.Formation{Name: testFormation}).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormation})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

		// WHEN
		actual, err := sut.DeleteFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, actual)
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
	})
}

func TestAssignFormation(t *testing.T) {
	formationInput := graphql.FormationInput{
		Name: testFormationName,
	}
	tnt := "tenant"
	externalTnt := "external-tenant"
	testObjectType := graphql.FormationObjectTypeTenant
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully assigned formation", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()

		mockService := &automock.Service{}
		mockConverter := &automock.Converter{}
		fetcherSvc := &automock.TenantFetcher{}
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormationName}).Return(&model.Formation{Name: testFormationName}, nil)

		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormationName})
		mockConverter.On("ToGraphQL", &model.Formation{Name: testFormationName}).Return(&graphql.Formation{Name: testFormationName})

		fetcherSvc.On("FetchOnDemand", "", tnt).Return(nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, fetcherSvc)

		// WHEN
		actual, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, testFormationName, actual.Name)
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter, fetcherSvc)
	})
	t.Run("returns error when objectType is tenant and cannot fetch its details", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntStartTransaction()

		mockService := &automock.Service{}
		mockConverter := &automock.Converter{}
		fetcherSvc := &automock.TenantFetcher{}

		fetcherSvc.On("FetchOnDemand", "", tnt).Return(testErr)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, fetcherSvc)

		// WHEN
		_, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter, fetcherSvc)
	})
	t.Run("returns error when can not load tenant from context", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		sut := formation.NewResolver(nil, nil, nil, nil)

		// WHEN
		_, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewCannotReadTenantError().Error())
	})
	t.Run("returns error when can not start db transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()

		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", "", tnt).Return(nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, nil, nil, fetcherSvc)

		// WHEN
		_, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, persist, transact, fetcherSvc)
	})
	t.Run("returns error when commit fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnCommit()

		mockService := &automock.Service{}
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormationName}).Return(&model.Formation{Name: testFormationName}, nil)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormationName})

		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", "", tnt).Return(nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, fetcherSvc)

		// WHEN
		_, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter, fetcherSvc)
	})
	t.Run("returns error when assign formation fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockService := &automock.Service{}
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormationName}).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormationName})

		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", "", tnt).Return(nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, fetcherSvc)

		// WHEN
		actual, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, actual)
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter, fetcherSvc)
	})
}

func TestUnassignFormation(t *testing.T) {
	formationInput := graphql.FormationInput{
		Name: testFormationName,
	}
	tnt := "tenant"
	externalTnt := "external-tenant"
	testObjectType := graphql.FormationObjectType("Application")
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully unassigned formation", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()

		mockService := &automock.Service{}
		mockConverter := &automock.Converter{}
		mockService.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormationName}).Return(&model.Formation{Name: testFormationName}, nil)

		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormationName})
		mockConverter.On("ToGraphQL", &model.Formation{Name: testFormationName}).Return(&graphql.Formation{Name: testFormationName})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

		// WHEN
		actual, err := sut.UnassignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, testFormationName, actual.Name)
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
	})
	t.Run("returns error when can not load tenant from context", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		sut := formation.NewResolver(nil, nil, nil, nil)

		// WHEN
		_, err := sut.UnassignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewCannotReadTenantError().Error())
	})
	t.Run("returns error when can not start db transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, nil, nil, nil)

		// WHEN
		_, err := sut.UnassignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, persist, transact)
	})
	t.Run("returns error when commit fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnCommit()

		mockService := &automock.Service{}
		mockService.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormationName}).Return(&model.Formation{Name: testFormationName}, nil)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormationName})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

		// WHEN
		_, err := sut.UnassignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
	})
	t.Run("returns error when assign formation fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockService := &automock.Service{}
		mockService.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, model.Formation{Name: testFormationName}).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(model.Formation{Name: testFormationName})

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

		// WHEN
		actual, err := sut.UnassignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, actual)
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
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
