package eventing

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResolver_SetDefaultEventingForApplication(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID.String())

	testErr := errors.New("this is a test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	defaultEveningURL := "https://eventing.domain.local"
	modelAppEventingCfg := fixModelApplicationEventingConfiguration(defaultEveningURL)
	gqlAppEventingCfg := fixGQLApplicationEventingConfiguration(defaultEveningURL)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		EventingSvcFn   func() *automock.EventingService
		ExpectedOutput  *graphql.ApplicationEventingConfiguration
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("SetAsDefaultForApplication", txtest.CtxWithDBMatcher(), runtimeID, applicationID).
					Return(modelAppEventingCfg, nil).Once()

				return eventingSvc
			},
			ExpectedOutput: gqlAppEventingCfg,
			ExpectedError:  nil,
		}, {
			Name:            "Error when setting the runtime for the application",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("SetAsDefaultForApplication", txtest.CtxWithDBMatcher(), runtimeID, applicationID).
					Return(nil, testErr).Once()

				return eventingSvc
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
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("SetAsDefaultForApplication", txtest.CtxWithDBMatcher(), runtimeID, applicationID).
					Return(modelAppEventingCfg, nil).Once()

				return eventingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			eventingSvc := testCase.EventingSvcFn()
			resolver := NewResolver(transact, eventingSvc)

			// WHEN
			result, err := resolver.SetDefaultEventingForApplication(ctx, applicationID.String(), runtimeID.String())

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, eventingSvc, transact, persist)
		})
	}

	t.Run("Error when runtime ID is not a valid UUID", func(t *testing.T) {
		// GIVEN
		resolver := NewResolver(nil, nil)

		// WHEN
		result, err := resolver.SetDefaultEventingForApplication(ctx, applicationID.String(), "abc")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while parsing runtime ID as UUID")
		assert.Nil(t, result)
	})

	t.Run("Error when application ID is not a valid UUID", func(t *testing.T) {
		// GIVEN
		resolver := NewResolver(nil, nil)

		// WHEN
		result, err := resolver.SetDefaultEventingForApplication(ctx, "abc", runtimeID.String())

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while parsing application ID as UUID")
		assert.Nil(t, result)
	})
}

func TestResolver_UnsetDefaultEventingForApplication(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID.String())

	testErr := errors.New("this is a test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	defaultEveningURL := "https://eventing.domain.local"
	modelAppEventingCfg := fixModelApplicationEventingConfiguration(defaultEveningURL)
	gqlAppEventingCfg := fixGQLApplicationEventingConfiguration(defaultEveningURL)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		EventingSvcFn   func() *automock.EventingService
		ExpectedOutput  *graphql.ApplicationEventingConfiguration
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("UnsetDefaultForApplication", txtest.CtxWithDBMatcher(), applicationID).
					Return(modelAppEventingCfg, nil).Once()

				return eventingSvc
			},
			ExpectedOutput: gqlAppEventingCfg,
			ExpectedError:  nil,
		}, {
			Name:            "Error when deleting the default eventing runtime for the application",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("UnsetDefaultForApplication", txtest.CtxWithDBMatcher(), applicationID).
					Return(nil, testErr).Once()

				return eventingSvc
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
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("UnsetDefaultForApplication", txtest.CtxWithDBMatcher(), applicationID).
					Return(modelAppEventingCfg, nil).Once()

				return eventingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			eventingSvc := testCase.EventingSvcFn()
			resolver := NewResolver(transact, eventingSvc)

			// WHEN
			result, err := resolver.UnsetDefaultEventingForApplication(ctx, applicationID.String())

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, eventingSvc, transact, persist)
		})
	}

	t.Run("Error when application ID is not a valid UUID", func(t *testing.T) {
		// GIVEN
		resolver := NewResolver(nil, nil)

		// WHEN
		result, err := resolver.UnsetDefaultEventingForApplication(ctx, "abc")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while parsing application ID as UUID")
		assert.Nil(t, result)
	})
}
