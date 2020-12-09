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

package director_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director/directorfakes"
	"github.com/stretchr/testify/assert"
)

func TestGraphQLClient_FetchApplications(t *testing.T) {
	type fields struct {
		getGCLI           func() *directorfakes.FakeClient
		inputGraphqlizer  director.GraphQLizer
		outputGraphqlizer director.GqlFieldsProvider
	}
	type testCase struct {
		name               string
		fields             fields
		expectedErr        bool
		expectedProperties map[string]int
	}

	inputGraphqlizer := &graphqlizer.Graphqlizer{}
	outputGraphqlizer := &graphqlizer.GqlFieldsProvider{}

	tests := []testCase{
		{
			name: "",
			fields: fields{
				getGCLI: func() *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(nil)
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			expectedProperties: map[string]int{
				"auths":         0,
				"webhooks":      0,
				"status":        0,
				"instanceAuths": 0,
				"documents":     0,
				"fetchRequest":  0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.fields.getGCLI()
			c := director.NewGraphQLClient(
				gcli,
				tt.fields.inputGraphqlizer,
				tt.fields.outputGraphqlizer,
			)
			_, err := c.FetchApplications(context.TODO())
			if tt.expectedErr {
				assert.Error(t, err)
			}
			assert.Equal(t, 1, gcli.DoCallCount())

			_, graphqlReq, _ := gcli.DoArgsForCall(0)
			query := graphqlReq.Query()
			for expectedProp, expectedCount := range tt.expectedProperties {
				fieldRegex := regexp.MustCompile(`\b` + expectedProp + `\b`)

				matches := fieldRegex.FindAllStringIndex(query, -1)
				actualCount := len(matches)

				assert.Equal(t, expectedCount, actualCount, expectedProp)
			}
		})
	}
}
