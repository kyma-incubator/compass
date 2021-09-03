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

type contextKey string

const ContextKey contextKey = "TenantCtxKey"

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
