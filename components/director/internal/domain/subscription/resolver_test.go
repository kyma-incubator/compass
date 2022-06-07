package subscription_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/subscription"
	"github.com/kyma-incubator/compass/components/director/internal/domain/subscription/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResolver_SubscribeTenant(t *testing.T) {
	// GIVEN
	subscriptionProviderID := "distinguish-value-123"
	providerSubaccountID := "provder-subaccount"
	subscribedSubaacountID := "subscribed-subaccount"
	subscriptionAppName := "app-name"
	region := "eu-1"
	testErr := errors.New("new-error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.SubscriptionService
		ExpectedOutput  bool
		ExpectedErr     error
	}{
		{
			Name:            "Success for Application flow",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.ApplicationTemplate, nil).Once()
				svc.On("SubscribeTenantToApplication", txtest.CtxWithDBMatcher(), subscriptionProviderID, region, providerSubaccountID, subscribedSubaacountID, subscriptionAppName).Return(true, nil).Once()
				svc.AssertNotCalled(t, "SubscribeTenantToRuntime")
				return svc
			},
			ExpectedOutput: true,
			ExpectedErr:    nil,
		},
		{
			Name:            "Success for Runtime flow",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.Runtime, nil).Once()
				svc.AssertNotCalled(t, "SubscribeTenantToApplication")
				svc.On("SubscribeTenantToRuntime", txtest.CtxWithDBMatcher(), subscriptionProviderID, subscribedSubaacountID, providerSubaccountID, region).Return(true, nil).Once()
				return svc
			},
			ExpectedOutput: true,
			ExpectedErr:    nil,
		},
		{
			Name:            "Error on transaction begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.AssertNotCalled(t, "SubscribeTenantToRuntime")
				svc.AssertNotCalled(t, "SubscribeTenantToApplication")
				svc.AssertNotCalled(t, "SubscribeTenantToRuntime")
				return svc
			},
			ExpectedOutput: false,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error on flow determination",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.ApplicationTemplate, testErr).Once()
				svc.AssertNotCalled(t, "SubscribeTenantToApplication")
				svc.AssertNotCalled(t, "SubscribeTenantToRuntime")
				return svc
			},
			ExpectedOutput: false,
			ExpectedErr:    errors.Wrapf(testErr, "while determining subscription flow"),
		},
		{
			Name:            "Error on subscription to applications fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.ApplicationTemplate, nil).Once()
				svc.On("SubscribeTenantToApplication", txtest.CtxWithDBMatcher(), subscriptionProviderID, region, providerSubaccountID, subscribedSubaacountID, subscriptionAppName).Return(false, testErr).Once()
				svc.AssertNotCalled(t, "SubscribeTenantToRuntime")
				return svc
			},
			ExpectedOutput: false,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error on subscription to runtime fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.Runtime, nil).Once()
				svc.On("SubscribeTenantToRuntime", txtest.CtxWithDBMatcher(), subscriptionProviderID, subscribedSubaacountID, providerSubaccountID, region).Return(false, testErr).Once()
				svc.AssertNotCalled(t, "SubscribeTenantToApplication")
				return svc
			},
			ExpectedOutput: false,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error on commit fail",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.Runtime, nil).Once()
				svc.On("SubscribeTenantToRuntime", txtest.CtxWithDBMatcher(), subscriptionProviderID, subscribedSubaacountID, providerSubaccountID, region).Return(true, nil).Once()
				svc.AssertNotCalled(t, "SubscribeTenantToApplication")
				return svc
			},
			ExpectedOutput: false,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			resolver := subscription.NewResolver(transact, svc)

			defer mock.AssertExpectationsForObjects(t, transact, persistTx, svc)

			// WHEN
			result, err := resolver.SubscribeTenant(context.TODO(), subscriptionProviderID, subscribedSubaacountID, providerSubaccountID, region, subscriptionAppName)

			// then
			assert.Equal(t, testCase.ExpectedOutput, result)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			}
		})
	}
}

func TestResolver_UnsubscribeTenant(t *testing.T) {
	// GIVEN
	subscriptionProviderID := "distinguish-value-123"
	providerSubaccountID := "provder-subaccount"
	subscribedSubaacountID := "subscribed-subaccount"
	region := "eu-1"
	testErr := errors.New("new-error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.SubscriptionService
		ExpectedOutput  bool
		ExpectedErr     error
	}{
		{
			Name:            "Success when flow is runtime",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.On("UnsubscribeTenantFromRuntime", txtest.CtxWithDBMatcher(), subscriptionProviderID, subscribedSubaacountID, providerSubaccountID, region).Return(true, nil).Once()
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.Runtime, nil).Once()
				svc.AssertNotCalled(t, "UnsubscribeTenantFromApplication")
				return svc
			},
			ExpectedOutput: true,
			ExpectedErr:    nil,
		},
		{
			Name:            "Success when flow is application",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.On("UnsubscribeTenantFromApplication", txtest.CtxWithDBMatcher(), subscriptionProviderID, region, providerSubaccountID).Return(true, nil).Once()
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.ApplicationTemplate, nil).Once()
				svc.AssertNotCalled(t, "UnsubscribeTenantFromRuntime")
				return svc
			},
			ExpectedOutput: true,
			ExpectedErr:    nil,
		},
		{
			Name:            "Error on transaction begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.AssertNotCalled(t, "UnsubscribeTenantFromRuntime")
				return svc
			},
			ExpectedOutput: false,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error determining flow type",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.AssertNotCalled(t, "UnsubscribeTenantFromRuntime")
				svc.AssertNotCalled(t, "UnsubscribeTenantFromApplication")
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.ApplicationTemplate, testErr).Once()
				return svc
			},
			ExpectedOutput: false,
			ExpectedErr:    errors.Wrapf(testErr, "while determining subscription flow"),
		},
		{
			Name:            "Error on unsubscription from runtime fail",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.On("UnsubscribeTenantFromRuntime", txtest.CtxWithDBMatcher(), subscriptionProviderID, subscribedSubaacountID, providerSubaccountID, region).Return(false, testErr).Once()
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.Runtime, nil).Once()

				return svc
			},
			ExpectedOutput: false,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error on unsubscription from application fail",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.On("UnsubscribeTenantFromApplication", txtest.CtxWithDBMatcher(), subscriptionProviderID, region, providerSubaccountID).Return(false, testErr).Once()
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.ApplicationTemplate, nil).Once()

				return svc
			},
			ExpectedOutput: false,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error on commit fail",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SubscriptionService {
				svc := &automock.SubscriptionService{}
				svc.On("DetermineSubscriptionFlow", txtest.CtxWithDBMatcher(), subscriptionProviderID, region).Return(resource.Runtime, nil).Once()
				svc.On("UnsubscribeTenantFromRuntime", txtest.CtxWithDBMatcher(), subscriptionProviderID, subscribedSubaacountID, providerSubaccountID, region).Return(true, nil).Once()

				return svc
			},
			ExpectedOutput: false,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			resolver := subscription.NewResolver(transact, svc)

			defer mock.AssertExpectationsForObjects(t, transact, persistTx, svc)

			// WHEN
			result, err := resolver.UnsubscribeTenant(context.TODO(), subscriptionProviderID, subscribedSubaacountID, providerSubaccountID, region)

			// then
			assert.Equal(t, testCase.ExpectedOutput, result)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			}
		})
	}
}
