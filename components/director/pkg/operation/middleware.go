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
	"github.com/pkg/errors"
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
func (m middleware) Validate(_ gqlgen.ExecutableSchema) error {
	return nil
}

// InterceptOperation saves an empty slice of async operations into the graphql operation context.
func (m *middleware) InterceptOperation(ctx context.Context, next gqlgen.OperationHandler) gqlgen.ResponseHandler {
	operations := make([]*Operation, 0)
	ctx = SaveToContext(ctx, &operations)

	return next(ctx)
}

// InterceptResponse enriches Async mutation responses with Operation URL location information and also empties the data property of the graphql response for such requests.
func (m *middleware) InterceptResponse(ctx context.Context, next gqlgen.ResponseHandler) *gqlgen.Response {
	resp := next(ctx)

	operations, ok := FromCtx(ctx)
	if !ok {
		return resp
	}

	locations := make([]string, 0)
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
					log.C(ctx).Errorf("Unable to prepare final response: gql field has unexpected type %T instead of *ast.Field", gqlSelection)
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

		newData, err := cleanupFields(resp, jsonPropsToDelete)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Unable to process and delete unnecessary bytes from response body: %v", err)
			return gqlgen.ErrorResponse(ctx, "failed to prepare response body")
		}

		resp.Data = newData
	}

	return resp
}

func cleanupFields(resp *gqlgen.Response, jsonPropsToDelete []string) ([]byte, error) {
	bytes, err := resp.Data.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling current data body")
	}

	for _, prop := range jsonPropsToDelete {
		var err error
		bytes, err = sjson.DeleteBytes(bytes, prop)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("while removing property %s from data body", prop))
		}
	}

	return bytes, nil
}
