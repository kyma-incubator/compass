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

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	OpCtxKey  = "OperationCtx"
	OpModeKey = "OperationModeCtx"
)

// OperationStatus denotes the different statuses that an Operation can be in
type OperationStatus string

const (
	OperationStatusSucceeded  OperationStatus = "SUCCEEDED"
	OperationStatusFailed     OperationStatus = "FAILED"
	OperationStatusInProgress OperationStatus = "IN_PROGRESS"
)

type OperationType string

const (
	OperationTypeCreate OperationType = "CREATE"
	OperationTypeUpdate OperationType = "UPDATE"
	OperationTypeDelete OperationType = "DELETE"
)

// OperationResponse defines the expected response format for the Operations API
type OperationResponse struct {
	*Operation
	Status OperationStatus `json:"status,omitempty"`
	Error  *string         `json:"error,omitempty"`
}

// Operation represents a GraphQL mutation which has associated HTTP requests (Webhooks) that need to be executed
// for the request to be completed fully. Objects of type Operation are meant to be constructed, enriched throughout
// the flow of the original mutation with information such as ResourceID and ResourceType and finally scheduled through
// a dedicated Scheduler implementation.
type Operation struct {
	OperationID       string        `json:"operation_id,omitempty"`
	OperationType     OperationType `json:"operation_type,omitempty"`
	OperationCategory string        `json:"operation_category,omitempty"`
	ResourceID        string        `json:"resource_id,omitempty"`
	ResourceType      resource.Type `json:"resource_type,omitempty"`
	CorrelationID     string        `json:"correlation_id,omitempty"`
	WebhookIDs        []string      `json:"webhook_ids,omitempty"`
	RequestData       string        `json:"request_data,omitempty"`
}

// Validate ensures that the constructed Operation has valid properties
func (op *Operation) Validate() error {
	return validation.ValidateStruct(op,
		validation.Field(&op.ResourceID, is.UUID),
		validation.Field(&op.ResourceType, validation.Required, validation.In(resource.Application)))
}

// SaveToContext saves Operation to the context
func SaveToContext(ctx context.Context, operations *[]*Operation) context.Context {
	if operations == nil {
		return ctx
	}

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
