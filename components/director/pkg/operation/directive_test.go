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

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/header"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/operation/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	gqlgen "github.com/99designs/gqlgen/graphql"

	tx_automock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/ast"
)

var resourceIdField = "id"

const (
	webhookID1 = "fe8ce7c6-919f-40f0-b78b-b1662dfbac64"
	webhookID2 = "4f40d0cf-5a33-4895-aa03-528ab0982fb2"
	webhookID3 = "dbd54239-5188-4bea-8826-bc04587a118e"
)

var whType = graphql.WebhookType(graphql.WebhookTypeRegisterApplication)
var whTypeApplicationUnregister = graphql.WebhookType(graphql.WebhookTypeUnregisterApplication)

var mockedHeaders = http.Header{
	"key": []string{"value"},
}

func TestHandleOperation(t *testing.T) {

	t.Run("missing operation mode param causes internal server error", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		rCtx := &gqlgen.ResolverContext{
			Object: "RegisterApplication",
			Field:  gqlgen.CollectedField{},
			Args: map[string]interface{}{
				operation.ModeParam: "notModeParam",
			},
			IsMethod: false,
		}
		ctx = gqlgen.WithResolverContext(ctx, rCtx)

		directive := operation.NewDirective(nil, nil, nil, nil, nil)

		// WHEN
		_, err := directive.HandleOperation(ctx, nil, nil, graphql.OperationTypeCreate, &whType, &resourceIdField)
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil)

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, nil, graphql.OperationTypeCreate, &whType, &resourceIdField)
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.ErrorResolve, graphql.OperationTypeCreate, &whType, nil)
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whType, nil)
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whType, nil)
		// THEN
		require.NoError(t, err)
		require.Equal(t, mockedNextResponse(), res)
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.ErrorResolve, graphql.OperationTypeCreate, &whType, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to process operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but response is not an Entity type should roll-back", func(t *testing.T) {
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.NonEntityResolve, graphql.OperationTypeCreate, &whType, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Failed to process operation")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but server Director fails to fetch webhooks should roll-back", func(t *testing.T) {
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

		directive := operation.NewDirective(mockedTransactioner, func(_ context.Context, _ string) ([]*model.Webhook, error) {
			return nil, mockedError()
		}, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.NonEntityResolve, graphql.OperationTypeCreate, &whType, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to retrieve webhooks")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but Director fails to prepare operation request due to missing tenant data should roll-back", func(t *testing.T) {
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

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, func(_ context.Context) (string, error) {
			return "", mockedError()
		}, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whType, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to prepare webhook request data")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but Director fails to prepare operation request due unsupported webhook provider type should roll-back", func(t *testing.T) {
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

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, mockedTenantLoaderFunc, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.NonWebhookProviderResolve, graphql.OperationTypeCreate, &whType, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to prepare webhook request data")
		require.Empty(t, res)
		require.Equal(t, graphql.OperationModeAsync, dummyResolver.finalCtx.Value(operation.OpModeKey))
	})

	t.Run("when mutation is in ASYNC mode, there is operation in context but Director fails to prepare operation request due failure to missing request headers should roll-back", func(t *testing.T) {
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

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, mockedTenantLoaderFunc, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whType, nil)
		// THEN
		require.Error(t, err, mockedError().Error(), "Unable to prepare webhook request data")
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
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)
		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return("", mockedError())
		defer mockedScheduler.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, mockedTenantLoaderFunc, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whType, nil)
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
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatFailsOnCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		testID := "test-id"
		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return(testID, nil)
		defer mockedScheduler.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, nil, mockedTenantLoaderFunc, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, &whType, nil)
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
		ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)

		mockedScheduler := &automock.Scheduler{}
		mockedScheduler.On("Schedule", mock.Anything, mock.Anything).Return(operationID, nil)
		defer mockedScheduler.AssertExpectations(t)

		webhookType := whType
		webhooks := []*model.Webhook{
			{ID: webhookID1, Type: model.WebhookType(webhookType)},
			{ID: webhookID2, Type: model.WebhookType(webhookType)},
			{ID: webhookID3, Type: model.WebhookType(webhookType)},
		}
		expectedWebhookIDs := make([]string, 0)
		for _, webhook := range webhooks {
			if graphql.WebhookType(webhook.Type) == webhookType {
				expectedWebhookIDs = append(expectedWebhookIDs, webhook.ID)
			}
		}

		testCases := []struct {
			Name               string
			Webhooks           []*model.Webhook
			ExpectedWebhookIDs []string
		}{
			{
				Name: "when all webhooks match their IDs should be present in the operation",
				Webhooks: []*model.Webhook{
					{ID: webhookID1, Type: model.WebhookType(webhookType)},
					{ID: webhookID2, Type: model.WebhookType(webhookType)},
					{ID: webhookID3, Type: model.WebhookType(webhookType)},
				},
				ExpectedWebhookIDs: []string{webhookID1, webhookID2, webhookID3},
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
			{
				Name: "when no webhooks match no IDs should be present in the operation",
				Webhooks: []*model.Webhook{
					{ID: webhookID1, Type: model.WebhookType(graphql.WebhookTypeUnregisterApplication)},
					{ID: webhookID2, Type: model.WebhookType(graphql.WebhookTypeUnregisterApplication)},
					{ID: webhookID3, Type: model.WebhookType(graphql.WebhookTypeUnregisterApplication)},
				},
				ExpectedWebhookIDs: []string{},
			},
		}

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(len(testCases))
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				directive := operation.NewDirective(mockedTransactioner, func(_ context.Context, _ string) ([]*model.Webhook, error) {
					return testCase.Webhooks, nil
				}, nil, mockedTenantLoaderFunc, mockedScheduler)

				dummyResolver := &dummyResolver{}

				// WHEN
				res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, operationType, &webhookType, nil)

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
				require.Equal(t, operation.OperationType(operationType), op.OperationType)
				require.Equal(t, operationCategory, op.OperationCategory)

				expectedRequestData := &graphql.RequestData{
					Application: *(mockedNextResponse().(*graphql.Application)),
					TenantID:    tenantID,
					Headers:     mockedHeaders,
				}

				expectedData, err := json.Marshal(expectedRequestData)
				require.NoError(t, err)

				require.Equal(t, string(expectedData), op.RequestData)

				require.Len(t, op.WebhookIDs, len(testCase.ExpectedWebhookIDs))
				require.Equal(t, testCase.ExpectedWebhookIDs, op.WebhookIDs)
			})
		}
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
				require.Contains(t, err.Error(), fmt.Sprintf("could not get idField: %q from request context", resourceIdField))
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
						CreatedAt: time.Now(),
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
						CreatedAt: time.Now(),
						DeletedAt: time.Now(),
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
			tenantLoaderFunc: tenantLoaderWithOptionalErr(nil),
			resourceFetcherFunc: func(ctx context.Context, tenant, id string) (model.Entity, error) {
				return &model.Application{
					BaseEntity: &model.BaseEntity{
						ID:        resourceID,
						Ready:     true,
						CreatedAt: time.Now(),
					},
				}, nil
			},
			resolverFunc: (&dummyResolver{}).SuccessResolve,
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
			rCtx := &gqlgen.ResolverContext{
				Object: test.mutation,
				Field: gqlgen.CollectedField{
					Field: &ast.Field{
						Name: test.mutation,
					},
				},
				Args:     test.resolverCtxArgs,
				IsMethod: false,
			}

			ctx = gqlgen.WithResolverContext(ctx, rCtx)
			ctx = context.WithValue(ctx, header.ContextKey, mockedHeaders)
			mockedTx, mockedTransactioner := test.transactionFunc()
			defer mockedTx.AssertExpectations(t)
			defer mockedTransactioner.AssertExpectations(t)

			if test.scheduler != nil {
				defer test.scheduler.AssertExpectations(t)
			}

			directive := operation.NewDirective(mockedTransactioner, func(ctx context.Context, resourceID string) ([]*model.Webhook, error) {
				return nil, nil
			}, test.resourceFetcherFunc, test.tenantLoaderFunc, test.scheduler)

			// WHEN
			res, err := directive.HandleOperation(ctx, nil, test.resolverFunc, graphql.OperationTypeDelete, &whTypeApplicationUnregister, &resourceIdField)
			// THEN
			test.validationFunc(t, res, err)
		})
	}

	t.Run("when idField is not present in the directive it should roll-back", func(t *testing.T) {
		// GIVEN
		operationCategory := "registerApplication"
		ctx := context.Background()
		rCtx := &gqlgen.ResolverContext{
			Object: "UnregisterApplication",
			Field: gqlgen.CollectedField{
				Field: &ast.Field{
					Name: operationCategory,
				},
			},
			Args:     resolverContextArgs(graphql.OperationModeAsync, resourceID),
			IsMethod: false,
		}

		ctx = gqlgen.WithResolverContext(ctx, rCtx)
		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil, nil)

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
	return []*model.Webhook{}, nil
}

func mockedTenantLoaderFunc(_ context.Context) (string, error) {
	return tenantID, nil
}

func mockedError() error {
	return errors.New("mocked error")
}

func resolverContextArgs(mode graphql.OperationMode, optionalResourceID string) map[string]interface{} {
	ctxArgs := map[string]interface{}{operation.ModeParam: &mode}
	if optionalResourceID != "" {
		ctxArgs[resourceIdField] = resourceID
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
