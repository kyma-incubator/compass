package systemauth_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var contextParam = txtest.CtxWithDBMatcher()

func TestResolver_GenericDeleteSystemAuth(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	objectID := "bar"
	objectType := model.RuntimeReference
	modelSystemAuth := fixModelSystemAuth(id, objectType, objectID, fixModelAuth())
	gqlSystemAuth := fixGQLSystemAuth(id, fixGQLAuth())

	testCases := []struct {
		Name               string
		PersistenceFn      func() *persistenceautomock.PersistenceTx
		TransactionerFn    func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn          func() *automock.SystemAuthService
		ConverterFn        func() *automock.SystemAuthConverter
		ExpectedSystemAuth *graphql.SystemAuth
		ExpectedErr        error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("Get", contextParam, id).Return(modelSystemAuth, nil).Once()
				svc.On("DeleteByIDForObject", contextParam, objectType, id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(gqlSystemAuth).Once()
				return conv
			},
			ExpectedSystemAuth: gqlSystemAuth,
			ExpectedErr:        nil,
		},
		{
			Name: "Error - Get SystemAuth",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("Get", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				return conv
			},
			ExpectedSystemAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name: "Error - Delete",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("Get", contextParam, id).Return(modelSystemAuth, nil).Once()
				svc.On("DeleteByIDForObject", contextParam, objectType, id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(gqlSystemAuth).Once()
				return conv
			},
			ExpectedSystemAuth: nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := systemauth.NewResolver(transact, svc, converter)

			// when
			fn := resolver.GenericDeleteSystemAuth(objectType)
			result, err := fn(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedSystemAuth, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persistTx.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}
