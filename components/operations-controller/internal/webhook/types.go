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

package webhook

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	web_hook "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"time"
)

// Request represents a webhook request to be executed
type Request struct {
	Webhook               graphql.Webhook
	Data                  web_hook.RequestData
	CorrelationID         string
	RetryInterval         time.Duration
	OperationCreationTime time.Time
}

func NewRequest(webhook graphql.Webhook, reqData string, correlationID string, opCreationTime time.Time) (*Request, error) {
	var data web_hook.RequestData
	if err := json.Unmarshal([]byte(reqData), &data); err != nil {
		return nil, err
	}

	var retryInterval time.Duration
	if webhook.RetryInterval != nil {
		retryInterval = time.Duration(*webhook.RetryInterval) * time.Minute // TODO: Align measuring units
	}

	return &Request{
		Webhook:               webhook,
		Data:                  data,
		CorrelationID:         correlationID,
		RetryInterval:         retryInterval,
		OperationCreationTime: opCreationTime,
	}, nil
}
