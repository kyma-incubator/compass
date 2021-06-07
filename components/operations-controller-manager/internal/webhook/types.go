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
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	web_hook "github.com/kyma-incubator/compass/components/director/pkg/webhook"
)

// Request represents a webhook request to be executed
type Request struct {
	Webhook       graphql.Webhook
	Object        web_hook.RequestObject
	CorrelationID string
}

// PollRequest represents a webhook poll request to be executed
type PollRequest struct {
	*Request
	PollURL string
}

// NewRequest constructs a webhook Request
func NewRequest(webhook graphql.Webhook, requestObject web_hook.RequestObject, correlationID string) *Request {
	return &Request{
		Webhook:       webhook,
		Object:        requestObject,
		CorrelationID: correlationID,
	}
}

// NewPollRequest constructs a webhook Request
func NewPollRequest(webhook graphql.Webhook, requestObject web_hook.RequestObject, correlationID string, pollURL string) *PollRequest {
	return &PollRequest{
		Request: NewRequest(webhook, requestObject, correlationID),
		PollURL: pollURL,
	}
}
