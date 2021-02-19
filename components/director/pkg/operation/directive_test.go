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
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/director/pkg/header"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/operation/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/ast"
)

const (
	webhookID1 = "fe8ce7c6-919f-40f0-b78b-b1662dfbac64"
	webhookID2 = "4f40d0cf-5a33-4895-aa03-528ab0982fb2"
	webhookID3 = "dbd54239-5188-4bea-8826-bc04587a118e"
)

func TestHandleOperation(t *testing.T) {
	var mockedHeaders = webhook.Header{
		"key": "value",
	}

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

		directive := operation.NewDirective(nil, nil, nil, nil)

		// WHEN
		_, err := directive.HandleOperation(ctx, nil, nil, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil)

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, nil, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.ErrorResolve, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.ErrorResolve, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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

		directive := operation.NewDirective(mockedTransactioner, nil, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.NonEntityResolve, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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
		}, nil, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.NonEntityResolve, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, func(_ context.Context) (string, error) {
			return "", mockedError()
		}, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, mockedTenantLoaderFunc, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.NonWebhookProviderResolve, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, mockedTenantLoaderFunc, nil)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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
		mockedScheduler.On("Schedule", mock.Anything).Return("", mockedError())
		defer mockedScheduler.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, mockedTenantLoaderFunc, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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
		mockedScheduler.On("Schedule", mock.Anything).Return(testID, nil)
		defer mockedScheduler.AssertExpectations(t)

		directive := operation.NewDirective(mockedTransactioner, mockedWebhooksResponse, mockedTenantLoaderFunc, mockedScheduler)

		dummyResolver := &dummyResolver{}

		// WHEN
		res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, graphql.OperationTypeCreate, graphql.WebhookTypeRegisterApplication)
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
		mockedScheduler.On("Schedule", mock.Anything).Return(operationID, nil)
		defer mockedScheduler.AssertExpectations(t)

		webhookType := graphql.WebhookTypeRegisterApplication
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
				}, mockedTenantLoaderFunc, mockedScheduler)

				dummyResolver := &dummyResolver{}

				// WHEN
				res, err := directive.HandleOperation(ctx, nil, dummyResolver.SuccessResolve, operationType, webhookType)

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

				expectedRequestData := &webhook.RequestData{
					Application: mockedNextResponse().(webhook.Resource),
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
