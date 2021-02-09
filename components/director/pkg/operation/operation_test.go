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

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func TestFromContext(t *testing.T) {

	op1 := &operation.Operation{
		OperationType:     operation.OperationTypeCreate,
		OperationCategory: "registerApplication",
		ResourceType:      resource.Application,
	}

	op2 := &operation.Operation{
		OperationType:     operation.OperationTypeCreate,
		OperationCategory: "registerApplication",
		ResourceType:      resource.Application,
	}

	initOperations := &[]*operation.Operation{op1}

	testCases := []struct {
		Name            string
		Context         context.Context
		OperationsToAdd *[]*operation.Operation

		ExpectedResult *[]*operation.Operation
		Exists         bool
	}{
		{
			Name:            "When Operation slice is set should append to it",
			OperationsToAdd: &[]*operation.Operation{op2},
			Context:         context.WithValue(context.TODO(), operation.OpCtxKey, initOperations),
			ExpectedResult:  &[]*operation.Operation{op1, op2},
			Exists:          true,
		},
		{
			Name:            "When Operation slice is not set save it in the context",
			OperationsToAdd: initOperations,
			Context:         context.TODO(),
			ExpectedResult:  initOperations,
			Exists:          true,
		},
		{
			Name:            "When Operation slice is not set it in the context at all",
			OperationsToAdd: nil,
			Context:         context.TODO(),
			ExpectedResult:  nil,
			Exists:          false,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			testCtx := operation.SaveToContext(testCase.Context, testCase.OperationsToAdd)

			// when
			result, exists := operation.FromCtx(testCtx)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.Exists, exists)
		})
	}
}

func TestSaveToContext(t *testing.T) {
	// given
	ctx := context.TODO()

	op := &operation.Operation{
		OperationType:     operation.OperationTypeCreate,
		OperationCategory: "registerApplication",
		ResourceType:      resource.Application,
	}
	operations := &[]*operation.Operation{op}

	// when
	result := operation.SaveToContext(ctx, operations)

	// then
	assert.Equal(t, operations, result.Value(operation.OpCtxKey))
}

func TestModeFromContext(t *testing.T) {

	testCases := []struct {
		Name    string
		Context context.Context

		ExpectedResult graphql.OperationMode
	}{
		{
			Name:           "When Operation Mode is explicitly set should return it",
			Context:        context.WithValue(context.TODO(), operation.OpModeKey, graphql.OperationModeAsync),
			ExpectedResult: graphql.OperationModeAsync,
		},
		{
			Name:           "When Operation Mode is not explicitly set, should return default (SYNC)",
			Context:        context.TODO(),
			ExpectedResult: graphql.OperationModeSync,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			result := operation.ModeFromCtx(testCase.Context)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
		})
	}
}

func TestSaveModeToContext(t *testing.T) {
	// given
	ctx := context.TODO()

	// when
	result := operation.SaveModeToContext(ctx, graphql.OperationModeAsync)

	// then
	assert.Equal(t, graphql.OperationModeAsync, result.Value(operation.OpModeKey))
}
