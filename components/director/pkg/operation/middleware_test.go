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
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/ast"
)

const (
	directorURL = "http://test-director/"
	operationID = "6188b606-5a60-451a-8065-d2d13b2245ff"
)

func TestExtensionHandlerOperation(t *testing.T) {

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
		resp := middleware.ExtensionHandler(context.Background(), dummyResolver.SuccessResolve)

		require.Equal(t, gqlResultResponse(gqlResults[0].resultName), resp)
	})

	t.Run("when an async operation is found in the context, location extension would be attached and data would be dropped", func(t *testing.T) {
		gqlResults := []gqlResult{
			{
				resultName:    "result",
				operationType: graphql.OperationModeAsync,
			},
		}

		rCtx := gqlRequestContextWithSelections(gqlResults...)
		ctx := gqlgen.WithRequestContext(context.Background(), rCtx)

		operations := &[]*operation.Operation{
			{
				OperationType:     operation.OperationTypeCreate,
				OperationCategory: "registerApplication",
				ResourceID:        operationID,
				ResourceType:      resource.Application,
			},
		}

		dummyResolver := dummyMiddlewareResolver{
			operationToAttach: operations,
			gqlResults:        gqlResults,
		}

		middleware := operation.NewMiddleware(directorURL)
		resp := middleware.ExtensionHandler(ctx, dummyResolver.SuccessResolve)

		require.Equal(t, "{}", string(resp))
		for _, op := range *operations {
			require.Contains(t, rCtx.Extensions[operation.LocationsParam], operationURL(op, directorURL))
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

		rCtx := gqlRequestContextWithSelections(gqlResults...)
		ctx := gqlgen.WithRequestContext(context.Background(), rCtx)

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

		dummyResolver := dummyMiddlewareResolver{
			operationToAttach: operations,
			gqlResults:        gqlResults,
		}

		middleware := operation.NewMiddleware(directorURL)
		resp := middleware.ExtensionHandler(ctx, dummyResolver.SuccessResolve)

		require.Equal(t, "{}", string(resp))
		for _, op := range *operations {
			require.Contains(t, rCtx.Extensions[operation.LocationsParam], operationURL(op, directorURL))
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

		rCtx := gqlRequestContextWithSelections(gqlResults...)
		ctx := gqlgen.WithRequestContext(context.Background(), rCtx)

		operations := &[]*operation.Operation{
			{
				OperationType:     operation.OperationTypeCreate,
				OperationCategory: "registerApplication",
				ResourceID:        operationID,
				ResourceType:      resource.Application,
			},
		}

		dummyResolver := dummyMiddlewareResolver{
			operationToAttach: operations,
			gqlResults:        gqlResults,
		}

		middleware := operation.NewMiddleware(directorURL)
		resp := middleware.ExtensionHandler(ctx, dummyResolver.SuccessResolve)

		require.Equal(t, fmt.Sprintf("{%s}", gqlResultItem(gqlResults[1].resultName)), string(resp))
		require.Len(t, rCtx.Extensions[operation.LocationsParam], 1)
		require.Contains(t, rCtx.Extensions[operation.LocationsParam], operationURL((*operations)[0], directorURL))
	})

	t.Run("when RegisterExtension fails, should return error message", func(t *testing.T) {
		reqCtx := &gqlgen.RequestContext{
			Extensions: map[string]interface{}{
				operation.LocationsParam: []string{"http://test-url/"},
			},
		}

		ctx := gqlgen.WithRequestContext(context.Background(), reqCtx)
		operations := &[]*operation.Operation{
			{
				OperationType:     operation.OperationTypeCreate,
				OperationCategory: "registerApplication",
				ResourceID:        operationID,
				ResourceType:      resource.Application,
			},
		}

		dummyResolver := dummyMiddlewareResolver{
			operationToAttach: operations,
		}

		middleware := operation.NewMiddleware(directorURL)
		resp := middleware.ExtensionHandler(ctx, dummyResolver.SuccessResolve)

		require.Equal(t, `{"error": "unable to finalize operation location"}`, string(resp))
	})

}

type dummyMiddlewareResolver struct {
	operationToAttach *[]*operation.Operation
	gqlResults        []gqlResult
}

func (d *dummyMiddlewareResolver) SuccessResolve(ctx context.Context) []byte {
	if d.operationToAttach != nil {
		operation.SaveToContext(ctx, d.operationToAttach)
	}

	body := ""
	for i, gqlResult := range d.gqlResults {
		body += gqlResultItem(gqlResult.resultName)

		if i != len(d.gqlResults)-1 {
			body += ","
		}
	}

	return []byte(fmt.Sprintf("{%s}", body))
}

func operationURL(op *operation.Operation, directorURL string) string {
	return fmt.Sprintf("%s/operations?%s=%s&%s=%s", directorURL, operation.ResourceIDParam, op.ResourceID, operation.ResourceTypeParam, op.ResourceType)
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

func gqlRequestContextWithSelections(results ...gqlResult) *gqlgen.RequestContext {
	reqCtx := &gqlgen.RequestContext{
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
