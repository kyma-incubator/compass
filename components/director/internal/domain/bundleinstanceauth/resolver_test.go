package bundleinstanceauth_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"

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
	// GIVEN
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
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			bndlConverter := testCase.BundleConverterFn()

			resolver := bundleinstanceauth.NewResolver(transact, svc, nil, converter, bndlConverter)

			// WHEN
			result, err := resolver.DeleteBundleInstanceAuth(context.TODO(), id)

			// THEN
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
	// GIVEN
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
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			bndlSvc := testCase.BndlServiceFn()
			converter := testCase.ConverterFn()
			bndlConverter := testCase.BundleConverterFn()

			resolver := bundleinstanceauth.NewResolver(transact, svc, bndlSvc, converter, bndlConverter)

			result, err := resolver.RequestBundleInstanceAuthCreation(context.TODO(), testBundleID, *gqlRequestInput)

			// THEN
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, persist, transact, svc, bndlSvc, converter)
		})
	}
}

func TestResolver_CreateBundleInstanceAuth(t *testing.T) {
	// GIVEN
	modelBundle := fixModelBundle(testBundleID, nil, nil)
	gqlCreateInput := *fixGQLCreateInput()
	modelCreateInput := *fixModelCreateInput()

	modelInstanceAuth := fixModelBundleInstanceAuthWithoutContextAndInputParams(testID, testBundleID, testTenant, nil, nil, &testRuntimeID)
	gqlInstanceAuth := fixGQLBundleInstanceAuthWithoutContextAndInputParams(testID, nil, nil, &testRuntimeID)

	txGen := txtest.NewTransactionContextGenerator(testError)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.Service
		BndlServiceFn   func() *automock.BundleService
		ConverterFn     func() *automock.Converter
		ExpectedResult  *graphql.BundleInstanceAuth
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("CreateBundleInstanceAuth", txtest.CtxWithDBMatcher(), testBundleID, modelCreateInput, modelInstanceAuth.InputParams).Return(testID, nil).Once()
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
				conv.On("CreateInputFromGraphQL", gqlCreateInput).Return(modelCreateInput, nil).Once()
				conv.On("ToGraphQL", modelInstanceAuth).Return(gqlInstanceAuth, nil).Once()
				return conv
			},
			ExpectedResult: gqlInstanceAuth,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ExpectedErr:     testError,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("CreateBundleInstanceAuth", txtest.CtxWithDBMatcher(), testBundleID, modelCreateInput, modelInstanceAuth.InputParams).Return(testID, nil).Once()
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
				conv.On("CreateInputFromGraphQL", gqlCreateInput).Return(modelCreateInput, nil).Once()
				return conv
			},
			ExpectedErr: testError,
		},
		{
			Name:            "Returns error when Instance Auth retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("CreateBundleInstanceAuth", txtest.CtxWithDBMatcher(), testBundleID, modelCreateInput, modelInstanceAuth.InputParams).Return(testID, nil).Once()
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
				conv.On("CreateInputFromGraphQL", gqlCreateInput).Return(modelCreateInput, nil).Once()
				return conv
			},
			ExpectedErr: testError,
		},
		{
			Name:            "Returns error when Instance Auth creation failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("CreateBundleInstanceAuth", txtest.CtxWithDBMatcher(), testBundleID, modelCreateInput, modelInstanceAuth.InputParams).Return("", testError).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("CreateInputFromGraphQL", gqlCreateInput).Return(modelCreateInput, nil).Once()
				return conv
			},
			ExpectedErr: testError,
		},
		{
			Name:            "Returns error when converting input to graphql failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("CreateInputFromGraphQL", gqlCreateInput).Return(model.BundleInstanceAuthCreateInput{}, testError).Once()
				return conv
			},
			ExpectedErr: testError,
		},
		{
			Name:            "Returns error when Bundle retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(nil, testError).Once()
				return svc
			},
			ExpectedErr: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := &automock.Service{}
			if testCase.ServiceFn != nil {
				svc = testCase.ServiceFn()
			}
			bndlSvc := &automock.BundleService{}
			if testCase.BndlServiceFn != nil {
				bndlSvc = testCase.BndlServiceFn()
			}
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}

			resolver := bundleinstanceauth.NewResolver(transact, svc, bndlSvc, converter, nil)

			result, err := resolver.CreateBundleInstanceAuth(context.TODO(), testBundleID, gqlCreateInput)

			// THEN
			if testCase.ExpectedErr == nil {
				require.Equal(t, testCase.ExpectedResult, result)
				require.Nil(t, err)
			} else {
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				require.Nil(t, result)
			}
			mock.AssertExpectationsForObjects(t, persist, transact, svc, bndlSvc, converter)
		})
	}
}

func TestResolver_UpdateBundleInstanceAuth(t *testing.T) {
	// GIVEN
	modelBundle := fixModelBundle(testBundleID, nil, nil)
	gqlUpdateInput := *fixGQLUpdateInput()
	modelUpdateInput := *fixModelUpdateInput()

	modelBundleWithSchema := fixModelBundle(testBundleID, str.Ptr("{\"type\": \"string\"}"), nil)
	invalidInputParams := graphql.JSON(`"{"`)
	gqlUpdateInputWithInvalidParams := graphql.BundleInstanceAuthUpdateInput{
		Context:     gqlUpdateInput.Context,
		InputParams: &invalidInputParams,
		Auth:        gqlUpdateInput.Auth,
	}
	modelUpdateInputWithInvalidParams := model.BundleInstanceAuthUpdateInput{
		Context:     modelUpdateInput.Context,
		InputParams: str.Ptr("{"),
		Auth:        modelUpdateInput.Auth,
	}

	modelInstanceAuth := fixModelBundleInstanceAuthWithoutContextAndInputParams(testID, testBundleID, testTenant, nil, nil, &testRuntimeID)
	updatedModelInstanceAuth := fixModelBundleInstanceAuthWithoutContextAndInputParams(testID, testBundleID, testTenant, nil, nil, &testRuntimeID)
	updatedModelInstanceAuth.Context = modelUpdateInput.Context
	updatedModelInstanceAuth.InputParams = modelUpdateInput.InputParams
	updatedModelInstanceAuth.Auth = modelUpdateInput.Auth.ToAuth()

	gqlInstanceAuth := fixGQLBundleInstanceAuthWithoutContextAndInputParams(testID, nil, nil, &testRuntimeID)

	txGen := txtest.NewTransactionContextGenerator(testError)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.Service
		BndlServiceFn   func() *automock.BundleService
		ConverterFn     func() *automock.Converter
		Input           graphql.BundleInstanceAuthUpdateInput
		ExpectedResult  *graphql.BundleInstanceAuth
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Update", txtest.CtxWithDBMatcher(), updatedModelInstanceAuth).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelInstanceAuth, nil).Twice()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("UpdateInputFromGraphQL", gqlUpdateInput).Return(modelUpdateInput, nil).Once()
				conv.On("ToGraphQL", modelInstanceAuth).Return(gqlInstanceAuth, nil).Once()
				return conv
			},
			Input:          gqlUpdateInput,
			ExpectedResult: gqlInstanceAuth,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			Input:           gqlUpdateInput,
			ExpectedErr:     testError,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Update", txtest.CtxWithDBMatcher(), updatedModelInstanceAuth).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelInstanceAuth, nil).Twice()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("UpdateInputFromGraphQL", gqlUpdateInput).Return(modelUpdateInput, nil).Once()
				return conv
			},
			Input:       gqlUpdateInput,
			ExpectedErr: testError,
		},
		{
			Name:            "Returns error when Instance Auth retrieval after update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Update", txtest.CtxWithDBMatcher(), updatedModelInstanceAuth).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelInstanceAuth, nil).Once()
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
				conv.On("UpdateInputFromGraphQL", gqlUpdateInput).Return(modelUpdateInput, nil).Once()
				return conv
			},
			Input:       gqlUpdateInput,
			ExpectedErr: testError,
		},
		{
			Name:            "Returns error when input params are not valid",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelInstanceAuth, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBundleWithSchema, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("UpdateInputFromGraphQL", gqlUpdateInputWithInvalidParams).Return(modelUpdateInputWithInvalidParams, nil).Once()
				return conv
			},
			Input:       gqlUpdateInputWithInvalidParams,
			ExpectedErr: errors.New("while validating BundleInstanceAuth"),
		},
		{
			Name:            "Returns error when Instance Auth update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("Update", txtest.CtxWithDBMatcher(), updatedModelInstanceAuth).Return(testError).Once()
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
				conv.On("UpdateInputFromGraphQL", gqlUpdateInput).Return(modelUpdateInput, nil).Once()
				return conv
			},
			Input:       gqlUpdateInput,
			ExpectedErr: testError,
		},
		{
			Name:            "Returns error when Instance Auth retrieval before update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
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
				conv.On("UpdateInputFromGraphQL", gqlUpdateInput).Return(modelUpdateInput, nil).Once()
				return conv
			},
			Input:       gqlUpdateInput,
			ExpectedErr: testError,
		},
		{
			Name:            "Returns error when converting input to graphql failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("UpdateInputFromGraphQL", gqlUpdateInput).Return(model.BundleInstanceAuthUpdateInput{}, testError).Once()
				return conv
			},
			Input:       gqlUpdateInput,
			ExpectedErr: testError,
		},
		{
			Name:            "Returns error when Bundle retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			BndlServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(nil, testError).Once()
				return svc
			},
			Input:       gqlUpdateInput,
			ExpectedErr: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := &automock.Service{}
			if testCase.ServiceFn != nil {
				svc = testCase.ServiceFn()
			}
			bndlSvc := &automock.BundleService{}
			if testCase.BndlServiceFn != nil {
				bndlSvc = testCase.BndlServiceFn()
			}
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}

			resolver := bundleinstanceauth.NewResolver(transact, svc, bndlSvc, converter, nil)

			result, err := resolver.UpdateBundleInstanceAuth(context.TODO(), testID, testBundleID, testCase.Input)

			// THEN
			if testCase.ExpectedErr == nil {
				require.Equal(t, testCase.ExpectedResult, result)
				require.Nil(t, err)
			} else {
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				require.Nil(t, result)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, svc, bndlSvc, converter)
		})
	}
}

func TestResolver_SetBundleInstanceAuth(t *testing.T) {
	// GIVEN

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
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			bndlConverter := testCase.BundleConverterFn()

			resolver := bundleinstanceauth.NewResolver(transact, svc, nil, converter, bndlConverter)

			// WHEN
			result, err := resolver.SetBundleInstanceAuth(context.TODO(), testAuthID, *gqlSetInput)

			// THEN
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, persist, transact, svc, converter)
		})
	}
}

func TestResolver_RequestBundleInstanceAuthDeletion(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "bar"
	modelInstanceAuth := fixSimpleModelBundleInstanceAuth(id)
	gqlInstanceAuth := fixSimpleGQLBundleInstanceAuth(id)

	modelBndl := &model.Bundle{
		DefaultInstanceAuth: fixModelAuth(),
		BaseEntity: &model.BaseEntity{
			ID: testBundleID,
		},
	}

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
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBndl, nil).Once()
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
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBndl, nil).Once()
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
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(nil, testErr).Once()
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
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBndl, nil).Once()
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
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBndl, nil).Once()
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
				svc.On("Get", txtest.CtxWithDBMatcher(), testBundleID).Return(modelBndl, nil).Once()
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
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			bundleSvc := testCase.BundleServiceFn()
			converter := testCase.ConverterFn()
			bndlConverter := testCase.BundleConverterFn()

			resolver := bundleinstanceauth.NewResolver(transact, svc, bundleSvc, converter, bndlConverter)

			// WHEN
			result, err := resolver.RequestBundleInstanceAuthDeletion(context.TODO(), id)

			// THEN
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persist, bundleSvc)
		})
	}
}
