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
	"fmt"
	"testing"

	gqlgen "github.com/99designs/gqlgen/graphql"
	panichandler "github.com/kyma-incubator/compass/components/director/internal/panic_handler"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const (
	directorURL = "http://test-director/"
	operationID = "6188b606-5a60-451a-8065-d2d13b2245ff"
)

func TestInterceptResponse(t *testing.T) {
	t.Run("when no operations are found in the context, no location extensions would be attached", func(t *testing.T) {
		gqlResults := []gqlResult{
			{
				resultName:    "result",
				operationType: graphql.OperationModeSync,
			},
		}
		dummyResolver := dummyMiddlewareResolver{
			gqlResults: gqlResults,
		}

		middleware := operation.NewMiddleware(directorURL)
		resp := middleware.InterceptResponse(context.Background(), dummyResolver.SuccessResolve)

		require.Equal(t, gqlResultResponse(gqlResults[0].resultName), []byte(resp.Data))
	})

	t.Run("when an async operation is found in the context, location extension would be attached and data would be dropped", func(t *testing.T) {
		gqlResults := []gqlResult{
			{
				resultName:    "result",
				operationType: graphql.OperationModeAsync,
			},
		}
		operations := &[]*operation.Operation{
			{
				OperationType:     operation.OperationTypeCreate,
				OperationCategory: "registerApplication",
				ResourceID:        operationID,
				ResourceType:      resource.Application,
			},
		}

		ctx := gqlContext(gqlResults, operations)
		dummyResolver := dummyMiddlewareResolver{
			gqlResults: gqlResults,
		}

		middleware := operation.NewMiddleware(directorURL)
		resp := middleware.InterceptResponse(ctx, dummyResolver.SuccessResolve)

		require.Equal(t, "{}", string(resp.Data))
		ext, ok := resp.Extensions[operation.LocationsParam]
		require.True(t, ok)
		require.Len(t, ext, len(*operations))

		for _, op := range *operations {
			assertOperationInResponseExtension(t, ext, op)
		}
	})

	t.Run("when multiple async operations are found in the context, multiple location extensions would be attached and data would dropped", func(t *testing.T) {
		gqlResults := []gqlResult{
			{
				resultName:    "result1",
				operationType: graphql.OperationModeAsync,
			}, {
				resultName:    "result2",
				operationType: graphql.OperationModeAsync,
			},
		}

		operations := &[]*operation.Operation{
			{
				OperationType:     operation.OperationTypeCreate,
				OperationCategory: "registerApplication",
				ResourceID:        operationID + "-1",
				ResourceType:      resource.Application,
			},
			{
				OperationType:     operation.OperationTypeCreate,
				OperationCategory: "registerApplication",
				ResourceID:        operationID + "-2",
				ResourceType:      resource.Application,
			},
		}

		ctx := gqlContext(gqlResults, operations)
		dummyResolver := dummyMiddlewareResolver{
			gqlResults: gqlResults,
		}

		middleware := operation.NewMiddleware(directorURL)
		resp := middleware.InterceptResponse(ctx, dummyResolver.SuccessResolve)

		require.Equal(t, "{}", string(resp.Data))
		ext, ok := resp.Extensions[operation.LocationsParam]
		require.True(t, ok)
		require.Len(t, ext, len(*operations))

		for _, op := range *operations {
			assertOperationInResponseExtension(t, ext, op)
		}
	})

	t.Run("when an async operation and sync operation are found in the context, single location extension would be attached and only the async data would dropped", func(t *testing.T) {
		gqlResults := []gqlResult{
			{
				resultName:    "result1",
				operationType: graphql.OperationModeAsync,
			}, {
				resultName:    "result2",
				operationType: graphql.OperationModeSync,
			},
		}

		operations := &[]*operation.Operation{
			{
				OperationType:     operation.OperationTypeCreate,
				OperationCategory: "registerApplication",
				ResourceID:        operationID,
				ResourceType:      resource.Application,
			},
		}

		ctx := gqlContext(gqlResults, operations)
		dummyResolver := dummyMiddlewareResolver{
			gqlResults: gqlResults,
		}

		middleware := operation.NewMiddleware(directorURL)
		resp := middleware.InterceptResponse(ctx, dummyResolver.SuccessResolve)

		require.Equal(t, fmt.Sprintf("{%s}", gqlResultItem(gqlResults[1].resultName)), string(resp.Data))

		ext, ok := resp.Extensions[operation.LocationsParam]
		require.True(t, ok)
		require.Len(t, ext, 1)
		assertOperationInResponseExtension(t, ext, (*operations)[0])
	})
}

func TestInterceptOperation(t *testing.T) {
	t.Run("adds empty operations slice to context", func(t *testing.T) {
		middleware := operation.NewMiddleware(directorURL)
		middleware.InterceptOperation(context.Background(), func(ctx context.Context) gqlgen.ResponseHandler {
			operations, ok := operation.FromCtx(ctx)
			require.True(t, ok)
			require.NotNil(t, operations)
			require.Len(t, *operations, 0)
			return nil
		})
	})
}

func assertOperationInResponseExtension(t *testing.T, ext interface{}, op *operation.Operation) {
	extArray, ok := ext.([]string)
	require.True(t, ok)
	require.Contains(t, extArray, operationURL(op, directorURL))
}

func gqlContext(results []gqlResult, operations *[]*operation.Operation) context.Context {
	ctx := operation.SaveToContext(context.Background(), operations)
	rCtx := gqlRequestContextWithSelections(results...)
	ctx = gqlgen.WithOperationContext(ctx, rCtx)
	ctx = gqlgen.WithResponseContext(ctx, func(ctx context.Context, err error) *gqlerror.Error { return nil }, panichandler.RecoverFn)
	return ctx
}

type dummyMiddlewareResolver struct {
	gqlResults []gqlResult
}

func (d *dummyMiddlewareResolver) SuccessResolve(_ context.Context) *gqlgen.Response {
	body := ""
	for i, gqlResult := range d.gqlResults {
		body += gqlResultItem(gqlResult.resultName)

		if i != len(d.gqlResults)-1 {
			body += ","
		}
	}
	return &gqlgen.Response{Data: []byte(fmt.Sprintf("{%s}", body))}
}

func operationURL(op *operation.Operation, directorURL string) string {
	return fmt.Sprintf("%s/%s/%s", directorURL, op.ResourceType, op.ResourceID)
}

func gqlResultItem(resultName string) string {
	return fmt.Sprintf(`"%s": { "id": "8b3340ff-b1e0-4e9d-8f3b-c36196279552", "name": "%s-app-name"}`, resultName, resultName)
}

func gqlResultResponse(resultName string) []byte {
	return []byte(fmt.Sprintf(`{"%s": { "id": "8b3340ff-b1e0-4e9d-8f3b-c36196279552", "name": "%s-app-name"}}`, resultName, resultName))
}

type gqlResult struct {
	resultName    string
	operationType graphql.OperationMode
}

func gqlRequestContextWithSelections(results ...gqlResult) *gqlgen.OperationContext {
	reqCtx := &gqlgen.OperationContext{
		RawQuery:      "",
		Variables:     nil,
		OperationName: "",
		Doc: &ast.QueryDocument{
			Operations: ast.OperationList{
				&ast.OperationDefinition{
					SelectionSet: ast.SelectionSet{},
				},
			},
		},
	}

	gqlOperation := reqCtx.Doc.Operations[0]

	for _, gqlResult := range results {
		gqlField := &ast.Field{
			Alias: gqlResult.resultName,
			Name:  "registerApplication",
			Arguments: ast.ArgumentList{
				&ast.Argument{
					Name: operation.ModeParam,
					Value: &ast.Value{
						Raw: string(gqlResult.operationType),
					},
				},
			},
		}

		gqlOperation.SelectionSet = append(gqlOperation.SelectionSet, gqlField)
	}

	return reqCtx
}
