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

package log

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

const ContextKey = "LogCtxKey"

var C = LoggerFromContext

// ContextWithLogger returns a new context with the provided logger
func ContextWithLogger(ctx context.Context, logger logr.Logger) context.Context {
	return context.WithValue(ctx, ContextKey, logger)
}

// LoggerFromContext retrieves the current logger from the context
func LoggerFromContext(ctx context.Context) logr.Logger {
	logger, ok := ctx.Value(ContextKey).(logr.Logger)
	if !ok {
		logger = ctrl.Log
	}
	return logger
}
