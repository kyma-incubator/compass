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
	"github.com/vektah/gqlparser/ast"
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

// ExtensionHandler enriches Async mutation responses with Operation URL location information and also empties the data property of the graphql response for such requests
func (m *middleware) ExtensionHandler(ctx context.Context, next func(ctx context.Context) []byte) []byte {
	operations := make([]*Operation, 0)
	ctx = SaveToContext(ctx, &operations)

	resp := next(ctx)

	locations := make([]string, 0)
	for _, operation := range operations {
		operationURL := fmt.Sprintf("%s/operations?%s=%s&%s=%s", m.directorURL, ResourceIDParam, operation.ResourceID, ResourceTypeParam, operation.ResourceType)
		locations = append(locations, operationURL)
	}

	if len(locations) > 0 {
		reqCtx := gqlgen.GetRequestContext(ctx)
		if err := reqCtx.RegisterExtension(LocationsParam, locations); err != nil {
			log.C(ctx).Errorf("Unable to attach %s extension: %s", LocationsParam, err.Error())
			return []byte(`{"error": "unable to finalize operation location"}`)
		}

		jsonPropsToDelete := make([]string, 0)
		for _, gqlOperation := range reqCtx.Doc.Operations {
			for _, gqlSelection := range gqlOperation.SelectionSet {
				gqlField, ok := gqlSelection.(*ast.Field)
				if !ok {
					return []byte(`{"error": "unable to prepare final response"}`)
				}

				mutationAlias := gqlField.Alias
				for _, gqlArgument := range gqlField.Arguments {
					if gqlArgument.Name == ModeParam && gqlArgument.Value.Raw == string(graphql.OperationModeAsync) {
						jsonPropsToDelete = append(jsonPropsToDelete, mutationAlias)
					}
				}
			}
		}

		for _, prop := range jsonPropsToDelete {
			var err error
			resp, err = sjson.DeleteBytes(resp, prop)
			if err != nil {
				log.C(ctx).Errorf("Unable to process and delete unnecessary bytes from response body: %s", err.Error())
				return []byte(`{"error": "failed to prepare response body"}`)
			}

		}
	}

	return resp
}
