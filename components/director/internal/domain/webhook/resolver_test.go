package webhook_test

import (
	"context"
	"errors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/stretchr/testify/require"

	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"

	"github.com/stretchr/testify/assert"
)

const (
	testCaseErrorOnStartingTransaction = "Returns error on starting transaction"
	testCaseErrorOnCommit              = "Returns error on committing transaction"
)

func TestResolver_AddWebhook(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	givenAppID := "foo"
	id := "bar"
	gqlWebhookInput := fixGQLWebhookInput("foo")
	modelWebhookInput := fixModelWebhookInput("foo")

	gqlWebhook := fixGQLWebhook(id, "", "")
	modelWebhook := fixModelWebhook(id, givenAppID, givenTenant(), "foo")

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.WebhookService
		AppServiceFn    func() *automock.ApplicationService
		ConverterFn     func() *automock.WebhookConverter
		ExpectedWebhook *graphql.Webhook
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), givenAppID, *modelWebhookInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelWebhook, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name:          testCaseErrorOnStartingTransaction,
			PersistenceFn: txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, givenError()).Once()
				return transact
			},
			ServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedErr: givenError(),
		},
		{
			Name: testCaseErrorOnCommit,
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(givenError()).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), givenAppID, *modelWebhookInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelWebhook, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedErr: givenError(),
		},
		{
			Name:            "Returns error when application not exist",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil, ExpectedErr: errors.New("cannot add Webhook to not existing Application"),
		},
		{
			Name:            "Returns error when application existence check failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(false, testErr).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when webhook creation failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), givenAppID, *modelWebhookInput).Return("", testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when webhook retrieval failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), givenAppID, *modelWebhookInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), givenAppID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			appSvc := testCase.AppServiceFn()
			converter := testCase.ConverterFn()

			persistTxMock := testCase.PersistenceFn()
			transactionerMock := testCase.TransactionerFn(persistTxMock)

			resolver := webhook.NewResolver(transactionerMock, svc, appSvc, converter)

			// when
			result, err := resolver.AddApplicationWebhook(context.TODO(), givenAppID, *gqlWebhookInput)

			// then
			assert.Equal(t, testCase.ExpectedWebhook, result)
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			svc.AssertExpectations(t)
			appSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTxMock.AssertExpectations(t)
			transactionerMock.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateWebhook(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	givenWebhookID := "bar"
	gqlWebhookInput := fixGQLWebhookInput("foo")
	modelWebhookInput := fixModelWebhookInput("foo")
	gqlWebhook := fixGQLWebhook(givenWebhookID, "", "")
	modelWebhook := fixModelWebhook(givenWebhookID, applicationID, givenTenant(), "foo")

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ExpectedWebhook *graphql.Webhook
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), givenWebhookID, *modelWebhookInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID).Return(modelWebhook, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name:          testCaseErrorOnStartingTransaction,
			PersistenceFn: txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, givenError()).Once()
				return transact
			},
			ServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedErr: givenError(),
		},
		{
			Name: testCaseErrorOnCommit,
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(givenError()).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), givenWebhookID, *modelWebhookInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID).Return(modelWebhook, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedErr: givenError(),
		},
		{
			Name:            "Returns error when webhook update failed",
			TransactionerFn: txtest.TransactionerThatSucceeds,
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), givenWebhookID, *modelWebhookInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when webhook retrieval failed",
			TransactionerFn: txtest.TransactionerThatSucceeds,
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), givenWebhookID, *modelWebhookInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			persistTxMock := testCase.PersistenceFn()
			transactionerMock := testCase.TransactionerFn(persistTxMock)

			resolver := webhook.NewResolver(transactionerMock, svc, nil, converter)

			// when
			result, err := resolver.UpdateApplicationWebhook(context.TODO(), givenWebhookID, *gqlWebhookInput)

			// then
			assert.Equal(t, testCase.ExpectedWebhook, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transactionerMock.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteWebhook(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	givenWebhookID := "bar"

	gqlWebhook := fixGQLWebhook(givenWebhookID, "", "")
	modelWebhook := fixModelWebhook(givenWebhookID, applicationID, givenTenant(), "foo")

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ExpectedWebhook *graphql.Webhook
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txtest.TransactionerThatSucceeds,
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID).Return(modelWebhook, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), givenWebhookID).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name:          testCaseErrorOnStartingTransaction,
			PersistenceFn: txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, givenError()).Once()
				return transact
			},
			ServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},

			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedErr: givenError(),
		},
		{
			Name: testCaseErrorOnCommit,
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(givenError()).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID).Return(modelWebhook, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), givenWebhookID).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ExpectedErr: givenError(),
		},
		{
			Name:            "Returns error when webhook retrieval failed",
			TransactionerFn: txtest.TransactionerThatSucceeds,
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when webhook deletion failed",
			TransactionerFn: txtest.TransactionerThatSucceeds,
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), givenWebhookID).Return(modelWebhook, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), givenWebhookID).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook, nil).Once()
				return conv
			},
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			persistTxMock := testCase.PersistenceFn()
			transactionerMock := testCase.TransactionerFn(persistTxMock)

			resolver := webhook.NewResolver(transactionerMock, svc, nil, converter)

			// when
			result, err := resolver.DeleteApplicationWebhook(context.TODO(), givenWebhookID)

			// then
			assert.Equal(t, testCase.ExpectedWebhook, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transactionerMock.AssertExpectations(t)
		})
	}
}
