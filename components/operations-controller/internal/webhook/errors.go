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

import "time"

// ReconcileError is an error type that provides error details for reconciliation failure which may or may not have to be requeued
type ReconcileError struct {
	Requeue      bool          `json:"requeue,omitempty"`
	RequeueAfter time.Duration `json:"requeue_after,omitempty"`
	Description  string        `json:"description,omitempty"`
}

// Error implements Error interface for ReconcileError
func (e *ReconcileError) Error() string {
	return e.Description
}
