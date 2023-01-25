package certsubjectmapping_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping"
	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
)

var (
	emptyCtx        = context.Background()
	txGen           = txtest.NewTransactionContextGenerator(testErr)
	testErr         = errors.New("test error")
	invalidInputErr = errors.New("subject: cannot be blank")
)

func TestResolver_CertificateSubjectMapping(t *testing.T) {
	testCases := []struct {
		Name                    string
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ConverterFn             func() *automock.Converter
		CertSubjectMappingSvcFn func() *automock.CertSubjectMappingService
		ExpectedOutput          *graphql.CertificateSubjectMapping
		ExpectedError           error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("ToGraphQL", CertSubjectMappingModel).Return(CertSubjectMappingGQLModel).Once()
				return conv
			},
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(CertSubjectMappingModel, nil).Once()
				return certSubjectMappingSvc
			},
			ExpectedOutput: CertSubjectMappingGQLModel,
		},
		{
			Name:            "Error when transaction fails to begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ExpectedOutput:  nil,
			ExpectedError:   testErr,
		},
		{
			Name:            "Error when getting certificate subject mapping fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(nil, testErr).Once()
				return certSubjectMappingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when committing the transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(CertSubjectMappingModel, nil).Once()
				return certSubjectMappingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := fixUnusedTransactioner()
			if testCase.TransactionerFn != nil {
				persist, transact = testCase.TransactionerFn()
			}

			conv := fixUnusedConverter()
			if testCase.ConverterFn != nil {
				conv = testCase.ConverterFn()
			}

			certSubjectMappingSvc := fixUnusedCertSubjectMappingSvc()
			if testCase.CertSubjectMappingSvcFn != nil {
				certSubjectMappingSvc = testCase.CertSubjectMappingSvcFn()
			}

			uidSvc := fixUnusedUIDService()

			defer mock.AssertExpectationsForObjects(t, persist, transact, conv, certSubjectMappingSvc, uidSvc)

			resolver := certsubjectmapping.NewResolver(transact, conv, certSubjectMappingSvc, uidSvc)

			// WHEN
			result, err := resolver.CertificateSubjectMapping(emptyCtx, TestID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}

func TestResolver_CertificateSubjectMappings(t *testing.T) {
	first := 2
	after := "testAfter"
	gqlAfter := graphql.PageCursor(after)

	testCases := []struct {
		Name                    string
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ConverterFn             func() *automock.Converter
		CertSubjectMappingSvcFn func() *automock.CertSubjectMappingService
		First                   *int
		ExpectedOutput          *graphql.CertificateSubjectMappingPage
		ExpectedErrorMsg        string
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("MultipleToGraphQL", CertificateSubjectMappingModelPage.Data).Return(CertificateSubjectMappingsGQL).Once()
				return conv
			},
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(CertificateSubjectMappingModelPage, nil).Once()
				return certSubjectMappingSvc
			},
			First:          &first,
			ExpectedOutput: CertificateSubjectMappingGQLPage,
		},
		{
			Name:             "Error when missing first parameter",
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "Invalid data [reason=missing required parameter 'first']",
		},
		{
			Name:             "Error when transaction fails to begin",
			TransactionerFn:  txGen.ThatFailsOnBegin,
			First:            &first,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:            "Error when getting certificate subject mapping fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(nil, testErr).Once()
				return certSubjectMappingSvc
			},
			First:            &first,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:            "Error when committing the transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(CertificateSubjectMappingModelPage, nil).Once()
				return certSubjectMappingSvc
			},
			First:            &first,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := fixUnusedTransactioner()
			if testCase.TransactionerFn != nil {
				persist, transact = testCase.TransactionerFn()
			}

			conv := fixUnusedConverter()
			if testCase.ConverterFn != nil {
				conv = testCase.ConverterFn()
			}

			certSubjectMappingSvc := fixUnusedCertSubjectMappingSvc()
			if testCase.CertSubjectMappingSvcFn != nil {
				certSubjectMappingSvc = testCase.CertSubjectMappingSvcFn()
			}

			uidSvc := fixUnusedUIDService()

			defer mock.AssertExpectationsForObjects(t, persist, transact, conv, certSubjectMappingSvc, uidSvc)

			resolver := certsubjectmapping.NewResolver(transact, conv, certSubjectMappingSvc, uidSvc)

			// WHEN
			result, err := resolver.CertificateSubjectMappings(emptyCtx, testCase.First, &gqlAfter)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}

func TestResolver_CreateCertificateSubjectMapping(t *testing.T) {
	testCases := []struct {
		Name                    string
		Input                   graphql.CertificateSubjectMappingInput
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ConverterFn             func() *automock.Converter
		CertSubjectMappingSvcFn func() *automock.CertSubjectMappingService
		UIDSvcFn                func() *automock.UIDService
		ExpectedOutput          *graphql.CertificateSubjectMapping
		ExpectedError           error
	}{
		{
			Name:            "Success",
			Input:           CertSubjectMappingGQLModelInput,
			TransactionerFn: txGen.ThatSucceeds,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("FromGraphql", TestID, CertSubjectMappingGQLModelInput).Return(CertSubjectMappingModel).Once()
				conv.On("ToGraphQL", CertSubjectMappingModel).Return(CertSubjectMappingGQLModel).Once()
				return conv
			},
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Create", txtest.CtxWithDBMatcher(), CertSubjectMappingModel).Return(TestID, nil).Once()
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(CertSubjectMappingModel, nil).Once()
				return certSubjectMappingSvc
			},
			UIDSvcFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(TestID).Once()
				return uidSvc
			},
			ExpectedOutput: CertSubjectMappingGQLModel,
		},
		{
			Name:            "Error when transaction fails to begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ExpectedOutput:  nil,
			ExpectedError:   testErr,
		},
		{
			Name:            "Error when certificate subject mapping input validation fails",
			Input:           CertSubjectMappingGQLInvalidInput,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ExpectedOutput:  nil,
			ExpectedError:   invalidInputErr,
		},
		{
			Name:            "Error when creating certificate subject mapping fails",
			Input:           CertSubjectMappingGQLModelInput,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("FromGraphql", TestID, CertSubjectMappingGQLModelInput).Return(CertSubjectMappingModel).Once()
				return conv
			},
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Create", txtest.CtxWithDBMatcher(), CertSubjectMappingModel).Return("", testErr).Once()
				return certSubjectMappingSvc
			},
			UIDSvcFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(TestID).Once()
				return uidSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when getting certificate subject mapping fails",
			Input:           CertSubjectMappingGQLModelInput,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("FromGraphql", TestID, CertSubjectMappingGQLModelInput).Return(CertSubjectMappingModel).Once()
				return conv
			},
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Create", txtest.CtxWithDBMatcher(), CertSubjectMappingModel).Return(TestID, nil).Once()
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(nil, testErr).Once()
				return certSubjectMappingSvc
			},
			UIDSvcFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(TestID).Once()
				return uidSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when committing the transaction fails",
			Input:           CertSubjectMappingGQLModelInput,
			TransactionerFn: txGen.ThatFailsOnCommit,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("FromGraphql", TestID, CertSubjectMappingGQLModelInput).Return(CertSubjectMappingModel).Once()
				return conv
			},
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Create", txtest.CtxWithDBMatcher(), CertSubjectMappingModel).Return(TestID, nil).Once()
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(CertSubjectMappingModel, nil).Once()
				return certSubjectMappingSvc
			},
			UIDSvcFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(TestID).Once()
				return uidSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := fixUnusedTransactioner()
			if testCase.TransactionerFn != nil {
				persist, transact = testCase.TransactionerFn()
			}

			conv := fixUnusedConverter()
			if testCase.ConverterFn != nil {
				conv = testCase.ConverterFn()
			}

			certSubjectMappingSvc := fixUnusedCertSubjectMappingSvc()
			if testCase.CertSubjectMappingSvcFn != nil {
				certSubjectMappingSvc = testCase.CertSubjectMappingSvcFn()
			}

			uidSvc := fixUnusedUIDService()
			if testCase.UIDSvcFn != nil {
				uidSvc = testCase.UIDSvcFn()
			}

			defer mock.AssertExpectationsForObjects(t, persist, transact, conv, certSubjectMappingSvc, uidSvc)

			resolver := certsubjectmapping.NewResolver(transact, conv, certSubjectMappingSvc, uidSvc)

			// WHEN
			result, err := resolver.CreateCertificateSubjectMapping(emptyCtx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}

func TestResolver_UpdateCertificateSubjectMapping(t *testing.T) {
	testCases := []struct {
		Name                    string
		Input                   graphql.CertificateSubjectMappingInput
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ConverterFn             func() *automock.Converter
		CertSubjectMappingSvcFn func() *automock.CertSubjectMappingService
		ExpectedOutput          *graphql.CertificateSubjectMapping
		ExpectedError           error
	}{
		{
			Name:            "Success",
			Input:           CertSubjectMappingGQLModelInput,
			TransactionerFn: txGen.ThatSucceeds,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("FromGraphql", TestID, CertSubjectMappingGQLModelInput).Return(CertSubjectMappingModel).Once()
				conv.On("ToGraphQL", CertSubjectMappingModel).Return(CertSubjectMappingGQLModel).Once()
				return conv
			},
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Update", txtest.CtxWithDBMatcher(), CertSubjectMappingModel).Return(nil).Once()
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(CertSubjectMappingModel, nil).Once()
				return certSubjectMappingSvc
			},
			ExpectedOutput: CertSubjectMappingGQLModel,
		},
		{
			Name:            "Error when transaction fails to begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ExpectedOutput:  nil,
			ExpectedError:   testErr,
		},
		{
			Name:            "Error when certificate subject mapping input validation fails",
			Input:           CertSubjectMappingGQLInvalidInput,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ExpectedOutput:  nil,
			ExpectedError:   invalidInputErr,
		},
		{
			Name:            "Error when updating certificate subject mapping fails",
			Input:           CertSubjectMappingGQLModelInput,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("FromGraphql", TestID, CertSubjectMappingGQLModelInput).Return(CertSubjectMappingModel).Once()
				return conv
			},
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Update", txtest.CtxWithDBMatcher(), CertSubjectMappingModel).Return(testErr).Once()
				return certSubjectMappingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when getting certificate subject mapping fails",
			Input:           CertSubjectMappingGQLModelInput,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("FromGraphql", TestID, CertSubjectMappingGQLModelInput).Return(CertSubjectMappingModel).Once()
				return conv
			},
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Update", txtest.CtxWithDBMatcher(), CertSubjectMappingModel).Return(nil).Once()
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(nil, testErr).Once()
				return certSubjectMappingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when committing the transaction fails",
			Input:           CertSubjectMappingGQLModelInput,
			TransactionerFn: txGen.ThatFailsOnCommit,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("FromGraphql", TestID, CertSubjectMappingGQLModelInput).Return(CertSubjectMappingModel).Once()
				return conv
			},
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Update", txtest.CtxWithDBMatcher(), CertSubjectMappingModel).Return(nil).Once()
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(CertSubjectMappingModel, nil).Once()
				return certSubjectMappingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := fixUnusedTransactioner()
			if testCase.TransactionerFn != nil {
				persist, transact = testCase.TransactionerFn()
			}

			conv := fixUnusedConverter()
			if testCase.ConverterFn != nil {
				conv = testCase.ConverterFn()
			}

			certSubjectMappingSvc := fixUnusedCertSubjectMappingSvc()
			if testCase.CertSubjectMappingSvcFn != nil {
				certSubjectMappingSvc = testCase.CertSubjectMappingSvcFn()
			}

			uidSvc := fixUnusedUIDService()

			defer mock.AssertExpectationsForObjects(t, persist, transact, conv, certSubjectMappingSvc, uidSvc)

			resolver := certsubjectmapping.NewResolver(transact, conv, certSubjectMappingSvc, uidSvc)

			// WHEN
			result, err := resolver.UpdateCertificateSubjectMapping(emptyCtx, TestID, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}

func TestResolver_DeleteCertificateSubjectMapping(t *testing.T) {
	testCases := []struct {
		Name                    string
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ConverterFn             func() *automock.Converter
		CertSubjectMappingSvcFn func() *automock.CertSubjectMappingService
		ExpectedOutput          *graphql.CertificateSubjectMapping
		ExpectedError           error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("ToGraphQL", CertSubjectMappingModel).Return(CertSubjectMappingGQLModel).Once()
				return conv
			},
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(CertSubjectMappingModel, nil).Once()
				certSubjectMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), TestID).Return(nil).Once()
				return certSubjectMappingSvc
			},
			ExpectedOutput: CertSubjectMappingGQLModel,
		},
		{
			Name:            "Error when transaction fails to begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ExpectedOutput:  nil,
			ExpectedError:   testErr,
		},
		{
			Name:            "Error when getting certificate subject mapping fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(nil, testErr).Once()
				return certSubjectMappingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when deleting certificate subject mapping fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(CertSubjectMappingModel, nil).Once()
				certSubjectMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), TestID).Return(testErr).Once()
				return certSubjectMappingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when committing the transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			CertSubjectMappingSvcFn: func() *automock.CertSubjectMappingService {
				certSubjectMappingSvc := &automock.CertSubjectMappingService{}
				certSubjectMappingSvc.On("Get", txtest.CtxWithDBMatcher(), TestID).Return(CertSubjectMappingModel, nil).Once()
				certSubjectMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), TestID).Return(nil).Once()
				return certSubjectMappingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := fixUnusedTransactioner()
			if testCase.TransactionerFn != nil {
				persist, transact = testCase.TransactionerFn()
			}

			conv := fixUnusedConverter()
			if testCase.ConverterFn != nil {
				conv = testCase.ConverterFn()
			}

			certSubjectMappingSvc := fixUnusedCertSubjectMappingSvc()
			if testCase.CertSubjectMappingSvcFn != nil {
				certSubjectMappingSvc = testCase.CertSubjectMappingSvcFn()
			}

			uidSvc := fixUnusedUIDService()

			defer mock.AssertExpectationsForObjects(t, persist, transact, conv, certSubjectMappingSvc, uidSvc)

			resolver := certsubjectmapping.NewResolver(transact, conv, certSubjectMappingSvc, uidSvc)

			// WHEN
			result, err := resolver.DeleteCertificateSubjectMapping(emptyCtx, TestID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}
