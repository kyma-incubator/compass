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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/header"
	tx_automock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/operation/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	webhookID1 = "fe8ce7c6-919f-40f0-b78b-b1662dfbac64"
	webhookID2 = "4f40d0cf-5a33-4895-aa03-528ab0982fb2"
	webhookID3 = "dbd54239-5188-4bea-8826-bc04587a118e"
)

var (
	resourceIDField             = "id"
	whTypeApplicationRegister   = graphql.WebhookTypeRegisterApplication
	whTypeApplicationUnregister = graphql.WebhookTypeUnregisterApplication

	mockedHeaders = http.Header{
		"key": []string{"value"},
	}
)

func TestHandleOperation(t *testing.T) {
	t.Run("missing operation mode param causes internal server error", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field:  gqlgen.CollectedField{},
			Args: map[string]interface{}{
				operation.ModeParam: "notModeParam",
			},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		directive := operation.NewDirective(nil, nil, nil, nil, nil, nil)

		// WHEN
		_, err := directive.HandleOperation(ctx, nil, nil, graphql.OperationTypeCreate, &whTypeApplicationRegister, &resourceIDField)
		// THEN
		require.Error(t, err, fmt.Sprintf("could not get %s parameter", operation.ModeParam))
	})

	t.Run("when mutation is in SYNC mode there is no operation in context but transaction fails to begin", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeSync
		rCtx := &gqlgen.FieldContext{
			Object:   "RegisterApplication",
			Field:    gqlgen.CollectedField{},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatFailsOnBegin()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil, nil)

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, nil, graphql.OperationTypeCreate, &whTypeApplicationRegister, &resourceIDField)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to initialize database operation")
		require.Empty(t, res)
	})

	t.Run("when mutation is in SYNC mode there is no operation in context but request fails should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeSync
		rCtx := &gqlgen.FieldContext{
			Object:   "RegisterApplication",
			Field:    gqlgen.CollectedField{},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.ErrorResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to process operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeSync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in SYNC mode there is no operation in context but transaction fails to commit should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeSync
		rCtx := &gqlgen.FieldContext{
			Object:   "RegisterApplication",
			Field:    gqlgen.CollectedField{},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatFailsOnCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to finalize database operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeSync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in SYNC mode there is no operation in context and finishes successfully", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeSync
		rCtx := &gqlgen.FieldContext{
			Object:   "RegisterApplication",
			Field:    gqlgen.CollectedField{},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.NoError(t, err)
		require.Equal(t, mockedNextResponse(), res)
		require.Equal(t, graphql.OperationModeSync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but request fails should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.ErrorResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to process operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		assertNoOperationsInCtx(t, opsFromCtx)
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but response is not an Entity type should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.NonEntityResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Failed to process operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		assertNoOperationsInCtx(t, opsFromCtx)
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but server Director fails to fetch webhooks should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, func(_ context.Context, _ string) ([]*model.Webhook, error) {
			return nil, mockedError()
		}, nil, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.NonEntityResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to retrieve webhooks")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		assertNoOperationsInCtx(t, opsFromCtx)
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but Director fails to prepare operation request due to missing tenant data should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, mockedEmptyResourceUpdaterFunc, func(_ context.Context) (string, error) {
			return "", mockedError()
		}, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to prepare webhook request data")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		assertNoOperationsInCtx(t, opsFromCtx)
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but Director fails to prepare operation request due unsupported webhook provider type should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, mockedEmptyResourceUpdaterFunc, mockedTenantLoaderFunc, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.NonWebhookProviderResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to prepare webhook request data")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		assertNoOperationsInCtx(t, opsFromCtx)
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but Director fails to prepare operation request due failure to missing request headers should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, mockedEmptyResourceUpdaterFunc, mockedTenantLoaderFunc, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to prepare webhook request data")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		assertNoOperationsInCtx(t, opsFromCtx)
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but Director fails to prepare operation request due missing webhooks", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, func(_ context.Context, _ string) ([]*model.Webhook, error) {
			return nil, mockedError()
		}, nil, mockedEmptyResourceUpdaterFunc, mockedTenantLoaderFunc, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, &resourceIDField)
		// THEN
		require.Error(t, err, "Unable to prepare webhooks")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but Director fails to prepare operation request due multiple webhooks found", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, func(_ context.Context, _ string) ([]*model.Webhook, error) {
			return []*model.Webhook{
				{ID: webhookID1, Type: model.WebhookTypeRegisterApplication},
				{ID: webhookID2, Type: model.WebhookTypeRegisterApplication},
				{ID: webhookID3, Type: model.WebhookTypeRegisterApplication},
			}, nil
		}, nil, mockedEmptyResourceUpdaterFunc, mockedTenantLoaderFunc, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, &resourceIDField)
		// THEN
		require.Error(t, err, "Unable to prepare webhooks")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but Scheduler fails to schedule should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)
		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return("", mockedError())
		defer mockedScheduler.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, mockedEmptyResourceUpdaterFunc, mockedTenantLoaderFunc, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to schedule operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		assertNoOperationsInCtx(t, opsFromCtx)
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but transaction commit fails should roll-back", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatFailsOnCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		testID := "test-id"
		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return(testID, nil)
		defer mockedScheduler.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, mockedEmptyResourceUpdaterFunc, mockedTenantLoaderFunc, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to finalize database operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		assertNoOperationsInCtx(t, opsFromCtx)
	})

	t.Run("when async mode is disabled it should roll-back", func(t *testing.T) {
		operationType := graphql.OperationTypeCreate
		operationCategory := "registerApplication"

		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)
		dummyResolver := &dummyResolver{}
		scheduler := &operation.DisabledScheduler{}
		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, mockedEmptyResourceUpdaterFunc, mockedTenantLoaderFunc, scheduler)

		// WHEN
		_, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, operationType, nil, nil)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Unable to schedule operation")
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but webhook fetcher fails should roll-back", func(t *testing.T) {
		operationType := graphql.OperationTypeCreate
		operationCategory := "registerApplication"

		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		dummyResolver := &dummyResolver{}

		errorWebhooksResponse := func(_ context.Context, _ string) ([]*model.Webhook, error) {
			return nil, errors.New("fail")
		}

		mockedScheduler := &automock.Scheduler{}
		defer mockedScheduler.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, errorWebhooksResponse, nil, mockedEmptyResourceUpdaterFunc, mockedTenantLoaderFunc, mockedScheduler)

		// WHEN
		_, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, operationType, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Unable to retrieve webhooks")
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context and finishes successfully", func(t *testing.T) {
		operationType := operation.OperationTypeCreate
		operationCategory := "registerApplication"

		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)

		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return(operationID, nil)
		defer mockedScheduler.AssertExpectations(t)

		webhookType := whTypeApplicationRegister

		testCases := []struct {
			Name               string
			Webhooks           []*model.Webhook
			ExpectedWebhookIDs []string
		}{
			{
				Name: "when all webhooks match their IDs should be present in the operation",
				Webhooks: []*model.Webhook{
					{ID: webhookID1, Type: model.WebhookType(webhookType)},
				},
				ExpectedWebhookIDs: []string{webhookID1},
			},
			{
				Name: "when a single webhook matches its ID should be present in the operation",
				Webhooks: []*model.Webhook{
					{ID: webhookID1, Type: model.WebhookType(webhookType)},
					{ID: webhookID2, Type: model.WebhookType(graphql.WebhookTypeUnregisterApplication)},
					{ID: webhookID3, Type: model.WebhookType(graphql.WebhookTypeUnregisterApplication)},
				},
				ExpectedWebhookIDs: []string{webhookID1},
			},
		}

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(len(testCases))
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				directive := operation.NewDirective(mockedTransactioner, func(_ context.Context, _ string) ([]*model.Webhook, error) {
					return testCase.Webhooks, nil
				}, nil, mockedEmptyResourceUpdaterFunc, mockedTenantLoaderFunc, mockedScheduler)

				dummyResolver := &dummyResolver{}

				// WHEN
				res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &webhookType, nil)

				// THEN
				require.NoError(t, err)
				require.Equal(t, mockedNextResponse(), res)
				require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

				opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
				operations, ok := opsFromCtx.(*[]*operation.Operation)
				require.True(t, ok)
				require.Len(t, *operations, 1)

				op := (*operations)[0]
				require.Equal(t, operationID, op.OperationID)
				require.Equal(t, operationType, op.OperationType)
				require.Equal(t, operationCategory, op.OperationCategory)

				headers := make(map[string]string)
				for key, value := range mockedHeaders {
					headers[key] = value[0]
				}

				expectedRequestObject := &webhook.ApplicationLifecycleWebhookRequestObject{
					Application: mockedNextResponse().(webhook.Resource),
					TenantID:    tenantID,
					Headers:     headers,
				}

				expectedObj, err := json.Marshal(expectedRequestObject)
				require.NoError(t, err)

				require.Equal(t, string(expectedObj), op.RequestObject)

				require.Len(t, op.WebhookIDs, len(testCase.ExpectedWebhookIDs))
				require.Equal(t, testCase.ExpectedWebhookIDs, op.WebhookIDs)
			})
		}
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context and resource updater func fails should return error", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: "registerApplication",
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode},
			IsMethod: false,
		}
		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, mockedResourceUpdaterFuncWithError, mockedTenantLoaderFunc, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, nil)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Unable to update resource application with id")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		assertNoOperationsInCtx(t, opsFromCtx)
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context and resource updater func is executed with CREATE operation type should finish successfully and update application status to CREATING", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		operationCategory := "registerApplication"
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode, resourceIDField: resourceID},
			IsMethod: false,
		}

		ctx = gqlgen.WithFieldContext(ctx, rCtx)
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)

		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return(operationID, nil)
		defer mockedScheduler.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, mockedResourceFetcherFunc, func(ctx context.Context, id string, ready bool, errorMsg *string, appStatusCondition model.ApplicationStatusCondition) error {
			require.NotNil(t, ctx)
			require.Equal(t, resourceID, id)
			require.Equal(t, false, ready)
			require.Nil(t, errorMsg)
			require.Equal(t, model.ApplicationStatusConditionCreating, appStatusCondition)
			return nil
		}, mockedTenantLoaderFunc, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, &resourceIDField)

		// THEN
		require.NoError(t, err)
		require.Equal(t, mockedNextResponse(), res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		operations, ok := opsFromCtx.(*[]*operation.Operation)
		require.True(t, ok)
		require.Len(t, *operations, 1)

		op := (*operations)[0]
		require.Equal(t, operationID, op.OperationID)
		require.Equal(t, operation.OperationTypeCreate, op.OperationType)
		require.Equal(t, operationCategory, op.OperationCategory)

		headers := make(map[string]string)
		for key, value := range mockedHeaders {
			headers[key] = value[0]
		}

		expectedRequestObject := &webhook.ApplicationLifecycleWebhookRequestObject{
			Application: mockedNextResponse().(webhook.Resource),
			TenantID:    tenantID,
			Headers:     headers,
		}

		expectedObj, err := json.Marshal(expectedRequestObject)
		require.NoError(t, err)

		require.Equal(t, string(expectedObj), op.RequestObject)

		require.Len(t, op.WebhookIDs, 1)
		require.Equal(t, webhookID1, op.WebhookIDs[0])
	})

	t.Run("when mutation is in ASYNC mode, and no webhooks are provided operation should finish successfully and update application status to CREATING", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		operationCategory := "registerApplication"
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode, resourceIDField: resourceID},
			IsMethod: false,
		}

		ctx = gqlgen.WithFieldContext(ctx, rCtx)
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)

		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return(operationID, nil)
		defer mockedScheduler.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedEmptyWebhooksResponse, mockedResourceFetcherFunc, func(ctx context.Context, id string, ready bool, errorMsg *string, appStatusCondition model.ApplicationStatusCondition) error {
			require.NotNil(t, ctx)
			require.Equal(t, resourceID, id)
			require.Equal(t, false, ready)
			require.Nil(t, errorMsg)
			require.Equal(t, model.ApplicationStatusConditionCreating, appStatusCondition)
			return nil
		}, mockedTenantLoaderFunc, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whTypeApplicationRegister, &resourceIDField)

		// THEN
		require.NoError(t, err)
		require.Equal(t, mockedNextResponse(), res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		operations, ok := opsFromCtx.(*[]*operation.Operation)
		require.True(t, ok)
		require.Len(t, *operations, 1)

		op := (*operations)[0]
		require.Equal(t, operationID, op.OperationID)
		require.Equal(t, operation.OperationTypeCreate, op.OperationType)
		require.Equal(t, operationCategory, op.OperationCategory)

		headers := make(map[string]string)
		for key, value := range mockedHeaders {
			headers[key] = value[0]
		}

		expectedRequestObject := &webhook.ApplicationLifecycleWebhookRequestObject{
			Application: mockedNextResponse().(webhook.Resource),
			TenantID:    tenantID,
			Headers:     headers,
		}

		expectedObj, err := json.Marshal(expectedRequestObject)
		require.NoError(t, err)
		require.Equal(t, string(expectedObj), op.RequestObject)
		require.Len(t, op.WebhookIDs, 0)
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context and resource updater func is executed with UPDATE operation type should finish successfully and update application status to UPDATING", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		operationCategory := "registerApplication"
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode, resourceIDField: resourceID},
			IsMethod: false,
		}

		ctx = gqlgen.WithFieldContext(ctx, rCtx)
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)

		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return(operationID, nil)
		defer mockedScheduler.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, mockedResourceFetcherFunc, func(ctx context.Context, id string, ready bool, errorMsg *string, appStatusCondition model.ApplicationStatusCondition) error {
			require.NotNil(t, ctx)
			require.Equal(t, resourceID, id)
			require.Equal(t, false, ready)
			require.Nil(t, errorMsg)
			require.Equal(t, model.ApplicationStatusConditionUpdating, appStatusCondition)
			return nil
		}, mockedTenantLoaderFunc, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeUpdate, &whTypeApplicationRegister, &resourceIDField)

		// THEN
		require.NoError(t, err)
		require.Equal(t, mockedNextResponse(), res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		operations, ok := opsFromCtx.(*[]*operation.Operation)
		require.True(t, ok)
		require.Len(t, *operations, 1)

		op := (*operations)[0]
		require.Equal(t, operationID, op.OperationID)
		require.Equal(t, operation.OperationTypeUpdate, op.OperationType)
		require.Equal(t, operationCategory, op.OperationCategory)

		headers := make(map[string]string)
		for key, value := range mockedHeaders {
			headers[key] = value[0]
		}

		expectedRequestObject := &webhook.ApplicationLifecycleWebhookRequestObject{
			Application: mockedNextResponse().(webhook.Resource),
			TenantID:    tenantID,
			Headers:     headers,
		}

		expectedObj, err := json.Marshal(expectedRequestObject)
		require.NoError(t, err)

		require.Equal(t, string(expectedObj), op.RequestObject)

		require.Len(t, op.WebhookIDs, 1)
		require.Equal(t, webhookID1, op.WebhookIDs[0])
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context and resource updater func is executed with DELETE operation type should finish successfully and update application status to DELETING", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		operationCategory := "registerApplication"
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode, resourceIDField: resourceID},
			IsMethod: false,
		}

		ctx = gqlgen.WithFieldContext(ctx, rCtx)
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)

		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return(operationID, nil)
		defer mockedScheduler.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, mockedResourceFetcherFunc, func(ctx context.Context, id string, ready bool, errorMsg *string, appStatusCondition model.ApplicationStatusCondition) error {
			require.NotNil(t, ctx)
			require.Equal(t, resourceID, id)
			require.Equal(t, false, ready)
			require.Nil(t, errorMsg)
			require.Equal(t, model.ApplicationStatusConditionDeleting, appStatusCondition)
			return nil
		}, mockedTenantLoaderFunc, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeDelete, &whTypeApplicationRegister, &resourceIDField)

		// THEN
		require.NoError(t, err)
		require.Equal(t, mockedNextResponse(), res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		operations, ok := opsFromCtx.(*[]*operation.Operation)
		require.True(t, ok)
		require.Len(t, *operations, 1)

		op := (*operations)[0]
		require.Equal(t, operationID, op.OperationID)
		require.Equal(t, operation.OperationTypeDelete, op.OperationType)
		require.Equal(t, operationCategory, op.OperationCategory)

		headers := make(map[string]string)
		for key, value := range mockedHeaders {
			headers[key] = value[0]
		}

		expectedRequestObject := &webhook.ApplicationLifecycleWebhookRequestObject{
			Application: mockedNextResponse().(webhook.Resource),
			TenantID:    tenantID,
			Headers:     headers,
		}

		expectedObj, err := json.Marshal(expectedRequestObject)
		require.NoError(t, err)

		require.Equal(t, string(expectedObj), op.RequestObject)

		require.Len(t, op.WebhookIDs, 1)
		require.Equal(t, webhookID1, op.WebhookIDs[0])
	})

	t.Run("when mutation is in ASYNC mode, there is operation without webhooks in context and resource updater func is executed with DELETE operation type should finish successfully and update application status to DELETING", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		operationCategory := "registerApplication"
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode, resourceIDField: resourceID},
			IsMethod: false,
		}

		ctx = gqlgen.WithFieldContext(ctx, rCtx)
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)

		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return(operationID, nil)
		defer mockedScheduler.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedEmptyWebhooksResponse, mockedResourceFetcherFunc, func(ctx context.Context, id string, ready bool, errorMsg *string, appStatusCondition model.ApplicationStatusCondition) error {
			require.NotNil(t, ctx)
			require.Equal(t, resourceID, id)
			require.Equal(t, false, ready)
			require.Nil(t, errorMsg)
			require.Equal(t, model.ApplicationStatusConditionDeleting, appStatusCondition)
			return nil
		}, mockedTenantLoaderFunc, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeDelete, &whTypeApplicationRegister, &resourceIDField)

		// THEN
		require.NoError(t, err)
		require.Equal(t, mockedNextResponse(), res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		operations, ok := opsFromCtx.(*[]*operation.Operation)
		require.True(t, ok)
		require.Len(t, *operations, 1)

		op := (*operations)[0]
		require.Equal(t, operationID, op.OperationID)
		require.Equal(t, operation.OperationTypeDelete, op.OperationType)
		require.Equal(t, operationCategory, op.OperationCategory)

		headers := make(map[string]string)
		for key, value := range mockedHeaders {
			headers[key] = value[0]
		}

		expectedRequestObject := &webhook.ApplicationLifecycleWebhookRequestObject{
			Application: mockedNextResponse().(webhook.Resource),
			TenantID:    tenantID,
			Headers:     headers,
		}

		expectedObj, err := json.Marshal(expectedRequestObject)
		require.NoError(t, err)
		require.Equal(t, string(expectedObj), op.RequestObject)
		require.Len(t, op.WebhookIDs, 0)
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context and resource updater func is executed with invalid operation type should return error", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		operationMode := graphql.OperationModeAsync
		operationCategory := "registerApplication"
		rCtx := &gqlgen.FieldContext{
			Object: "RegisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     map[string]interface{}{operation.ModeParam: &operationMode, resourceIDField: resourceID},
			IsMethod: false,
		}

		ctx = gqlgen.WithFieldContext(ctx, rCtx)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, mockedResourceFetcherFunc, mockedEmptyResourceUpdaterFunc, mockedTenantLoaderFunc, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, "invalid", &whTypeApplicationRegister, &resourceIDField)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Invalid status condition")
		require.Nil(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))

		opsFromCtx := dummyResolver.finalCtx.Value(operation.OpCtxKey)
		assertNoOperationsInCtx(t, opsFromCtx)
	})
}

func TestHandleOperation_ConcurrencyCheck(t *testing.T) {
	type testCase struct {
		description         string
		mutation            string
		scheduler           *automock.Scheduler
		tenantLoaderFunc    func(ctx context.Context) (string, error)
		resourceFetcherFunc func(ctx context.Context, tenant, id string) (model.Entity, error)
		validationFunc      func(t *testing.T, res interface{}, err error)
		resolverFunc        func(ctx context.Context) (res interface{}, err error)
		resolverCtxArgs     map[string]interface{}
		transactionFunc     func() (*tx_automock.PersistenceTx, *tx_automock.Transactioner)
	}

	testCases := []testCase{
		{
			description:     "when resource ID is not present in the resolver context it should roll-back",
			mutation:        "UnregisterApplication",
			resolverCtxArgs: resolverContextArgs(graphql.OperationModeAsync, ""),
			transactionFunc: func() (*tx_automock.PersistenceTx, *tx_automock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			validationFunc: func(t *testing.T, res interface{}, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), fmt.Sprintf("could not get idField: %q from request context", resourceIDField))
				require.Empty(t, res)
			},
		},
		{
			description:      "when tenant fetching fails it should roll-back",
			mutation:         "UnregisterApplication",
			resolverCtxArgs:  resolverContextArgs(graphql.OperationModeAsync, resourceID),
			tenantLoaderFunc: tenantLoaderWithOptionalErr(mockedError()),
			transactionFunc: func() (*tx_automock.PersistenceTx, *tx_automock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			validationFunc: func(t *testing.T, res interface{}, err error) {
				require.Error(t, err)
				require.True(t, apperrors.IsTenantRequired(err))
				require.Empty(t, res)
			},
		},
		{
			description:     "when resource is not found it should roll-back",
			mutation:        "UnregisterApplication",
			resolverCtxArgs: resolverContextArgs(graphql.OperationModeAsync, resourceID),
			transactionFunc: func() (*tx_automock.PersistenceTx, *tx_automock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			tenantLoaderFunc: tenantLoaderWithOptionalErr(nil),
			resourceFetcherFunc: func(ctx context.Context, tenant, id string) (model.Entity, error) {
				return nil, apperrors.NewNotFoundError(resource.Application, resourceID)
			},
			validationFunc: func(t *testing.T, res interface{}, err error) {
				require.Error(t, err)
				require.True(t, apperrors.IsNotFoundError(err))
				require.Empty(t, res)
			},
		},
		{
			description:     "when resource fetching fails it should roll-back",
			mutation:        "UnregisterApplication",
			resolverCtxArgs: resolverContextArgs(graphql.OperationModeAsync, resourceID),
			transactionFunc: func() (*tx_automock.PersistenceTx, *tx_automock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			tenantLoaderFunc: tenantLoaderWithOptionalErr(nil),
			resourceFetcherFunc: func(ctx context.Context, tenant, id string) (model.Entity, error) {
				return nil, mockedError()
			},
			validationFunc: func(t *testing.T, res interface{}, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), fmt.Sprintf("failed to fetch resource with id %s", resourceID))
				require.Empty(t, res)
			},
		},
		{
			description:     "when concurrent create operation is running it should roll-back",
			mutation:        "UnregisterApplication",
			resolverCtxArgs: resolverContextArgs(graphql.OperationModeAsync, resourceID),
			transactionFunc: func() (*tx_automock.PersistenceTx, *tx_automock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			tenantLoaderFunc: tenantLoaderWithOptionalErr(nil),
			resourceFetcherFunc: func(ctx context.Context, tenant, id string) (model.Entity, error) {
				return &model.Application{
					BaseEntity: &model.BaseEntity{
						ID:        resourceID,
						Ready:     false,
						CreatedAt: timeToTimePtr(time.Now()),
					},
				}, nil
			},
			validationFunc: func(t *testing.T, res interface{}, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "create operation is in progress")
				require.Empty(t, res)
			},
		},
		{
			description:     "when concurrent delete operation is running it should roll-back",
			mutation:        "UnregisterApplication",
			resolverCtxArgs: resolverContextArgs(graphql.OperationModeAsync, resourceID),
			transactionFunc: func() (*tx_automock.PersistenceTx, *tx_automock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			tenantLoaderFunc: tenantLoaderWithOptionalErr(nil),
			resourceFetcherFunc: func(ctx context.Context, tenant, id string) (model.Entity, error) {
				return &model.Application{
					BaseEntity: &model.BaseEntity{
						ID:        resourceID,
						Ready:     false,
						CreatedAt: timeToTimePtr(time.Now()),
						DeletedAt: timeToTimePtr(time.Now()),
					},
				}, nil
			},
			validationFunc: func(t *testing.T, res interface{}, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "delete operation is in progress")
				require.Empty(t, res)
			},
		},
		{
			description:     "when there are no concurrent operations it should finish successfully",
			mutation:        "UnregisterApplication",
			scheduler:       scheduler(operationID, nil),
			resolverCtxArgs: resolverContextArgs(graphql.OperationModeAsync, resourceID),
			transactionFunc: func() (*tx_automock.PersistenceTx, *tx_automock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
			},
			tenantLoaderFunc:    tenantLoaderWithOptionalErr(nil),
			resourceFetcherFunc: mockedResourceFetcherFunc,
			resolverFunc:        (&dummyResolver{}).SuccessResolve,
			validationFunc: func(t *testing.T, res interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, mockedNextResponse(), res)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.description, func(t *testing.T) {
			// GIVEN
			ctx := context.Background()
			rCtx := &gqlgen.FieldContext{
				Object: test.mutation,
				Field: gqlgen.CollectedField{
					Field: &ast.Field{
						Name: test.mutation,
					},
				},
				Args:     test.resolverCtxArgs,
				IsMethod: false,
			}

			ctx = gqlgen.WithFieldContext(ctx, rCtx)
			ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)
			mockedTx, mockedTransactioner := test.transactionFunc()
			defer mockedTx.AssertExpectations(t)
			defer mockedTransactioner.AssertExpectations(t)

			if test.scheduler != nil {
				defer test.scheduler.AssertExpectations(t)
			}

			directive := operation.NewDirective(mockedTransactioner, func(ctx context.Context, resourceID string) ([]*model.Webhook, error) {
				return nil, nil
			}, test.resourceFetcherFunc, mockedEmptyResourceUpdaterFunc, test.tenantLoaderFunc, test.scheduler)

			// WHEN
			res, err := directive.HandleOperation(ctx, nil, test.resolverFunc, graphql.OperationTypeDelete, nil, &resourceIDField)
			// THEN
			test.validationFunc(t, res, err)
		})
	}

	t.Run("when idField is not present in the directive it should roll-back", func(t *testing.T) {
		// GIVEN
		operationCategory := "registerApplication"
		ctx := context.Background()
		rCtx := &gqlgen.FieldContext{
			Object: "UnregisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     resolverContextArgs(graphql.OperationModeAsync, resourceID),
			IsMethod: false,
		}

		ctx = gqlgen.WithFieldContext(ctx, rCtx)
		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil, nil)

		// WHEN
		_, err := directive.HandleOperation(ctx, nil, nil, graphql.OperationTypeDelete, &whTypeApplicationUnregister, nil)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "idField from context should not be empty")
	})
}

type dummyResolver struct {
	finalCtx context.Context
}

func (d *dummyResolver) SuccessResolve(ctx context.Context) (res interface{}, err error) {
	d.finalCtx = ctx
	return mockedNextResponse(), nil
}

func (d *dummyResolver) ErrorResolve(ctx context.Context) (res interface{}, err error) {
	d.finalCtx = ctx
	return nil, mockedError()
}

func (d *dummyResolver) NonEntityResolve(ctx context.Context) (res interface{}, err error) {
	d.finalCtx = ctx
	return &graphql.Runtime{}, nil
}

func (d *dummyResolver) NonWebhookProviderResolve(ctx context.Context) (res interface{}, err error) {
	d.finalCtx = ctx
	return &graphql.Bundle{BaseEntity: &graphql.BaseEntity{ID: resourceID}}, nil
}

func mockedNextResponse() interface{} {
	return &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: resourceID}}
}

func mockedWebhooksResponse(_ context.Context, _ string) ([]*model.Webhook, error) {
	return []*model.Webhook{
		{ID: webhookID1, Type: model.WebhookTypeRegisterApplication},
	}, nil
}

func mockedEmptyWebhooksResponse(_ context.Context, _ string) ([]*model.Webhook, error) {
	return nil, nil
}

func mockedResourceFetcherFunc(context.Context, string, string) (model.Entity, error) {
	return &model.Application{
		BaseEntity: &model.BaseEntity{
			ID:        resourceID,
			Ready:     true,
			CreatedAt: timeToTimePtr(time.Now()),
		},
	}, nil
}

func mockedTenantLoaderFunc(_ context.Context) (string, error) {
	return tenantID, nil
}

func mockedResourceUpdaterFuncWithError(context.Context, string, bool, *string, model.ApplicationStatusCondition) error {
	return mockedError()
}

func mockedEmptyResourceUpdaterFunc(context.Context, string, bool, *string, model.ApplicationStatusCondition) error {
	return nil
}

func mockedError() error {
	return errors.New("mocked error")
}

func resolverContextArgs(mode graphql.OperationMode, optionalResourceID string) map[string]interface{} {
	ctxArgs := map[string]interface{}{operation.ModeParam: &mode}
	if optionalResourceID != "" {
		ctxArgs[resourceIDField] = resourceID
	}

	return ctxArgs
}

func tenantLoaderWithOptionalErr(optionalErr error) func(ctx context.Context) (string, error) {
	if optionalErr != nil {
		return func(ctx context.Context) (string, error) { return "", optionalErr }
	}

	return func(ctx context.Context) (string, error) { return tenantID, nil }
}

func scheduler(operationID string, err error) *automock.Scheduler {
	mockedScheduler := &automock.Scheduler{}
	mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return(operationID, err)
	return mockedScheduler
}

func assertNoOperationsInCtx(t *testing.T, ops interface{}) {
	operations, ok := ops.(*[]*operation.Operation)
	require.True(t, ok)
	require.Len(t, *operations, 0)
}
