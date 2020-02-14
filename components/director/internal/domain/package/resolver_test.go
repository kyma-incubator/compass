package mp_package_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"

	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/package/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
)

func TestResolver_AddPackage(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	appId := "1"
	desc := "bar"
	name := "baz"

	modelPackage := fixPackageModel(t, name, desc)
	gqlPackage := fixGQLPackage(id, name, desc)
	gqlPackageInput := fixGQLPackageCreateInput(name, desc)
	modelPackageInput := fixModelPackageCreateInput(t, name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.PackageService
		ConverterFn     func() *automock.PackageConverter
		ExpectedPackage *graphql.Package
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelPackageInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(modelPackageInput, nil).Once()
				conv.On("ToGraphQL", modelPackage).Return(gqlPackage, nil).Once()
				return conv
			},
			ExpectedPackage: gqlPackage,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when converting input from GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(model.PackageCreateInput{}, testErr).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when adding Package failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelPackageInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(modelPackageInput, nil).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when Package retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelPackageInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(modelPackageInput, nil).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelPackageInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(modelPackageInput, nil).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelPackageInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(modelPackageInput, nil).Once()
				conv.On("ToGraphQL", modelPackage).Return(nil, testErr).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_package.NewResolver(transact, svc, converter)

			// when
			result, err := resolver.AddPackage(context.TODO(), appId, gqlPackageInput)

			// then
			assert.Equal(t, testCase.ExpectedPackage, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "id"
	name := "foo"
	desc := "bar"
	gqlPackageUpdateInput := fixGQLPackageUpdateInput(name, desc)
	modelPackageUpdateInput := fixModelPackageUpdateInput(t, name, desc)
	gqlPackage := fixGQLPackage(id, name, desc)
	modelPackage := fixPackageModel(t, name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.PackageService
		ConverterFn     func() *automock.PackageConverter
		InputPackage    graphql.PackageUpdateInput
		ExpectedPackage *graphql.Package
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelPackageUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(modelPackageUpdateInput, nil).Once()
				conv.On("ToGraphQL", modelPackage).Return(gqlPackage, nil).Once()
				return conv
			},
			InputPackage:    gqlPackageUpdateInput,
			ExpectedPackage: gqlPackage,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			InputPackage:    gqlPackageUpdateInput,
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when converting from GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(model.PackageUpdateInput{}, testErr).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when Package update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelPackageUpdateInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(modelPackageUpdateInput, nil).Once()
				return conv
			},
			InputPackage:    gqlPackageUpdateInput,
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when Package retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelPackageUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(modelPackageUpdateInput, nil).Once()
				return conv
			},
			InputPackage:    gqlPackageUpdateInput,
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelPackageUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(modelPackageUpdateInput, nil).Once()
				return conv
			},
			InputPackage:    gqlPackageUpdateInput,
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelPackageUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(modelPackageUpdateInput, nil).Once()
				conv.On("ToGraphQL", modelPackage).Return(nil, testErr).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_package.NewResolver(transact, svc, converter)

			// when
			result, err := resolver.UpdatePackage(context.TODO(), id, gqlPackageUpdateInput)

			// then
			assert.Equal(t, testCase.ExpectedPackage, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeletePackage(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "id"
	name := "foo"
	desc := "desc"
	modelPackage := fixPackageModel(t, name, desc)
	gqlPackage := fixGQLPackage(id, name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.PackageService
		ConverterFn     func() *automock.PackageConverter
		ExpectedPackage *graphql.Package
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("ToGraphQL", modelPackage).Return(gqlPackage, nil).Once()
				return conv
			},
			ExpectedPackage: gqlPackage,
			ExpectedErr:     nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when Package retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when Package deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("ToGraphQL", modelPackage).Return(nil, testErr).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_package.NewResolver(transact, svc, converter)

			// when
			result, err := resolver.DeletePackage(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedPackage, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transact.AssertExpectations(t)
			persist.AssertExpectations(t)
		})
	}
}
