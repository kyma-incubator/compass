package formation_test

import (
	"context"
	"errors"
	"testing"

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
		mockConverter.On("ToGraphQL", &modelFormation).Return(&graphqlFormation)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

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
		mockConverter.On("ToGraphQL", &modelFormation).Return(&graphqlFormation)

		ctx := tenant.SaveToContext(context.TODO(), tnt, externalTnt)
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

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

		sut := formation.NewResolver(nil, nil, nil, nil)

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
		sut := formation.NewResolver(transact, nil, nil, nil)

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
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

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
		sut := formation.NewResolver(transact, mockService, mockConverter, nil)

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
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, modelFormation).Return(&modelFormation, nil)

		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)
		mockConverter.On("ToGraphQL", &modelFormation).Return(&graphqlFormation)

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
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, modelFormation).Return(&modelFormation, nil)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)

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
		mockService.On("AssignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, modelFormation).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)

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
		mockService.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, modelFormation).Return(&modelFormation, nil)

		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)
		mockConverter.On("ToGraphQL", &modelFormation).Return(&graphqlFormation)

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
		mockService.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, modelFormation).Return(&modelFormation, nil)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)

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
		mockService.On("UnassignFormation", contextThatHasTenant(tnt), tnt, "", testObjectType, modelFormation).Return(nil, testErr)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromGraphQL", formationInput).Return(modelFormation)

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
				conv.On("ToGraphQL", &modelFormation).Return(&graphqlFormation).Once()
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

			resolver := formation.NewResolver(transact, service, converter, nil)

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
				conv.On("MultipleToGraphQL", modelFormations).Return(graphqlFormations).Once()
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

			resolver := formation.NewResolver(transact, service, converter, nil)

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
		resolver := formation.NewResolver(nil, nil, nil, nil)

		// WHEN
		f, err := resolver.Formations(ctx, nil, &after)

		// THEN
		require.Error(t, err)
		require.Nil(t, f)
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
