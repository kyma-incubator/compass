package packageinstanceauth_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestResolver_DeletePackageInstanceAuth(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelInstanceAuth := fixModelPackageInstanceAuth(id)
	gqlInstanceAuth := fixGQLPackageInstanceAuth(id)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.Service
		ConverterFn     func() *automock.Converter
		ExpectedResult  *graphql.PackageInstanceAuth
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("ToGraphQL", modelInstanceAuth).Return(gqlInstanceAuth).Once()
				return conv
			},
			ExpectedResult: gqlInstanceAuth,
			ExpectedErr:    nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Instance Auth retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Instance Auth deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := packageinstanceauth.NewResolver(transact, svc, converter)

			// when
			result, err := resolver.DeletePackageInstanceAuth(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transact.AssertExpectations(t)
			persist.AssertExpectations(t)
		})
	}
}

func fixGQLPackageInstanceAuth(id string) *graphql.PackageInstanceAuth {
	return &graphql.PackageInstanceAuth{
		ID: id,
	}
}
