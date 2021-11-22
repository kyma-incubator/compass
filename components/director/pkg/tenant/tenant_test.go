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

package tenant_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromContext(t *testing.T) {
	value := "foo"

	testCases := []struct {
		Name    string
		Context context.Context

		ExpectedResult     string
		ExpectedErrMessage string
	}{
		{
			Name:               "Success",
			Context:            context.WithValue(context.TODO(), tenant.ContextKey, value),
			ExpectedResult:     value,
			ExpectedErrMessage: "",
		},
		{
			Name:               "Error empty tenant value",
			Context:            context.WithValue(context.TODO(), tenant.ContextKey, ""),
			ExpectedResult:     "",
			ExpectedErrMessage: "Tenant is required",
		},
		{
			Name:               "Error missing tenant",
			Context:            context.TODO(),
			ExpectedResult:     "",
			ExpectedErrMessage: "cannot read tenant from context",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			result, err := tenant.LoadFromContext(testCase.Context)

			// then
			if testCase.ExpectedErrMessage != "" {
				require.Equal(t, testCase.ExpectedErrMessage, err.Error())
				return
			}

			assert.Equal(t, testCase.ExpectedResult, result)
		})
	}
}

func TestSaveToLoadFromContext(t *testing.T) {
	// given
	value := "foo"

	// when
	result := tenant.SaveToContext(context.TODO(), value)

	// then
	assert.Equal(t, value, result.Value(tenant.ContextKey))
}

func TestLoadIsolationTypeFromContext(t *testing.T) {
	validValue := tenant.RecursiveIsolationType
	invalidValue := "foo"

	testCases := []struct {
		Name           string
		Context        context.Context
		ExpectedResult string
	}{
		{
			Name:           "Success",
			Context:        context.WithValue(context.TODO(), tenant.IsolationTypeKey, validValue),
			ExpectedResult: string(tenant.RecursiveIsolationType),
		},
		{
			Name:           "Default Recursive when isolation type value is invalid",
			Context:        context.WithValue(context.TODO(), tenant.IsolationTypeKey, invalidValue),
			ExpectedResult: string(tenant.RecursiveIsolationType),
		},
		{
			Name:           "Default Recursive when isolation type value is empty",
			Context:        context.WithValue(context.TODO(), tenant.IsolationTypeKey, ""),
			ExpectedResult: string(tenant.RecursiveIsolationType),
		},
		{
			Name:           "Default Recursive when isolation type value is missing",
			Context:        context.TODO(),
			ExpectedResult: string(tenant.RecursiveIsolationType),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			//when
			result := tenant.LoadIsolationTypeFromContext(testCase.Context)

			// then
			assert.Equal(t, tenant.IsolationType(testCase.ExpectedResult), result)
		})
	}
}

func TestSaveIsolationTypeToContext(t *testing.T) {
	// given
	value := "foo"

	// when
	result := tenant.SaveIsolationTypeToContext(context.TODO(), value)

	// then
	assert.Equal(t, tenant.IsolationType(value), result.Value(tenant.IsolationTypeKey))
}
