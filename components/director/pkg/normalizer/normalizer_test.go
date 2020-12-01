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

package normalizer_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	"github.com/stretchr/testify/assert"
)

func TestNormalizer(t *testing.T) {
	const expectedName = "mp-test-application"

	// given
	testCases := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name:     "Contains upper case characters",
			Input:    "tEsT-ApPlIcAtIoN",
			Expected: expectedName,
		},
		{
			Name:     "Contains dot non-alphanumeric character in the middle",
			Input:    "test.application",
			Expected: expectedName,
		},
		{
			Name:     "Contains $ non-alphanumeric character in the middle",
			Input:    "test$application",
			Expected: expectedName,
		},
		{
			Name:     "Contains & non-alphanumeric character in the middle",
			Input:    "test&application",
			Expected: expectedName,
		},
		{
			Name:     "Contains @ non-alphanumeric character in the middle",
			Input:    "test@application",
			Expected: expectedName,
		},
		{
			Name:     "Contains a sequence of non-alphanumeric character in the middle",
			Input:    "test%.-&application",
			Expected: expectedName,
		},
		{
			Name:     "Contains non-alphanumeric character at the beginning",
			Input:    "!test-application",
			Expected: expectedName,
		},
		{
			Name:     "Contains sequence of non-alphanumeric character at the beginning",
			Input:    "!@#$%test-application",
			Expected: expectedName,
		},
		{
			Name:     "Contains non-alphanumeric character at the end",
			Input:    "test-application%",
			Expected: expectedName,
		},
		{
			Name:     "Contains sequence of non-alphanumeric character at the end",
			Input:    "test-application&*^%",
			Expected: expectedName,
		},
		{
			Name:     "Contains non-alphanumeric character at the beginning, end and in the middle",
			Input:    "[test@application]",
			Expected: expectedName,
		},
		{
			Name:     "Contains multiple dash characters in the middle",
			Input:    "test---application",
			Expected: expectedName,
		},
		{
			Name:     "Contains multiple dash characters at the beginning",
			Input:    "---test-application",
			Expected: expectedName,
		},
		{
			Name:     "Contains multiple dash characters at the end",
			Input:    "test-application---",
			Expected: expectedName,
		},
		{
			Name:     "Contains multiple dash characters at the beginning, end and in the middle",
			Input:    "---test---application---",
			Expected: expectedName,
		},
		{
			Name:     "Contains a single dash character at the end",
			Input:    "test-application-",
			Expected: expectedName,
		},
		{
			Name:     "Contains a combination of dashes and non-dash non-alphanumeric characters at the end",
			Input:    "test-application&--*^-%",
			Expected: expectedName,
		},
		{
			Name:     "Contains upper case characters, non-alphanumeric characters and multiple dashes",
			Input:    "---tEsT@ApPlIcAtIoN.!$",
			Expected: expectedName,
		},
		{
			Name:     "Doesn't contain upper case characters, non-alphanumeric characters or any redundant dashes",
			Input:    "test-application",
			Expected: expectedName,
		},
		{
			Name:     "Already normalized",
			Input:    expectedName,
			Expected: expectedName,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			defaultNormalizer := normalizer.DefaultNormalizator{}

			// when
			normalizedResult := defaultNormalizer.Normalize(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, normalizedResult)
		})
	}
}
