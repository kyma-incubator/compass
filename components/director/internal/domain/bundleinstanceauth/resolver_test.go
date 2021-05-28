package bundleinstanceauth_test

import (
	"context"
	"testing"

	pkgmock "github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResolver_DeleteBundleInstanceAuth(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelInstanceAuth := fixSimpleModelBundleInstanceAuth(id)
	gqlInstanceAuth := fixSimpleGQLBundleInstanceAuth(id)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.Service
		ConverterFn       func() *automock.Converter
		BundleConverterFn func() *pkgmock.BundleConverter
		ExpectedResult    *graphql.BundleInstanceAuth
		ExpectedErr       error
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
				conv.On("ToGraphQL", modelInstanceAuth).Return(gqlInstanceAuth, nil).Once()
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
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
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
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
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
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
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
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
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
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
			bndlConverter := testCase.BundleConverterFn()

			resolver := bundleinstanceauth.NewResolver(transact, svc, nil, converter, bndlConverter)

			// when
			result, err := resolver.DeleteBundleInstanceAuth(context.TODO(), id)

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

func TestResolver_RequestBundleInstanceAuthCreation(t *testing.T) {
	// given
	modelBundle := fixModelBundle(testBundleID, nil, nil)
	gqlRequestInput := fixGQLRequestInput()
	modelRequestInput := fixModelRequestInput()

	modelInstanceAuth := fixModelBundleInstanceAuthWithoutContextAndInputParams(testID, testBundleID, testTenant, nil, nil, &testRuntimeID)
	gqlInstanceAuth := fixGQLBundleInstanceAuthWithoutContextAndInputParams(testID, nil, nil, &testRuntimeID)

	txGen := txtest.NewTransactionContextGenerator(testError)

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.Service
		BndlServiceFn     func() *automock.BundleService
		ConverterFn       func() *automock.Converter
		BundleConverterFn func() *pkgmock.BundleConverter
		ExpectedResult    *graphql.BundleInstanceAuth
		ExpectedErr       error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Create", txtest.CtxWithDBMatcher(), testBundleID, *modelRequestInput, modelInstanceAuth.Auth, modelInstanceAuth.InputParams).Return(testID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelInstanceAuth, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("RequestInputFromGraphQL", *gqlRequestInput).Return(*modelRequestInput).Once()
				conv.On("ToGraphQL", modelInstanceAuth).Return(gqlInstanceAuth, nil).Once()
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
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
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Create", txtest.CtxWithDBMatcher(), testBundleID, *modelRequestInput, modelInstanceAuth.Auth, modelInstanceAuth.InputParams).Return(testID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelInstanceAuth, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("RequestInputFromGraphQL", *gqlRequestInput).Return(*modelRequestInput, nil).Once()
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when Bundle retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(nil, testError).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when Instance Auth creation failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Create", txtest.CtxWithDBMatcher(), testBundleID, *modelRequestInput, modelInstanceAuth.Auth, modelInstanceAuth.InputParams).Return("", testError).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("RequestInputFromGraphQL", *gqlRequestInput).Return(*modelRequestInput, nil).Once()
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when Instance Auth retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Create", txtest.CtxWithDBMatcher(), testBundleID, *modelRequestInput, modelInstanceAuth.Auth, modelInstanceAuth.InputParams).Return(testID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("RequestInputFromGraphQL", *gqlRequestInput).Return(*modelRequestInput, nil).Once()
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			bndlSvc := testCase.BndlServiceFn()
			converter := testCase.ConverterFn()
			bndlConverter := testCase.BundleConverterFn()

			resolver := bundleinstanceauth.NewResolver(transact, svc, bndlSvc, converter, bndlConverter)

			result, err := resolver.RequestBundleInstanceAuthCreation(context.TODO(), testBundleID, *gqlRequestInput)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, persist, transact, svc, bndlSvc, converter)
		})
	}
}

func TestResolver_SetBundleInstanceAuth(t *testing.T) {
	// given

	testAuthID := "foo"

	modelSetInput := fixModelSetInput()
	gqlSetInput := fixGQLSetInput()

	modelInstanceAuth := fixModelBundleInstanceAuthWithoutContextAndInputParams(testID, testBundleID, testTenant, nil, nil, nil)
	gqlInstanceAuth := fixGQLBundleInstanceAuthWithoutContextAndInputParams(testID, nil, nil, nil)

	txGen := txtest.NewTransactionContextGenerator(testError)

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.Service
		ConverterFn       func() *automock.Converter
		BundleConverterFn func() *pkgmock.BundleConverter
		ExpectedResult    *graphql.BundleInstanceAuth
		ExpectedErr       error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("SetAuth", txtest.CtxWithDBMatcher(), testID, *modelSetInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelInstanceAuth, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("SetInputFromGraphQL", *gqlSetInput).Return(*modelSetInput, nil).Once()
				conv.On("ToGraphQL", modelInstanceAuth).Return(gqlInstanceAuth, nil).Once()
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
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
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("SetAuth", txtest.CtxWithDBMatcher(), testID, *modelSetInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelInstanceAuth, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("SetInputFromGraphQL", *gqlSetInput).Return(*modelSetInput, nil).Once()
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when setting Instance Auth failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("SetAuth", txtest.CtxWithDBMatcher(), testID, *modelSetInput).Return(testError).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("SetInputFromGraphQL", *gqlSetInput).Return(*modelSetInput, nil).Once()
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when Instance Auth retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("SetAuth", txtest.CtxWithDBMatcher(), testID, *modelSetInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("SetInputFromGraphQL", *gqlSetInput).Return(*modelSetInput, nil).Once()
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			bndlConverter := testCase.BundleConverterFn()

			resolver := bundleinstanceauth.NewResolver(transact, svc, nil, converter, bndlConverter)

			// when
			result, err := resolver.SetBundleInstanceAuth(context.TODO(), testAuthID, *gqlSetInput)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, persist, transact, svc, converter)
		})
	}
}

func TestResolver_RequestBundleInstanceAuthDeletion(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelInstanceAuth := fixSimpleModelBundleInstanceAuth(id)
	gqlInstanceAuth := fixSimpleGQLBundleInstanceAuth(id)

	modelBndl := &model.Bundle{DefaultInstanceAuth: fixModelAuth()}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.Service
		BundleServiceFn   func() *automock.BundleService
		ConverterFn       func() *automock.Converter
		BundleConverterFn func() *pkgmock.BundleConverter
		ExpectedResult    *graphql.BundleInstanceAuth
		ExpectedErr       error
	}{
		{
			Name:            "Success - Deleted",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Once()
				svc.On("RequestDeletion", txtest.CtxWithDBMatcher(), modelInstanceAuth, modelBndl.DefaultInstanceAuth).Return(true, nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(modelBndl, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("ToGraphQL", modelInstanceAuth).Return(gqlInstanceAuth, nil).Once()
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: gqlInstanceAuth,
			ExpectedErr:    nil,
		},
		{
			Name:            "Success - Not Deleted",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Twice()
				svc.On("RequestDeletion", txtest.CtxWithDBMatcher(), modelInstanceAuth, modelBndl.DefaultInstanceAuth).Return(false, nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(modelBndl, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("ToGraphQL", modelInstanceAuth).Return(gqlInstanceAuth, nil).Once()
				return conv
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: gqlInstanceAuth,
			ExpectedErr:    nil,
		},
		{
			Name:            "Error - Get Instance Auth",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				return &automock.BundleService{}
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error - Get Bundle",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error - Request Deletion",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Once()
				svc.On("RequestDeletion", txtest.CtxWithDBMatcher(), modelInstanceAuth, modelBndl.DefaultInstanceAuth).Return(false, testErr).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(modelBndl, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error - Get After Setting Status",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				svc.On("RequestDeletion", txtest.CtxWithDBMatcher(), modelInstanceAuth, modelBndl.DefaultInstanceAuth).Return(false, nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(modelBndl, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error - Transaction Begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.Service {
				return &automock.Service{}
			},
			BundleServiceFn: func() *automock.BundleService {
				return &automock.BundleService{}
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error - Transaction Commit",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Once()
				svc.On("RequestDeletion", txtest.CtxWithDBMatcher(), modelInstanceAuth, modelBndl.DefaultInstanceAuth).Return(true, nil).Once()
				return svc
			},
			BundleServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(modelBndl, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
			},
			BundleConverterFn: func() *pkgmock.BundleConverter {
				return &pkgmock.BundleConverter{}
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
			bundleSvc := testCase.BundleServiceFn()
			converter := testCase.ConverterFn()
			bndlConverter := testCase.BundleConverterFn()

			resolver := bundleinstanceauth.NewResolver(transact, svc, bundleSvc, converter, bndlConverter)

			// when
			result, err := resolver.RequestBundleInstanceAuthDeletion(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persist, bundleSvc)
		})
	}
}
