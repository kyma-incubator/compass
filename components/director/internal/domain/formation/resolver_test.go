package formation_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"

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
		Name: testFormationName,
	}

	testTemplateName := "test-template-name"
	formationInputWithTemplateName := graphql.FormationInput{
		Name:         testFormationName,
		TemplateName: &testTemplateName,
	}

	tnt := "tenant"
	externalTnt := "external-tenant"
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	t.Run("successfully created formation with provided template name", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()

		mockService := &automock.Service{}
		mockConverter := &automock.Converter{}
		mockService.On("CreateFormation", contextThatHasTenant(tnt), tnt, modelFormation, testTemplateName).Return(&modelFormation, nil)

		mockConverter.On("FromGraphQL", formationInputWithTemplateName).Return(modelFormation)
		mockConverter.On("ToGraphQL", &modelFormation).Return(&graphqlFormation, nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil, nil, nil)

		// WHEN
		actual, err := sut.CreateFormation(ctx, formationInputWithTemplateName)

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
		mockService.On("CreateFormation", contextThatHasTenant(tnt), tnt, modelFormation, model.DefaultTemplateName).Return(&modelFormation, nil)

		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)
		mockConverter.On("ToGraphQL", &modelFormation).Return(&graphqlFormation, nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil, nil, nil)

		// WHEN
		actual, err := sut.CreateFormation(ctx, formationInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, testFormationName, actual.Name)
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
	})

	t.Run("returns error when can not load tenant from context", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		sut := formation.NewResolver(nil, nil, nil, nil, nil, nil)

		// WHEN
		_, err := sut.CreateFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewCannotReadTenantError().Error())
	})

	t.Run("returns error when can not start db transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, nil, nil, nil, nil, nil)

		// WHEN
		_, err := sut.CreateFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, persist, transact)
	})

	t.Run("returns error when commit fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnCommit()

		mockService := &automock.Service{}
		mockService.On("CreateFormation", contextThatHasTenant(tnt), tnt, modelFormation, model.DefaultTemplateName).Return(&modelFormation, nil)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil, nil, nil)

		// WHEN
		_, err := sut.CreateFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, persist, transact, mockService, mockConverter)
	})

	t.Run("returns error when create formation fails", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatDoesntExpectCommit()

		mockService := &automock.Service{}
		mockService.On("CreateFormation", contextThatHasTenant(tnt), tnt, modelFormation, model.DefaultTemplateName).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil, nil, nil)

		// WHEN
		actual, err := sut.CreateFormation(ctx, formationInput)

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
		mockConverter.On("ToGraphQL", &model.Formation{Name: testFormation}).Return(&graphql.Formation{Name: testFormation}, nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil, nil, nil)

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

		sut := formation.NewResolver(nil, nil, nil, nil, nil, nil)

		// WHEN
		_, err := sut.DeleteFormation(ctx, formationInput)

		// THEN
		require.Error(t, err)
	})
	t.Run("returns error when can not start db transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, nil, nil, nil, nil, nil)

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
		sut := formation.NewResolver(transact, mockService, mockConverter, nil, nil, nil)

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
		sut := formation.NewResolver(transact, mockService, mockConverter, nil, nil, nil)

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
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, modelFormation).Return(&modelFormation, nil)

		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)
		mockConverter.On("ToGraphQL", &modelFormation).Return(&graphqlFormation, nil)

		fetcherSvc.On("FetchOnDemand", contextThatHasTenant(tnt), "", tnt).Return(nil)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil, nil, fetcherSvc)

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

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		fetcherSvc.On("FetchOnDemand", ctx, "", tnt).Return(testErr)

		sut := formation.NewResolver(transact, mockService, mockConverter, nil, nil, fetcherSvc)

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

		sut := formation.NewResolver(nil, nil, nil, nil, nil, nil)

		// WHEN
		_, err := sut.AssignFormation(ctx, "", testObjectType, formationInput)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewCannotReadTenantError().Error())
	})
	t.Run("returns error when can not start db transaction", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatFailsOnBegin()
		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", ctx, "", tnt).Return(nil)

		sut := formation.NewResolver(transact, nil, nil, nil, nil, fetcherSvc)

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
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, modelFormation).Return(&modelFormation, nil)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", ctx, "", tnt).Return(nil)

		sut := formation.NewResolver(transact, mockService, mockConverter, nil, nil, fetcherSvc)

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
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, modelFormation).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)

		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", ctx, "", tnt).Return(nil)

		sut := formation.NewResolver(transact, mockService, mockConverter, nil, nil, fetcherSvc)

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
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	testCases := []struct {
		Name                     string
		TxFn                     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn                func() *automock.Service
		ConverterFn              func() *automock.Converter
		FormationAssignmentSvcFn func() *automock.FormationAssignmentService
		InputID                  string
		ObjectType               graphql.FormationObjectType
		Context                  context.Context
		ExpectedFormation        *graphql.Formation
		ExpectedError            error
	}{
		{
			Name: "successfully unassigned formation",
			TxFn: txGen.ThatSucceeds,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("FromGraphQL", formationInput).Return(modelFormation)
				conv.On("ToGraphQL", &modelFormation).Return(&graphqlFormation, nil)
				return conv
			},
			Context:    tenant.SaveToContext(context.TODO(), tnt, externalTnt),
			ObjectType: graphql.FormationObjectTypeTenant,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", graphql.FormationObjectTypeTenant, modelFormation).Return(&modelFormation, nil)
				return svc
			},
			ExpectedFormation: &graphqlFormation,
		},
		{
			Name:          "fails when transaction fails to open",
			TxFn:          txGen.ThatFailsOnBegin,
			ConverterFn:   unusedConverter,
			Context:       tenant.SaveToContext(context.TODO(), tnt, externalTnt),
			ObjectType:    graphql.FormationObjectTypeApplication,
			ServiceFn:     unusedService,
			InputID:       ApplicationID,
			ExpectedError: testErr,
		},
		{
			Name:          "returns error when can not load tenant from context",
			TxFn:          txGen.ThatDoesntExpectCommit,
			Context:       context.TODO(),
			ObjectType:    graphql.FormationObjectTypeTenant,
			ConverterFn:   unusedConverter,
			ServiceFn:     unusedService,
			ExpectedError: apperrors.NewCannotReadTenantError(),
		}, {
			Name:          "returns error when can not start db transaction",
			TxFn:          txGen.ThatFailsOnBegin,
			Context:       tenant.SaveToContext(context.TODO(), tnt, externalTnt),
			ObjectType:    graphql.FormationObjectTypeTenant,
			ConverterFn:   unusedConverter,
			ServiceFn:     unusedService,
			ExpectedError: testErr,
		}, {
			Name:       "returns error when commit fails",
			TxFn:       txGen.ThatFailsOnCommit,
			Context:    tenant.SaveToContext(context.TODO(), tnt, externalTnt),
			ObjectType: graphql.FormationObjectTypeTenant,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("FromGraphQL", formationInput).Return(modelFormation)
				return conv
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", graphql.FormationObjectTypeTenant, modelFormation).Return(&modelFormation, nil)
				return svc
			},
			ExpectedError: testErr,
		}, {
			Name:       "returns error when assign formation fails",
			TxFn:       txGen.ThatDoesntExpectCommit,
			Context:    tenant.SaveToContext(context.TODO(), tnt, externalTnt),
			ObjectType: graphql.FormationObjectTypeTenant,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("FromGraphQL", formationInput).Return(modelFormation)
				return conv
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", graphql.FormationObjectTypeTenant, modelFormation).Return(nil, testErr)
				return svc
			},
			ExpectedError: testErr,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TxFn()
			service := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			formationAssignmentSvc := &automock.FormationAssignmentService{}
			if testCase.FormationAssignmentSvcFn != nil {
				formationAssignmentSvc = testCase.FormationAssignmentSvcFn()
			}

			resolver := formation.NewResolver(transact, service, converter, formationAssignmentSvc, nil, nil)

			// WHEN
			f, err := resolver.UnassignFormation(testCase.Context, testCase.InputID, testCase.ObjectType, formationInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedFormation, f)

			mock.AssertExpectationsForObjects(t, persist, service, converter, formationAssignmentSvc)
		})
	}
}

func TestUnassignFormationGlobal(t *testing.T) {
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	testCases := []struct {
		Name                     string
		TxFn                     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn                func() *automock.Service
		ConverterFn              func() *automock.Converter
		FormationAssignmentSvcFn func() *automock.FormationAssignmentService
		ExpectedFormation        *graphql.Formation
		ExpectedError            error
	}{
		{
			Name: "successfully unassigned formation",
			TxFn: txGen.ThatSucceeds,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("ToGraphQL", &modelFormationWithTenant).Return(&graphqlFormation, nil)
				return conv
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("GetGlobalByID", mock.Anything, FormationID).Return(&modelFormationWithTenant, nil)
				svc.On("UnassignFormation", mock.Anything, TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, modelFormationWithTenant).Return(&modelFormationWithTenant, nil)
				return svc
			},
			ExpectedFormation: &graphqlFormation,
		},
		{
			Name:          "fails when transaction fails to open",
			TxFn:          txGen.ThatFailsOnBegin,
			ConverterFn:   unusedConverter,
			ServiceFn:     unusedService,
			ExpectedError: testErr,
		},
		{
			Name:          "returns error when can not start db transaction",
			TxFn:          txGen.ThatFailsOnBegin,
			ConverterFn:   unusedConverter,
			ServiceFn:     unusedService,
			ExpectedError: testErr,
		}, {
			Name:        "returns error when commit fails",
			TxFn:        txGen.ThatFailsOnCommit,
			ConverterFn: unusedConverter,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("GetGlobalByID", mock.Anything, FormationID).Return(&modelFormationWithTenant, nil)
				svc.On("UnassignFormation", mock.Anything, TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, modelFormationWithTenant).Return(&modelFormationWithTenant, nil)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:        "returns error when unassign formation fails",
			TxFn:        txGen.ThatDoesntExpectCommit,
			ConverterFn: unusedConverter,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("GetGlobalByID", mock.Anything, FormationID).Return(&modelFormationWithTenant, nil)
				svc.On("UnassignFormation", mock.Anything, TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, modelFormationWithTenant).Return(nil, testErr)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:        "returns error when get formation fails",
			TxFn:        txGen.ThatDoesntExpectCommit,
			ConverterFn: unusedConverter,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("GetGlobalByID", mock.Anything, FormationID).Return(nil, testErr)
				return svc
			},
			ExpectedError: testErr,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TxFn()
			service := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			formationAssignmentSvc := &automock.FormationAssignmentService{}
			if testCase.FormationAssignmentSvcFn != nil {
				formationAssignmentSvc = testCase.FormationAssignmentSvcFn()
			}

			resolver := formation.NewResolver(transact, service, converter, formationAssignmentSvc, nil, nil)

			// WHEN
			f, err := resolver.UnassignFormationGlobal(emptyCtx, ApplicationID, graphql.FormationObjectTypeApplication, FormationID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedFormation, f)

			mock.AssertExpectationsForObjects(t, persist, service, converter, formationAssignmentSvc)
		})
	}
}

func TestFormation(t *testing.T) {
	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	ctx := context.TODO()

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.Service
		ConverterFn       func() *automock.Converter
		FetcherFn         func() *automock.TenantFetcher
		InputID           string
		ExpectedFormation *graphql.Formation
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("Get", txtest.CtxWithDBMatcher(), FormationID).Return(&modelFormation, nil).Once()
				return service
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("ToGraphQL", &modelFormation).Return(&graphqlFormation, nil).Once()
				return conv
			},
			InputID:           FormationID,
			ExpectedFormation: &graphqlFormation,
			ExpectedError:     nil,
		},
		{
			Name: "Returns error when getting formation fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("Get", txtest.CtxWithDBMatcher(), FormationID).Return(nil, testErr).Once()
				return service
			},
			ConverterFn:       unusedConverter,
			InputID:           FormationID,
			ExpectedFormation: nil,
			ExpectedError:     testErr,
		},
		{
			Name:              "Returns error when can't start transaction",
			TxFn:              txGen.ThatFailsOnBegin,
			ServiceFn:         unusedService,
			ConverterFn:       unusedConverter,
			InputID:           FormationID,
			ExpectedFormation: nil,
			ExpectedError:     testErr,
		},
		{
			Name: "Returns error when can't commit transaction",
			TxFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("Get", txtest.CtxWithDBMatcher(), FormationID).Return(formationModel, nil).Once()
				return service
			}, ConverterFn: unusedConverter,
			InputID:           FormationID,
			ExpectedFormation: nil,
			ExpectedError:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TxFn()
			service := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := formation.NewResolver(transact, service, converter, nil, nil, nil)

			// WHEN
			f, err := resolver.Formation(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedFormation, f)

			mock.AssertExpectationsForObjects(t, persist, service, converter)
		})
	}
}

func TestFormationByName(t *testing.T) {
	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	tnt := "tenant"
	externalTnt := "external-tenant"
	ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
	emptyCtx := context.Background()

	testCases := []struct {
		Name              string
		Ctx               context.Context
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.Service
		ConverterFn       func() *automock.Converter
		FetcherFn         func() *automock.TenantFetcher
		InputName         string
		ExpectedFormation *graphql.Formation
		ExpectedError     error
	}{
		{
			Name: "Success",
			Ctx:  ctx,
			TxFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("GetFormationByName", contextThatHasTenant(tnt), testFormationName, tnt).Return(&modelFormation, nil).Once()
				return service
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("ToGraphQL", &modelFormation).Return(&graphqlFormation, nil).Once()
				return conv
			},
			InputName:         testFormationName,
			ExpectedFormation: &graphqlFormation,
			ExpectedError:     nil,
		},
		{
			Name:              "Returns error when getting tenant fails",
			Ctx:               emptyCtx,
			TxFn:              txGen.ThatDoesntExpectCommit,
			ServiceFn:         unusedService,
			ConverterFn:       unusedConverter,
			InputName:         testFormationName,
			ExpectedFormation: nil,
			ExpectedError:     errors.New("cannot read tenant from context"),
		},
		{
			Name: "Returns error when getting formation fails",
			Ctx:  ctx,
			TxFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("GetFormationByName", txtest.CtxWithDBMatcher(), testFormationName, tnt).Return(nil, testErr).Once()
				return service
			},
			ConverterFn:       unusedConverter,
			InputName:         testFormationName,
			ExpectedFormation: nil,
			ExpectedError:     testErr,
		},
		{
			Name:              "Returns error when can't start transaction",
			Ctx:               ctx,
			TxFn:              txGen.ThatFailsOnBegin,
			ServiceFn:         unusedService,
			ConverterFn:       unusedConverter,
			InputName:         testFormationName,
			ExpectedFormation: nil,
			ExpectedError:     testErr,
		},
		{
			Name: "Returns error when can't commit transaction",
			Ctx:  ctx,
			TxFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("GetFormationByName", txtest.CtxWithDBMatcher(), testFormationName, tnt).Return(formationModel, nil).Once()
				return service
			}, ConverterFn: unusedConverter,
			InputName:         testFormationName,
			ExpectedFormation: nil,
			ExpectedError:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			testCtx := testCase.Ctx
			persist, transact := testCase.TxFn()
			service := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := formation.NewResolver(transact, service, converter, nil, nil, nil)

			// WHEN
			f, err := resolver.FormationByName(testCtx, testCase.InputName)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedFormation, f)

			mock.AssertExpectationsForObjects(t, persist, service, converter)
		})
	}
}
func TestFormations(t *testing.T) {
	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	ctx := context.TODO()

	first := 100
	afterStr := "after"
	after := graphql.PageCursor(afterStr)

	modelFormations := []*model.Formation{&modelFormation}

	graphqlFormations := []*graphql.Formation{&graphqlFormation}

	modelPage := &model.FormationPage{
		Data: modelFormations,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: 1,
	}

	expectedOutput := &graphql.FormationPage{
		Data: graphqlFormations,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: 1,
	}

	testCases := []struct {
		Name               string
		TxFn               func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn          func() *automock.Service
		ConverterFn        func() *automock.Converter
		FetcherFn          func() *automock.TenantFetcher
		InputID            string
		ExpectedFormations *graphql.FormationPage
		ExpectedError      error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("List", txtest.CtxWithDBMatcher(), first, afterStr).Return(modelPage, nil).Once()
				return service
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("MultipleToGraphQL", modelFormations).Return(graphqlFormations, nil).Once()
				return conv
			},
			InputID:            FormationID,
			ExpectedFormations: expectedOutput,
			ExpectedError:      nil,
		},
		{
			Name: "Returns error when listing formations fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("List", txtest.CtxWithDBMatcher(), first, afterStr).Return(nil, testErr).Once()
				return service
			},
			ConverterFn:        unusedConverter,
			InputID:            FormationID,
			ExpectedFormations: nil,
			ExpectedError:      testErr,
		},
		{
			Name: "Returns error when converting formations to graphql fails",
			TxFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("List", txtest.CtxWithDBMatcher(), first, afterStr).Return(modelPage, nil).Once()
				return service
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("MultipleToGraphQL", modelFormations).Return(nil, testErr).Once()
				return conv
			},
			InputID:            FormationID,
			ExpectedFormations: nil,
			ExpectedError:      testErr,
		},
		{
			Name:               "Returns error when can't start transaction",
			TxFn:               txGen.ThatFailsOnBegin,
			ServiceFn:          unusedService,
			ConverterFn:        unusedConverter,
			InputID:            FormationID,
			ExpectedFormations: nil,
			ExpectedError:      testErr,
		},
		{
			Name: "Returns error when can't commit transaction",
			TxFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("List", txtest.CtxWithDBMatcher(), first, afterStr).Return(modelPage, nil).Once()
				return service
			},
			ConverterFn:        unusedConverter,
			InputID:            FormationID,
			ExpectedFormations: nil,
			ExpectedError:      testErr,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TxFn()
			service := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := formation.NewResolver(transact, service, converter, nil, nil, nil)

			// WHEN
			f, err := resolver.Formations(ctx, &first, &after)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedFormations, f)

			mock.AssertExpectationsForObjects(t, persist, service, converter)
		})
	}

	t.Run("Returns error when 'first' is nil", func(t *testing.T) {
		resolver := formation.NewResolver(nil, nil, nil, nil, nil, nil)

		// WHEN
		f, err := resolver.Formations(ctx, nil, &after)

		// THEN
		require.Error(t, err)
		require.Nil(t, f)
	})
}

func TestFormationsForObject(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	ctx := context.TODO()

	modelFormations := []*model.Formation{&modelFormation}

	graphqlFormations := []*graphql.Formation{&graphqlFormation}

	testCases := []struct {
		Name               string
		TxFn               func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn          func() *automock.Service
		ConverterFn        func() *automock.Converter
		InputID            string
		ExpectedFormations []*graphql.Formation
		ExpectedError      error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("ListFormationsForObject", txtest.CtxWithDBMatcher(), ApplicationID).Return(modelFormations, nil).Once()
				return service
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("MultipleToGraphQL", modelFormations).Return(graphqlFormations, nil).Once()
				return conv
			},
			InputID:            ApplicationID,
			ExpectedFormations: graphqlFormations,
			ExpectedError:      nil,
		},
		{
			Name: "Returns error when listing formations fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("ListFormationsForObject", txtest.CtxWithDBMatcher(), ApplicationID).Return(nil, testErr).Once()
				return service
			},
			ConverterFn:        unusedConverter,
			InputID:            ApplicationID,
			ExpectedFormations: nil,
			ExpectedError:      testErr,
		},
		{
			Name: "Returns error when converting formations to graphql fails",
			TxFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("ListFormationsForObject", txtest.CtxWithDBMatcher(), ApplicationID).Return(modelFormations, nil).Once()
				return service
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("MultipleToGraphQL", modelFormations).Return(nil, testErr).Once()
				return conv
			},
			InputID:            ApplicationID,
			ExpectedFormations: nil,
			ExpectedError:      testErr,
		},
		{
			Name:               "Returns error when can't start transaction",
			TxFn:               txGen.ThatFailsOnBegin,
			ServiceFn:          unusedService,
			ConverterFn:        unusedConverter,
			InputID:            ApplicationID,
			ExpectedFormations: nil,
			ExpectedError:      testErr,
		},
		{
			Name: "Returns error when can't commit transaction",
			TxFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				service := &automock.Service{}
				service.On("ListFormationsForObject", txtest.CtxWithDBMatcher(), ApplicationID).Return(modelFormations, nil).Once()
				return service
			},
			ConverterFn:        unusedConverter,
			InputID:            ApplicationID,
			ExpectedFormations: nil,
			ExpectedError:      testErr,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TxFn()
			service := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := formation.NewResolver(transact, service, converter, nil, nil, nil)

			// WHEN
			f, err := resolver.FormationsForObject(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedFormations, f)

			mock.AssertExpectationsForObjects(t, persist, service, converter)
		})
	}
}

func TestResolver_FormationAssignment(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")
	notFoundErr := apperrors.NewNotFoundError(resource.FormationAssignment, FormationAssignmentID)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	gqlFormation := fixGqlFormation()
	gqlFormationAssignment := fixGqlFormationAssignment(FormationAssignmentState, &TestConfigValueStr)
	formationAssignmentModel := fixFormationAssignmentModel(FormationAssignmentState, TestConfigValueRawJSON)

	testCases := []struct {
		Name                        string
		TransactionerFn             func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn                   func() *automock.FormationAssignmentService
		ConverterFn                 func() *automock.FormationAssignmentConverter
		Formation                   *graphql.Formation
		FormationAssignmentID       string
		ExpectedFormationAssignment *graphql.FormationAssignment
		ExpectedErrMsg              string
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetForFormation", txtest.CtxWithDBMatcher(), FormationAssignmentID, FormationID).Return(formationAssignmentModel, nil).Once()
				return faSvc
			},
			ConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToGraphQL", formationAssignmentModel).Return(gqlFormationAssignment, nil).Once()
				return faConv
			},
			Formation:                   gqlFormation,
			FormationAssignmentID:       FormationAssignmentID,
			ExpectedFormationAssignment: gqlFormationAssignment,
		},
		{
			Name:                        "Return error when formation object is nil",
			TransactionerFn:             txGen.ThatDoesntStartTransaction,
			ExpectedFormationAssignment: nil,
			ExpectedErrMsg:              "Formation cannot be empty",
		},
		{
			Name:                        "Returns error when commit begin fails",
			TransactionerFn:             txGen.ThatFailsOnBegin,
			Formation:                   gqlFormation,
			ExpectedFormationAssignment: nil,
			ExpectedErrMsg:              testErr.Error(),
		},
		{
			Name:            "Returns error when formation assignment retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetForFormation", txtest.CtxWithDBMatcher(), FormationAssignmentID, FormationID).Return(nil, testErr).Once()
				return faSvc
			},
			Formation:                   gqlFormation,
			FormationAssignmentID:       FormationAssignmentID,
			ExpectedFormationAssignment: nil,
			ExpectedErrMsg:              testErr.Error(),
		},
		{
			Name:            "Returns error when formation assignment for formation is not found",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetForFormation", txtest.CtxWithDBMatcher(), FormationAssignmentID, FormationID).Return(nil, notFoundErr).Once()
				return faSvc
			},
			Formation:                   gqlFormation,
			FormationAssignmentID:       FormationAssignmentID,
			ExpectedFormationAssignment: nil,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetForFormation", txtest.CtxWithDBMatcher(), FormationAssignmentID, FormationID).Return(formationAssignmentModel, nil).Once()
				return faSvc
			},
			Formation:                   gqlFormation,
			FormationAssignmentID:       FormationAssignmentID,
			ExpectedFormationAssignment: nil,
			ExpectedErrMsg:              testErr.Error(),
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetForFormation", txtest.CtxWithDBMatcher(), FormationAssignmentID, FormationID).Return(formationAssignmentModel, nil).Once()
				return faSvc
			},
			ConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToGraphQL", formationAssignmentModel).Return(nil, testErr).Once()
				return faConv
			},
			Formation:                   gqlFormation,
			FormationAssignmentID:       FormationAssignmentID,
			ExpectedFormationAssignment: nil,
			ExpectedErrMsg:              testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()

			faSvc := &automock.FormationAssignmentService{}
			if testCase.ServiceFn != nil {
				faSvc = testCase.ServiceFn()
			}

			faConv := &automock.FormationAssignmentConverter{}
			if testCase.ConverterFn != nil {
				faConv = testCase.ConverterFn()
			}

			resolver := formation.NewResolver(transact, nil, nil, faSvc, faConv, nil)

			// WHEN
			fa, err := resolver.FormationAssignment(ctx, testCase.Formation, testCase.FormationAssignmentID)

			// THEN
			require.Equal(t, testCase.ExpectedFormationAssignment, fa)
			if testCase.ExpectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, faSvc, faConv, transact, persist)
		})
	}
}

func TestResolver_FormationAssignments(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	formationIDs := []string{FormationID, FormationID + "2"}

	// Formation Assignments model fixtures
	faModelFirst := fixFormationAssignmentModel(FormationAssignmentState, TestConfigValueRawJSON)
	faModelSecond := fixFormationAssignmentModelWithSuffix(FormationAssignmentState, TestConfigValueRawJSON, nil, "-2")

	fasFirst := []*model.FormationAssignment{faModelFirst}
	fasSecond := []*model.FormationAssignment{faModelSecond}

	faPageFirst := fixFormationAssignmentPage(fasFirst)
	faPageSecond := fixFormationAssignmentPage(fasSecond)
	faPages := []*model.FormationAssignmentPage{faPageFirst, faPageSecond}

	// Formation Assignments GraphQL fixtures
	gqlFormationAssignmentFirst := fixGqlFormationAssignment(FormationAssignmentState, &TestConfigValueStr)
	gqlFormationAssignmentSecond := fixGqlFormationAssignmentWithSuffix(FormationAssignmentState, &TestConfigValueStr, "-2")

	gqlFAFist := []*graphql.FormationAssignment{gqlFormationAssignmentFirst}
	gqlFASecond := []*graphql.FormationAssignment{gqlFormationAssignmentSecond}

	gqlFAPageFirst := fixGQLFormationAssignmentPage(gqlFAFist)
	gqlFAPageSecond := fixGQLFormationAssignmentPage(gqlFASecond)
	gqlFAPages := []*graphql.FormationAssignmentPage{gqlFAPageFirst, gqlFAPageSecond}

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.FormationAssignmentService
		ConverterFn     func() *automock.FormationAssignmentConverter
		ExpectedResult  []*graphql.FormationAssignmentPage
		ExpectedErr     []error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("ListByFormationIDs", txtest.CtxWithDBMatcher(), formationIDs, first, after).Return(faPages, nil).Once()
				return faSvc
			},
			ConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("MultipleToGraphQL", fasFirst).Return(gqlFAFist, nil).Once()
				faConv.On("MultipleToGraphQL", fasSecond).Return(gqlFASecond, nil).Once()
				return faConv
			},
			ExpectedResult: gqlFAPages,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ExpectedResult:  nil,
			ExpectedErr:     []error{testErr},
		},
		{
			Name:            "Returns error when listing formations",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("ListByFormationIDs", txtest.CtxWithDBMatcher(), formationIDs, first, after).Return(nil, testErr).Once()
				return faSvc
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("ListByFormationIDs", txtest.CtxWithDBMatcher(), formationIDs, first, after).Return(faPages, nil).Once()
				return faSvc
			},
			ConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("MultipleToGraphQL", fasFirst).Return(gqlFAFist, nil).Once()
				faConv.On("MultipleToGraphQL", fasSecond).Return(gqlFASecond, nil).Once()
				return faConv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("ListByFormationIDs", txtest.CtxWithDBMatcher(), formationIDs, first, after).Return(faPages, nil).Once()
				return faSvc
			},
			ConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("MultipleToGraphQL", fasFirst).Return(nil, testErr).Once()
				return faConv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()

			faSvc := &automock.FormationAssignmentService{}
			if testCase.ServiceFn != nil {
				faSvc = testCase.ServiceFn()
			}

			faConv := &automock.FormationAssignmentConverter{}
			if testCase.ConverterFn != nil {
				faConv = testCase.ConverterFn()
			}

			resolver := formation.NewResolver(transact, nil, nil, faSvc, faConv, nil)
			firstFormationAssignmentParams := dataloader.ParamFormationAssignment{ID: FormationID, Ctx: ctx, First: &first, After: &gqlAfter}
			secondFormationAssignmentParams := dataloader.ParamFormationAssignment{ID: FormationID + "2", Ctx: ctx, First: &first, After: &gqlAfter}
			keys := []dataloader.ParamFormationAssignment{firstFormationAssignmentParams, secondFormationAssignmentParams}

			// WHEN
			result, err := resolver.FormationAssignmentsDataLoader(keys)

			// THEN
			require.Equal(t, testCase.ExpectedResult, result)
			require.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, faSvc, faConv, transact, persist)
		})
	}

	t.Run("Returns error when there are no formations IDs", func(t *testing.T) {
		resolver := formation.NewResolver(nil, nil, nil, nil, nil, nil)

		// WHEN
		_, errs := resolver.FormationAssignmentsDataLoader([]dataloader.ParamFormationAssignment{})

		// THEN
		require.Error(t, errs[0])
		require.EqualError(t, errs[0], apperrors.NewInternalError("No Formations found").Error())
	})

	t.Run("Returns error when start cursor is nil", func(t *testing.T) {
		firstFormationAssignmentParams := dataloader.ParamFormationAssignment{ID: FormationID, Ctx: ctx, First: nil, After: &gqlAfter}
		keys := []dataloader.ParamFormationAssignment{firstFormationAssignmentParams}

		resolver := formation.NewResolver(nil, nil, nil, nil, nil, nil)

		// WHEN
		_, errs := resolver.FormationAssignmentsDataLoader(keys)

		// THEN
		require.Error(t, errs[0])
		require.EqualError(t, errs[0], apperrors.NewInvalidDataError("missing required parameter 'first'").Error())
	})
}

func TestResolver_Status(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	formationIDs := []string{FormationID, FormationID + "2", FormationID + "3", FormationID + "4"}

	// Formation Assignments model fixtures
	faModelReady := fixFormationAssignmentModel(string(model.ReadyAssignmentState), TestConfigValueRawJSON)
	faModelInitial := fixFormationAssignmentModelWithSuffix(string(model.InitialAssignmentState), TestConfigValueRawJSON, nil, "-2")
	faModelError := fixFormationAssignmentModelWithSuffix(string(model.CreateErrorAssignmentState), nil, json.RawMessage(`{"error": {"message": "failure", "errorCode": 1}}`), "-3")
	faModelEmptyError := fixFormationAssignmentModelWithSuffix(string(model.CreateErrorAssignmentState), nil, nil, "-4")

	fasReady := []*model.FormationAssignment{faModelReady}                                         // all are READY -> READY condition
	fasInProgress := []*model.FormationAssignment{faModelInitial, faModelReady}                    // no errors, but one is INITIAL -> IN_PROGRESS condition
	fasError := []*model.FormationAssignment{faModelReady, faModelError, faModelInitial}           // have error -> ERROR condition
	fasEmptyError := []*model.FormationAssignment{faModelReady, faModelEmptyError, faModelInitial} // should handle empty Value and have ERROR condition

	fasPerFormation := [][]*model.FormationAssignment{fasReady, fasInProgress, fasError, fasEmptyError}

	fasUnmarshallable := []*model.FormationAssignment{fixFormationAssignmentModelWithSuffix(string(model.DeleteErrorAssignmentState), nil, json.RawMessage(`unmarshallable structure`), "-4")}

	faPagesWithUnmarshallableError := [][]*model.FormationAssignment{fasUnmarshallable, fasReady, fasInProgress, fasError}

	// Formation Assignments GraphQL fixtures

	gqlStatusFirst := graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil}
	gqlStatusSecond := graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil}
	gqlStatusThird := graphql.FormationStatus{
		Condition: graphql.FormationStatusConditionError,
		Errors: []*graphql.FormationStatusError{{
			AssignmentID: addSuffix(FormationAssignmentID, "-3"),
			Message:      "failure",
			ErrorCode:    1,
		}},
	}
	gqlStatusFourth := graphql.FormationStatus{
		Condition: graphql.FormationStatusConditionError,
		Errors: []*graphql.FormationStatusError{{
			AssignmentID: addSuffix(FormationAssignmentID, "-4"),
		}},
	}

	gqlStatuses := []*graphql.FormationStatus{&gqlStatusFirst, &gqlStatusSecond, &gqlStatusThird, &gqlStatusFourth}

	emptyFaPage := [][]*model.FormationAssignment{nil, nil, nil, nil}

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.FormationAssignmentService
		Params          []dataloader.ParamFormationStatus
		ExpectedResult  []*graphql.FormationStatus
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("ListByFormationIDsNoPaging", txtest.CtxWithDBMatcher(), formationIDs).Return(fasPerFormation, nil).Once()
				return faSvc
			},
			Params:         []dataloader.ParamFormationStatus{firstFormationStatusParams, secondFormationStatusParams, thirdFormationStatusParams, fourthPageFormations},
			ExpectedResult: gqlStatuses,
			ExpectedErr:    nil,
		},
		{
			Name:            "Success when there are no formation assignments",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("ListByFormationIDsNoPaging", txtest.CtxWithDBMatcher(), []string{FormationID}).Return(emptyFaPage, nil).Once()
				return faSvc
			},
			Params:         []dataloader.ParamFormationStatus{firstFormationStatusParams},
			ExpectedResult: []*graphql.FormationStatus{&gqlStatusFirst},
			ExpectedErr:    nil,
		},
		{
			Name:            "Success when there are no FAs and the formation is in error state",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("ListByFormationIDsNoPaging", txtest.CtxWithDBMatcher(), []string{FormationID}).Return(emptyFaPage, nil).Once()
				return faSvc
			},
			Params: []dataloader.ParamFormationStatus{{ID: FormationID, State: string(model.CreateErrorFormationState), Message: "failure", ErrorCode: 1}},
			ExpectedResult: []*graphql.FormationStatus{{
				Condition: graphql.FormationStatusConditionError,
				Errors: []*graphql.FormationStatusError{{
					Message:   "failure",
					ErrorCode: 1,
				}},
			}},
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			Params:          []dataloader.ParamFormationStatus{firstFormationStatusParams, secondFormationStatusParams, thirdFormationStatusParams, fourthPageFormations},
			ExpectedResult:  nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when listing formations",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("ListByFormationIDsNoPaging", txtest.CtxWithDBMatcher(), formationIDs).Return(nil, testErr).Once()
				return faSvc
			},
			Params:         []dataloader.ParamFormationStatus{firstFormationStatusParams, secondFormationStatusParams, thirdFormationStatusParams, fourthPageFormations},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when can't unmarshal assignment value",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("ListByFormationIDsNoPaging", txtest.CtxWithDBMatcher(), formationIDs).Return(faPagesWithUnmarshallableError, nil).Once()
				return faSvc
			},
			Params:         []dataloader.ParamFormationStatus{firstFormationStatusParams, secondFormationStatusParams, thirdFormationStatusParams, fourthPageFormations},
			ExpectedResult: nil,
			ExpectedErr:    errors.New("while unmarshalling formation assignment error with assignment ID \"FormationAssignmentID-4\""),
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("ListByFormationIDsNoPaging", txtest.CtxWithDBMatcher(), formationIDs).Return(fasPerFormation, nil).Once()
				return faSvc
			},
			Params:         []dataloader.ParamFormationStatus{firstFormationStatusParams, secondFormationStatusParams, thirdFormationStatusParams, fourthPageFormations},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()

			faSvc := &automock.FormationAssignmentService{}
			if testCase.ServiceFn != nil {
				faSvc = testCase.ServiceFn()
			}

			resolver := formation.NewResolver(transact, nil, nil, faSvc, nil, nil)

			params := make([]dataloader.ParamFormationStatus, 0, len(testCase.Params))
			for _, param := range testCase.Params {
				param.Ctx = ctx
				params = append(params, param)
			}

			// WHEN
			result, errs := resolver.StatusDataLoader(params)

			// THEN
			require.EqualValues(t, testCase.ExpectedResult, result)
			if errs != nil {
				require.Contains(t, errs[0].Error(), testCase.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, faSvc, transact, persist)
		})
	}

	t.Run("Returns error when there are no formations IDs", func(t *testing.T) {
		resolver := formation.NewResolver(nil, nil, nil, nil, nil, nil)

		// WHEN
		_, errs := resolver.StatusDataLoader([]dataloader.ParamFormationStatus{})

		// THEN
		require.Error(t, errs[0])
		require.EqualError(t, errs[0], apperrors.NewInternalError("No Formations found").Error())
	})
}

func TestResynchronizeFormationNotifications(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	ctx := tenant.SaveToContext(context.TODO(), TntInternalID, TntExternalID)
	shouldReset := true
	shouldNotReset := false

	testCases := []struct {
		Name              string
		Context           context.Context
		ShouldReset       *bool
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationService  func() *automock.Service
		Converter         func() *automock.Converter
		ExpectedFormation *graphql.Formation
		ExpectedErrorMsg  string
	}{
		{
			Name: "successfully resynchronized formation notifications",
			TxFn: txGen.ThatSucceeds,
			FormationService: func() *automock.Service {
				svc := &automock.Service{}

				svc.On("ResynchronizeFormationNotifications", contextThatHasTenant(TntInternalID), FormationID, false).Return(&modelFormation, nil).Once()

				return svc
			},
			Converter: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("ToGraphQL", &modelFormation).Return(&graphqlFormation, nil).Once()
				return conv
			},
			ExpectedFormation: &graphqlFormation,
		},
		{
			Name:        "successfully resynchronized formation notifications with reset false",
			TxFn:        txGen.ThatSucceeds,
			ShouldReset: &shouldNotReset,
			FormationService: func() *automock.Service {
				svc := &automock.Service{}

				svc.On("ResynchronizeFormationNotifications", contextThatHasTenant(TntInternalID), FormationID, false).Return(&modelFormation, nil).Once()

				return svc
			},
			Converter: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("ToGraphQL", &modelFormation).Return(&graphqlFormation, nil).Once()
				return conv
			},
			ExpectedFormation: &graphqlFormation,
		},
		{
			Name:        "successfully resynchronized formation notifications with reset",
			TxFn:        txGen.ThatSucceeds,
			ShouldReset: &shouldReset,
			FormationService: func() *automock.Service {
				svc := &automock.Service{}

				svc.On("ResynchronizeFormationNotifications", contextThatHasTenant(TntInternalID), FormationID, true).Return(&modelFormation, nil).Once()

				return svc
			},
			Converter: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("ToGraphQL", &modelFormation).Return(&graphqlFormation, nil).Once()
				return conv
			},
			ExpectedFormation: &graphqlFormation,
		},
		{
			Name: "failed during resynchronizing",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationService: func() *automock.Service {
				svc := &automock.Service{}

				svc.On("ResynchronizeFormationNotifications", contextThatHasTenant(TntInternalID), FormationID, false).Return(nil, testErr)

				return svc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name: "failed to commit after resynchronizing",
			TxFn: txGen.ThatFailsOnCommit,
			FormationService: func() *automock.Service {
				svc := &automock.Service{}

				svc.On("ResynchronizeFormationNotifications", contextThatHasTenant(TntInternalID), FormationID, false).Return(&modelFormation, nil).Once()

				return svc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "returns error when can not start db transaction",
			TxFn:             txGen.ThatFailsOnBegin,
			ExpectedErrorMsg: testErr.Error(),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := unusedConverter()
			if testCase.Converter != nil {
				conv = testCase.Converter()
			}
			formationService := unusedService()
			if testCase.FormationService != nil {
				formationService = testCase.FormationService()
			}
			persist, transact := testCase.TxFn()

			resolver := formation.NewResolver(transact, formationService, conv, nil, nil, nil)

			// WHEN
			resultFormationModel, err := resolver.ResynchronizeFormationNotifications(ctx, FormationID, testCase.ShouldReset)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
				require.Nil(t, resultFormationModel)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedFormation, resultFormationModel)
			}
			mock.AssertExpectationsForObjects(t, conv, formationService, persist, transact)
		})
	}
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

func addSuffix(str, suffix string) *string {
	res := str + suffix
	return &res
}
