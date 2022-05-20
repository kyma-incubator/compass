package scenarioassignment_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResolverCreateAutomaticScenarioAssignment(t *testing.T) {
	testErr := errors.New("test err")
	givenInput := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: scenarioName,
		Selector: &graphql.LabelSelectorInput{
			Key:   scenarioassignment.SubaccountIDKey,
			Value: externalTargetTenantID,
		},
	}
	expectedOutput := graphql.AutomaticScenarioAssignment{
		ScenarioName: scenarioName,
		Selector: &graphql.Label{
			Key:   scenarioassignment.SubaccountIDKey,
			Value: externalTargetTenantID,
		},
	}

	ctx := tenant.SaveToContext(context.TODO(), tenantID, "")

	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))

	t.Run("happy path", func(t *testing.T) {
		tx, transact := txGen.ThatSucceeds()

		mockConverter := &automock.GqlConverter{}
		mockConverter.On("FromInputGraphQL", givenInput, targetTenantID).Return(fixModel()).Once()
		mockConverter.On("ToGraphQL", fixModel(), externalTargetTenantID).Return(expectedOutput).Once()

		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", givenInput.Selector.Value).Return(nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return(targetTenantID, nil).Once()

		formationSvc := &automock.FormationService{}
		formationSvc.On("CreateAutomaticScenarioAssignment", mock.Anything, fixModel()).Return(fixModel(), nil).Once()

		sut := scenarioassignment.NewResolver(transact, nil, mockConverter, tenantSvc, fetcherSvc, formationSvc)

		// WHEN
		actual, err := sut.CreateAutomaticScenarioAssignment(ctx, givenInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, &expectedOutput, actual)
		mock.AssertExpectationsForObjects(t, tx, transact, mockConverter, tenantSvc, fetcherSvc, formationSvc)
	})

	t.Run("error when tenant creation fail", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntStartTransaction()

		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", mock.Anything).Return(testErr).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, nil, nil, nil, fetcherSvc, formationSvc)

		// WHEN
		_, err := sut.CreateAutomaticScenarioAssignment(ctx, givenInput)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, tx, transact, fetcherSvc)
	})

	t.Run("error when tenant conversion fail", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()

		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", mock.Anything).Return(nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return("", testErr).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, nil, nil, tenantSvc, fetcherSvc, formationSvc)

		// WHEN
		_, err := sut.CreateAutomaticScenarioAssignment(ctx, givenInput)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, tx, transact, tenantSvc, fetcherSvc)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", "subaccountID").Return(nil).Once()

		tx, transact := txGen.ThatFailsOnBegin()
		defer mock.AssertExpectationsForObjects(t, tx, transact, fetcherSvc)
		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, nil, nil, nil, fetcherSvc, formationSvc)

		// WHEN
		_, err := sut.CreateAutomaticScenarioAssignment(context.TODO(),
			graphql.AutomaticScenarioAssignmentSetInput{Selector: &graphql.LabelSelectorInput{
				Key:   "key",
				Value: "subaccountID",
			}})

		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
	})

	t.Run("error on creating assignment by service", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()
		mockConverter := &automock.GqlConverter{}
		mockConverter.On("FromInputGraphQL", givenInput, targetTenantID).Return(fixModel()).Once()

		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", mock.Anything).Return(nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return(targetTenantID, nil).Once()

		formationSvc := &automock.FormationService{}
		formationSvc.On("CreateAutomaticScenarioAssignment", mock.Anything, fixModel()).Return(model.AutomaticScenarioAssignment{}, fixError()).Once()
		sut := scenarioassignment.NewResolver(transact, nil, mockConverter, tenantSvc, fetcherSvc, formationSvc)

		// WHEN
		_, err := sut.CreateAutomaticScenarioAssignment(ctx, givenInput)

		// THEN
		assert.EqualError(t, err, fmt.Sprintf("while creating Assignment: %s", errMsg))
		mock.AssertExpectationsForObjects(t, tx, transact, formationSvc, mockConverter, tenantSvc)
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnCommit()
		mockConverter := &automock.GqlConverter{}
		mockConverter.On("FromInputGraphQL", givenInput, targetTenantID).Return(fixModel()).Once()

		fetcherSvc := &automock.TenantFetcher{}
		fetcherSvc.On("FetchOnDemand", mock.Anything).Return(nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return(targetTenantID, nil).Once()

		formationSvc := &automock.FormationService{}
		formationSvc.On("CreateAutomaticScenarioAssignment", mock.Anything, fixModel()).Return(fixModel(), nil).Once()

		sut := scenarioassignment.NewResolver(transact, nil, mockConverter, tenantSvc, fetcherSvc, formationSvc)

		// WHEN
		_, err := sut.CreateAutomaticScenarioAssignment(ctx, givenInput)

		// THEN
		assert.EqualError(t, err, "while committing transaction: some persistence error")
		mock.AssertExpectationsForObjects(t, tx, transact, formationSvc, mockConverter, tenantSvc)
	})
}

func TestResolver_GetAutomaticScenarioAssignmentByScenario(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))
	expectedOutput := fixGQL()

	t.Run("happy path", func(t *testing.T) {
		tx, transact := txGen.ThatSucceeds()

		mockConverter := &automock.GqlConverter{}
		mockConverter.On("ToGraphQL", fixModel(), externalTargetTenantID).Return(expectedOutput).Once()

		mockSvc := &automock.AsaService{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(fixModel(), nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetExternalTenant", mock.Anything, targetTenantID).Return(externalTargetTenantID, nil).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter, tenantSvc, nil, formationSvc)

		// WHEN
		actual, err := sut.GetAutomaticScenarioAssignmentForScenarioName(context.TODO(), scenarioName)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, &expectedOutput, actual)
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter, tenantSvc)
	})

	t.Run("error when GetExternalTenant fail", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()

		mockSvc := &automock.AsaService{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(fixModel(), nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetExternalTenant", mock.Anything, targetTenantID).Return("", fixError()).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, tenantSvc, nil, formationSvc)

		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentForScenarioName(context.TODO(), scenarioName)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fixError().Error())
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, tenantSvc)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnBegin()
		defer mock.AssertExpectationsForObjects(t, tx, transact)
		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, nil, nil, nil, nil, formationSvc)

		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentForScenarioName(context.TODO(), scenarioName)

		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
	})

	t.Run("error on receiving assignment by service", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()
		mockSvc := &automock.AsaService{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(model.AutomaticScenarioAssignment{}, fixError()).Once()
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc)
		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, nil, nil, formationSvc)

		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentForScenarioName(context.TODO(), scenarioName)

		// THEN
		assert.EqualError(t, err, fmt.Sprintf("while getting Assignment: %s", errMsg))
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnCommit()
		mockSvc := &automock.AsaService{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(fixModel(), nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetExternalTenant", mock.Anything, targetTenantID).Return(externalTargetTenantID, nil).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, tenantSvc, nil, formationSvc)

		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentForScenarioName(context.TODO(), scenarioName)

		// THEN
		assert.EqualError(t, err, "while committing transaction: some persistence error")
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, tenantSvc)
	})
}

func TestResolver_AutomaticScenarioAssignmentsForSelector(t *testing.T) {
	givenInput := graphql.LabelSelectorInput{
		Key:   scenarioassignment.SubaccountIDKey,
		Value: externalTargetTenantID,
	}

	expectedModels := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   scenarioName,
			TargetTenantID: targetTenantID,
		},
		{
			ScenarioName:   "scenario-B",
			TargetTenantID: targetTenantID,
		},
	}

	expectedOutput := []*graphql.AutomaticScenarioAssignment{
		{
			ScenarioName: scenarioName,
			Selector: &graphql.Label{
				Key:   scenarioassignment.SubaccountIDKey,
				Value: externalTargetTenantID,
			},
		},
		{
			ScenarioName: "scenario-B",
			Selector: &graphql.Label{
				Key:   scenarioassignment.SubaccountIDKey,
				Value: externalTargetTenantID,
			},
		},
	}

	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))

	t.Run("happy path", func(t *testing.T) {
		tx, transact := txGen.ThatSucceeds()

		mockConverter := &automock.GqlConverter{}
		mockConverter.On("ToGraphQL", *expectedModels[0], externalTargetTenantID).Return(*expectedOutput[0]).Once()
		mockConverter.On("ToGraphQL", *expectedModels[1], externalTargetTenantID).Return(*expectedOutput[1]).Once()

		mockSvc := &automock.AsaService{}
		mockSvc.On("ListForTargetTenant", mock.Anything, targetTenantID).Return(expectedModels, nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return(targetTenantID, nil).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter, tenantSvc, nil, formationSvc)

		// WHEN
		actual, err := sut.AutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedOutput, actual)
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter, tenantSvc)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnBegin()
		defer mock.AssertExpectationsForObjects(t, tx, transact)
		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, nil, nil, nil, nil, formationSvc)

		// WHEN
		_, err := sut.AutomaticScenarioAssignmentsForSelector(context.TODO(), graphql.LabelSelectorInput{})

		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
	})

	t.Run("error on getting assignments by service", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return(targetTenantID, nil).Once()

		mockSvc := &automock.AsaService{}
		mockSvc.On("ListForTargetTenant", mock.Anything, targetTenantID).Return(nil, fixError()).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, tenantSvc, nil, formationSvc)

		// WHEN
		actual, err := sut.AutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.Nil(t, actual)
		require.EqualError(t, err, "while getting the assignments: some error")
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, tenantSvc)
	})

	t.Run("error on getting assignments by service", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return("", fixError()).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, nil, nil, tenantSvc, nil, formationSvc)

		// WHEN
		actual, err := sut.AutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.Nil(t, actual)
		require.EqualError(t, err, "while converting tenant: some error")
		mock.AssertExpectationsForObjects(t, tx, transact, tenantSvc)
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnCommit()

		mockSvc := &automock.AsaService{}
		mockSvc.On("ListForTargetTenant", mock.Anything, targetTenantID).Return(expectedModels, nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return(targetTenantID, nil).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, tenantSvc, nil, formationSvc)

		// WHEN
		actual, err := sut.AutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.EqualError(t, err, "while committing transaction: some persistence error")
		require.Nil(t, actual)
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, tenantSvc)
	})
}

func TestResolver_AutomaticScenarioAssignments(t *testing.T) {
	testErr := errors.New("test error")

	mod1 := fixModelWithScenarioName("foo")
	mod2 := fixModelWithScenarioName("bar")
	modItems := []*model.AutomaticScenarioAssignment{
		&mod1, &mod2,
	}
	modelPage := fixModelPageWithItems(modItems)

	gql1 := fixGQLWithScenarioName("foo")
	gql2 := fixGQLWithScenarioName("bar")
	gqlItems := []*graphql.AutomaticScenarioAssignment{
		&gql1, &gql2,
	}
	gqlPage := fixGQLPageWithItems(gqlItems)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.AsaService
		TenantSvcFn     func() *automock.TenantService
		ConverterFn     func() *automock.GqlConverter
		ExpectedResult  *graphql.AutomaticScenarioAssignmentPage
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.AsaService {
				svc := &automock.AsaService{}
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(&modelPage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.GqlConverter {
				conv := &automock.GqlConverter{}
				conv.On("ToGraphQL", mod1, externalTargetTenantID).Return(gql1).Once()
				conv.On("ToGraphQL", mod2, externalTargetTenantID).Return(gql2).Once()
				return conv
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetExternalTenant", mock.Anything, targetTenantID).Return(externalTargetTenantID, nil).Once()
				tenantSvc.On("GetExternalTenant", mock.Anything, targetTenantID).Return(externalTargetTenantID, nil).Once()
				return tenantSvc
			},
			ExpectedResult: &gqlPage,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.AsaService {
				svc := &automock.AsaService{}
				return svc
			},
			ConverterFn: func() *automock.GqlConverter {
				conv := &automock.GqlConverter{}
				return conv
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				return tenantSvc
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Assignments listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.AsaService {
				svc := &automock.AsaService{}
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.GqlConverter {
				conv := &automock.GqlConverter{}
				return conv
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				return tenantSvc
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when GetExternalTenant failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.AsaService {
				svc := &automock.AsaService{}
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(&modelPage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.GqlConverter {
				conv := &automock.GqlConverter{}
				return conv
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetExternalTenant", mock.Anything, targetTenantID).Return("", testErr).Once()
				return tenantSvc
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.AsaService {
				svc := &automock.AsaService{}
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(&modelPage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.GqlConverter {
				conv := &automock.GqlConverter{}
				conv.On("ToGraphQL", mod1, externalTargetTenantID).Return(gql1).Once()
				conv.On("ToGraphQL", mod2, externalTargetTenantID).Return(gql2).Once()
				return conv
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetExternalTenant", mock.Anything, targetTenantID).Return(externalTargetTenantID, nil).Once()
				tenantSvc.On("GetExternalTenant", mock.Anything, targetTenantID).Return(externalTargetTenantID, nil).Once()
				return tenantSvc
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			tenantSvc := testCase.TenantSvcFn()

			resolver := scenarioassignment.NewResolver(transact, svc, converter, tenantSvc, nil, nil)

			// WHEN
			result, err := resolver.AutomaticScenarioAssignments(context.TODO(), &first, &gqlAfter)

			// THEN
			assert.Equal(t, testCase.ExpectedResult, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testErr.Error())
			}

			mock.AssertExpectationsForObjects(t, persist, transact, svc, converter, tenantSvc)
		})
	}
}

func TestResolver_DeleteAutomaticScenarioAssignmentsForSelector(t *testing.T) {
	givenInput := graphql.LabelSelectorInput{
		Key:   scenarioassignment.SubaccountIDKey,
		Value: externalTargetTenantID,
	}

	scenarioNameA := "scenario-A"
	scenarioNameB := "scenario-B"
	expectedModels := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   scenarioNameA,
			TargetTenantID: targetTenantID,
		},
		{
			ScenarioName:   scenarioNameB,
			TargetTenantID: targetTenantID,
		},
	}

	expectedOutput := []*graphql.AutomaticScenarioAssignment{
		{
			ScenarioName: scenarioNameA,
			Selector: &graphql.Label{
				Key:   scenarioassignment.SubaccountIDKey,
				Value: externalTargetTenantID,
			},
		},
		{
			ScenarioName: scenarioNameB,
			Selector: &graphql.Label{
				Key:   scenarioassignment.SubaccountIDKey,
				Value: externalTargetTenantID,
			},
		},
	}

	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatSucceeds()

		mockConverter := &automock.GqlConverter{}
		mockConverter.On("ToGraphQL", *expectedModels[0], externalTargetTenantID).Return(*expectedOutput[0]).Once()
		mockConverter.On("ToGraphQL", *expectedModels[1], externalTargetTenantID).Return(*expectedOutput[1]).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return(targetTenantID, nil).Once()

		mockSvc := &automock.AsaService{}
		mockSvc.On("ListForTargetTenant", txtest.CtxWithDBMatcher(), targetTenantID).Return(expectedModels, nil).Once()

		formationSvc := &automock.FormationService{}
		formationSvc.On("DeleteManyForSameTargetTenant", txtest.CtxWithDBMatcher(), expectedModels).Return(nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter, tenantSvc, nil, formationSvc)

		// WHEN
		actual, err := sut.DeleteAutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedOutput, actual)
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter, tenantSvc, formationSvc)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnBegin()
		defer mock.AssertExpectationsForObjects(t, tx, transact)
		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, nil, nil, nil, nil, formationSvc)

		// WHEN
		_, err := sut.DeleteAutomaticScenarioAssignmentsForSelector(context.TODO(), graphql.LabelSelectorInput{})

		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
	})

	t.Run("error on converting tenant", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatDoesntExpectCommit()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return("", fixError()).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, nil, nil, tenantSvc, nil, formationSvc)

		// WHEN
		actual, err := sut.DeleteAutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.Nil(t, actual)
		require.EqualError(t, err, fmt.Sprintf("while converting tenant: %s", errMsg))
		mock.AssertExpectationsForObjects(t, tx, transact, tenantSvc)
	})

	t.Run("error on getting assignments by service", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatDoesntExpectCommit()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return(targetTenantID, nil).Once()

		mockSvc := &automock.AsaService{}
		mockSvc.On("ListForTargetTenant", txtest.CtxWithDBMatcher(), targetTenantID).Return(nil, fixError()).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, tenantSvc, nil, formationSvc)

		// WHEN
		actual, err := sut.DeleteAutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.Nil(t, actual)
		require.EqualError(t, err, fmt.Sprintf("while getting the Assignments for target tenant [targetTenantID]: %s", errMsg))
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, tenantSvc)
	})

	t.Run("error on deleting assignments by service", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatDoesntExpectCommit()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return(targetTenantID, nil).Once()

		mockSvc := &automock.AsaService{}
		mockSvc.On("ListForTargetTenant", txtest.CtxWithDBMatcher(), targetTenantID).Return(expectedModels, nil).Once()

		formationSvc := &automock.FormationService{}
		formationSvc.On("DeleteManyForSameTargetTenant", txtest.CtxWithDBMatcher(), expectedModels).Return(fixError()).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, tenantSvc, nil, formationSvc)

		// WHEN
		actual, err := sut.DeleteAutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.Nil(t, actual)
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignments for target tenant [targetTenantID]: %s", errMsg))
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, tenantSvc, formationSvc)
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatFailsOnCommit()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetInternalTenant", mock.Anything, externalTargetTenantID).Return(targetTenantID, nil).Once()

		mockSvc := &automock.AsaService{}
		mockSvc.On("ListForTargetTenant", txtest.CtxWithDBMatcher(), targetTenantID).Return(expectedModels, nil).Once()

		formationSvc := &automock.FormationService{}
		formationSvc.On("DeleteManyForSameTargetTenant", txtest.CtxWithDBMatcher(), expectedModels).Return(nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, tenantSvc, nil, formationSvc)

		// WHEN
		actual, err := sut.DeleteAutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.EqualError(t, err, "while committing transaction: some persistence error")
		require.Nil(t, actual)
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, tenantSvc, formationSvc)
	})
}

func TestResolver_DeleteAutomaticScenarioAssignmentForScenario(t *testing.T) {
	expectedModel := fixModel()
	expectedOutput := fixGQL()

	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatSucceeds()

		mockConverter := &automock.GqlConverter{}
		mockConverter.On("ToGraphQL", expectedModel, externalTargetTenantID).Return(expectedOutput).Once()

		mockSvc := &automock.AsaService{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(expectedModel, nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetExternalTenant", mock.Anything, targetTenantID).Return(externalTargetTenantID, nil).Once()

		formationSvc := &automock.FormationService{}
		formationSvc.On("DeleteAutomaticScenarioAssignment", txtest.CtxWithDBMatcher(), expectedModel).Return(nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter, tenantSvc, nil, formationSvc)

		// WHEN
		actual, err := sut.DeleteAutomaticScenarioAssignmentForScenario(fixCtxWithTenant(), scenarioName)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, &expectedOutput, actual)
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter, tenantSvc)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatFailsOnBegin()
		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, nil, nil, nil, nil, formationSvc)

		// WHEN
		_, err := sut.DeleteAutomaticScenarioAssignmentForScenario(context.TODO(), scenarioName)

		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
		mock.AssertExpectationsForObjects(t, tx, transact)
	})

	t.Run("error on getting assignment by service", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatDoesntExpectCommit()

		mockSvc := &automock.AsaService{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(model.AutomaticScenarioAssignment{}, fixError()).Once()

		formationSvc := &automock.FormationService{}

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, nil, nil, formationSvc)

		// WHEN
		_, err := sut.DeleteAutomaticScenarioAssignmentForScenario(fixCtxWithTenant(), scenarioName)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while getting the Assignment for scenario [name=%s]: %s", scenarioName, errMsg))
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc)
	})

	t.Run("error on deleting assignments by service", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatDoesntExpectCommit()

		mockSvc := &automock.AsaService{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(expectedModel, nil).Once()

		formationSvc := &automock.FormationService{}
		formationSvc.On("DeleteAutomaticScenarioAssignment", txtest.CtxWithDBMatcher(), expectedModel).Return(fixError()).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, nil, nil, formationSvc)

		// WHEN
		_, err := sut.DeleteAutomaticScenarioAssignmentForScenario(fixCtxWithTenant(), scenarioName)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignment for scenario [name=%s]: %s", scenarioName, errMsg))
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, formationSvc)
	})

	t.Run("error when converting tenant", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatDoesntExpectCommit()

		mockSvc := &automock.AsaService{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(expectedModel, nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetExternalTenant", mock.Anything, targetTenantID).Return("", fixError()).Once()

		formationSvc := &automock.FormationService{}
		formationSvc.On("DeleteAutomaticScenarioAssignment", txtest.CtxWithDBMatcher(), expectedModel).Return(nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, tenantSvc, nil, formationSvc)

		// WHEN
		_, err := sut.DeleteAutomaticScenarioAssignmentForScenario(fixCtxWithTenant(), scenarioName)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while converting tenant: %s", errMsg))
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, tenantSvc, formationSvc)
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatFailsOnCommit()

		mockSvc := &automock.AsaService{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(expectedModel, nil).Once()

		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetExternalTenant", mock.Anything, targetTenantID).Return(externalTargetTenantID, nil).Once()

		formationSvc := &automock.FormationService{}
		formationSvc.On("DeleteAutomaticScenarioAssignment", txtest.CtxWithDBMatcher(), expectedModel).Return(nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil, tenantSvc, nil, formationSvc)

		// WHEN
		_, err := sut.DeleteAutomaticScenarioAssignmentForScenario(fixCtxWithTenant(), scenarioName)

		// THEN
		require.EqualError(t, err, "while committing transaction: some persistence error")
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, tenantSvc)
	})
}
