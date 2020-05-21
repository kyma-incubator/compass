package eventing

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResolver_SetDefaultEventingForApplication(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID.String(), externalTenantID.String())

	testErr := errors.New("this is a test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	app := fixApplicationModel("test-app")

	defaultEveningURL := "https://eventing.domain.local"
	modelAppEventingCfg := fixModelApplicationEventingConfiguration(t, defaultEveningURL)
	gqlAppEventingCfg := fixGQLApplicationEventingConfiguration(defaultEveningURL)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		EventingSvcFn   func() *automock.EventingService
		AppSvcFn        func() *automock.ApplicationService
		ExpectedOutput  *graphql.ApplicationEventingConfiguration
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("SetForApplication", txtest.CtxWithDBMatcher(), runtimeID, app).
					Return(modelAppEventingCfg, nil).Once()
				return eventingSvc
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), applicationID.String()).Return(&app, nil)
				return appSvc
			},
			ExpectedOutput: gqlAppEventingCfg,
			ExpectedError:  nil,
		}, {
			Name:            "Error when setting the runtime for the application",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("SetForApplication", txtest.CtxWithDBMatcher(), runtimeID, app).
					Return(nil, testErr).Once()

				return eventingSvc
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), applicationID.String()).Return(&app, nil)
				return appSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when getting the application",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				return eventingSvc
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), applicationID.String()).Return(nil, testErr)
				return appSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				return eventingSvc
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("SetForApplication", txtest.CtxWithDBMatcher(), runtimeID, app).
					Return(modelAppEventingCfg, nil).Once()

				return eventingSvc
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), applicationID.String()).Return(&app, nil)
				return appSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			eventingSvc := testCase.EventingSvcFn()
			appSvc := testCase.AppSvcFn()
			resolver := NewResolver(transact, eventingSvc, appSvc)

			// WHEN
			result, err := resolver.SetEventingForApplication(ctx, applicationID.String(), runtimeID.String())

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, eventingSvc, appSvc, transact, persist)
		})
	}

	t.Run("Error when runtime ID is not a valid UUID", func(t *testing.T) {
		// GIVEN
		resolver := NewResolver(nil, nil, nil)

		// WHEN
		result, err := resolver.SetEventingForApplication(ctx, applicationID.String(), "abc")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while parsing runtime ID as UUID")
		assert.Nil(t, result)
	})

	t.Run("Error when application ID is not a valid UUID", func(t *testing.T) {
		// GIVEN
		resolver := NewResolver(nil, nil, nil)

		// WHEN
		result, err := resolver.SetEventingForApplication(ctx, "abc", runtimeID.String())

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while parsing application ID as UUID")
		assert.Nil(t, result)
	})
}

func TestResolver_UnsetDefaultEventingForApplication(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID.String(), externalTenantID.String())

	app := fixApplicationModel("test-app")

	testErr := errors.New("this is a test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	defaultEveningURL := "https://eventing.domain.local/test-app/events/v1"
	modelAppEventingCfg := fixModelApplicationEventingConfiguration(t, defaultEveningURL)
	gqlAppEventingCfg := fixGQLApplicationEventingConfiguration(defaultEveningURL)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		EventingSvcFn   func() *automock.EventingService
		AppSvcFn        func() *automock.ApplicationService
		ExpectedOutput  *graphql.ApplicationEventingConfiguration
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("UnsetForApplication", txtest.CtxWithDBMatcher(), app).
					Return(modelAppEventingCfg, nil).Once()

				return eventingSvc
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), applicationID.String()).Return(&app, nil)
				return appSvc
			},
			ExpectedOutput: gqlAppEventingCfg,
			ExpectedError:  nil,
		}, {
			Name:            "Error when deleting the default eventing runtime for the application",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("UnsetForApplication", txtest.CtxWithDBMatcher(), app).
					Return(nil, testErr).Once()

				return eventingSvc
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), applicationID.String()).Return(&app, nil)
				return appSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when getting application",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				return eventingSvc
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), applicationID.String()).Return(nil, testErr)
				return appSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				return eventingSvc
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("UnsetForApplication", txtest.CtxWithDBMatcher(), app).
					Return(modelAppEventingCfg, nil).Once()
				return eventingSvc
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), applicationID.String()).Return(&app, nil).Once()
				return appSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			eventingSvc := testCase.EventingSvcFn()
			appSvc := testCase.AppSvcFn()
			resolver := NewResolver(transact, eventingSvc, appSvc)

			// WHEN
			result, err := resolver.UnsetEventingForApplication(ctx, applicationID.String())

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, eventingSvc, appSvc, transact, persist)
		})
	}

	t.Run("Error when application ID is not a valid UUID", func(t *testing.T) {
		// GIVEN
		resolver := NewResolver(nil, nil, nil)

		// WHEN
		result, err := resolver.UnsetEventingForApplication(ctx, "abc")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while parsing application ID as UUID")
		assert.Nil(t, result)
	})
}
