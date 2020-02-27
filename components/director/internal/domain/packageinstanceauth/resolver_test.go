package packageinstanceauth_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResolver_DeletePackageInstanceAuth(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelInstanceAuth := fixSimpleModelPackageInstanceAuth(id)
	gqlInstanceAuth := fixSimpleGQLPackageInstanceAuth(id)

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

			resolver := packageinstanceauth.NewResolver(transact, svc, nil, converter)

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

func TestResolver_RequestPackageInstanceAuthCreation(t *testing.T) {
	// given
	modelPackage := fixModelPackage(testPackageID, nil, nil)
	gqlRequestInput := fixGQLRequestInput()
	modelRequestInput := fixModelRequestInput()

	modelInstanceAuth := fixModelPackageInstanceAuthWithoutContextAndInputParams(testID, testPackageID, testTenant, nil, nil)
	gqlInstanceAuth := fixGQLPackageInstanceAuthWithoutContextAndInputParams(testID, nil, nil)

	txGen := txtest.NewTransactionContextGenerator(testError)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.Service
		PkgServiceFn    func() *automock.PackageService
		ConverterFn     func() *automock.Converter
		ExpectedResult  *graphql.PackageInstanceAuth
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Create", txtest.CtxWithDBMatcher(), testPackageID, *modelRequestInput, modelInstanceAuth.Auth, modelInstanceAuth.InputParams).Return(testID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelInstanceAuth, nil).Once()
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testPackageID).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("RequestInputFromGraphQL", *gqlRequestInput).Return(*modelRequestInput).Once()
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
			PkgServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Create", txtest.CtxWithDBMatcher(), testPackageID, *modelRequestInput, modelInstanceAuth.Auth, modelInstanceAuth.InputParams).Return(testID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelInstanceAuth, nil).Once()
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testPackageID).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("RequestInputFromGraphQL", *gqlRequestInput).Return(*modelRequestInput).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when Package retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testPackageID).Return(nil, testError).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when Instance Auth creation failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Create", txtest.CtxWithDBMatcher(), testPackageID, *modelRequestInput, modelInstanceAuth.Auth, modelInstanceAuth.InputParams).Return("", testError).Once()
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testPackageID).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("RequestInputFromGraphQL", *gqlRequestInput).Return(*modelRequestInput).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when Instance Auth retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Create", txtest.CtxWithDBMatcher(), testPackageID, *modelRequestInput, modelInstanceAuth.Auth, modelInstanceAuth.InputParams).Return(testID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testPackageID).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("RequestInputFromGraphQL", *gqlRequestInput).Return(*modelRequestInput).Once()
				return conv
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
			pkgSvc := testCase.PkgServiceFn()
			converter := testCase.ConverterFn()

			resolver := packageinstanceauth.NewResolver(transact, svc, pkgSvc, converter)

			// when
			result, err := resolver.RequestPackageInstanceAuthCreation(context.Background(), testPackageID, *gqlRequestInput)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, persist, transact, svc, pkgSvc, converter)
		})
	}
}

func TestResolver_SetPackageInstanceAuth(t *testing.T) {
	// given

	testAuthID := "foo"

	modelSetInput := fixModelSetInput()
	gqlSetInput := fixGQLSetInput()

	modelInstanceAuth := fixModelPackageInstanceAuthWithoutContextAndInputParams(testID, testPackageID, testTenant, nil, nil)
	gqlInstanceAuth := fixGQLPackageInstanceAuthWithoutContextAndInputParams(testID, nil, nil)

	txGen := txtest.NewTransactionContextGenerator(testError)

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
				svc.On("SetAuth", txtest.CtxWithDBMatcher(), testID, *modelSetInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelInstanceAuth, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("SetInputFromGraphQL", *gqlSetInput).Return(*modelSetInput).Once()
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
				conv.On("SetInputFromGraphQL", *gqlSetInput).Return(*modelSetInput).Once()
				return conv
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
				conv.On("SetInputFromGraphQL", *gqlSetInput).Return(*modelSetInput).Once()
				return conv
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
				conv.On("SetInputFromGraphQL", *gqlSetInput).Return(*modelSetInput).Once()
				return conv
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

			resolver := packageinstanceauth.NewResolver(transact, svc, nil, converter)

			// when
			result, err := resolver.SetPackageInstanceAuth(context.Background(), testAuthID, *gqlSetInput)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, persist, transact, svc, converter)
		})
	}
}

func TestResolver_RequestPackageInstanceAuthDeletion(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelInstanceAuth := fixSimpleModelPackageInstanceAuth(id)
	gqlInstanceAuth := fixSimpleGQLPackageInstanceAuth(id)

	modelPkg := &model.Package{DefaultInstanceAuth: fixModelAuth()}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name             string
		TransactionerFn  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn        func() *automock.Service
		PackageServiceFn func() *automock.PackageService
		ConverterFn      func() *automock.Converter
		ExpectedResult   *graphql.PackageInstanceAuth
		ExpectedErr      error
	}{
		{
			Name:            "Success - Deleted",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Once()
				svc.On("RequestDeletion", txtest.CtxWithDBMatcher(), modelInstanceAuth, modelPkg.DefaultInstanceAuth).Return(true, nil).Once()
				return svc
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(modelPkg, nil).Once()
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
			Name:            "Success - Not Deleted",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Twice()
				svc.On("RequestDeletion", txtest.CtxWithDBMatcher(), modelInstanceAuth, modelPkg.DefaultInstanceAuth).Return(false, nil).Once()
				return svc
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(modelPkg, nil).Once()
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
			Name:            "Error - Get Instance Auth",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Error - Get Package",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelInstanceAuth, nil).Once()
				return svc
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
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
				svc.On("RequestDeletion", txtest.CtxWithDBMatcher(), modelInstanceAuth, modelPkg.DefaultInstanceAuth).Return(false, testErr).Once()
				return svc
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(modelPkg, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
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
				svc.On("RequestDeletion", txtest.CtxWithDBMatcher(), modelInstanceAuth, modelPkg.DefaultInstanceAuth).Return(false, nil).Once()
				return svc
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(modelPkg, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
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
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
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
				svc.On("RequestDeletion", txtest.CtxWithDBMatcher(), modelInstanceAuth, modelPkg.DefaultInstanceAuth).Return(true, nil).Once()
				return svc
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("GetByInstanceAuthID", txtest.CtxWithDBMatcher(), id).Return(modelPkg, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				return &automock.Converter{}
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
			packageSvc := testCase.PackageServiceFn()
			converter := testCase.ConverterFn()

			resolver := packageinstanceauth.NewResolver(transact, svc, packageSvc, converter)

			// when
			result, err := resolver.RequestPackageInstanceAuthDeletion(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persist, packageSvc)
		})
	}
}
