/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package operation_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/operation/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/ast"
)

func TestHandleOperation(t *testing.T) {
	t.Run("missing operation mode param causes internal server error", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		rCtx := &gqlgen.ResolverContext{
			Object:   "RegisterApplication",
			Field:    gqlgen.CollectedField{},
			Args:     map[string]interface{}{},
			IsMethod: false,
		}
		ctx = gqlgen.WithResolverContext(ctx, rCtx)

		directive := operation.NewDirective(nil, nil)

		// WHEN
		_, err := directive.HandleOperation(ctx, nil, nil, graphql.OperationTypeCreate)
		// THEN
		require.Error(t, err, fmt.Sprintf("could not get %s parameter", operation.ModeParam))
	})

	t.Run("when mutation is in SYNC mode there is no operation in context but transaction fails to begin", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeSync
		rCtx := &gqlgen.ResolverContext{
			Object:   "RegisterApplication",
			Field:    gqlgen.CollectedField{},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithResolverContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatFailsOnBegin()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil)

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, nil, graphql.OperationTypeCreate)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to initialize database operation")
		require.Empty(t, res)
	})

	t.Run("when mutation is in SYNC mode there is no operation in context but request fails should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeSync
		rCtx := &gqlgen.ResolverContext{
			Object:   "RegisterApplication",
			Field:    gqlgen.CollectedField{},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithResolverContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.ErrorResolve, graphql.OperationTypeCreate)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to process operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeSync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in SYNC mode there is no operation in context but transaction fails to commit should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeSync
		rCtx := &gqlgen.ResolverContext{
			Object:   "RegisterApplication",
			Field:    gqlgen.CollectedField{},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithResolverContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatFailsOnCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to finalize database operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeSync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in SYNC mode there is no operation in context and finishes successfully", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeSync
		rCtx := &gqlgen.ResolverContext{
			Object:   "RegisterApplication",
			Field:    gqlgen.CollectedField{},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithResolverContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate)
		// THEN
		require.NoError(t, err)
		require.Equal(t, mockedNextOutput(), res)
		require.Equal(t, graphql.OperationModeSync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but request fails should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.ResolverContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithResolverContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.ErrorResolve, graphql.OperationTypeCreate)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to process operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but Scheduler fails to schedule should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.ResolverContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithResolverContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything).Return("", mockedError())
		defer mockedScheduler.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to schedule operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but transaction commit fails should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.ResolverContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithResolverContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatFailsOnCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		testID := "test-id"
		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything).Return(testID, nil)
		defer mockedScheduler.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to finalize database operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context and finishes successfully", func(t *testing.T) {
		operationID := "test-id"
		operationType := graphql.OperationTypeCreate
		operationCategory := "registerApplication"

		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.ResolverContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithResolverContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything).Return(operationID, nil)
		defer mockedScheduler.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, operationType)

		// THEN
		require.NoError(t, err)
		require.Equal(t, mockedNextOutput(), res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		operations, ok := opsFromCtx.(*[]*operation.Operation)
		require.True(t, ok)
		require.Len(t, *operations, 1)

		operation := (*operations)[0]
		require.Equal(t, operationID, operation.OperationID)
		require.Equal(t, operationType, operation.OperationType)
		require.Equal(t, operationCategory, operation.OperationCategory)
	})

}

type dummyResolver struct {
	finalCtx context.Context
}

func (d *dummyResolver) SuccessResolve(ctx context.Context) (res interface{}, err error) {
	d.finalCtx = ctx
	return mockedNextOutput(), nil
}

func (d *dummyResolver) ErrorResolve(ctx context.Context) (res interface{}, err error) {
	d.finalCtx = ctx
	return nil, mockedError()
}

func mockedNextOutput() string {
	return "nextOutput"
}

func mockedError() error {
	return errors.New("mocked error")
}
