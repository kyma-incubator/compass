package runtime_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

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

	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name            string
		TransactionerFn func() *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		Input           graphql.RuntimeInput
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
	}{
		{
			Name: "Success",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(modelRuntime, nil).Once()
				svc.On("Create", ctxWithPersistenceTx, modelInput).Return("foo", nil).Once()
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
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Create", ctxWithPersistenceTx, modelInput).Return("", testErr).Once()
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
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Create", ctxWithPersistenceTx, modelInput).Return("foo", nil).Once()
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(nil, testErr).Once()
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
			transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, appCtx, svc, converter)

			// when
			result, err := resolver.CreateRuntime(ctx, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
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

	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name            string
		TransactionerFn func() *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		RuntimeID       string
		Input           graphql.RuntimeInput
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
	}{
		{
			Name: "Success",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(modelRuntime, nil).Once()
				svc.On("Update", ctxWithPersistenceTx, runtimeID, modelInput).Return(nil).Once()
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
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Update", ctxWithPersistenceTx, runtimeID, modelInput).Return(testErr).Once()
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
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Update", ctxWithPersistenceTx, runtimeID, modelInput).Return(nil).Once()
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(nil, testErr).Once()
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
			transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, appCtx, svc, converter)

			// when
			result, err := resolver.UpdateRuntime(context.TODO(), testCase.RuntimeID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteRuntime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime("foo", "tenant-foo", "Foo", "Bar")
	gqlRuntime := fixGQLRuntime("foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name            string
		TransactionerFn func() *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		InputID         string
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
	}{
		{
			Name: "Success",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", ctxWithPersistenceTx, "foo").Return(nil).Once()
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
			Name: "Returns error when runtime deletion failed",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", ctxWithPersistenceTx, "foo").Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Returns error when runtime retrieval failed",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(nil, testErr).Once()
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
			transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, appCtx, svc, converter)

			// when
			result, err := resolver.DeleteRuntime(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Runtime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime("foo", "tenant-foo", "Foo", "Bar")
	gqlRuntime := fixGQLRuntime("foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name            string
		TransactionerFn func() *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		InputID         string
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
	}{
		{
			Name: "Success",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(modelRuntime, nil).Once()

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
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(nil, testErr).Once()

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
			transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, appCtx, svc, converter)

			// when
			result, err := resolver.Runtime(ctx, testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
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

	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name              string
		TransactionerFn   func() *persistenceautomock.Transactioner
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
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("List", ctxWithPersistenceTx, filter, &first, &after).Return(fixRuntimePage(modelRuntimes), nil).Once()
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
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("List", ctxWithPersistenceTx, filter, &first, &after).Return(nil, testErr).Once()
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
			transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, appCtx, svc, converter)

			// when
			result, err := resolver.Runtimes(ctx, testCase.InputLabelFilters, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
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
	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name            string
		TransactionerFn func() *persistenceautomock.Transactioner
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
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("SetLabel", ctxWithPersistenceTx, modelLabel).Return(nil).Once()
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
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("SetLabel", ctxWithPersistenceTx, modelLabel).Return(testErr).Once()
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
			transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, appCtx, svc, converter)

			// when
			result, err := resolver.SetRuntimeLabel(ctx, testCase.InputRuntimeID, testCase.InputKey, testCase.InputValue)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
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

	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name            string
		TransactionerFn func() *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		InputRuntimeID  string
		InputKey        string
		ExpectedLabel   *graphql.Label
		ExpectedErr     error
	}{
		{
			Name: "Success",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", ctxWithPersistenceTx, runtimeID, gqlLabel.Key).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", ctxWithPersistenceTx, runtimeID, gqlLabel.Key).Return(nil).Once()
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
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", ctxWithPersistenceTx, runtimeID, gqlLabel.Key).Return(nil, testErr).Once()
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
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", ctxWithPersistenceTx, runtimeID, gqlLabel.Key).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", ctxWithPersistenceTx, runtimeID, gqlLabel.Key).Return(testErr).Once()
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
			transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, appCtx, svc, converter)

			// when
			result, err := resolver.DeleteRuntimeLabel(ctx, testCase.InputRuntimeID, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}
