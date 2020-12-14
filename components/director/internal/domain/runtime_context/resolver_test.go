package runtime_context_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var contextParam = mock.MatchedBy(func(ctx context.Context) bool {
	persistenceOp, err := persistence.FromCtx(ctx)
	return err == nil && persistenceOp != nil
})

func TestResolver_CreateRuntimeContext(t *testing.T) {
	// given
	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"
	tenant := "tenant"

	modelRuntimeContext := &model.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Tenant:    tenant,
		Key:       key,
		Value:     val,
	}
	gqlRuntimeContext := &graphql.RuntimeContext{
		ID:    id,
		Key:   key,
		Value: val,
	}
	testErr := errors.New("Test error")

	gqlInput := graphql.RuntimeContextInput{
		Key:   key,
		Value: val,
	}
	modelInput := model.RuntimeContextInput{
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeContextService
		ConverterFn     func() *automock.RuntimeContextConverter

		Input                  graphql.RuntimeContextInput
		ExpectedRuntimeContext *graphql.RuntimeContext
		ExpectedErr            error
		Consumer               *consumer.Consumer
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
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput, runtimeID).Return(modelInput).Once()
				conv.On("ToGraphQL", modelRuntimeContext).Return(gqlRuntimeContext).Once()
				return conv
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: gqlRuntimeContext,
			ExpectedErr:            nil,
		},
		{
			Name: "Returns error when consumer type is application",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				return &persistenceautomock.Transactioner{}
			},
			ServiceFn: func() *automock.RuntimeContextService {
				return &automock.RuntimeContextService{}
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            apperrors.NewUnauthorizedError("runtime context access is restricted to runtimes only"),
			Consumer: &consumer.Consumer{
				ConsumerID:   runtimeID,
				ConsumerType: consumer.Application,
			},
		},
		{
			Name: "Returns error when consumer type is integration system",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				return &persistenceautomock.Transactioner{}
			},
			ServiceFn: func() *automock.RuntimeContextService {
				return &automock.RuntimeContextService{}
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            apperrors.NewUnauthorizedError("runtime context access is restricted to runtimes only"),
			Consumer: &consumer.Consumer{
				ConsumerID:   runtimeID,
				ConsumerType: consumer.IntegrationSystem,
			},
		},
		{
			Name: "Returns error when runtime context creation failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Create", contextParam, modelInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput, runtimeID).Return(modelInput).Once()
				return conv
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name: "Returns error when runtime context retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput, runtimeID).Return(modelInput).Once()
				return conv
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime_context.NewResolver(transact, svc, converter)

			c := testCase.Consumer
			if c == nil {
				c = &consumer.Consumer{
					ConsumerID:   runtimeID,
					ConsumerType: consumer.Runtime,
				}
			}
			ctx := consumer.SaveToContext(context.TODO(), *c)

			// when
			result, err := resolver.RegisterRuntimeContext(ctx, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeContext, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_UpdateRuntimeContext(t *testing.T) {
	// given
	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"
	tenant := "tenant"

	modelRuntimeContext := &model.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Tenant:    tenant,
		Key:       key,
		Value:     val,
	}
	gqlRuntimeContext := &graphql.RuntimeContext{
		ID:    id,
		Key:   key,
		Value: val,
	}
	testErr := errors.New("Test error")

	gqlInput := graphql.RuntimeContextInput{
		Key:   key,
		Value: val,
	}
	modelInput := model.RuntimeContextInput{
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	testCases := []struct {
		Name                   string
		PersistenceFn          func() *persistenceautomock.PersistenceTx
		TransactionerFn        func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn              func() *automock.RuntimeContextService
		ConverterFn            func() *automock.RuntimeContextConverter
		RuntimeContextID       string
		Input                  graphql.RuntimeContextInput
		ExpectedRuntimeContext *graphql.RuntimeContext
		ExpectedErr            error
		Consumer               *consumer.Consumer
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
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Update", contextParam, id, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput, runtimeID).Return(modelInput).Once()
				conv.On("ToGraphQL", modelRuntimeContext).Return(gqlRuntimeContext).Once()
				return conv
			},
			RuntimeContextID:       id,
			Input:                  gqlInput,
			ExpectedRuntimeContext: gqlRuntimeContext,
			ExpectedErr:            nil,
		},
		{
			Name: "Returns error when consumer id is different from owner id",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Update", contextParam, id, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput, "different-runtime-id").Return(modelInput).Once()
				return conv
			},
			Input:                  gqlInput,
			RuntimeContextID:       id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            apperrors.NewUnauthorizedError("runtime context not accessible"),
			Consumer: &consumer.Consumer{
				ConsumerID:   "different-runtime-id",
				ConsumerType: consumer.Runtime,
			},
		},
		{
			Name: "Returns error when consumer type is application",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				return &persistenceautomock.Transactioner{}
			},
			ServiceFn: func() *automock.RuntimeContextService {
				return &automock.RuntimeContextService{}
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			Input:                  gqlInput,
			RuntimeContextID:       id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            apperrors.NewUnauthorizedError("runtime context access is restricted to runtimes only"),
			Consumer: &consumer.Consumer{
				ConsumerID:   runtimeID,
				ConsumerType: consumer.Application,
			},
		},
		{
			Name: "Returns error when consumer type is integration system",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				return &persistenceautomock.Transactioner{}
			},
			ServiceFn: func() *automock.RuntimeContextService {
				return &automock.RuntimeContextService{}
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            apperrors.NewUnauthorizedError("runtime context access is restricted to runtimes only"),
			Consumer: &consumer.Consumer{
				ConsumerID:   runtimeID,
				ConsumerType: consumer.IntegrationSystem,
			},
		},
		{
			Name: "Returns error when runtime context update failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Update", contextParam, id, modelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput, runtimeID).Return(modelInput).Once()
				return conv
			},
			RuntimeContextID:       id,
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name: "Returns error when runtime context retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Update", contextParam, id, modelInput).Return(nil).Once()
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput, runtimeID).Return(modelInput).Once()
				return conv
			},
			RuntimeContextID:       id,
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime_context.NewResolver(transact, svc, converter)

			c := testCase.Consumer
			if c == nil {
				c = &consumer.Consumer{
					ConsumerID:   runtimeID,
					ConsumerType: consumer.Runtime,
				}
			}
			ctx := consumer.SaveToContext(context.TODO(), *c)

			// when
			result, err := resolver.UpdateRuntimeContext(ctx, testCase.RuntimeContextID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeContext, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_DeleteRuntimeContext(t *testing.T) {
	// given
	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"
	tenant := "tenant"

	modelRuntimeContext := &model.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Tenant:    tenant,
		Key:       key,
		Value:     val,
	}
	gqlRuntimeContext := &graphql.RuntimeContext{
		ID:    id,
		Key:   key,
		Value: val,
	}
	testErr := errors.New("Test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                   string
		TransactionerFn        func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn              func() *automock.RuntimeContextService
		ConverterFn            func() *automock.RuntimeContextConverter
		InputID                string
		ExpectedRuntimeContext *graphql.RuntimeContext
		ExpectedErr            error
		Consumer               *consumer.Consumer
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("ToGraphQL", modelRuntimeContext).Return(gqlRuntimeContext).Once()
				return conv
			},
			InputID:                id,
			ExpectedRuntimeContext: gqlRuntimeContext,
			ExpectedErr:            nil,
		},
		{
			Name: "Returns error when consumer id is different from owner id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()
				return persistTx, transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            apperrors.NewUnauthorizedError("runtime context not accessible"),
			Consumer: &consumer.Consumer{
				ConsumerID:   "different-runtime-id",
				ConsumerType: consumer.Runtime,
			},
		},
		{
			Name: "Returns error when consumer type is application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return &persistenceautomock.PersistenceTx{}, &persistenceautomock.Transactioner{}
			},
			ServiceFn: func() *automock.RuntimeContextService {
				return &automock.RuntimeContextService{}
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            apperrors.NewUnauthorizedError("runtime context access is restricted to runtimes only"),
			Consumer: &consumer.Consumer{
				ConsumerID:   runtimeID,
				ConsumerType: consumer.Application,
			},
		},
		{
			Name: "Returns error when consumer type is integration system",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return &persistenceautomock.PersistenceTx{}, &persistenceautomock.Transactioner{}
			},
			ServiceFn: func() *automock.RuntimeContextService {
				return &automock.RuntimeContextService{}
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            apperrors.NewUnauthorizedError("runtime context access is restricted to runtimes only"),
			Consumer: &consumer.Consumer{
				ConsumerID:   runtimeID,
				ConsumerType: consumer.IntegrationSystem,
			},
		},
		{
			Name:            "Returns error when runtime context deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("ToGraphQL", modelRuntimeContext).Return(gqlRuntimeContext).Once()
				return conv
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name:            "Returns error when runtime context retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				return conv
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name:            "Returns error when transaction starting failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				return conv
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Delete", contextParam, modelRuntimeContext.ID).Return(nil)
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("ToGraphQL", modelRuntimeContext).Return(gqlRuntimeContext).Once()
				return conv
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime_context.NewResolver(transact, svc, converter)

			c := testCase.Consumer
			if c == nil {
				c = &consumer.Consumer{
					ConsumerID:   runtimeID,
					ConsumerType: consumer.Runtime,
				}
			}
			ctx := consumer.SaveToContext(context.TODO(), *c)

			// when
			result, err := resolver.DeleteRuntimeContext(ctx, testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeContext, result)
			if testCase.ExpectedErr != nil {
				assert.EqualError(t, testCase.ExpectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_RuntimeContext(t *testing.T) {
	// given
	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"
	tenant := "tenant"

	modelRuntimeContext := &model.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Tenant:    tenant,
		Key:       key,
		Value:     val,
	}
	gqlRuntimeContext := &graphql.RuntimeContext{
		ID:    id,
		Key:   key,
		Value: val,
	}
	testErr := errors.New("Test error")

	testCases := []struct {
		Name                   string
		PersistenceFn          func() *persistenceautomock.PersistenceTx
		TransactionerFn        func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn              func() *automock.RuntimeContextService
		ConverterFn            func() *automock.RuntimeContextConverter
		InputID                string
		ExpectedRuntimeContext *graphql.RuntimeContext
		ExpectedErr            error
		Consumer               *consumer.Consumer
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
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntimeContext, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("ToGraphQL", modelRuntimeContext).Return(gqlRuntimeContext).Once()
				return conv
			},
			InputID:                id,
			ExpectedRuntimeContext: gqlRuntimeContext,
			ExpectedErr:            nil,
		},
		{
			Name: "Returns error when consumer id is different from owner id",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntimeContext, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            apperrors.NewUnauthorizedError("runtime context not accessible"),
			Consumer: &consumer.Consumer{
				ConsumerID:   "different-runtime-id",
				ConsumerType: consumer.Runtime,
			},
		},
		{
			Name: "Returns error when consumer type is application",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				return &persistenceautomock.Transactioner{}
			},
			ServiceFn: func() *automock.RuntimeContextService {
				return &automock.RuntimeContextService{}
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            apperrors.NewUnauthorizedError("runtime context access is restricted to runtimes only"),
			Consumer: &consumer.Consumer{
				ConsumerID:   runtimeID,
				ConsumerType: consumer.Application,
			},
		},
		{
			Name: "Returns error when consumer type is integration system",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				return &persistenceautomock.Transactioner{}
			},
			ServiceFn: func() *automock.RuntimeContextService {
				return &automock.RuntimeContextService{}
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            apperrors.NewUnauthorizedError("runtime context access is restricted to runtimes only"),
			Consumer: &consumer.Consumer{
				ConsumerID:   runtimeID,
				ConsumerType: consumer.IntegrationSystem,
			},
		},
		{
			Name: "Success when runtime context not found returns nil",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntimeContext, apperrors.NewNotFoundError(resource.RuntimeContext, "foo")).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				return conv
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            nil,
		},
		{
			Name: "Returns error when runtime context retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				return conv
			},
			InputID:                id,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime_context.NewResolver(transact, svc, converter)

			c := testCase.Consumer
			if c == nil {
				c = &consumer.Consumer{
					ConsumerID:   runtimeID,
					ConsumerType: consumer.Runtime,
				}
			}
			ctx := consumer.SaveToContext(context.TODO(), *c)

			// when
			result, err := resolver.RuntimeContext(ctx, testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeContext, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_RuntimeContexts(t *testing.T) {
	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"
	tenant := "tenant"

	testErr := errors.New("Test error")

	// given
	modelRuntimeContexts := []*model.RuntimeContext{
		{
			ID:        id,
			RuntimeID: runtimeID,
			Tenant:    tenant,
			Key:       key,
			Value:     val,
		},
		{
			ID:        id + "2",
			RuntimeID: runtimeID + "2",
			Tenant:    tenant + "2",
			Key:       key + "2",
			Value:     val + "2",
		},
	}

	gqlRuntimeContexts := []*graphql.RuntimeContext{
		{
			ID:    id,
			Key:   key,
			Value: val,
		},
		{
			ID:    id + "2",
			Key:   key + "2",
			Value: val + "2",
		},
	}

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	filter := []*labelfilter.LabelFilter{{Key: ""}}
	gqlFilter := []*graphql.LabelFilter{{Key: ""}}

	testCases := []struct {
		Name              string
		PersistenceFn     func() *persistenceautomock.PersistenceTx
		TransactionerFn   func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn         func() *automock.RuntimeContextService
		ConverterFn       func() *automock.RuntimeContextConverter
		InputLabelFilters []*graphql.LabelFilter
		InputFirst        *int
		InputAfter        *graphql.PageCursor
		ExpectedResult    *graphql.RuntimeContextPage
		ExpectedErr       error
		Consumer          *consumer.Consumer
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
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("List", contextParam, runtimeID, filter, first, after).Return(&model.RuntimeContextPage{
					Data: modelRuntimeContexts,
					PageInfo: &pagination.Page{
						StartCursor: "start",
						EndCursor:   "end",
						HasNextPage: false,
					},
					TotalCount: len(modelRuntimeContexts),
				}, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("MultipleToGraphQL", modelRuntimeContexts).Return(gqlRuntimeContexts).Once()
				return conv
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult: &graphql.RuntimeContextPage{
				Data: gqlRuntimeContexts,
				PageInfo: &graphql.PageInfo{
					StartCursor: "start",
					EndCursor:   "end",
					HasNextPage: false,
				},
				TotalCount: len(gqlRuntimeContexts),
			},
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when consumer type is application",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				return &persistenceautomock.Transactioner{}
			},
			ServiceFn: func() *automock.RuntimeContextService {
				return &automock.RuntimeContextService{}
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    nil,
			ExpectedErr:       apperrors.NewUnauthorizedError("runtime context access is restricted to runtimes only"),
			Consumer: &consumer.Consumer{
				ConsumerID:   runtimeID,
				ConsumerType: consumer.Application,
			},
		},
		{
			Name: "Returns error when consumer type is integration system",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				return &persistenceautomock.Transactioner{}
			},
			ServiceFn: func() *automock.RuntimeContextService {
				return &automock.RuntimeContextService{}
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				return &automock.RuntimeContextConverter{}
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    nil,
			ExpectedErr:       apperrors.NewUnauthorizedError("runtime context access is restricted to runtimes only"),
			Consumer: &consumer.Consumer{
				ConsumerID:   runtimeID,
				ConsumerType: consumer.IntegrationSystem,
			},
		},
		{
			Name: "Returns error when runtime context listing failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("List", contextParam, runtimeID, filter, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
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

			resolver := runtime_context.NewResolver(transact, svc, converter)

			c := testCase.Consumer
			if c == nil {
				c = &consumer.Consumer{
					ConsumerID:   runtimeID,
					ConsumerType: consumer.Runtime,
				}
			}
			ctx := consumer.SaveToContext(context.TODO(), *c)

			// when
			result, err := resolver.RuntimeContexts(ctx, testCase.InputLabelFilters, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_Labels(t *testing.T) {
	// given
	id := "foo"
	tenant := "tenant"
	labelKey := "key"
	labelValue := "val"

	key := "key"
	val := "value"

	gqlRuntimeContext := &graphql.RuntimeContext{
		ID:    id,
		Key:   key,
		Value: val,
	}
	testErr := errors.New("Test error")

	modelLabels := map[string]*model.Label{
		"abc": {
			ID:         "abc",
			Tenant:     tenant,
			Key:        labelKey,
			Value:      labelValue,
			ObjectID:   id,
			ObjectType: model.RuntimeContextLabelableObject,
		},
		"def": {
			ID:         "def",
			Tenant:     tenant,
			Key:        labelKey,
			Value:      labelValue,
			ObjectID:   id,
			ObjectType: model.RuntimeContextLabelableObject,
		},
	}

	gqlLabels := &graphql.Labels{
		labelKey: labelValue,
		labelKey: labelValue,
	}

	testCases := []struct {
		Name                string
		PersistenceFn       func() *persistenceautomock.PersistenceTx
		TransactionerFn     func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn           func() *automock.RuntimeContextService
		InputRuntimeContext *graphql.RuntimeContext
		InputKey            string
		ExpectedResult      *graphql.Labels
		ExpectedErr         error
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
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("ListLabels", contextParam, id).Return(modelLabels, nil).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: gqlLabels,
			ExpectedErr:    nil,
		},
		{
			Name: "Success returns nil when labels not found",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("ListLabels", contextParam, id).Return(nil, errors.New("doesn't exist")).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: nil,
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
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
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

			resolver := runtime_context.NewResolver(transact, svc, nil)

			// when
			result, err := resolver.Labels(context.TODO(), gqlRuntimeContext, &testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, transact, persistTx)
		})
	}
}
