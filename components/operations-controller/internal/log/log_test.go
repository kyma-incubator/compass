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

package log_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/log"
	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestLoadFromContext(t *testing.T) {
	logger := ctrl.Log

	testCases := []struct {
		Name    string
		Context context.Context

		ExpectedResult logr.Logger
	}{
		{
			Name:           "Success",
			Context:        context.WithValue(context.TODO(), log.ContextKey, logger),
			ExpectedResult: logger,
		},
		{
			Name:           "Success when logger is not set",
			Context:        context.TODO(),
			ExpectedResult: logger,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			result := log.LoggerFromContext(testCase.Context)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)

		})
	}
}

func TestContextWithLogger(t *testing.T) {
	// given
	logger := ctrl.Log

	// when
	result := log.ContextWithLogger(context.TODO(), logger)

	// then
	assert.Equal(t, logger, result.Value(log.ContextKey))
}
