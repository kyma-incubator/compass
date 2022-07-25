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

package errors

import (
	"github.com/pkg/errors"
)

var (
	ErrReconciliationTimeoutReached = errors.New("reconciliation timeout reached")
	ErrWebhookTimeoutReached        = errors.New("webhook timeout reached")
	ErrWebhookPollTimeExpired       = errors.New("polling time has expired")
	ErrFailedWebhookStatus          = errors.New("webhook operation has finished with failed status")
	ErrUnsupportedWebhookMode       = errors.New("unsupported webhook mode")
)

// FatalReconcileErr represents an error type which denotes a failure to proceed with the reconciliation of an Operation CR.
// The error should result in ceasing with the requeue of the specific resource.
type FatalReconcileErr struct {
	error
}

// NewFatalReconcileError constructs a new FatalReconcileErr with the given error message
func NewFatalReconcileError(message string) *FatalReconcileErr {
	return &FatalReconcileErr{
		error: errors.New(message),
	}
}

// NewFatalReconcileError constructs a new FatalReconcileErr based on the provided error
func NewFatalReconcileErrorFromExisting(err error) *FatalReconcileErr {
	return &FatalReconcileErr{
		error: err,
	}
}

// IsWebhookStatusGoneErr check whether an error is a WebhookStatusGoneErr
// and returns true if so.
func IsWebhookStatusGoneErr(err error) (ok bool) {
	_, ok = err.(webhookclient.WebhookStatusGoneErr)
	return
}
