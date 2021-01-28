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
	"github.com/tidwall/sjson"
	"github.com/vektah/gqlparser/ast"
)

type middleware struct {
	directorURL string
}

func NewMiddleware(directorURL string) middleware {
	return middleware{
		directorURL: directorURL,
	}
}

func (m *middleware) Handler(ctx context.Context, next func(ctx context.Context) []byte) []byte {
	operations := make([]*Operation, 0)
	ctx = SaveToContext(ctx, &operations)

	resp := next(ctx)

	locations := make([]string, 0)
	for _, operation := range operations {
		operationURL := fmt.Sprintf("%s/operations?resourceID=%s&resourceType=%s", m.directorURL, operation.ResourceID, operation.ResourceType)
		locations = append(locations, operationURL)
	}

	if len(locations) > 0 {
		reqCtx := gqlgen.GetRequestContext(ctx)
		if err := reqCtx.RegisterExtension("locations", locations); err != nil {
			panic(err)
		}

		jsonPropsToDelete := make([]string, 0)
		for _, gqlOperation := range reqCtx.Doc.Operations {
			for _, gqlSelection := range gqlOperation.SelectionSet {
				gqlField := gqlSelection.(*ast.Field)
				mutationAlias := gqlField.Alias
				for _, gqlArgument := range gqlField.Arguments {
					if gqlArgument.Name == "mode" && gqlArgument.Value.Raw == string(graphql.OperationModeAsync) {
						jsonPropsToDelete = append(jsonPropsToDelete, mutationAlias)
					}
				}
			}
		}

		for _, prop := range jsonPropsToDelete {
			var err error
			resp, err = sjson.DeleteBytes(resp, prop)
			if err != nil {
				panic(err)
			}

		}
	}

	return resp
}
