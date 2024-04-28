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
	"net/http"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	"github.com/sirupsen/logrus"
)

// RequestLogger returns middleware that setups request scoped logging.
// URL paths starting with pathsToLogOnDebug will be logged on debug instead of info.
func RequestLogger(pathsToLogOnDebug ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			entry := LoggerWithCorrelationID(r)
			entry = LoggerWithTraceID(r, entry)
			entry = LoggerWithSpanID(r, entry)
			entry = LoggerWithParentSpanID(r, entry)

			ctx = ContextWithLogger(ctx, entry)
			ctx = ContextWithMdc(ctx)
			r = r.WithContext(ctx)

			start := time.Now()

			remoteAddr := r.RemoteAddr
			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				remoteAddr = realIP
			}

			logOnDebug := shouldLogOnDebug(r.URL.Path, pathsToLogOnDebug)

			beforeLogger := entry.WithFields(logrus.Fields{
				"request": r.RequestURI,
				"method":  r.Method,
				"remote":  remoteAddr,
			})
			beforeLogFunc := beforeLogger.Info
			if logOnDebug {
				beforeLogFunc = beforeLogger.Debug
			}
			beforeLogFunc("Started handling request...")

			lrw := newLoggingResponseWriter(rw)
			next.ServeHTTP(lrw, r)

			duration := time.Since(start)

			afterLogger := entry.WithFields(logrus.Fields{
				"status_code": lrw.statusCode,
				"took":        duration,
			})

			if mdc := MdcFromContext(ctx); nil != mdc {
				afterLogger = mdc.appendFields(afterLogger)
			}

			afterLogFunc := afterLogger.Info
			if logOnDebug {
				afterLogFunc = afterLogger.Debug
			}
			afterLogFunc("Finished handling request...")
		})
	}
}

func shouldLogOnDebug(requestPath string, pathsToLogOnDebug []string) bool {
	for _, path := range pathsToLogOnDebug {
		if strings.HasPrefix(requestPath, path) {
			return true
		}
	}
	return false
}

// LoggerWithCorrelationID missing godoc
func LoggerWithCorrelationID(r *http.Request) *logrus.Entry {
	ctx := r.Context()
	entry := C(ctx)
	if correlationID := correlation.CorrelationIDForRequest(r); correlationID != "" {
		entry = entry.WithField(FieldRequestID, correlationID)
	}

	return entry
}

// LoggerWithTraceID adds the trace ID to the logger
func LoggerWithTraceID(r *http.Request, entry *logrus.Entry) *logrus.Entry {
	if entry == nil {
		ctx := r.Context()
		entry = C(ctx)
	}

	if traceID := correlation.TraceIDForRequest(r); traceID != "" {
		entry = entry.WithField(FieldTraceID, traceID)
	}

	return entry
}

// LoggerWithSpanID adds the span ID to the logger
func LoggerWithSpanID(r *http.Request, entry *logrus.Entry) *logrus.Entry {
	if entry == nil {
		ctx := r.Context()
		entry = C(ctx)
	}

	if spanID := correlation.SpanIDForRequest(r); spanID != "" {
		entry = entry.WithField(FieldSpanID, spanID)
	}

	return entry
}

// LoggerWithParentSpanID adds the parent span ID to the logger
func LoggerWithParentSpanID(r *http.Request, entry *logrus.Entry) *logrus.Entry {
	if entry == nil {
		ctx := r.Context()
		entry = C(ctx)
	}

	if parentSpanID := correlation.ParentSpanIDForRequest(r); parentSpanID != "" {
		entry = entry.WithField(FieldParentSpanID, parentSpanID)
	}

	return entry
}

// AddCorrelationIDsToLogger adds all correlation IDs to the logger
func AddCorrelationIDsToLogger(ctx context.Context, correlationID, traceID, spanID, parentSpanID string) *logrus.Entry {
	logger := C(ctx).WithField(FieldRequestID, correlationID)
	logger = logger.WithField(FieldTraceID, traceID)
	logger = logger.WithField(FieldSpanID, spanID)
	logger = logger.WithField(FieldParentSpanID, parentSpanID)

	return logger
}
