package runtime_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var contextParam = mock.MatchedBy(func(ctx context.Context) bool {
	persistenceOp, err := persistence.FromCtx(ctx)
	return err == nil && persistenceOp != nil
})

func TestResolver_CreateRuntime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime("foo", "tenant-foo", "Foo", "Lorem ipsum")
	gqlRuntime := fixGQLRuntime("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter

		Input           graphql.RuntimeInput
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			Input:           gqlInput,
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when runtime creation failed",
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Create", contextParam, modelInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Returns error when runtime retrieval failed",
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.CreateRuntime(context.TODO(), testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateRuntime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime("foo", "tenant-foo", "Foo", "Lorem ipsum")
	gqlRuntime := fixGQLRuntime("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}
	runtimeID := "foo"

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		RuntimeID       string
		Input           graphql.RuntimeInput
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Update", contextParam, runtimeID, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			RuntimeID:       runtimeID,
			Input:           gqlInput,
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when runtime update failed",
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Update", contextParam, runtimeID, modelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			RuntimeID:       runtimeID,
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Returns error when runtime retrieval failed",
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Update", contextParam, runtimeID, modelInput).Return(nil).Once()
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			RuntimeID:       runtimeID,
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.UpdateRuntime(context.TODO(), testCase.RuntimeID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteRuntime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime("foo", "tenant-foo", "Foo", "Bar")
	gqlRuntime := fixGQLRuntime("foo", "Foo", "Bar")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	testAuths := fixOAuths()

	testCases := []struct {
		Name             string
		TransactionerFn  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn        func() *automock.RuntimeService
		SysAuthServiceFn func() *automock.SystemAuthService
		OAuth20ServiceFn func() *automock.OAuth20Service
		ConverterFn      func() *automock.RuntimeConverter
		InputID          string
		ExpectedRuntime  *graphql.Runtime
		ExpectedErr      error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)

				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns error when runtime deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)

				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when runtime retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when transaction starting failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, modelRuntime.ID).Return(nil)
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Return error when listing all auths failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(nil, testErr)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Return error when removing oauth from hydra",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(testErr)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			sysAuthSvc := testCase.SysAuthServiceFn()
			oAuth20Svc := testCase.OAuth20ServiceFn()

			resolver := runtime.NewResolver(transact, svc, sysAuthSvc, oAuth20Svc, converter, nil)

			// when
			result, err := resolver.DeleteRuntime(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			if testCase.ExpectedErr != nil {
				assert.EqualError(t, testCase.ExpectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
			transact.AssertExpectations(t)
			sysAuthSvc.AssertExpectations(t)
			oAuth20Svc.AssertExpectations(t)
		})
	}
}

func TestResolver_Runtime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime("foo", "tenant-foo", "Foo", "Bar")
	gqlRuntime := fixGQLRuntime("foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		InputID         string
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when runtime retrieval failed",
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.Runtime(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_Runtimes(t *testing.T) {
	// given
	modelRuntimes := []*model.Runtime{
		fixModelRuntime("foo", "tenant-foo", "Foo", "Lorem Ipsum"),
		fixModelRuntime("bar", "tenant-bar", "Bar", "Lorem Ipsum"),
	}

	gqlRuntimes := []*graphql.Runtime{
		fixGQLRuntime("foo", "Foo", "Lorem Ipsum"),
		fixGQLRuntime("bar", "Bar", "Lorem Ipsum"),
	}

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	filter := []*labelfilter.LabelFilter{{Key: ""}}
	gqlFilter := []*graphql.LabelFilter{{Key: ""}}
	testErr := errors.New("Test error")

	testCases := []struct {
		Name              string
		PersistenceFn     func() *persistenceautomock.PersistenceTx
		TransactionerFn   func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn         func() *automock.RuntimeService
		ConverterFn       func() *automock.RuntimeConverter
		InputLabelFilters []*graphql.LabelFilter
		InputFirst        *int
		InputAfter        *graphql.PageCursor
		ExpectedResult    *graphql.RuntimePage
		ExpectedErr       error
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("List", contextParam, filter, first, after).Return(fixRuntimePage(modelRuntimes), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("MultipleToGraphQL", modelRuntimes).Return(gqlRuntimes).Once()
				return conv
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    fixGQLRuntimePage(gqlRuntimes),
			ExpectedErr:       nil,
		},
		{
			Name: "Returns error when runtime listing failed",
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("List", contextParam, filter, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    nil,
			ExpectedErr:       testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.Runtimes(context.TODO(), testCase.InputLabelFilters, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_SetRuntimeLabel(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	runtimeID := "foo"
	gqlLabel := &graphql.Label{
		Key:   "key",
		Value: []string{"foo", "bar"},
	}
	modelLabel := &model.LabelInput{
		Key:        "key",
		Value:      []string{"foo", "bar"},
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		InputRuntimeID  string
		InputKey        string
		InputValue      interface{}
		ExpectedLabel   *graphql.Label
		ExpectedErr     error
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("SetLabel", contextParam, modelLabel).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			InputValue:     gqlLabel.Value,
			ExpectedLabel:  gqlLabel,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when adding label to runtime failed",
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("SetLabel", contextParam, modelLabel).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			InputValue:     gqlLabel.Value,
			ExpectedLabel:  nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.SetRuntimeLabel(context.TODO(), testCase.InputRuntimeID, testCase.InputKey, testCase.InputValue)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}

	t.Run("Returns error when Label input validation failed", func(t *testing.T) {
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil)

		// when
		result, err := resolver.SetRuntimeLabel(context.TODO(), "", "", "")

		// then
		require.Nil(t, result)
		require.Error(t, err)
		assert.EqualError(t, err, "validation error for type LabelInput: key: cannot be blank; value: cannot be blank.")
	})
}

func TestResolver_DeleteRuntimeLabel(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	runtimeID := "foo"

	gqlLabel := &graphql.Label{
		Key:   "key",
		Value: []string{"foo", "bar"},
	}
	modelLabel := &model.Label{
		Key:   "key",
		Value: []string{"foo", "bar"},
	}

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		InputRuntimeID  string
		InputKey        string
		ExpectedLabel   *graphql.Label
		ExpectedErr     error
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, gqlLabel.Key).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", contextParam, runtimeID, gqlLabel.Key).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			ExpectedLabel:  gqlLabel,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when label retrieval failed",
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, gqlLabel.Key).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			ExpectedLabel:  nil,
			ExpectedErr:    testErr,
		},
		{
			Name: "Returns error when deleting runtime's label failed",
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, gqlLabel.Key).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", contextParam, runtimeID, gqlLabel.Key).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			ExpectedLabel:  nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.DeleteRuntimeLabel(context.TODO(), testCase.InputRuntimeID, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_Labels(t *testing.T) {
	// given

	id := "foo"
	tenant := "tenant"
	labelKey := "key"
	labelValue := "val"

	gqlRuntime := fixGQLRuntime(id, "name", "desc")

	modelLabels := map[string]*model.Label{
		"abc": {
			ID:         "abc",
			Tenant:     tenant,
			Key:        labelKey,
			Value:      labelValue,
			ObjectID:   id,
			ObjectType: model.ApplicationLabelableObject,
		},
		"def": {
			ID:         "def",
			Tenant:     tenant,
			Key:        labelKey,
			Value:      labelValue,
			ObjectID:   id,
			ObjectType: model.ApplicationLabelableObject,
		},
	}

	gqlLabels := graphql.Labels{
		labelKey: labelValue,
		labelKey: labelValue,
	}

	testErr := errors.New("Test error")

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		InputRuntime    *graphql.Runtime
		InputKey        string
		ExpectedResult  graphql.Labels
		ExpectedErr     error
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListLabels", contextParam, id).Return(modelLabels, nil).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: gqlLabels,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when label listing failed",
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
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListLabels", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			svc := testCase.ServiceFn()
			transact := testCase.TransactionerFn(persistTx)

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, nil)

			// when
			result, err := resolver.Labels(context.TODO(), gqlRuntime, &testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			transact.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_Auths(t *testing.T) {
	// GIVEN
	tnt := "tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	parentRuntime := fixGQLRuntime("foo", "bar", "baz")

	modelSysAuths := []model.SystemAuth{
		fixModelSystemAuth("bar", tnt, parentRuntime.ID, fixModelAuth()),
		fixModelSystemAuth("baz", tnt, parentRuntime.ID, fixModelAuth()),
		fixModelSystemAuth("faz", tnt, parentRuntime.ID, fixModelAuth()),
	}

	gqlSysAuths := []*graphql.SystemAuth{
		fixGQLSystemAuth("bar", fixGQLAuth()),
		fixGQLSystemAuth("baz", fixGQLAuth()),
		fixGQLSystemAuth("faz", fixGQLAuth()),
	}

	testErr := errors.New("this is a test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		SysAuthSvcFn    func() *automock.SystemAuthService
		SysAuthConvFn   func() *automock.SystemAuthConverter
		ExpectedOutput  []*graphql.SystemAuth
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), model.RuntimeReference, parentRuntime.ID).Return(modelSysAuths, nil).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				sysAuthConv.On("ToGraphQL", &modelSysAuths[0]).Return(gqlSysAuths[0]).Once()
				sysAuthConv.On("ToGraphQL", &modelSysAuths[1]).Return(gqlSysAuths[1]).Once()
				sysAuthConv.On("ToGraphQL", &modelSysAuths[2]).Return(gqlSysAuths[2]).Once()
				return sysAuthConv
			},
			ExpectedOutput: gqlSysAuths,
			ExpectedError:  nil,
		},
		{
			Name:            "Error when listing for object",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), model.RuntimeReference, parentRuntime.ID).Return(nil, testErr).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), model.RuntimeReference, parentRuntime.ID).Return(modelSysAuths, nil).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			sysAuthSvc := testCase.SysAuthSvcFn()
			sysAuthConv := testCase.SysAuthConvFn()

			resolver := runtime.NewResolver(transact, nil, sysAuthSvc, nil, nil, sysAuthConv)

			// WHEN
			result, err := resolver.Auths(ctx, parentRuntime)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			sysAuthSvc.AssertExpectations(t)
			sysAuthConv.AssertExpectations(t)
		})
	}

	t.Run("Error when parent object is nil", func(t *testing.T) {
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil)

		// WHEN
		result, err := resolver.Auths(context.TODO(), nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Runtime cannot be empty")
		assert.Nil(t, result)
	})
}

func fixOAuths() []model.SystemAuth {
	return []model.SystemAuth{
		{
			ID:       "foo",
			TenantID: "foo",
			Value: &model.Auth{
				Credential: model.CredentialData{
					Basic: nil,
					Oauth: &model.OAuthCredentialData{
						ClientID:     "foo",
						ClientSecret: "foo",
						URL:          "foo",
					},
				},
			},
		},
		{
			ID:       "bar",
			TenantID: "bar",
			Value:    nil,
		},
		{
			ID:       "test",
			TenantID: "test",
			Value: &model.Auth{
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
						Username: "test",
						Password: "test",
					},
					Oauth: nil,
				},
			},
		},
	}
}
