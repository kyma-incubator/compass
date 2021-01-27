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

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const OpCtxKey = "OperationCtx"

type Operation struct {
	OperationID       string                `json:"id"`
	OperationType     graphql.OperationType `json:"type"`
	OperationCategory string                `json:"operation_category"`
	ResourceID        string                `json:"resource_id"`
	ResourceType      string                `json:"resource_type"`
	CorrelationID     string                `json:"correlation_id"`
	WebhookID         string                `json:"webhook_id"`
	RelatedResources  []RelatedResource     `json:"related_resources"`
	RequestData       string                `json:"request_data"`
}

type RelatedResource struct {
	ResourceType string
	ResourceID   string
}

// SaveToContext saves Operation to the context
func SaveToContext(ctx context.Context, op Operation) context.Context {
	return context.WithValue(ctx, OpCtxKey, op)
}

// FromCtx extracts Operation from context
func FromCtx(ctx context.Context) (Operation, error) {
	opCtx := ctx.Value(OpCtxKey)

	if op, ok := opCtx.(Operation); ok {
		return op, nil
	}

	return Operation{}, apperrors.NewInternalError("unable to fetch operation from context")
}
