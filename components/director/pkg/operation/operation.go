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

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	OpCtxKey  = "OperationCtx"
	OpModeKey = "OperationModeCtx"
)

type OperationStatus string

const (
	OperationStatusSucceeded  OperationStatus = "SUCCEEDED"
	OperationStatusFailed     OperationStatus = "FAILED"
	OperationStatusInProgress OperationStatus = "IN_PROGRESS"
)

type Operation struct {
	OperationID       string                `json:"id"`
	OperationType     graphql.OperationType `json:"type"`
	OperationCategory string                `json:"operation_category"`
	ResourceID        string                `json:"resource_id"`
	ResourceType      string                `json:"resource_type"`
	CorrelationID     string                `json:"correlation_id"`
	WebhookID         string                `json:"webhook_id"`
	RequestData       string                `json:"request_data"`
}

// SaveToContext saves Operation to the context
func SaveToContext(ctx context.Context, operations *[]*Operation) context.Context {
	operationsFromCtx, exists := FromCtx(ctx)
	if exists {
		*operationsFromCtx = append(*operationsFromCtx, *operations...)
		return ctx
	}

	return context.WithValue(ctx, OpCtxKey, operations)
}

// FromCtx extracts Operation from context
func FromCtx(ctx context.Context) (*[]*Operation, bool) {
	opCtx := ctx.Value(OpCtxKey)

	if operations, ok := opCtx.(*[]*Operation); ok {
		return operations, true
	}

	return nil, false
}

// SaveModeToContext saves operation mode to the context
func SaveModeToContext(ctx context.Context, opMode graphql.OperationMode) context.Context {
	return context.WithValue(ctx, OpModeKey, opMode)
}

// ModeFromCtx extracts operation mode from context
func ModeFromCtx(ctx context.Context) graphql.OperationMode {
	opCtx := ctx.Value(OpModeKey)

	if opMode, ok := opCtx.(graphql.OperationMode); ok {
		return opMode
	}

	return graphql.OperationModeSync
}
