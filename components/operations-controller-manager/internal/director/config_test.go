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
	"testing"

	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	var tests = []struct {
		Msg            string
		ConfigProvider func() *director.Config
		ExpectValid    bool
	}{
		{
			Msg: "Operation endpoint is not set should be invalid",
			ConfigProvider: func() *director.Config {
				config := director.DefaultConfig()
				config.OperationEndpoint = ""
				return config
			},
			ExpectValid: false,
		},
		{
			Msg: "Actual operation endpoint set should be valid",
			ConfigProvider: func() *director.Config {
				config := director.DefaultConfig()
				config.OperationEndpoint = "http://127.0.0.1:3002/operation"
				return config
			},
			ExpectValid: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Msg, func(t *testing.T) {
			err := test.ConfigProvider().Validate()
			if test.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
