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
)

// DisabledScheduler missing godoc
// DisabledScheduler defines a Scheduler implementation that can be used when asynchronous operations are disabled
type DisabledScheduler struct{}

// Schedule returns an error when called
func (d *DisabledScheduler) Schedule(ctx context.Context, _ *Operation) (string, error) {
	return "", apperrors.NewInvalidOperationError("operation scheduling is currently disabled")
}
