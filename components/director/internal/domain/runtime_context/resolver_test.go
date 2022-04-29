package runtimectx_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/stretchr/testify/mock"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var contextParam = mock.MatchedBy(func(ctx context.Context) bool {
	persistenceOp, err := persistence.FromCtx(ctx)
	return err == nil && persistenceOp != nil
})

func TestResolver_CreateRuntimeContext(t *testing.T) {
	// GIVEN
	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"

	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	modelRuntimeContext := &model.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Key:       key,
		Value:     val,
	}
	gqlRuntimeContext := &graphql.RuntimeContext{
		ID:    id,
		Key:   key,
		Value: val,
	}

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
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.RuntimeContextService
		ConverterFn     func() *automock.RuntimeContextConverter

		Input                  graphql.RuntimeContextInput
		ExpectedRuntimeContext *graphql.RuntimeContext
		ExpectedErr            error
		Consumer               *consumer.Consumer
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("GetByID", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQLWithRuntimeID", gqlInput, runtimeID).Return(modelInput).Once()
				conv.On("ToGraphQL", modelRuntimeContext).Return(gqlRuntimeContext).Once()
				return conv
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: gqlRuntimeContext,
			ExpectedErr:            nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQLWithRuntimeID", gqlInput, runtimeID).Return(modelInput).Once()
				return conv
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name:            "Returns error when runtime context creation failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Create", contextParam, modelInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQLWithRuntimeID", gqlInput, runtimeID).Return(modelInput).Once()
				return conv
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name:            "Returns error when runtime context retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				svc.On("GetByID", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQLWithRuntimeID", gqlInput, runtimeID).Return(modelInput).Once()
				return conv
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("GetByID", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQLWithRuntimeID", gqlInput, runtimeID).Return(modelInput).Once()
				return conv
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtimectx.NewResolver(transact, svc, converter)

			c := testCase.Consumer
			if c == nil {
				c = &consumer.Consumer{
					ConsumerID:   runtimeID,
					ConsumerType: consumer.Runtime,
				}
			}
			ctx := consumer.SaveToContext(context.TODO(), *c)

			// WHEN
			result, err := resolver.RegisterRuntimeContext(ctx, runtimeID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeContext, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_UpdateRuntimeContext(t *testing.T) {
	// GIVEN
	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"

	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	modelRuntimeContext := &model.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Key:       key,
		Value:     val,
	}
	gqlRuntimeContext := &graphql.RuntimeContext{
		ID:    id,
		Key:   key,
		Value: val,
	}

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
		TransactionerFn        func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn              func() *automock.RuntimeContextService
		ConverterFn            func() *automock.RuntimeContextConverter
		RuntimeContextID       string
		Input                  graphql.RuntimeContextInput
		ExpectedRuntimeContext *graphql.RuntimeContext
		ExpectedErr            error
		Consumer               *consumer.Consumer
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("GetByID", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Update", contextParam, id, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelRuntimeContext).Return(gqlRuntimeContext).Once()
				return conv
			},
			RuntimeContextID:       id,
			Input:                  gqlInput,
			ExpectedRuntimeContext: gqlRuntimeContext,
			ExpectedErr:            nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name:            "Returns error when runtime context update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Update", contextParam, id, modelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			RuntimeContextID:       id,
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name:            "Returns error when runtime context retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("Update", contextParam, id, modelInput).Return(nil).Once()
				svc.On("GetByID", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			RuntimeContextID:       id,
			Input:                  gqlInput,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("GetByID", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Update", contextParam, id, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
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
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtimectx.NewResolver(transact, svc, converter)

			c := testCase.Consumer
			if c == nil {
				c = &consumer.Consumer{
					ConsumerID:   runtimeID,
					ConsumerType: consumer.Runtime,
				}
			}
			ctx := consumer.SaveToContext(context.TODO(), *c)

			// WHEN
			result, err := resolver.UpdateRuntimeContext(ctx, testCase.RuntimeContextID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeContext, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_DeleteRuntimeContext(t *testing.T) {
	// GIVEN
	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"

	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	modelRuntimeContext := &model.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Key:       key,
		Value:     val,
	}
	gqlRuntimeContext := &graphql.RuntimeContext{
		ID:    id,
		Key:   key,
		Value: val,
	}

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
				svc.On("GetByID", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
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
			Name:            "Returns error when runtime context deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("GetByID", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(testErr).Once()
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
			Name:            "Returns error when runtime context retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("GetByID", contextParam, "foo").Return(nil, testErr).Once()
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
				svc.On("GetByID", contextParam, "foo").Return(modelRuntimeContext, nil).Once()
				svc.On("Delete", contextParam, modelRuntimeContext.ID).Return(nil)
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
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtimectx.NewResolver(transact, svc, converter)

			c := testCase.Consumer
			if c == nil {
				c = &consumer.Consumer{
					ConsumerID:   runtimeID,
					ConsumerType: consumer.Runtime,
				}
			}
			ctx := consumer.SaveToContext(context.TODO(), *c)

			// WHEN
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

func TestResolver_Labels(t *testing.T) {
	// GIVEN
	id := "foo"
	labelKey1 := "key1"
	labelValue1 := "val1"
	labelKey2 := "key2"
	labelValue2 := "val2"

	key := "key1"
	val := "value1"

	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	gqlRuntimeContext := &graphql.RuntimeContext{
		ID:    id,
		Key:   key,
		Value: val,
	}

	modelLabels := map[string]*model.Label{
		"abc": {
			ID:         "abc",
			Key:        labelKey1,
			Value:      labelValue1,
			ObjectID:   id,
			ObjectType: model.RuntimeContextLabelableObject,
		},
		"def": {
			ID:         "def",
			Key:        labelKey2,
			Value:      labelValue2,
			ObjectID:   id,
			ObjectType: model.RuntimeContextLabelableObject,
		},
	}

	gqlLabels := graphql.Labels{
		labelKey1: labelValue1,
		labelKey2: labelValue2,
	}

	gqlLabels1 := graphql.Labels{
		labelKey1: labelValue1,
	}

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn           func() *automock.RuntimeContextService
		InputRuntimeContext *graphql.RuntimeContext
		InputKey            *string
		ExpectedResult      graphql.Labels
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("ListLabels", contextParam, id).Return(modelLabels, nil).Once()
				return svc
			},
			InputKey:       nil,
			ExpectedResult: gqlLabels,
			ExpectedErr:    nil,
		},
		{
			Name:            "Success when labels are filtered",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("ListLabels", contextParam, id).Return(modelLabels, nil).Once()
				return svc
			},
			InputKey:       &labelKey1,
			ExpectedResult: gqlLabels1,
			ExpectedErr:    nil,
		},
		{
			Name:            "Success returns nil when labels not found",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("ListLabels", contextParam, id).Return(nil, errors.New("doesn't exist")).Once()
				return svc
			},
			InputKey:       &labelKey1,
			ExpectedResult: nil,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				return svc
			},
			InputKey:       &labelKey1,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when label listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("ListLabels", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			InputKey:       &labelKey1,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("ListLabels", contextParam, id).Return(modelLabels, nil).Once()
				return svc
			},
			InputKey:       nil,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()

			resolver := runtimectx.NewResolver(transact, svc, nil)

			// WHEN
			result, err := resolver.Labels(context.TODO(), gqlRuntimeContext, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, transact, persistTx)
		})
	}
}
