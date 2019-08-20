package webhook_test

import (
	"context"
	"errors"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"

	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
)

func TestResolver_AddWebhook(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	gqlWebhookInput := fixGQLWebhookInput("foo")
	modelWebhookInput := fixModelWebhookInput("foo")

	gqlWebhook := fixGQLWebhook(id, "", "")
	modelWebhook := fixModelWebhook(id, applicationID, givenTenant(), "foo")

	testCases := []struct {
		Name               string
		PersistenceFn      func() *persistenceautomock.PersistenceTx
		TransactionerFn    func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn          func() *automock.WebhookService
		AppServiceFn       func() *automock.ApplicationService
		ConverterFn        func() *automock.WebhookConverter
		InputApplicationID string
		InputWebhook       graphql.WebhookInput
		ExpectedWebhook    *graphql.Webhook
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			PersistenceFn:   mockPersistenceContextThatExpectsCommit,
			TransactionerFn: mockTransactionerThatSucceed,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", ctxWithDBMatcher(), applicationID, *modelWebhookInput).Return(id, nil).Once()
				svc.On("Get", ctxWithDBMatcher(), id).Return(modelWebhook, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", ctxWithDBMatcher(), applicationID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    gqlWebhook,
			ExpectedErr:        nil,
		},
		{
			Name:            "Returns error when application not exist",
			PersistenceFn:   mockPersistenceContextThatDontExpectCommit,
			TransactionerFn: mockTransactionerThatSucceed,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", ctxWithDBMatcher(), applicationID).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    nil, ExpectedErr: errors.New("Cannot add Webhook to not existing Application"),
		},
		{
			Name:            "Returns error when application existence check failed",
			PersistenceFn:   mockPersistenceContextThatDontExpectCommit,
			TransactionerFn: mockTransactionerThatSucceed,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", ctxWithDBMatcher(), applicationID).Return(false, testErr).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Returns error when webhook creation failed",
			PersistenceFn:   mockPersistenceContextThatDontExpectCommit,
			TransactionerFn: mockTransactionerThatSucceed,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", ctxWithDBMatcher(), applicationID, *modelWebhookInput).Return("", testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", ctxWithDBMatcher(), applicationID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Returns error when webhook retrieval failed",
			PersistenceFn:   mockPersistenceContextThatDontExpectCommit,
			TransactionerFn: mockTransactionerThatSucceed,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", ctxWithDBMatcher(), applicationID, *modelWebhookInput).Return(id, nil).Once()
				svc.On("Get", ctxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", ctxWithDBMatcher(), applicationID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    nil,
			ExpectedErr:        testErr,
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
			result, err := resolver.AddApplicationWebhook(context.TODO(), testCase.InputApplicationID, testCase.InputWebhook)

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
	id := "bar"
	gqlWebhookInput := fixGQLWebhookInput("foo")
	modelWebhookInput := fixModelWebhookInput("foo")
	gqlWebhook := fixGQLWebhook(id, "", "")
	modelWebhook := fixModelWebhook(id, applicationID, givenTenant(), "foo")

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		InputWebhookID  string
		InputWebhook    graphql.WebhookInput
		ExpectedWebhook *graphql.Webhook
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			PersistenceFn:   mockPersistenceContextThatExpectsCommit,
			TransactionerFn: mockTransactionerThatSucceed,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", ctxWithDBMatcher(), id, *modelWebhookInput).Return(nil).Once()
				svc.On("Get", ctxWithDBMatcher(), id).Return(modelWebhook, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns error when webhook update failed",
			TransactionerFn: mockTransactionerThatSucceed,
			PersistenceFn:   mockPersistenceContextThatDontExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", ctxWithDBMatcher(), id, *modelWebhookInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when webhook retrieval failed",
			TransactionerFn: mockTransactionerThatSucceed,
			PersistenceFn:   mockPersistenceContextThatDontExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", ctxWithDBMatcher(), id, *modelWebhookInput).Return(nil).Once()
				svc.On("Get", ctxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
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
			result, err := resolver.UpdateApplicationWebhook(context.TODO(), testCase.InputWebhookID, testCase.InputWebhook)

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
	id := "bar"

	gqlWebhookInput := fixGQLWebhookInput("foo")

	gqlWebhook := fixGQLWebhook(id, "", "")
	modelWebhook := fixModelWebhook(id, applicationID, givenTenant(), "foo")

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		InputWebhookID  string
		InputWebhook    graphql.WebhookInput
		ExpectedWebhook *graphql.Webhook
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: mockTransactionerThatSucceed,
			PersistenceFn:   mockPersistenceContextThatExpectsCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", ctxWithDBMatcher(), id).Return(modelWebhook, nil).Once()
				svc.On("Delete", ctxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns error when webhook retrieval failed",
			TransactionerFn: mockTransactionerThatSucceed,
			PersistenceFn:   mockPersistenceContextThatDontExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", ctxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when webhook deletion failed",
			TransactionerFn: mockTransactionerThatSucceed,
			PersistenceFn:   mockPersistenceContextThatDontExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", ctxWithDBMatcher(), id).Return(modelWebhook, nil).Once()
				svc.On("Delete", ctxWithDBMatcher(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
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
			result, err := resolver.DeleteApplicationWebhook(context.TODO(), testCase.InputWebhookID)

			// then
			assert.Equal(t, testCase.ExpectedWebhook, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transactionerMock.AssertExpectations(t)
		})
	}
}

func mockPersistenceContextThatExpectsCommit() *persistenceautomock.PersistenceTx {
	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil).Once()
	return persistTx
}

func mockPersistenceContextThatDontExpectCommit() *persistenceautomock.PersistenceTx {
	persistTx := &persistenceautomock.PersistenceTx{}
	return persistTx
}

func mockTransactionerThatSucceed(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
	transact := &persistenceautomock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Once()
	transact.On("RollbackUnlessCommited", persistTx).Return().Once()
	return transact
}

func ctxWithDBMatcher() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		_, err := persistence.FromCtx(ctx)
		return err == nil
	})
}

// TODO test another cases!!!
