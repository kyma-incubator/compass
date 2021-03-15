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

package operation

import (
	"context"
	"fmt"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/tidwall/sjson"
	"github.com/vektah/gqlparser/v2/ast"
)

const LocationsParam = "locations"

type middleware struct {
	directorURL string
}

// NewMiddleware creates a new handler struct responsible for enriching the response of Async mutations with Operation URL location information
func NewMiddleware(directorURL string) *middleware {
	return &middleware{
		directorURL: directorURL,
	}
}


// ExtensionName should be a CamelCase string version of the extension which may be shown in stats and logging.
func (m middleware) ExtensionName() string {
	return "OperationsExtension"
}
// Validate is called when adding an extension to the server, it allows validation against the servers schema.
func (m middleware) Validate(schema gqlgen.ExecutableSchema) error {
	return nil
}

func (m *middleware) InterceptOperation(ctx context.Context, next gqlgen.OperationHandler) gqlgen.ResponseHandler {
	operations := make([]*Operation, 0)
	ctx = SaveToContext(ctx, &operations)

	return next(ctx)
}


// InterceptResponse enriches Async mutation responses with Operation URL location information and also empties the data property of the graphql response for such requests
func (m *middleware) InterceptResponse(ctx context.Context, next gqlgen.ResponseHandler) *gqlgen.Response {
	resp := next(ctx)

	locations := make([]string, 0)
	operations, _ := FromCtx(ctx)
	for _, operation := range *operations {
		operationURL := fmt.Sprintf("%s/%s/%s", m.directorURL, operation.ResourceType, operation.ResourceID)
		locations = append(locations, operationURL)
	}

	if len(locations) > 0 {
		reqCtx := gqlgen.GetOperationContext(ctx)
		gqlgen.RegisterExtension(ctx, LocationsParam, locations)
		resp.Extensions = gqlgen.GetExtensions(ctx)

		jsonPropsToDelete := make([]string, 0)
		for _, gqlOperation := range reqCtx.Doc.Operations {
			for _, gqlSelection := range gqlOperation.SelectionSet {
				gqlField, ok := gqlSelection.(*ast.Field)
				if !ok {
					return gqlgen.ErrorResponse(ctx, "unable to prepare final response")
				}

				mutationAlias := gqlField.Alias
				for _, gqlArgument := range gqlField.Arguments {
					if gqlArgument.Name == ModeParam && gqlArgument.Value.Raw == string(graphql.OperationModeAsync) {
						jsonPropsToDelete = append(jsonPropsToDelete, mutationAlias)
					}
				}
			}
		}

		bytes, _ := resp.Data.MarshalJSON()
		for _, prop := range jsonPropsToDelete {
			var err error

			bytes, err = sjson.DeleteBytes(bytes, prop)
			if err != nil {
				log.C(ctx).Errorf("Unable to process and delete unnecessary bytes from response body: %s", err.Error())
				return gqlgen.ErrorResponse(ctx, "failed to prepare response body")				//return []byte(`{"error": "failed to prepare response body"}`)
			}

		}
		resp.Data = bytes
	}

	return resp
}
