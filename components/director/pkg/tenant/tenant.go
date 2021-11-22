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

package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type key string

const (
	// ContextKey missing godoc
	ContextKey key = "TenantCtxKey"

	// IsolationTypeKey is the key under which the IsolationType is saved in a given context.Context.
	IsolationTypeKey key = "IsolationTypeKey"
)

// IsolationType represents the isolation type.
type IsolationType string

const (
	// SimpleIsolationType is an isolation type that works by matching resources to tenants directly.
	SimpleIsolationType IsolationType = "simple"

	// RecursiveIsolationType is an isolation type that works by matching tenants recursively.
	RecursiveIsolationType IsolationType = "recursive"
)

// LoadFromContext retrieves the tenantID from the provided context or returns error if missing
func LoadFromContext(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value(ContextKey).(string)

	if !ok {
		return "", apperrors.NewCannotReadTenantError()
	}

	if tenantID == "" {
		return "", apperrors.NewTenantRequiredError()
	}

	return tenantID, nil
}

// SaveToContext saves the provided tenantID into the respective context
func SaveToContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, ContextKey, tenantID)
}

// LoadIsolationTypeFromContext loads the tenant isolation type from context.
// If no valid isolation type is set, consider the recursive type as the default one.
func LoadIsolationTypeFromContext(ctx context.Context) IsolationType {
	if isolationType, ok := ctx.Value(IsolationTypeKey).(IsolationType); ok && isolationType.IsValid() {
		return isolationType
	}
	return RecursiveIsolationType
}

// SaveIsolationTypeToContext saves the isolation type into the provided context.
func SaveIsolationTypeToContext(ctx context.Context, isolationTypeString string) context.Context {
	return context.WithValue(ctx, IsolationTypeKey, IsolationType(isolationTypeString))
}

// IsValid checks whether the isolation type is a valid value.
func (it IsolationType) IsValid() bool {
	return it == SimpleIsolationType ||
		it == RecursiveIsolationType
}
