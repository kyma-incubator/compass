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

package webhook_test

import (
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	var tests = []struct {
		Msg            string
		ConfigProvider func() *webhook.Config
		ExpectValid    bool
	}{
		{
			Msg: "Default config should be valid",
			ConfigProvider: func() *webhook.Config {
				return webhook.DefaultConfig()
			},
			ExpectValid: true,
		},
		{
			Msg: "Negative Timeout Factor should be invalid",
			ConfigProvider: func() *webhook.Config {
				config := webhook.DefaultConfig()
				config.TimeoutFactor = -1
				return config
			},
		},
		{
			Msg: "Zero Timeout should be invalid",
			ConfigProvider: func() *webhook.Config {
				config := webhook.DefaultConfig()
				config.TimeoutFactor = 0
				return config
			},
		},
		{
			Msg: "Negative WebhookTimeout should be invalid",
			ConfigProvider: func() *webhook.Config {
				config := webhook.DefaultConfig()
				config.WebhookTimeout = -1
				return config
			},
		},
		{
			Msg: "Negative RequeueInterval should be invalid",
			ConfigProvider: func() *webhook.Config {
				config := webhook.DefaultConfig()
				config.RequeueInterval = -1
				return config
			},
		},
		{
			Msg: "Time Layout different from RFC3339Nano should be invalid",
			ConfigProvider: func() *webhook.Config {
				config := webhook.DefaultConfig()
				config.TimeLayout = time.RFC822
				return config
			},
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
