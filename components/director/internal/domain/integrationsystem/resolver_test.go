package integrationsystem_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_IntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	modelIntSys := fixModelIntegrationSystem(testID, testName)
	gqlIntSys := fixGQLIntegrationSystem(testID, testName)

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		IntSysSvcFn    func() *automock.IntegrationSystemService
		IntSysConvFn   func() *automock.IntegrationSystemConverter
		ExpectedOutput *graphql.IntegrationSystem
		ExpectedError  error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelIntSys, nil).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("ToGraphQL", modelIntSys).Return(gqlIntSys).Once()
				return intSysConv
			},
			ExpectedOutput: gqlIntSys,
		},
		{
			Name: "Returns nil when integration system not found",
			TxFn: txGen.ThatDoesntExpectCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, apperrors.NewNotFoundError("")).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			ExpectedOutput: nil,
		},
		{
			Name: "Returns error when getting integration system failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelIntSys, nil).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			intSysSvc := testCase.IntSysSvcFn()
			intSysConv := testCase.IntSysConvFn()

			resolver := integrationsystem.NewResolver(transact, intSysSvc, nil, nil, intSysConv, nil)

			// WHEN
			result, err := resolver.IntegrationSystem(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			intSysSvc.AssertExpectations(t)
			intSysConv.AssertExpectations(t)
		})
	}
}

func TestResolver_IntegrationSystems(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)
	txGen := txtest.NewTransactionContextGenerator(testError)
	modelIntSys := []*model.IntegrationSystem{
		fixModelIntegrationSystem("i1", "n1"),
		fixModelIntegrationSystem("i2", "n2"),
	}
	modelPage := fixModelIntegrationSystemPage(modelIntSys)
	gqlIntSys := []*graphql.IntegrationSystem{
		fixGQLIntegrationSystem("i1", "n1"),
		fixGQLIntegrationSystem("i2", "n2"),
	}
	gqlPage := fixGQLIntegrationSystemPage(gqlIntSys)
	first := 2
	after := "test"
	gqlAfter := graphql.PageCursor(after)

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		IntSysSvcFn    func() *automock.IntegrationSystemService
		IntSysConvFn   func() *automock.IntegrationSystemConverter
		ExpectedOutput *graphql.IntegrationSystemPage
		ExpectedError  error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(modelPage, nil).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("MultipleToGraphQL", modelIntSys).Return(gqlIntSys).Once()
				return intSysConv
			},
			ExpectedOutput: &gqlPage,
		},
		{
			Name: "Returns error when getting integration system failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(model.IntegrationSystemPage{}, testError).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(modelPage, nil).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			intSysSvc := testCase.IntSysSvcFn()
			intSysConv := testCase.IntSysConvFn()

			resolver := integrationsystem.NewResolver(transact, intSysSvc, nil, nil, intSysConv, nil)

			// WHEN
			result, err := resolver.IntegrationSystems(ctx, &first, &gqlAfter)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			intSysSvc.AssertExpectations(t)
			intSysConv.AssertExpectations(t)
		})
	}
}

func TestResolver_CreateIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	modelIntSys := fixModelIntegrationSystem(testID, testName)
	modelIntSysInput := fixModelIntegrationSystemInput(testName)
	gqlIntSys := fixGQLIntegrationSystem(testID, testName)
	gqlIntSysInput := fixGQLIntegrationSystemInput(testName)

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		IntSysSvcFn    func() *automock.IntegrationSystemService
		IntSysConvFn   func() *automock.IntegrationSystemConverter
		ExpectedOutput *graphql.IntegrationSystem
		ExpectedError  error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Create", txtest.CtxWithDBMatcher(), modelIntSysInput).Return(modelIntSys.ID, nil).Once()
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelIntSys, nil).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("InputFromGraphQL", gqlIntSysInput).Return(modelIntSysInput).Once()
				intSysConv.On("ToGraphQL", modelIntSys).Return(gqlIntSys).Once()
				return intSysConv
			},
			ExpectedOutput: gqlIntSys,
		},
		{
			Name: "Returns error when creating integration system failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Create", txtest.CtxWithDBMatcher(), modelIntSysInput).Return("", testError).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("InputFromGraphQL", gqlIntSysInput).Return(modelIntSysInput).Once()
				return intSysConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when getting integration system failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Create", txtest.CtxWithDBMatcher(), modelIntSysInput).Return(modelIntSys.ID, nil).Once()
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("InputFromGraphQL", gqlIntSysInput).Return(modelIntSysInput).Once()
				return intSysConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("InputFromGraphQL", gqlIntSysInput).Return(modelIntSysInput).Once()
				return intSysConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Create", txtest.CtxWithDBMatcher(), modelIntSysInput).Return(modelIntSys.ID, nil).Once()
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelIntSys, nil).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("InputFromGraphQL", gqlIntSysInput).Return(modelIntSysInput).Once()
				return intSysConv
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			intSysSvc := testCase.IntSysSvcFn()
			intSysConv := testCase.IntSysConvFn()

			resolver := integrationsystem.NewResolver(transact, intSysSvc, nil, nil, intSysConv, nil)

			// WHEN
			result, err := resolver.CreateIntegrationSystem(ctx, gqlIntSysInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			intSysSvc.AssertExpectations(t)
			intSysConv.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	modelIntSys := fixModelIntegrationSystem(testID, testName)
	modelIntSysInput := fixModelIntegrationSystemInput(testName)
	gqlIntSys := fixGQLIntegrationSystem(testID, testName)
	gqlIntSysInput := fixGQLIntegrationSystemInput(testName)

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		IntSysSvcFn    func() *automock.IntegrationSystemService
		IntSysConvFn   func() *automock.IntegrationSystemConverter
		ExpectedOutput *graphql.IntegrationSystem
		ExpectedError  error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Update", txtest.CtxWithDBMatcher(), testID, modelIntSysInput).Return(nil).Once()
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelIntSys, nil).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("InputFromGraphQL", gqlIntSysInput).Return(modelIntSysInput).Once()
				intSysConv.On("ToGraphQL", modelIntSys).Return(gqlIntSys).Once()
				return intSysConv
			},
			ExpectedOutput: gqlIntSys,
		},
		{
			Name: "Returns error when updating integration system failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Update", txtest.CtxWithDBMatcher(), testID, modelIntSysInput).Return(testError).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("InputFromGraphQL", gqlIntSysInput).Return(modelIntSysInput).Once()
				return intSysConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when getting integration system failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Update", txtest.CtxWithDBMatcher(), testID, modelIntSysInput).Return(nil).Once()
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("InputFromGraphQL", gqlIntSysInput).Return(modelIntSysInput).Once()
				return intSysConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Update", txtest.CtxWithDBMatcher(), testID, modelIntSysInput).Return(nil).Once()
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelIntSys, nil).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("InputFromGraphQL", gqlIntSysInput).Return(modelIntSysInput).Once()
				return intSysConv
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			intSysSvc := testCase.IntSysSvcFn()
			intSysConv := testCase.IntSysConvFn()

			resolver := integrationsystem.NewResolver(transact, intSysSvc, nil, nil, intSysConv, nil)

			// WHEN
			result, err := resolver.UpdateIntegrationSystem(ctx, testID, gqlIntSysInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			intSysSvc.AssertExpectations(t)
			intSysConv.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	modelIntSys := fixModelIntegrationSystem(testID, testName)
	gqlIntSys := fixGQLIntegrationSystem(testID, testName)
	testAuth := fixOauth()

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		IntSysSvcFn    func() *automock.IntegrationSystemService
		IntSysConvFn   func() *automock.IntegrationSystemConverter
		SysAuthSvcFn   func() *automock.SystemAuthService
		OAuth20SvcFn   func() *automock.Oauth20Service
		ExpectedOutput *graphql.IntegrationSystem
		ExpectedError  error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelIntSys, nil).Once()
				intSysSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				intSysConv.On("ToGraphQL", modelIntSys).Return(gqlIntSys).Once()
				return intSysConv
			},
			SysAuthSvcFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), model.IntegrationSystemReference, modelIntSys.ID).Return(nil, nil)
				return svc
			},
			OAuth20SvcFn: func() *automock.Oauth20Service {
				svc := &automock.Oauth20Service{}
				return svc
			},
			ExpectedOutput: gqlIntSys,
		},
		{
			Name: "Returns error when getting integration system failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			SysAuthSvcFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20SvcFn: func() *automock.Oauth20Service {
				svc := &automock.Oauth20Service{}
				return svc
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when deleting integration system failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelIntSys, nil).Once()
				intSysSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(testError).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			SysAuthSvcFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), model.IntegrationSystemReference, modelIntSys.ID).Return(nil, nil)
				return svc
			},
			OAuth20SvcFn: func() *automock.Oauth20Service {
				svc := &automock.Oauth20Service{}
				return svc
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			SysAuthSvcFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20SvcFn: func() *automock.Oauth20Service {
				svc := &automock.Oauth20Service{}
				return svc
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelIntSys, nil).Once()
				intSysSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return intSysSvc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				intSysConv := &automock.IntegrationSystemConverter{}
				return intSysConv
			},
			SysAuthSvcFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), model.IntegrationSystemReference, modelIntSys.ID).Return(nil, nil)
				return svc
			},
			OAuth20SvcFn: func() *automock.Oauth20Service {
				svc := &automock.Oauth20Service{}
				return svc
			},
			ExpectedError: testError,
		},
		{
			Name: "Return error when listing all auths failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				svc := &automock.IntegrationSystemService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), "foo").Return(modelIntSys, nil).Once()
				return svc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				conv := &automock.IntegrationSystemConverter{}
				return conv
			},
			SysAuthSvcFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), model.IntegrationSystemReference, modelIntSys.ID).Return(nil, testError)
				return svc
			},
			OAuth20SvcFn: func() *automock.Oauth20Service {
				svc := &automock.Oauth20Service{}
				return svc
			},
			ExpectedError: testError,
		},
		{
			Name: "Return error when li failed ",
			TxFn: txGen.ThatDoesntExpectCommit,
			IntSysSvcFn: func() *automock.IntegrationSystemService {
				svc := &automock.IntegrationSystemService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), "foo").Return(modelIntSys, nil).Once()
				return svc
			},
			IntSysConvFn: func() *automock.IntegrationSystemConverter {
				conv := &automock.IntegrationSystemConverter{}
				return conv
			},
			SysAuthSvcFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), model.IntegrationSystemReference, modelIntSys.ID).Return([]model.SystemAuth{testAuth}, nil)
				return svc
			},
			OAuth20SvcFn: func() *automock.Oauth20Service {
				svc := &automock.Oauth20Service{}
				svc.On("DeleteClientCredentials", txtest.CtxWithDBMatcher(), testAuth.Value.Credential.Oauth.ClientID).Return(testError)
				return svc
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			intSysSvc := testCase.IntSysSvcFn()
			intSysConv := testCase.IntSysConvFn()
			sysAuthSvc := testCase.SysAuthSvcFn()
			OAuth20Svc := testCase.OAuth20SvcFn()
			resolver := integrationsystem.NewResolver(transact, intSysSvc, sysAuthSvc, OAuth20Svc, intSysConv, nil)

			// WHEN
			result, err := resolver.DeleteIntegrationSystem(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			intSysSvc.AssertExpectations(t)
			intSysConv.AssertExpectations(t)
		})
	}
}

func TestResolver_Auths(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)

	parentIntegrationSystem := fixGQLIntegrationSystem(testID, testName)

	modelSysAuths := []model.SystemAuth{
		fixModelSystemAuth("bar", parentIntegrationSystem.ID, fixModelAuth()),
		fixModelSystemAuth("baz", parentIntegrationSystem.ID, fixModelAuth()),
		fixModelSystemAuth("faz", parentIntegrationSystem.ID, fixModelAuth()),
	}

	gqlSysAuths := []*graphql.SystemAuth{
		fixGQLSystemAuth("bar", fixGQLAuth()),
		fixGQLSystemAuth("baz", fixGQLAuth()),
		fixGQLSystemAuth("faz", fixGQLAuth()),
	}

	txGen := txtest.NewTransactionContextGenerator(testError)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		SysAuthSvcFn    func() *automock.SystemAuthService
		SysAuthConvFn   func() *automock.SystemAuthConverter
		ExpectedOutput  []*graphql.SystemAuth
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), model.IntegrationSystemReference, parentIntegrationSystem.ID).Return(modelSysAuths, nil).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				sysAuthConv.On("ToGraphQL", &modelSysAuths[0]).Return(gqlSysAuths[0]).Once()
				sysAuthConv.On("ToGraphQL", &modelSysAuths[1]).Return(gqlSysAuths[1]).Once()
				sysAuthConv.On("ToGraphQL", &modelSysAuths[2]).Return(gqlSysAuths[2]).Once()
				return sysAuthConv
			},
			ExpectedOutput: gqlSysAuths,
		},
		{
			Name:            "Error when listing for object",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), model.IntegrationSystemReference, parentIntegrationSystem.ID).Return(nil, testError).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			ExpectedError: testError,
		},
		{
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			ExpectedError: testError,
		},
		{
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), model.IntegrationSystemReference, parentIntegrationSystem.ID).Return(modelSysAuths, nil).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			sysAuthSvc := testCase.SysAuthSvcFn()
			sysAuthConv := testCase.SysAuthConvFn()

			resolver := integrationsystem.NewResolver(transact, nil, sysAuthSvc, nil, nil, sysAuthConv)

			// WHEN
			result, err := resolver.Auths(ctx, parentIntegrationSystem)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			sysAuthSvc.AssertExpectations(t)
			sysAuthConv.AssertExpectations(t)
		})
	}

	t.Run("Error when parent object is nil", func(t *testing.T) {
		resolver := integrationsystem.NewResolver(nil, nil, nil, nil, nil, nil)

		// WHEN
		result, err := resolver.Auths(context.TODO(), nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Integration System cannot be empty")
		assert.Nil(t, result)
	})
}

func fixOauth() model.SystemAuth {
	return model.SystemAuth{
		ID:       "foo",
		TenantID: "foo",
		Value: &model.Auth{
			Credential: model.CredentialData{
				Basic: nil,
				Oauth: &model.OAuthCredentialData{
					ClientID:     "foo",
					ClientSecret: "foo",
					URL:          "foo",
				},
			},
		},
	}
}
